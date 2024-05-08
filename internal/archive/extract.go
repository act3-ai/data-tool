package archive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

func makeTarReadPipestream(path string) (PipeReader, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening tar path: %w", err)
	}
	dec := &PipeIn{}
	dec.ConnectIn(fp)
	return dec, nil
}

func makeZstReadPipestream(path string) (PipeReader, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	dec := NewPipeZstdDec()
	dec.ConnectIn(fp)
	return dec, nil
}

func makeOutfilePipestream(path string) (PipeWriter, error) {
	fp, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}
	out := &PipeOut{W: fp}
	return out, nil
}

// ReadPipeStreamCreator defines the signature for pipe stream reader creation functions for
// various encoding and compression algorithms.
type ReadPipeStreamCreator func(path string) (PipeReader, error)

func extractWithPipeStream(ctx context.Context, archivePath string, destPath string, makePipeStream ReadPipeStreamCreator, makeParents bool) error {
	log := logger.FromContext(ctx).With("src", archivePath, "dest", destPath)
	log.DebugContext(ctx, "Extracting archive")

	log.DebugContext(ctx, "creating unarchiver")
	in, err := makePipeStream(archivePath)
	if err != nil {
		return err
	}
	defer in.Close()

	// ensure destPath exists
	// TODO seems like this should be done by the caller
	if err := os.MkdirAll(destPath, 0o777); err != nil {
		return fmt.Errorf("error creating archivePath: %w", err)
	}

	tar := TarExtractor{
		OverwriteExisting:     true,
		MakeParentDirectories: makeParents,
	}

	log.DebugContext(ctx, "Unarchiving")
	err = tar.UnarchiveRead(in, destPath)
	if err != nil {
		return err
	}

	log.DebugContext(ctx, "Extract completed")
	return in.Close()
}

// ExtractTarZstd uses the mholt archiver library (v3.3 or greater)
// to inflate and unarchive a tar+zst file.  Any existing files
// are overwritten.
func ExtractTarZstd(ctx context.Context, archivePath string, destPath string) error {
	return extractWithPipeStream(ctx, archivePath, destPath, makeZstReadPipestream, false)
}

// ExtractTarZstdCompat uses the mholt archiver library (v3.3 or greater)
// to inflate and unarchive a tar+zst file.  Any existing files
// are overwritten
// parent directories are created as needed.
func ExtractTarZstdCompat(ctx context.Context, archivePath string, destPath string) error {
	return extractWithPipeStream(ctx, archivePath, destPath, makeZstReadPipestream, true)
}

// ExtractTar uses the mholt archiver library to unarchive a tar file.
// Any existing files are overwritten.
func ExtractTar(ctx context.Context, archivePath string, destPath string) error {
	return extractWithPipeStream(ctx, archivePath, destPath, makeTarReadPipestream, false)
}

// ExtractTarCompat uses the mholt archiver library to unarchive a tar file.
// Any existing files are overwritten
// parent directories are created as needed.
func ExtractTarCompat(ctx context.Context, archivePath string, destPath string) error {
	return extractWithPipeStream(ctx, archivePath, destPath, makeTarReadPipestream, true)
}

// ExtractZstd uses the mholt archiver library (v3.3 or greater)
// to inflate a .zst file.  Any existing files
// are overwritten. If delSource is true, the archive file itself
// is removed upon success.  To avoid naming conflicts, the source
// file can be renamed (renameSource option) prior to extraction.
func ExtractZstd(ctx context.Context, srcPath string, destPath string) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "Extracting from file", "src", srcPath, "dest", destPath)
	// defer timer.Measure(time.Now(), "Extract ["+srcPath+"]:")
	// rename the file to prepend a `.`, allowing single files to
	//  be archived and extracted without conflict.
	file := filepath.Base(srcPath)
	if file == "." || file == "/" {
		return fmt.Errorf("invalid file name specified for extraction")
	}

	arcName := srcPath

	log.DebugContext(ctx, "Removing destination file (if it exists)")
	// Remove the destination file if it already exists
	err := os.Remove(destPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("error removing destination file: %w", err)
	}

	log.DebugContext(ctx, "Decompressing with default zstd decompressor", "src", arcName, "dest", destPath)
	// Create a default zst decompressor
	in, err := makeZstReadPipestream(arcName)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := makeOutfilePipestream(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("error extracting zstd file: %w", err)
	}

	log.DebugContext(ctx, "Extract completed", "dest", destPath)
	if err := out.Close(); err != nil {
		return err
	}
	return in.Close()
}
