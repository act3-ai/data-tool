package archive

import (
	"context"
	"fmt"
	"io"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

func ExtractTarFromReader(ctx context.Context, rc io.ReadCloser, destPath string) error {
	dec := &PipeIn{}
	dec.ConnectIn(rc)
	return extractWithPipeStreamFromReader(ctx, dec, destPath, false)
}

// parent directories are created as needed.
func ExtractTarCompatFromReader(ctx context.Context, rc io.ReadCloser, destPath string) error {
	dec := &PipeIn{}
	dec.ConnectIn(rc)
	return extractWithPipeStreamFromReader(ctx, dec, destPath, true)
}

func ExtractTarZstdFromReader(ctx context.Context, rc io.ReadCloser, destPath string) error {
	dec := &PipeZstdDec{}
	dec.ConnectIn(rc)
	return extractWithPipeStreamFromReader(ctx, dec, destPath, false)
}

// parent directories are created as needed.
func ExtractTarZstdCompatFromReader(ctx context.Context, rc io.ReadCloser, destPath string) error {
	dec := &PipeZstdDec{}
	dec.ConnectIn(rc)
	return extractWithPipeStreamFromReader(ctx, dec, destPath, true)
}

func ExtractZstdFromReader(ctx context.Context, rc io.ReadCloser, destPath string) error {
	log := logger.FromContext(ctx)

	log.DebugContext(ctx, "Decompressing with default zstd decompressor", "dest", destPath)
	dec := &PipeZstdDec{}
	dec.ConnectIn(rc)
	defer dec.Close()

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

func extractWithPipeStreamFromReader(ctx context.Context, in io.ReadCloser, destPath string, makeParents bool) error {
	log := logger.FromContext(ctx).With("dest", destPath)
	log.DebugContext(ctx, "Extracting archive")

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
