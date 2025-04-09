// Package orasutil is a collection of oras extensions
package orasutil

import (
	"context"
	"io"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	"github.com/act3-ai/data-tool/internal/descriptor"
)

// NewNoPushStorage is an oras content.Storage adapter that turns the Push() func into a no-op.
func NewNoPushStorage(storage content.Storage) content.Storage {
	return &noPushStorage{storage}
}

type noPushStorage struct {
	content.Storage
}

func (s *noPushStorage) Push(ctx context.Context, expected ocispec.Descriptor, r io.Reader) error {
	return nil
}

// NewExistenceCachedStorage adapts a content.Storage to cache results from Exists(), Push(), Fetch() to avoid unnecessary round-trips.
// This is sound so long as there are no concurrent deletes issued to the Storage through another interface.
func NewExistenceCachedStorage(storage content.Storage) content.Storage {
	return &ExistenceCachedStorage{
		Storage: storage,
	}
}

// ExistenceCachedStorage checks and populates the cache on calls to Exists(), Push(), Fetch().
type ExistenceCachedStorage struct {
	content.Storage

	// descriptor to existence (positive and negative)
	exists sync.Map // map[descriptor.Descriptor]bool
}

// Exists attempts to load the cached results of a previous call, falling back to the storage if not loaded.
func (s *ExistenceCachedStorage) Exists(ctx context.Context, desc ocispec.Descriptor) (bool, error) {
	d := descriptor.FromOCI(desc)

	if v, ok := s.exists.Load(d); ok {
		return v.(bool), nil
	}

	v, err := s.Storage.Exists(ctx, desc)
	if err != nil {
		return v, err //nolint:wrapcheck
	}

	// This Store() does not need to be atomic with Load()
	s.exists.Store(d, v)
	return v, nil
}

// Push is instrumented with a cached existence check, used before pushing to the storage.
func (s *ExistenceCachedStorage) Push(ctx context.Context, desc ocispec.Descriptor, r io.Reader) error {
	d := descriptor.FromOCI(desc)

	if v, ok := s.exists.Load(d); ok {
		if v.(bool) {
			return errdef.ErrAlreadyExists
		}
	}

	err := s.Storage.Push(ctx, desc, r)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// This Store() does not need to be atomic with Load()
	s.exists.Store(d, true)
	return nil
}

// Fetch is instrumented with a cached existence check, used before fetching from the storage.
func (s *ExistenceCachedStorage) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	d := descriptor.FromOCI(desc)

	if v, ok := s.exists.Load(d); ok {
		if !v.(bool) {
			return nil, errdef.ErrNotFound
		}
	}

	r, err := s.Storage.Fetch(ctx, desc)
	if err != nil {
		return r, err //nolint:wrapcheck
	}

	// This Store() does not need to be atomic with Load()
	s.exists.Store(d, true)
	return r, nil
}
