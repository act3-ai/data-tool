package cache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orascontent "oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	"github.com/act3-ai/go-common/pkg/logger"
)

// fileMounter wraps an oras content.Storage push func with file linking.
// Must only be used with content.Storage interfaces that use the local
// filesystem to store blobs.
type fileMounter struct {
	orascontent.Storage

	root string
}

// NewFileMounter returns a FileMounter as an oras content.Storage.
//
// fileMounter wraps an oras content.Storage push func with file linking.
// Must only be used with content.Storage interfaces that use the local
// filesystem to store blobs.
func NewFileMounter(root string, storage orascontent.Storage) (orascontent.Storage, error) {
	// TODO: Type assert input storage. If an oci.Memory, remote.Repository, or
	// the like then return an error.
	return &fileMounter{
		Storage: storage,
		root:    root,
	}, nil
}

// Push provides an optimization if the io.Reader is an open os.File, otherwise
// it uses the underlying content.Storage Push func.
func (fm *fileMounter) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	log := logger.FromContext(ctx)
	// optimization
	fd, ok := content.(*os.File)
	switch {
	case ok:
		newPath, err := blobPath(expected.Digest)
		if err != nil {
			// fail early, as we would fail again later during the cache push anyways
			return fmt.Errorf("determining blob path: %w", err)
		}

		fullNewPath := filepath.Join(fm.root, newPath)

		err = os.MkdirAll(filepath.Dir(fullNewPath), 0777)
		if err != nil {
			return fmt.Errorf("initializing blob path %s: %w", fullNewPath, err)
		}

		err = os.Link(fd.Name(), fullNewPath)
		if err == nil || errors.Is(err, os.ErrExist) {
			return nil
		}

		log.ErrorContext(ctx, "mounting file into file storage", "error", err)
		fallthrough
	default:
		return fm.Storage.Push(ctx, expected, content) //nolint
	}
}

// Same as oras-go
// Source: https://github.com/oras-project/oras-go/blob/v2.5.0/content/oci/readonlystorage.go#L93
func blobPath(dgst digest.Digest) (string, error) {
	if err := dgst.Validate(); err != nil {
		return "", fmt.Errorf("cannot calculate blob path from invalid digest %s: %w: %w",
			dgst.String(), errdef.ErrInvalidDigest, err)
	}
	return filepath.Join(ocispec.ImageBlobsDir, dgst.Algorithm().String(), dgst.Encoded()), nil
}
