package testing

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type filesystemHelper struct {
	fsys fs.FS

	// seen is a set of filenames that have been depended upon by another template.
	// It is an error to process a file that has already been depended upon.
	seen map[string]struct{}
}

func (fsh *filesystemHelper) fileContents(filename string) ([]byte, error) {
	if fsh.seen == nil {
		fsh.seen = map[string]struct{}{}
	}
	fsh.seen[filename] = struct{}{}
	// fmt.Println("dependency: ", filename)
	return fs.ReadFile(fsh.fsys, filename) //nolint
}

// FileDescriptor constructs a OCI descriptor from a file.
func (fsh *filesystemHelper) FileDescriptor(filename string, mediaType string, algorithm digest.Algorithm) (*ocispec.Descriptor, error) {
	data, err := fsh.fileContents(filename)
	if err != nil {
		return nil, err
	}

	return fsh.fileDescriptorFromData(data, mediaType, algorithm)
}

func (fsh *filesystemHelper) fileDescriptorFromData(data []byte, mediaType string, algorithm digest.Algorithm) (*ocispec.Descriptor, error) {
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}

	_, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return nil, fmt.Errorf("invalid media type %q: %w", mediaType, err)
	}

	if algorithm == "" {
		algorithm = digest.Canonical
	}

	if !algorithm.Available() {
		return nil, fmt.Errorf("digest \"%s\" not available", algorithm)
	}

	dgst := algorithm.FromBytes(data)

	desc := ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    dgst,
		Size:      int64(len(data)),
	}

	return &desc, nil
}

// FileDigest constructs a OCI descriptor from a file.
func (fsh *filesystemHelper) FileDigest(filename string, algorithm digest.Algorithm) (digest.Digest, error) {
	if !algorithm.Available() {
		return "", fmt.Errorf("digest \"%s\" not available", algorithm)
	}

	data, err := fsh.fileContents(filename)
	if err != nil {
		return "", err
	}

	dgst := algorithm.FromBytes(data)

	return dgst, nil
}

// tar archives the given files.
func (fsh *filesystemHelper) Tar(filenames ...string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, filename := range filenames {
		if err := fsh.tarFile(tw, filename); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("closing archiver: %w", err)
	}

	return buf.Bytes(), nil
}

func (fsh *filesystemHelper) tarFile(tw *tar.Writer, filename string) error {
	data, err := fsh.fileContents(filename)
	if err != nil {
		return err
	}
	size := int64(len(data))

	hdr := tar.Header{
		Name:     filename,
		Size:     int64(len(data)),
		Mode:     0666,
		ModTime:  time.Unix(0, 0).UTC(),
		Typeflag: tar.TypeReg,
	}

	if err := tw.WriteHeader(&hdr); err != nil {
		return fmt.Errorf("writing file header: %w", err)
	}

	n, err := tw.Write(data)
	if err != nil {
		return fmt.Errorf("copying data into archive: %w", err)
	}
	// covering size mismatch from descriptor to what we actually downloaded
	if n != len(data) {
		return fmt.Errorf("copied %d B but expected %d B: %w", n, size, io.ErrShortWrite)
	}

	return nil
}

// gzipHelper compresses the input with GZIP.
func gzipHelper(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	// Setting the Header fields is optional.
	// zw.Name = "filename.txt"
	// zw.Comment = "templated"
	// zw.ModTime = time.Date(1977, time.May, 25, 0, 0, 0, 0, time.UTC)

	if _, err := zw.Write(input); err != nil {
		return nil, fmt.Errorf("compressing: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("closing compressor: %w", err)
	}

	return buf.Bytes(), nil
}

// zstdHelper compresses the input with Zstd.
func zstdHelper(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw, err := zstd.NewWriter(&buf)
	if err != nil {
		return nil, fmt.Errorf("creating compressing: %w", err)
	}

	if _, err = zw.Write(input); err != nil {
		return nil, fmt.Errorf("compressing: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("closing compressor: %w", err)
	}

	return buf.Bytes(), nil
}
