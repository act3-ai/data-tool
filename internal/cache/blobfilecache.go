package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/djherbis/atime"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orascontent "oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/data/tool/internal/util"
	"git.act3-ace.com/ace/go-common/pkg/fsutil"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// FileCacher implements oras' content.GraphStorage as well as registry.Mounter.
type FileCacher interface {
	// content.GraphStorage
	Exists(ctx context.Context, target ocispec.Descriptor) (bool, error)
	Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error)
	Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error
	Predecessors(ctx context.Context, node ocispec.Descriptor) ([]ocispec.Descriptor, error)
	// registry.Mounter
	Mount(ctx context.Context, desc ocispec.Descriptor, path string, getContent func() (io.ReadCloser, error)) error
}

// FileCacheManager extends FileCacher with ability to remove blobs.
type FileCacheManager interface {
	FileCacher

	Delete(ctx context.Context, desc ocispec.Descriptor) error
	Prune(ctx context.Context, maxSize int64) error
}

// FileCache is an implementation of FileCacheManager. A persistent file-based oras content.Storage,
// with an optional in-memory PredecessorFinder (enabling it to be used as a content.GraphStorage).
// Predecessors() returns an error if NewFileCache() was not provided the WithPredecessors() option.
type FileCache struct {
	root         string    // cache root directory
	pendingBlobs *sync.Map // pending data in progress of being written to cache; digest.Digest : pendingBlob

	// predecessors is a map of subject to referring descriptors
	predecessors map[digest.Digest][]ocispec.Descriptor
	pMux         sync.RWMutex

	fallbackCache orascontent.ReadOnlyStorage // A secondary cache capable of supplementing files not found in the primary.
}

// pendingBlob simply wraps blob with a channel, notifying potential
// duplicate pushes the it has completed. Simultaneous pushes are de-duplicated such that
// the second push reads and discards from its input reader, blocking until the first completes.
// This ensures graph integrity and safe calls of oras.CopyGraph().
type pendingBlob struct {
	blob

	sync.Once
	done chan struct{}
}

// NewFileCache creates a new FileCache object for managing data blobs.
func NewFileCache(path string, opts ...FileCacheOpt) (*FileCache, error) {
	fc := &FileCache{
		root:         path,
		pendingBlobs: &sync.Map{},
	}
	if _, err := os.Stat(fc.root); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(fc.root, 0777); err != nil {
			return nil, fmt.Errorf("failed to create cache path: %w", err)
		}
	}
	for _, o := range opts {
		o(fc)
	}
	return fc, nil
}

// FileCacheOpt is a functional option type for configuring a FileCache
// on creation.
type FileCacheOpt func(*FileCache)

// WithPredecessors enables (in-memory) blob predecessors. Predecessors are
// established during calls to Fetch (if the blob exists in the cache or fallback cache) or
// Push.
func WithPredecessors() FileCacheOpt {
	return func(fc *FileCache) {
		fc.predecessors = make(map[digest.Digest][]ocispec.Descriptor)
		fc.pMux = sync.RWMutex{}
	}
}

// WithFallbackCache enables a fallback content.Storage that can provide a secondary source for blobs not currently in the
// cache, such as an HTTP source for retrieval.  This only used for a read source cache, and additionally does not
// add items to the primary cache location when the fallback cache is used as the stand in reader.
func WithFallbackCache(store orascontent.Storage) FileCacheOpt {
	return func(fc *FileCache) {
		fc.fallbackCache = store
	}
}

