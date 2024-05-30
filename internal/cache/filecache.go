package cache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/opencontainers/go-digest"

	"gitlab.com/act3-ai/asce/data/tool/internal/util"
)

// BottleFileCache is an implementation of MoteCache for dealing with
// data bottle caching of blobs.  No metadata is maintained for the motes,
// relying on file existence to identify items present in the cache.
type BottleFileCache struct {
	localPath    string   // local path for the index file
	pendingMotes sync.Map // pending data in progress of being written to cache

	DigestAlgorithm digest.Algorithm

	LinkOnReadDisabled bool      // Whether or not to create a hard link during read operations to prevent locking files
	FallbackCache      MoteCache // A secondary cache to check if the primary cache does not contain a requested file.  This fallback is considered only on read operations. Caution, avoid graph cycles!
}

// IsDirty always returns false for BottleFileCache -- no way to detect if there have
// been changes.
func (bc *BottleFileCache) IsDirty() bool {
	return false
}

// Refresh is a null action for BottleFileCache, false is returned to indicate that
// no update process needs to be performed.
func (bc *BottleFileCache) Refresh() (bool, error) {
	return false, nil
}

// Update does nothing for BottleFileCache, as there is no state to update.
func (bc *BottleFileCache) Update() error {
	return nil
}

// Initialize checks if a cache exists at the cache path path, and creates
// one if necessary.
func (bc *BottleFileCache) Initialize() error {
	if _, err := os.Stat(bc.localPath); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(bc.localPath, 0o777); err != nil {
			return fmt.Errorf("failed to create cache path: %w", err)
		}
	}

	return nil
}

// MoteExists returns true if the given item appears in the cache,
// This does not check the pending motes.
func (bc *BottleFileCache) MoteExists(dgst digest.Digest) bool {
	path := bc.MoteRef(dgst)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		if bc.FallbackCache != nil {
			return bc.FallbackCache.MoteExists(dgst)
		}
		return false
	}
	return true
}

// Find returns the mote information interface for a known mote.  since
// BottleFileCache does not retain metadata information for motes, this
// will only contain digest and file size for the mote.
func (bc *BottleFileCache) Find(dgst digest.Digest) (MoteInfo, bool) {
	path := bc.MoteRef(dgst)
	finfo, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		if bc.FallbackCache != nil {
			return bc.FallbackCache.Find(dgst)
		}
		return Mote{}, false
	}

	m := Mote{
		Digest:   dgst,
		DataSize: finfo.Size(),
	}
	return m, true
}

// MoteRef for bottlefilecache returns the expected file path for a mote
// digest.  The file is not checked for existence.
func (bc *BottleFileCache) MoteRef(dgst digest.Digest) string {
	return filepath.Join(bc.localPath, dgst.Algorithm().String(), dgst.Encoded())
}

// CommitMote adds or finalizes an existing file into the cache.
// Use CommitCopy mode to keep the original file on commit
// Use CommitMove mode to remove the original file (using Rename)
// WARNING: this function is re-entered on CommitCopy, because MoteWriter
// is used to synchronize writes to the cache, which in turn calls this
// function with CommitMove mode to finalize the update.
func (bc *BottleFileCache) CommitMote(srcpath string, mote Mote, commitMode int) error {
	dgst := mote.GetDigest()
	// Check if the mote is already in the cache, if so no update is required.
	if bc.MoteExists(dgst) {
		return nil
	}

	// Copy uses a MoteWriter to copy data from the source to the mote.
	// MoteWriter uses CommitMote to finalize, so this function is reentered
	// with mode CommitMove
	if commitMode == CommitCopy {
		return bc.copyToMote(srcpath, dgst)
	}

	// lastly, rename the source path to the destination blob, moving temp
	// files to final cache location
	prepsrc := srcpath
	if !strings.HasPrefix(srcpath, bc.localPath) {
		// in the case of a move between two filesystems, os.Rename will cause
		// a file copy behind the scenes, which is nonatomic.  Thus, if the file
		// is not being renamed within the cache dir, we do a 'double rename'
		// with a temporary file intermediary to guard against a non atomic
		// copy
		// TODO check if they are on the same filesystem to avoid this copy (not just if they are in the same directory).
		// Due to bind mounts this check is not sufficient anyway.

		tmpfn := tempFileName(bc.localPath)
		err := os.Rename(srcpath, tmpfn)
		if err != nil {
			return fmt.Errorf("error performing double rename: %w", err)
		}
		defer os.Remove(tmpfn)
		prepsrc = tmpfn
		// Since the temp rename may have taken some time, check to make sure the
		// destination blob hasn't arrived in the interim.
		if bc.MoteExists(dgst) {
			return nil
		}
	}

	if err := os.Chmod(prepsrc, 0o644); err != nil {
		return fmt.Errorf("error changing file permissions: %w", err)
	}

	blobpath := bc.MoteRef(dgst)
	if err := util.CreatePathForFile(blobpath); err != nil {
		return err
	}
	if err := os.Rename(prepsrc, blobpath); err != nil {
		return fmt.Errorf("error renaming cache file: %w", err)
	}
	return nil
}

