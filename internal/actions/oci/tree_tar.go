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

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// localArchiveReference prefers an existing local regular file over a repository reference.
// Try the whole argument first so a filename containing ':' or '@' is not split.
func localArchiveReference(raw string) (path, ref string, found bool, err error) {
	// check if the raw string is a file
    if info, statErr := os.Stat(raw); statErr == nil {
        if info.Mode().IsRegular() {
            return raw, "", true, nil
        }
        // A directory named like a registry component is not an archive.
        return "", "", false, nil
    } else if !errors.Is(statErr, os.ErrNotExist) {
        return "", "", false, fmt.Errorf("checking local path %q: %w", raw, statErr)
    }

    // Only split at a delimiter whose left side is an existing regular file.
    for _, i := range []int{strings.LastIndex(raw, "@"), strings.LastIndex(raw, ":")} {
        if i <= 0 || i == len(raw)-1 {
            continue // skip if delimiter is at the beginning or end.
        }
        candidate := raw[:i]
        info, statErr := os.Stat(candidate)
        if errors.Is(statErr, os.ErrNotExist) {
            continue // Canidate file DNE, try next delimiter
        }
        if statErr != nil {
            return "", "", false, fmt.Errorf("checking local path %q: %w", candidate, statErr)
        }
        if !info.Mode().IsRegular() {
            return "", "", false, fmt.Errorf("local path %q is not a regular file", candidate)
        }
		//Return the file path and tag/digest reference on the right side
        return candidate, raw[i+1:], true, nil
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
        return nil, fmt.Errorf("OCI archive index has no root descriptors")
    }
    return index.Manifests, nil
}

// extractOCILayoutTar extracts an OCI layout rooted at the archive root.
// The caller owns the returned directory and must remove it.
func extractOCILayoutTar(ctx context.Context, archivePath string) (dir string, err error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	// Read the first two bytes of a file
	buffered := bufio.NewReader(f)
	magic, err := buffered.Peek(2)
	if err != nil && err != io.EOF {
		return "", err
	}
	// decompress the tar.gz
	var archive io.Reader = buffered
	if len(magic) == 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		gz, err := gzip.NewReader(buffered)
		if err != nil {
			return "", fmt.Errorf("opening gzip stream: %w", err)
		}
		defer gz.Close()
		archive = gz
	}
	// create a temporary directory for extraction
	dir, err = os.MkdirTemp("", "ace-dt-oci-tree-*")
	if err != nil {
		return "", err
	}
	// if theres an error, delete the temporary direcotry
	defer func() {
		if err != nil {
			os.RemoveAll(dir)
		}
	}()

	// prevent tarbomb exploitation
	const maxExtractedBytes int64 = 100 << 30 // 100 GiB
	const maxEntries = 100000
	var extractedBytes int64
	var entries int
	

	tr := tar.NewReader(tarContextReader{ctx: ctx, reader: archive})
	for {
		hdr, nextErr := tr.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return "", fmt.Errorf("reading tar header: %w", nextErr)
		}
		entries++
		// checker for the maximum entries in the tarball
		if entries > maxEntries {
			return "", fmt.Errorf("tarball exceeds %d entries", maxEntries)
		}

		// Tar paths use forward slashes. Reject backslashes to avoid platform-dependent paths.
		if strings.Contains(hdr.Name, "\\") || filepath.IsAbs(hdr.Name) || filepath.VolumeName(hdr.Name) != "" {
			return "", fmt.Errorf("unsafe tar entry path %q", hdr.Name)
		}
		rel := filepath.Clean(filepath.FromSlash(hdr.Name))
		// check for paths that try to escape the extraction directory
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("unsafe tar entry path %q", hdr.Name)
		}
		if rel == "." {
			if hdr.Typeflag == tar.TypeDir {
				continue
			}
			return "", fmt.Errorf("unsafe tar entry path %q", hdr.Name)
		}
		path := filepath.Join(dir, rel)

		switch hdr.Typeflag {
		case tar.TypeDir:
			// ensure the directory is empty and has zero size
			if hdr.Size != 0 {
				return "", fmt.Errorf("directory %q has nonzero size", hdr.Name)
			}

			if err := os.MkdirAll(path, 0700); err != nil {
				return "", fmt.Errorf("creating directory %q: %w", hdr.Name, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if hdr.Size < 0 || hdr.Size > maxExtractedBytes-extractedBytes {
				return "", fmt.Errorf("tarball exceeds %d extracted bytes", maxExtractedBytes)
			}
			extractedBytes += hdr.Size
			// ensure the parent directory exists
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				return "", fmt.Errorf("creating parent for %q: %w", hdr.Name, err)
			}
			//uses create excl so it doesn't overwrite an existing file
			flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
			if rel == "index.json" {
				//handles multiple index.json files
				flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			}
			out, err := os.OpenFile(path, flags, 0600)
			if err != nil {
				return "", fmt.Errorf("creating file %q: %w", hdr.Name, err)
			}

			// uses CopyN so it only copies the same number of bytes as in the header
			_, copyErr := io.CopyN(out, tr, hdr.Size)
			closeErr := out.Close()
			if copyErr != nil {
				return "", fmt.Errorf("extracting file %q: %w", hdr.Name, copyErr)
			}
			if closeErr != nil {
				return "", fmt.Errorf("closing file %q: %w", hdr.Name, closeErr)
			}
		default:
			// ignores hardlinks and symlinks
			return "", fmt.Errorf("unsupported tar entry type %q for %q", hdr.Typeflag, hdr.Name)
		}
	}

	// checks that the directory follows the OCI layout spec
	for _, name := range []string{"oci-layout", "index.json"} {
		info, statErr := os.Stat(filepath.Join(dir, name))
		if statErr != nil || !info.Mode().IsRegular() {
			return "", fmt.Errorf("tarball must contain a regular %s at its root", name)
		}
	}
	// looks for blobs directory
	info, statErr := os.Stat(filepath.Join(dir, "blobs"))
	if statErr != nil || !info.IsDir() {
		return "", fmt.Errorf("tarball must contain a blobs directory at its root")
	}

	return dir, nil
}
// uses the io.Reader to check for cancellation
type tarContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r tarContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}