// Exists returns true if the given item appears in the cache, as a file.
// This does not check the pending blobs.
func (fc *FileCache) Exists(ctx context.Context, target ocispec.Descriptor) (bool, error) {
	if err := target.Digest.Validate(); err != nil {
		return false, fmt.Errorf("validating digest: %w", err)
	}

	path := fc.blobPath(target.Digest)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		if fc.fallbackCache != nil {
			fallbackExists, err := fc.fallbackCache.Exists(ctx, target)
			if err != nil {
				return fallbackExists, fmt.Errorf("checking blob existence in fallback cache: %w", err)
			}
			return fallbackExists, nil
		}
		return false, nil
	}

	// although we don't actually access the file, we assume checking it's existence is meaningful
	if err := os.Chtimes(path, time.Now(), time.Time{}); err != nil {
		// not fatal
		logger.FromContext(ctx).ErrorContext(ctx, "updating cached blob access time", "path", path, "error", err)
	}

	// oras.ExtendedCopyGraph will expect predecessors to be setup
	if target.MediaType == ocispec.MediaTypeImageManifest ||
		target.MediaType == ocispec.MediaTypeImageIndex {
		m, err := fc.blobReader(ctx, target.Digest)
		if err != nil {
			return true, err
		}
		defer m.Close()
		blob, err := io.ReadAll(m)
		if err != nil {
			return true, fmt.Errorf("reading image manifest from cache: %w", err)
		}
		if err := m.Close(); err != nil {
			return true, fmt.Errorf("closing blob file: %w", err)
		}

		if err := fc.addAsPredecessor(ctx, blob, target); err != nil {
			// json.Unmarshal error, which isn't fatal here but
			// likely will be downstream. Perhaps they know something we don't,
			// so we log the error and eat it
			logger.FromContext(ctx).ErrorContext(ctx, "adding blob as predecessor", "error", err)
		}
	}

	return true, nil
}

// Fetch returns a ReadCloser for reading from a cache blob.  Before
// opening the blob, the file is hard linked to a temporary file name, and
// the temporary file is removed upon close.
func (fc *FileCache) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	if err := desc.Digest.Validate(); err != nil {
		return nil, fmt.Errorf("validating digest: %w", err)
	}

	path := fc.blobPath(desc.Digest)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		if fc.fallbackCache != nil {
			rc, err := fc.fallbackCache.Fetch(ctx, desc)
			if err != nil {
				return nil, fmt.Errorf("fetching from fallback cache: %w", err)
			}
			return rc, nil
		}
		return nil, errdef.ErrNotFound
	}

	if fc.predecessors != nil {
		if desc.MediaType == ocispec.MediaTypeImageManifest ||
			desc.MediaType == ocispec.MediaTypeImageIndex {
			m, err := fc.blobReader(ctx, desc.Digest)
			if err != nil {
				return nil, err
			}
			defer m.Close()
			blob, err := io.ReadAll(m)
			if err != nil {
				return nil, fmt.Errorf("reading image manifest from cache: %w", err)
			}
			if err := m.Close(); err != nil {
				return nil, fmt.Errorf("closing blob file: %w", err)
			}

			if err := fc.addAsPredecessor(ctx, blob, desc); err != nil {
				// json.Unmarshal error, which isn't fatal here but
				// likely will be downstream. Perhaps they know something we don't,
				// so we log the error and eat it
				logger.FromContext(ctx).ErrorContext(ctx, "adding blob as predecessor", "error", err)
			}
			return io.NopCloser(bytes.NewReader(blob)), nil
		}
	}

	m, err := fc.blobReader(ctx, desc.Digest)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Push copies blob content to a file if there is not already a copy in progress.
func (fc *FileCache) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	if err := expected.Digest.Validate(); err != nil {
		return fmt.Errorf("validating digest: %w", err)
	}

	log := logger.V(logger.FromContext(ctx), 1)
	b := &pendingBlob{
		blob: blob{
			Descriptor: expected,
		},
		Once: sync.Once{},
		done: make(chan struct{}),
	}

	actual, loaded := fc.pendingBlobs.LoadOrStore(b.Digest, b)
	blob := actual.(*pendingBlob)
	if loaded {
		log.InfoContext(ctx, "blob push in progress by another goroutine, skipping duplicate push to cache", "digest", b.Digest)
		// discard data, ensuring tee'd reads are successful
		_, err := io.Copy(io.Discard, content)
		if err != nil {
			return fmt.Errorf("discarding duplicate blob: %w", err)
		}
		<-blob.done // ensure in-progress push is completed
		return nil
	}
	defer func() {
		close(blob.done)
		fc.pendingBlobs.Delete(blob.Digest)
	}()

	var doErr error
	blob.Once.Do(func() {
		doErr = fc.pushOnce(ctx, expected, content)
	})
	return doErr
}