// copyToMote copies existing data to the cache using a mote writer (which
// creates a temporary file until the copy is finished).
func (bc *BottleFileCache) copyToMote(srcpath string, dgst digest.Digest) error {
	infile, err := os.Open(srcpath)
	if err != nil {
		return fmt.Errorf("error opening temp tile: %w", err)
	}
	defer infile.Close()

	writer, err := bc.MoteWriter(dgst)
	if err != nil {
		return err
	}
	defer writer.Close()

	if _, err = io.Copy(writer, infile); err != nil {
		return fmt.Errorf("error copying mote to cache: %w", err)
	}

	return errors.Join(
		infile.Close(),
		writer.Close(),
	)
}

// MoteWriter returns a MoteWriter for writing to a cache blob. The write is
// done to a temporary location, and an enclosed Close handler is used to
// finalize the cache addition.
func (bc *BottleFileCache) MoteWriter(dgst digest.Digest) (io.WriteCloser, error) {
	dat, ok := bc.pendingMotes.Load(dgst)
	if !ok {
		return nil, ErrNotFound
	}
	m := dat.(Mote)

	f, err := os.CreateTemp(bc.localPath, "in_")
	if err != nil {
		return nil, fmt.Errorf("error opening temp tile: %w", err)
	}
	tmppath := f.Name()

	var closed bool
	finish := func() error {
		if closed {
			return nil
		}
		closed = true

		finfo, err := f.Stat()
		if err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}
		if finfo.Size() != m.Size() {
			return ErrSizeMismatch
		}
		return bc.CommitMote(tmppath, m, CommitMove)
	}

	m.writeFn = f.Write
	m.closeFn = finish
	return m, nil
}

// makeTempLink creates a new temporary hardlink to the provided blob as
// identified by dgst.  This temproary hardlink filename can be used to
// read from the blob, and is intended to be removed after closing it,
// allowing the original blob to be safely and atomically removed at
// any time from the cache.
func (bc *BottleFileCache) makeTempLink(dgst digest.Digest) (string, error) {
	blobpath := bc.MoteRef(dgst)
	temppath := tempFileName(bc.localPath)
	if err := os.Link(blobpath, temppath); err != nil {
		return temppath, fmt.Errorf("error creating link: %w", err)
	}
	return temppath, nil
}

// makeReaderPath either resolves the reader path to original data, or
// calls makeTempLink to generate a link file path to return.
func (bc *BottleFileCache) makeReaderPath(dgst digest.Digest) (string, error) {
	if bc.LinkOnReadDisabled {
		return bc.MoteRef(dgst), nil
	}
	return bc.makeTempLink(dgst)
}

func (bc *BottleFileCache) openMote(dgst digest.Digest, useReadAt bool) (*Mote, error) {
	blobpath, err := bc.makeReaderPath(dgst)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(blobpath)
	if err != nil {
		return nil, fmt.Errorf("error opening blob file: %w", err)
	}
	finfo, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("error stat-ing blob file: %w", err)
	}

	m := &Mote{
		Digest:   dgst,
		DataSize: finfo.Size(),
	}
	if useReadAt {
		m.readAtFn = f.ReadAt
	} else {
		m.readFn = f.Read
	}
	m.closeFn = f.Close
	if !bc.LinkOnReadDisabled {
		m.closeFn = func() error {
			return errors.Join(f.Close(), os.Remove(blobpath))
		}
	}

	return m, nil
}

// MoteReader returns a MoteReader for reading from a cache blob.  Before
// opening the blob, the file is hard linked to a temporary file name, and
// the temporary file is removed upon close.
func (bc *BottleFileCache) MoteReader(dgst digest.Digest) (io.ReadCloser, error) {
	path := bc.MoteRef(dgst)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		if bc.FallbackCache != nil {
			return bc.FallbackCache.MoteReader(dgst)
		}
		return nil, ErrNotFound
	}
	m, err := bc.openMote(dgst, false)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// MoteReaderAt returns an io.ReaderAt for offset based reading.
