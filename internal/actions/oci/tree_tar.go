package oci

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

// localArchiveReference prefers an existing local regular file over a repository reference.
// Try the whole argument first so a filename containing ':' or '@' is not split.
func localArchiveReference(raw string) (path, ref string, found bool, err error) {
	// check if the raw string is a file
	info, statErr := os.Stat(raw)
	if statErr == nil && info.Mode().IsRegular() {
		return raw, "", true, nil
	}
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return "", "", false, fmt.Errorf("checking local path %q: %w", raw, statErr)
	}

	if _, err := registry.ParseReference(raw); err != nil {
		return "", "", false, fmt.Errorf("invalid registry reference %q: %w", raw, err)
	}
	return "", "", false, nil
}

// archiveRoots returns every root descriptor from the final index.json in index order.
func archiveRoots(dir string) ([]ocispec.Descriptor, error) {
	// There must be a root.json at the root.
	data, err := os.ReadFile(filepath.Join(dir, "index.json"))
	if err != nil {
		return nil, fmt.Errorf("reading OCI index: %w", err)
	}
	var index ocispec.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("parsing OCI index: %w", err)
	}
	if len(index.Manifests) == 0 {
		return nil, errors.New("OCI archive index has no root descriptors")
	}
	return index.Manifests, nil
}

// extractOCILayoutTar extracts an OCI layout rooted at the archive root into dir.
// The caller owns dir and must remove it, including when extraction fails.
func extractOCILayoutTar(ctx context.Context, archivePath, dir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("opening archive %q: %w", archivePath, err)
	}
	defer f.Close()
	// Read the first four bytes of a file.
	buffered := bufio.NewReader(f)
	magic, err := buffered.Peek(4)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("reading archive header: %w", err)
	}
	// decompress the tar.gz or tar.zst.
	var archive io.Reader = buffered
	switch {
	case len(magic) >= 2 && magic[0] == 0x1f && magic[1] == 0x8b:
		gz, err := gzip.NewReader(buffered)
		if err != nil {
			return fmt.Errorf("opening gzip stream: %w", err)
		}
		defer gz.Close()
		archive = gz
	case len(magic) == 4 && magic[0] == 0x28 && magic[1] == 0xb5 && magic[2] == 0x2f && magic[3] == 0xfd:
		zr, err := zstd.NewReader(buffered)
		if err != nil {
			return fmt.Errorf("opening zstd stream: %w", err)
		}
		defer zr.Close()
		archive = zr
	}

	tr := tar.NewReader(tarContextReader{canceled: ctx.Err, reader: archive})
	if err := extractTarEntries(tr, dir); err != nil {
		return err
	}
	return validateExtractedOCILayout(dir)

}

// extractTarEntries enforces the archive-wide extracted-size limit.
func extractTarEntries(tr *tar.Reader, dir string) error {
	// prevent tarbomb exploitation.
	const maxExtractedBytes int64 = 100 << 30 // 100 GiB
	var extractedBytes int64

	for {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("reading tar header: %w", err)
		}

		rel, err := safeTarEntryPath(hdr.Name)
		if err != nil {
			return err
		}
		if rel == "." && hdr.Typeflag == tar.TypeDir {
			continue
		}
		if rel == "." {
			return fmt.Errorf("unsafe tar entry path %q", hdr.Name)
		}

		if hdr.Typeflag == tar.TypeReg || hdr.Typeflag == 0 {
			if hdr.Size < 0 || hdr.Size > maxExtractedBytes-extractedBytes {
				return fmt.Errorf("OCI archive exceeds the %d-byte extraction limit; extract it yourself and pass the OCI layout directory instead", maxExtractedBytes)
			}
		}
		if err := extractTarEntry(tr, hdr, filepath.Join(dir, rel), rel, &extractedBytes, maxExtractedBytes); err != nil {
			return err
		}
	}
}

// safeTarEntryPath rejects entries outside the extraction directory.
func safeTarEntryPath(name string) (string, error) {
	// Tar paths use forward slashes. Reject backslashes to avoid platform-dependent paths.
	if strings.Contains(name, "\\") || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return "", fmt.Errorf("unsafe tar entry path %q", name)
	}
	rel := filepath.Clean(filepath.FromSlash(name))
	// check for paths that try to escape the extraction directory
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe tar entry path %q", name)
	}
	return rel, nil
}

// extractTarEntry writes one regular file or creates one directory.
func extractTarEntry(tr *tar.Reader, hdr *tar.Header, path, rel string, extractedBytes *int64, maxExtractedBytes int64) error {
	switch hdr.Typeflag {
	case tar.TypeDir:
		// ensure the directory is empty and has zero size.
		if hdr.Size != 0 {
			return fmt.Errorf("directory %q has nonzero size", hdr.Name)
		}
		if err := os.MkdirAll(path, 0700); err != nil {
			return fmt.Errorf("creating directory %q: %w", hdr.Name, err)
		}
		return nil
	case tar.TypeReg, 0:
		if hdr.Size < 0 || hdr.Size > maxExtractedBytes-*extractedBytes {
			return fmt.Errorf("tarball exceeds %d extracted bytes", maxExtractedBytes)
		}
		*extractedBytes += hdr.Size
		// ensure the parent directory exists.
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return fmt.Errorf("creating parent for %q: %w", hdr.Name, err)
		}
		// Uses create excl so it doesn't overwrite an existing file.
		flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
		if rel == "index.json" {
			// handles multiple index.json files.
			flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		}
		out, err := os.OpenFile(path, flags, 0600)
		if err != nil {
			return fmt.Errorf("creating file %q: %w", hdr.Name, err)
		}

		// uses CopyN so it only copies the same number of bytes as in the header.
		_, copyErr := io.CopyN(out, tr, hdr.Size)
		closeErr := out.Close()
		if copyErr != nil {
			return fmt.Errorf("extracting file %q: %w", hdr.Name, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("closing file %q: %w", hdr.Name, closeErr)
		}
		return nil
	default:
		// ignores hardlinks and symlinks.
		return fmt.Errorf("unsupported tar entry type %q for %q", hdr.Typeflag, hdr.Name)
	}
}

// validateExtractedOCILayout checks the required root entries.
func validateExtractedOCILayout(dir string) error {
	// checks that the directory follows the OCI layout spec.
	for _, name := range []string{"oci-layout", "index.json"} {
		info, statErr := os.Stat(filepath.Join(dir, name))
		if statErr != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("tarball must contain a regular %s at its root", name)
		}
	}
	// looks for blobs directory.
	info, statErr := os.Stat(filepath.Join(dir, "blobs"))
	if statErr != nil || !info.IsDir() {
		return errors.New("tarball must contain a blobs directory at its root")

	}
	return nil
}

// uses the io.Reader to check for cancellation.
type tarContextReader struct {
	canceled func() error
	reader   io.Reader
}

func (r tarContextReader) Read(p []byte) (int, error) {
	if err := r.canceled(); err != nil {
		return 0, err
	}

	//nolint:wrapcheck // Preserve io.Reader errors, including io.EOF, unchanged.
	return r.reader.Read(p)

}