// pushOnce performs the actual pushing of the blob to the FileCache. It's intended to be used once via
// blob.Once.Do().
func (fc *FileCache) pushOnce(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	log := logger.FromContext(ctx)
	if fc.predecessors != nil {
		if expected.MediaType == ocispec.MediaTypeImageManifest ||
			expected.MediaType == ocispec.MediaTypeImageIndex {
			blob, err := io.ReadAll(content)
			if err != nil {
				return fmt.Errorf("reading image manifest: %w", err)
			}
			content = bytes.NewReader(blob)

			if err := fc.addAsPredecessor(ctx, blob, expected); err != nil {
				return fmt.Errorf("adding blob as predecessor: %w", err)
			}
		}
	}

	exists, _ := fc.Exists(ctx, expected)
	if exists {
		log.InfoContext(ctx, "blob already cached", "digest", expected.Digest)
		// discard data, ensuring tee'd reads are successful
		_, err := io.Copy(io.Discard, content)
		if err != nil {
			return fmt.Errorf("discarding duplicate blob: %w", err)
		}
		return nil
	}

	// if for some reason the cache gets deleted, rebuild our file structure
	err := os.MkdirAll(fc.root, 0777)
	if err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	wc, err := fc.blobWriter(ctx, expected.Digest)
	if err != nil {
		return fmt.Errorf("preparing blob writer: %w", err)
	}
	defer wc.Close()

	// cleanup properly when an error occurs
	cleanupOnErr := func() error {
		// ensure the blob is committed to it's final destination so we delete the correct file
		if err := wc.Close(); err != nil {
			return fmt.Errorf("closing blob file: %w", err)

		}
		if err := fc.Delete(ctx, expected); err != nil {
			return fmt.Errorf("deleting malformed blob from cache: %w", err)
		}
		return nil
	}

	vr := orascontent.NewVerifyReader(content, expected)
	_, err = io.Copy(wc, vr)
	if err != nil {
		return errors.Join(fmt.Errorf("copying blob: %w", err), cleanupOnErr())
	}

	if err := vr.Verify(); err != nil {
		return errors.Join(fmt.Errorf("verifying blob: %w", err), cleanupOnErr())
	}
	return nil
}

// Mount creates a hardlink from path to the underlying cache filesystem. A full descriptor is not required,
// however the digest field is. Mount implements oras registry.Mounter interface.
func (fc *FileCache) Mount(ctx context.Context, desc ocispec.Descriptor, path string, getContent func() (io.ReadCloser, error)) error {
	if err := desc.Digest.Validate(); err != nil {
		return fmt.Errorf("validating digest: %w", err)
	}
	log := logger.FromContext(ctx).With("digest", desc.Digest, "mountPath", path)

	exists, err := fc.Exists(ctx, desc)
	switch {
	case exists:
		log.InfoContext(ctx, "cache entry already exists")
		return nil
	case err != nil:
		log.ErrorContext(ctx, "failed to check existence in cache", "error", err)
		fallthrough
	default:
		path, err := filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("evaluating potential symlink: %w", err)
		}

		blobPath := fc.blobPath(desc.Digest)
		blobDir := filepath.Dir(blobPath)
		if err := os.MkdirAll(blobDir, 0777); err != nil {
			log.ErrorContext(ctx, "initializing file cache directories", "error", err)
		}

		err = os.Link(path, blobPath)
		if err == nil {
			log.InfoContext(ctx, "mount successful")
			break
		}
		// hardlink may fail if file is not within the same filesystem.
		log.ErrorContext(ctx, "mount failed, falling back to push", "error", err)
		rc, err := getContent()
		if err != nil {
			return fmt.Errorf("retrieving content: %w", err)
		}

		if err := fc.Push(ctx, desc, rc); err != nil {
			return errors.Join(fmt.Errorf("pushing content: %w", err), rc.Close())
		}

		if err := rc.Close(); err != nil {
			return fmt.Errorf("closing getContent source: %w", err)
		}
	}

	return nil
}

// ErrPredecessorsDisabled is returned by FileCache.Predecessors if the in-memory
// predecessors graph is not initialized, i.e. NewFileCache was not called with the
// WithPredecessors option.
var ErrPredecessorsDisabled = errors.New("predecessors not enabled")