func (bc *BottleFileCache) MoteReaderAt(dgst digest.Digest) (ReaderAt, error) {
	path := bc.MoteRef(dgst)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		if bc.FallbackCache != nil {
			return bc.FallbackCache.MoteReaderAt(dgst)
		}
		return nil, ErrNotFound
	}
	m, err := bc.openMote(dgst, true)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// CreateMote returns a new pending Mote based on the BottleFileCache settings.
func (bc *BottleFileCache) CreateMote(dgst digest.Digest, mediaType string, size int64) (Mote, error) {
	m := Mote{
		Digest:    dgst,
		MediaType: mediaType,
		DataSize:  size,
	}

	bc.pendingMotes.Store(dgst, m)

	return m, nil
}

// RemoveMote removes a mote from the collection of cached motes.
func (bc *BottleFileCache) RemoveMote(dgst digest.Digest) error {
	blobpath := bc.MoteRef(dgst)
	var err error
	if bc.MoteExists(dgst) {
		bc.pendingMotes.Delete(dgst)
		err = os.Remove(blobpath)
	}
	if err != nil {
		return fmt.Errorf("error removing mote: %w", err)
	}
	return nil
}

// Prune removes bottle items until the total size of the cache is less than or
// equal to maxSize.
func (bc *BottleFileCache) Prune(ctx context.Context, maxSize int64) error {
	// TODO: cached files are now stored in the digest algorithm subdirectories.  The below code should be adjusted to
	// consider all existing algorithm subdirs instead of just sha256
	algPath := filepath.Join(bc.localPath, "sha256")
	fsys := os.DirFS(algPath)
	curSize, err := util.DirSize(fsys)
	if err != nil {
		return err
	}
	// short circuit
	if curSize <= maxSize {
		return nil
	}
	var nonCacheSize int64

	// file prunefile()
	infos, err := util.ReadDirSortedByAccessTime(fsys, ".") // TODO this is inefficient and could use the results ([]FileInfo) of the dirSize() above.
	if err != nil {
		return err
	}
	for _, info := range infos {
		if err := ctx.Err(); err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".boltdb") {
			nonCacheSize += info.Size()
			continue
		}
		sz := info.Size()
		err := os.RemoveAll(filepath.Join(algPath, info.Name()))
		// permission denied occurs if file is locked, just skip to the next file
		if errors.Is(err, fs.ErrPermission) {
			continue
		}
		// some other error during removal
		if err != nil {
			return fmt.Errorf("error removing cache: %w", err)
		}
		curSize -= sz
		if curSize <= maxSize {
			break
		}
	}

	if curSize > maxSize+nonCacheSize {
		return fmt.Errorf("failed to reduce cache below max size of %d, size reduced to %d", maxSize, curSize)
	}
	return nil
}

// tempFileName creates a temporary file name, similar to os.TempFile, but without
// opening the file. this is more suited to the file path archiver library.
func tempFileName(basepath string) string {
	i := rand.Int63()
	return filepath.Join(basepath, fmt.Sprintf("tmp_%d", i))
}

// BottleFileCacheOpt is a functional option type for configuring a bottle cache
// on creation.
type BottleFileCacheOpt func(*BottleFileCache)

// NewBottleFileCache creates a new bottle cache object for managing data bottle
// cache.  This does not load or create the cache on disk, use Initialize to
// do so.
func NewBottleFileCache(path string, opts ...BottleFileCacheOpt) *BottleFileCache {
	bc := &BottleFileCache{
		DigestAlgorithm: digest.SHA256,
		localPath:       path,
	}
	for _, o := range opts {
		o(bc)
	}
	return bc
}

// WithKeyAlgorithm allows an override to create a cache using a different
// hashing algorithm than the default.  The algorithm is one of the
// open container go-digest algorithm constants (strings).
func WithKeyAlgorithm(alg digest.Algorithm) BottleFileCacheOpt {
	return func(bc *BottleFileCache) {
		bc.DigestAlgorithm = alg
	}
}

// DisableLinkOnRead causes the BottleFileCache to not create hard links when creating a mote reader.
func DisableLinkOnRead() BottleFileCacheOpt {
	return func(bc *BottleFileCache) {
		bc.LinkOnReadDisabled = true
	}
}

// WithFallbackCache enables a fallback MoteCache that can provide a secondary source for motes not currently in the
// cache, such as an HTTP source for retrieval.  This only used for a read source cache, and additionally does not
// add items to the primary cache location when the fallback cache is used as the stand in reader.
func WithFallbackCache(cache MoteCache) BottleFileCacheOpt {
	return func(bc *BottleFileCache) {
		bc.FallbackCache = cache
	}
}
