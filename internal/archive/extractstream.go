package archive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// ExtractTar unarchives a tar file.
// Existing files are overwritten.
func ExtractTar(ctx context.Context, rc io.ReadCloser, destPath string) error {
	dec := &PipeIn{}
	dec.ConnectIn(rc)
	return extractWithPipeStream(ctx, dec, destPath, false)
}

// ExtractTarCompat unarchives a tar file.
// Existing files are overwritten. Parent directories are created as needed.
func ExtractTarCompat(ctx context.Context, rc io.ReadCloser, destPath string) error {
	dec := &PipeIn{}
	dec.ConnectIn(rc)
	return extractWithPipeStream(ctx, dec, destPath, true)
}

// ExtractTarZstd decompresses and unarchives a tar+zst file.
// Existing files are overwritten.
func ExtractTarZstd(ctx context.Context, rc io.ReadCloser, destPath string) error {
	dec := &PipeZstdDec{}
	dec.ConnectIn(rc)
	return extractWithPipeStream(ctx, dec, destPath, false)
}

// ExtractTarZstdCompat decompresses and unarchives a tar+zst file.
// Existing files are overwritten. Parent directories are created as needed.
func ExtractTarZstdCompat(ctx context.Context, rc io.ReadCloser, destPath string) error {
	dec := &PipeZstdDec{}
	dec.ConnectIn(rc)
	return extractWithPipeStream(ctx, dec, destPath, true)
}

// ExtractZstd decompresses a .zst file.
// Existing files are overwritten.
func ExtractZstd(ctx context.Context, rc io.ReadCloser, destPath string) error {
	log := logger.FromContext(ctx)

	log.DebugContext(ctx, "Decompressing with default zstd decompressor", "dest", destPath)
	dec := &PipeZstdDec{}
	dec.ConnectIn(rc)
	defer dec.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0777); err != nil {
		return fmt.Errorf("initializing part parent directories: %w", err)
	}

	out, err := makeOutfilePipestream(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, dec)
	if err != nil {
		return fmt.Errorf("error extracting zstd file: %w", err)
	}

	log.DebugContext(ctx, "Extract completed", "dest", destPath)
	if err := out.Close(); err != nil {
		return err
	}

	if err := dec.Close(); err != nil {
		return fmt.Errorf("closing zst decompressor: %w", err)
	}
	return nil
}

// extractWithPipeStream unarchives from the provided reader.
func extractWithPipeStream(ctx context.Context, in io.ReadCloser, destPath string, makeParents bool) error {
	log := logger.FromContext(ctx).With("dest", destPath)
	log.DebugContext(ctx, "Extracting archive")

	if err := os.MkdirAll(destPath, 0777); err != nil {
		return fmt.Errorf("creating archivePath: %w", err)
	}

	tar := TarExtractor{
		OverwriteExisting:     true,
		MakeParentDirectories: makeParents,
	}

	err := tar.UnarchiveRead(in, destPath)
	if err != nil {
		return err
	}

	log.DebugContext(ctx, "Extract completed")
	return in.Close()
}

func makeOutfilePipestream(path string) (PipeWriter, error) {
	fp, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}
	out := &PipeOut{W: fp}
	return out, nil
}
