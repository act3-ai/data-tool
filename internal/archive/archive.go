// Package archive implements methods for working with archives, including archiving and compressing tar archives using
// tar and zstd libraries.  Additionally, a stream pipeline is defined that enables archival, compression, and digesting
// operations to occur in a modular fashion with one pass through a file.
package archive

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// TarToStream archives a list of files
// and writes the archive to the provided stream.
func TarToStream(ctx context.Context, fsys fs.FS, ostream io.Writer, options ...TarArchiverOption) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Creating tar archiver")

	ar, err := NewArchiver(ostream, options...)
	if err != nil {
		return err
	}
	defer ar.Close()
	log.InfoContext(ctx, "Archiving and compressing")
	defer log.InfoContext(ctx, "Archive completed")

	// related accepted proposal https://github.com/golang/go/issues/49580

	var walkFn func(path string, d os.DirEntry, err error) error
	walkFn = func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// We use fs.Stat (instead of d.Info()) so that the FileInfo is the info for the actual file (data) and not the symlink
		info, err := fs.Stat(fsys, path)
		if err != nil {
			return fmt.Errorf("unable to get the actual file info: %w", err)
		}

		if info.IsDir() {
			if d.Name() == "." {
				// We do not include the top level "./" directory in the archive
				// That is the job of the bottle to ensure it exists before this tar archive is extracted
				return nil
			}

			if strings.HasPrefix(d.Name(), ".") {
				// ignore hidden directories
				return fs.SkipDir
			}

			dinfo, err := d.Info()
			if err != nil {
				return fmt.Errorf("error getting tar directory info: %w", err)
			}
			// check to see if it is a symlink
			if dinfo.Mode().Type()&fs.ModeSymlink != 0 {
				// symlink to directory
				// manually recurse because fs.WalkDir does not do this for us
				// TODO what about infinite recursion (I suspect there are cases where it could happen).
				// Something like this approach might work well https://github.com/edwardrf/symwalk
				return fs.WalkDir(fsys, path, walkFn)
			}

			// regular directory
			return ar.WriteEntry(fsys, path, info)
		}

		// file
		if strings.HasPrefix(d.Name(), ".") {
			// ignore hidden files
			return nil
		}
		return ar.WriteEntry(fsys, path, info)
	}

	// this walks in lexicographic order so it is deterministic
	return fs.WalkDir(fsys, ".", walkFn)
}
