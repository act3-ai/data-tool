package util

import (
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/act3-ai/go-common/pkg/fsutil"
)

// DirSize returns the size of a directory.
func DirSize(fsys fs.FS) (int64, error) {
	var size int64
	seen := make(map[uint64]string)

	return size, fs.WalkDir(fsys, ".", func(path string, d os.DirEntry, err error) error {
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

// GetDirLastUpdate returns the last update time of a directory.
func GetDirLastUpdate(fsys fs.FS) (time.Time, error) {
	var lastTime time.Time

	return lastTime, fs.WalkDir(fsys, ".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("error getting file info: %w", err)
		}
		mtime := info.ModTime()
		if mtime.After(lastTime) {
			lastTime = mtime
		}
		return nil
	})
}

// GetDirUpdatedPaths returns a list of paths that have been updated since the given time.
func GetDirUpdatedPaths(fsys fs.FS, earliest time.Time) ([]string, error) {
	var paths []string

	return paths, fs.WalkDir(fsys, ".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("error getting file info: %w", err)
		}
		if info.IsDir() {
			return nil
		}
		if info.ModTime().After(earliest) {
			paths = append(paths, path)
		}
		return nil
	})
}
