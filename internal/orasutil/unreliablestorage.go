package orasutil

import (
	"context"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
)

// UnreliableStorage overwrites the Exists function to consistently fetch manifests
// if the underlying content.GraphStorage cannot guarantee all successors also
// exist. Useful in a oras.CopyGraph context.
type UnreliableStorage struct {
	content.Storage
}

// Exists overwrites the underlying content.Storage's Exists by returning false
// if a descriptor has a manifest or index mediatype.
func (u *UnreliableStorage) Exists(ctx context.Context, target ocispec.Descriptor) (bool, error) {
	if target.MediaType == ocispec.MediaTypeImageManifest ||
		target.MediaType == ocispec.MediaTypeImageIndex {
		return false, nil
	}

	return u.Storage.Exists(ctx, target) //nolint
}