// Predecessors finds the nodes directly pointing to a given node of a directed acyclic graph. In other
// words, returns the "parents" of the current descriptor. Predecessors implements oras content.PredecessorFinder, and
// is an extension of oras conent.Storage.
//
// Predecessors returns an error if the FileCache was not initialized with the WithPredecessors Option.
func (fc *FileCache) Predecessors(ctx context.Context, node ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	if err := node.Digest.Validate(); err != nil {
		return nil, fmt.Errorf("validating digest: %w", err)
	}

	if fc.predecessors != nil {
		fc.pMux.RLock()
		predecessors, ok := fc.predecessors[node.Digest]
		fc.pMux.RUnlock()
		if !ok {
			return []ocispec.Descriptor{}, nil
		}
		return predecessors, nil
	}
	return nil, ErrPredecessorsDisabled
}

// Delete removes a specific blob from the collection of cached blobs.
func (fc *FileCache) Delete(ctx context.Context, desc ocispec.Descriptor) error {
	if err := desc.Digest.Validate(); err != nil {
		return fmt.Errorf("validating digest: %w", err)
	}

	blobpath := fc.blobPath(desc.Digest)
	exists, err := fc.Exists(ctx, desc)
	switch {
	case err != nil:
		return fmt.Errorf("unable to check blob existence in cache: %w", err)
	case !exists:
		return nil
	default:
		err = os.Remove(blobpath)
		if err != nil {
			return fmt.Errorf("error removing mote: %w", err)
		}
	}

	return nil
}

// Prune removes bottle items until the total size of the cache is less than or
// equal to maxSize.
func (fc *FileCache) Prune(ctx context.Context, maxSize int64) error {
	// walk the blobs dir
	path := filepath.Join(fc.root, ocispec.ImageBlobsDir)
	curSize, finfos, err := evalDir(os.DirFS(path))
	if err != nil {
		return fmt.Errorf("evaluating cache directory: %w", err)
	}
	// short circuit
	if curSize <= maxSize {
		return nil
	}

	// TODO: we could use other metrics to determine which cached files are more valuable than others;
	// such as size, frequency of use, commonality of digest algorithm, etc. For now, we simply use
	// a last access time, which is updated on Exists().
	slices.SortFunc(finfos, func(a, b fileInfoWithDir) int {
		if atime.Get(a.FileInfo).Before(atime.Get(b.FileInfo)) {
			return -1
		}
		return 1
	})

	var nonCacheSize int64
Pruning:
	for _, info := range finfos {
		if err := ctx.Err(); err != nil {
			return err
		}
		// don't delete the blobinfocache
		if strings.HasSuffix(info.Name(), ".boltdb") {
			nonCacheSize += info.Size()
			continue
		}
		sz := info.Size()
		d := ocispec.Descriptor{
			MediaType: "application/octet-stream",
			Digest:    digest.Digest(filepath.Base(filepath.Dir(info.path)) + ":" + info.Name()),
			Size:      sz,
		}
		err = fc.Delete(ctx, d)
		switch {
		case errors.Is(err, fs.ErrPermission):
			// permission denied occurs if file is locked, just skip to the next file
			continue
		case err != nil:
			return fmt.Errorf("error removing cache entry: %w", err)
		default:
			curSize -= sz
			if curSize <= maxSize {
				break Pruning
			}
		}
	}

	if curSize > maxSize+nonCacheSize {
		return fmt.Errorf("failed to reduce cache below max size of %d, size reduced to %d", maxSize, curSize)
	}
	return nil
}

