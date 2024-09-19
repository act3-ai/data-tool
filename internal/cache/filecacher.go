package cache

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orascontent "oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
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
	// optimization
	fd, ok := content.(*os.File)
	switch {
	case ok:
		newPath, err := blobPath(expected.Digest)
		if err != nil {
			// fail early, as we would fail again later during the cache push anyways
			return fmt.Errorf("determining blob path: %w", err)
		}

		err = os.Link(fd.Name(), newPath)
		if err == nil {
			return nil
		}
		logger.FromContext(ctx).ErrorContext(ctx, "mounting file into file storage", "error", err)
		fallthrough
	default:
		return fm.Storage.Push(ctx, expected, content) //nolint
	}
}

// Same as oras-go
// Source: https://github.com/oras-project/oras-go/blob/v2.5.0/content/oci/readonlystorage.go#L93
func blobPath(dgst digest.Digest) (string, error) {
	if err := dgst.Validate(); err != nil {
		return "", fmt.Errorf("cannot calculate blob path from invalid digest %s: %w: %v",
			dgst.String(), errdef.ErrInvalidDigest, err)
	}
	return path.Join(ocispec.ImageBlobsDir, dgst.Algorithm().String(), dgst.Encoded()), nil
}
