package cache

import (
	"context"
	"errors"
	"fmt"
	"io"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// NilCache implements oras content.GraphStorage with empty functionality for cases
// when caching is disabled.
type NilCache struct {
}

// Exists returns false.
func (nc *NilCache) Exists(ctx context.Context, target ocispec.Descriptor) (bool, error) {
	return false, nil
}

// Fetch is not supported.
func (nc *NilCache) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	return nil, errors.ErrUnsupported
}

// Push prevents potential io blocking by discarding the provided reader.
func (nc *NilCache) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	_, err := io.Copy(io.Discard, content)
	if err != nil {
		return fmt.Errorf("discarding duplicate blob: %w", err)
	}
	return nil
}

// Predecessors always returns an empty set.
func (nc *NilCache) Predecessors(ctx context.Context, node ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	return []ocispec.Descriptor{}, nil
}