// addAsPredecessors attempts to add a blob as a predecessor of its subject, if set, in the
// in-memory predecessor map. It is safe to call for all mediatypes, but only establishes
// predecessors for mediatypes: ocispec.MediaTypeImageManifest and ocispec.MediaTypeImageIndex.
func (fc *FileCache) addAsPredecessor(ctx context.Context, blob []byte, desc ocispec.Descriptor) error {
	// addAsPredecessors could take a reader instead, allowing us to once again tee our reader stream.
	// Since we end up storing the manifest in memery anyhow, we can slightly reduce complexity by reading
	// it into memory earlier.
	log := logger.FromContext(ctx).With("blobDigest", desc.Digest)

	var subjectDigest digest.Digest
	var subjectMediaType string
	switch {
	case desc.MediaType == ocispec.MediaTypeImageManifest:
		var manifest ocispec.Manifest
		err := json.Unmarshal(blob, &manifest)
		if err != nil {
			return fmt.Errorf("failed to decode manifest blob: %w", err)
		}

		if manifest.Subject != nil {
			subjectDigest = manifest.Subject.Digest
			subjectMediaType = manifest.Subject.MediaType
		}
	case desc.MediaType == ocispec.MediaTypeImageIndex:
		var index ocispec.Index
		err := json.Unmarshal(blob, &index)
		if err != nil {
			return fmt.Errorf("failed to decode index manifest blob: %w", err)
		}

		if index.Subject != nil {
			subjectDigest = index.Subject.Digest
			subjectMediaType = index.Subject.MediaType
		}
	default:
		log.InfoContext(ctx, "unknown mediatype, skipping evaluation of subject status", "mediatype", desc.MediaType)
		return nil
	}

	if subjectDigest != "" {
		fc.pMux.Lock()
		existingList, ok := fc.predecessors[subjectDigest]
		if ok {
			for _, desc := range existingList {
				if desc.Digest == subjectDigest && desc.MediaType == subjectMediaType {
					fc.pMux.Unlock()
					return nil // blob is already known to be a predecessor
				}
			}
			fc.predecessors[subjectDigest] = append(fc.predecessors[subjectDigest], desc)

		} else {
			fc.predecessors[subjectDigest] = []ocispec.Descriptor{desc}
		}
		fc.pMux.Unlock()
		log.InfoContext(ctx, "adding blob manifest to subject's predecessors", "subjectDigest", subjectDigest)
	}

	return nil
}

// blobPath returns the expected file path for a blob
// digest. The file is not checked for existence.
func (fc *FileCache) blobPath(dgst digest.Digest) string {
	return filepath.Join(fc.root, ocispec.ImageBlobsDir, dgst.Algorithm().String(), dgst.Encoded())
}

// commitBlob finalizes a blob write to the cache. It renames srcpath to the destination path, committing the data
// atomically; unless the files are across filesystems.
func (fc *FileCache) commitBlob(ctx context.Context, srcpath string, blob *blob) error {
	// re-check if the blob was committed by another process while we were writing to the temp file
	exists, err := fc.Exists(ctx, blob.Descriptor)
	if err != nil {
		return fmt.Errorf("checking blob existence before committing: %w", err)
	}
	if exists {
		return nil
	}

	// lastly, rename the source path to the destination blob, moving temp
	// files to final cache location
	prepsrc := srcpath
	if !strings.HasPrefix(srcpath, fc.root) {
		// in the case of a move between two filesystems, os.Rename will cause
		// a file copy behind the scenes, which is nonatomic.  Thus, if the file
		// is not being renamed within the cache dir, we do a 'double rename'
		// with a temporary file intermediary to guard against a non atomic
		// copy
		// TODO check if they are on the same filesystem to avoid this copy (not just if they are in the same directory).
		// Due to bind mounts this check is not sufficient anyway.

		tmpFile, err := os.CreateTemp(fc.root, "tmp_write_*")
		if err != nil {
			return fmt.Errorf("creating temporary file: %w", err)
		}
		tmpPath := tmpFile.Name()
		if err := tmpFile.Close(); err != nil {
			return fmt.Errorf("closing temporary file: %w", err)
		}

		err = os.Rename(srcpath, tmpPath)
		if err != nil {
			return fmt.Errorf("error performing double rename: %w", err)
		}
		defer os.Remove(tmpPath)
		prepsrc = tmpPath
		// Since the temp rename may have taken some time, check to make sure the
		// destination blob hasn't arrived in the interim.
		exists, err := fc.Exists(ctx, blob.Descriptor)
		if err != nil {
			return fmt.Errorf("checking blob existence before committing: %w", err)
		}
		if exists {
			return nil
		}
	}

	if err := os.Chmod(prepsrc, 0644); err != nil {
		return fmt.Errorf("error changing file permissions: %w", err)
	}

	blobpath := fc.blobPath(blob.Digest)
	if err := util.CreatePathForFile(blobpath); err != nil {
		return err
	}
	if err := os.Rename(prepsrc, blobpath); err != nil {
		return fmt.Errorf("error renaming cache file: %w", err)
	}
	return nil
}

