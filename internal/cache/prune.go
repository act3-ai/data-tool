package cache

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/djherbis/atime"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/go-common/pkg/fsutil"
)

// Prune removes files until the total size of the cache is less than or
// equal to maxSize.
func Prune(ctx context.Context, root string, maxSize int64) error {
	// sanity
	if root == "" {
		return fmt.Errorf("invalid cache root directory: %s", root)
	}
	// short circuit
	if maxSize <= 0 {
		// don't remove blobinfocache, as this will incapacitate any bottles with
		// virtual parts
		err := os.RemoveAll(filepath.Join(root, "blobs"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("removing entire cache: %w", err)
		}
		return nil
	}

	// walk the blobs dir
	path := filepath.Join(root, ocispec.ImageBlobsDir)
	path = filepath.ToSlash(path)
	curSize, finfos, err := evalDir(path)
	if err != nil {
		return fmt.Errorf("evaluating cache directory: %w", err)
	}
	// short circuit
	if curSize <= maxSize {
		return nil
	}

	// TODO: we could use other metrics to determine which cached files are more valuable than others;
	// such as size, frequency of use, commonality of digest algorithm, etc. For now, we simply use
	// a last access time.
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

		err = deleteBlob(root, digest.Digest(filepath.Base(filepath.Dir(info.path))+":"+info.Name()))
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

func deleteBlob(root string, dgst digest.Digest) error {
	path, err := blobPath(dgst)
	if err != nil {
		return fmt.Errorf("%s: %w", dgst, errdef.ErrInvalidDigest)
	}
	targetPath := filepath.Join(root, path)
	err = os.Remove(targetPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w", dgst, errdef.ErrNotFound)
		}
		return fmt.Errorf("removing blob from cache: %w", err)
	}
	return nil
}

// evalDir evaluates a directory returning the total size of all files seen as well as their file infos.
func evalDir(fullPath string) (int64, []fileInfoWithDir, error) {
	var size int64
	seen := make(map[uint64]string)
	infos := make([]fileInfoWithDir, 0)

	fsys := os.DirFS(fullPath)
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
		wrappedFI := fileInfoWithDir{fi, filepath.Join(fullPath, path)}
		infos = append(infos, wrappedFI)

		inode, err := fsutil.GetInode(wrappedFI)
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

func (f fileInfoWithDir) Name() string {
	return f.path
}
