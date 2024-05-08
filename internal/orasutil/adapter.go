// Package orasutil is a collection of oras extensions
package orasutil

import (
	"context"
	"io"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/data/tool/internal/descriptor"
)

// NewNoPushTarget is an oras.Target adapter that turns the Push() and Tag() into no-ops.
func NewNoPushTarget(target oras.Target) oras.Target {
	return &noPushTarget{target}
}

type noPushTarget struct {
	oras.Target
}

func (s *noPushTarget) Push(ctx context.Context, expected ocispec.Descriptor, r io.Reader) error {
	return nil
}

func (s *noPushTarget) Tag(ctx context.Context, desc ocispec.Descriptor, reference string) error {
	return nil
}

// NewExistenceCachedTarget adapts a oras.Target to cache results from Exists(), Push(), Fetch() to avoid unnecessary round-trips.
// This is sound so long as there are no concurrent deletes issued to the Target through another interface.
func NewExistenceCachedTarget(target oras.Target) oras.Target {
	return &existenceCachedTarget{
		Target: target,
	}
}

// Checks and populates the cache on calls to Exists(), Push(), Fetch().
type existenceCachedTarget struct {
	oras.Target

	// descriptor to existence (positive and negative)
	exists sync.Map // map[descriptor.Descriptor]bool
}

func (s *existenceCachedTarget) Exists(ctx context.Context, desc ocispec.Descriptor) (bool, error) {
	d := descriptor.FromOCI(desc)
	if v, ok := s.exists.Load(d); ok {
		return v.(bool), nil
	}

	v, err := s.Target.Exists(ctx, desc)
	if err != nil {
		return v, err //nolint:wrapcheck
	}

	// This Store() does not need to be atomic with Load()
	s.exists.Store(d, v)
	return v, nil
}

func (s *existenceCachedTarget) Push(ctx context.Context, desc ocispec.Descriptor, r io.Reader) error {
	d := descriptor.FromOCI(desc)
	if v, ok := s.exists.Load(d); ok {
		if v.(bool) {
			return errdef.ErrAlreadyExists
		}
	}

	err := s.Target.Push(ctx, desc, r)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// This Store() does not need to be atomic with Load()
	s.exists.Store(d, true)
	return nil
}

func (s *existenceCachedTarget) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	d := descriptor.FromOCI(desc)
	if v, ok := s.exists.Load(d); ok {
		if !v.(bool) {
			return nil, errdef.ErrNotFound
		}
	}

	r, err := s.Target.Fetch(ctx, desc)
	if err != nil {
		return r, err //nolint:wrapcheck
	}

	// This Store() does not need to be atomic with Load()
	s.exists.Store(d, true)
	return r, nil
}