// blob represents a data entry in the cache. The blobReader and blobWriter
// functions prepare the blob file for concurrency safe reads and writes.
type blob struct {
	ocispec.Descriptor

	file    *os.File
	closeFn func() error // must always be set
}

// Read implements the io.Reader interface.
func (b *blob) Read(p []byte) (n int, err error) {
	return b.file.Read(p) //nolint
}

// Write implements the io.Writer interface.
func (b *blob) Write(p []byte) (n int, err error) {
	return b.file.Write(p) //nolint
}

// Close implements the io.Closer interface.
func (b *blob) Close() error {
	return b.closeFn()
}

// blobReader provides concurrency safe read access to a blob file.
func (fc *FileCache) blobReader(ctx context.Context, dgst digest.Digest) (io.ReadCloser, error) {
	blobPath := fc.blobPath(dgst)

	// create a temporary duplicate via a hard-link
	tf, err := os.CreateTemp(fc.root, "tmp_read_*")
	if err != nil {
		return nil, fmt.Errorf("creating temporary file: %w", err)
	}
	tmpPath := tf.Name()
	if err := tf.Close(); err != nil {
		return nil, fmt.Errorf("closing temporary file: %w", err)
	}
	if err := os.Remove(tmpPath); err != nil {
		return nil, fmt.Errorf("removing temporary file: %w", err)
	}
	if err := os.Link(blobPath, tmpPath); err != nil {
		return nil, fmt.Errorf("error creating link: %w", err)
	}

	tmpFile, err := os.Open(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("error opening blob file: %w", err)
	}

	finfo, err := tmpFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("error stat-ing blob file: %w", err)
	}

	return &blob{
		Descriptor: ocispec.Descriptor{
			Digest:    dgst,
			Size:      finfo.Size(),
			MediaType: "application/octet-stream",
		},
		file: tmpFile,
		closeFn: func() error {
			return errors.Join(tmpFile.Close(), os.Remove(tmpPath))
		},
	}, nil
}

// blobWriter provides write access to a blob file. It first writes to a temporary file
// later renaming it to the final destination.
func (fc *FileCache) blobWriter(ctx context.Context, dgst digest.Digest) (io.WriteCloser, error) {
	tmpFile, err := os.CreateTemp(fc.root, "in_")
	if err != nil {
		return nil, fmt.Errorf("error opening temp tile: %w", err)
	}
	tmpPath := tmpFile.Name()

	b := &blob{
		Descriptor: ocispec.Descriptor{
			Digest:    dgst,
			MediaType: "application/octet-stream",
		},
		file: tmpFile,
	}
	b.closeFn = func() error {
		if err := tmpFile.Close(); err != nil {
			return err
		}

		return fc.commitBlob(ctx, tmpPath, b)
	}

	return b, nil
}

// evalDir evaluates a directory returning the total size of all files seen as well as their file infos.
func evalDir(fsys fs.FS) (int64, []fileInfoWithDir, error) {
	var size int64
	seen := make(map[uint64]string)
	infos := make([]fileInfoWithDir, 0)

	return size, infos, fs.WalkDir(fsys, ".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			return fmt.Errorf("error getting file info: %w", err)
		}
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		infos = append(infos, fileInfoWithDir{fi, path})

		inode, err := fsutil.GetInode(fi)
		if err != nil {
			return fmt.Errorf("error getting inode: %w", err)
		}

		_, ok := seen[inode]
		if ok {
			// duplicate inode number, skip
			return nil
		}
		seen[inode] = path
		size += fi.Size()

		return nil
	})
}

// fileInfoWithDir wraps fs.FileInfo with a field containing the full file path.
type fileInfoWithDir struct {
	fs.FileInfo

	path string
}
