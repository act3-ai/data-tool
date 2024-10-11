// Package registry provides options for remote OCI registry config and caching.
package registry

import (
	"context"

	"oras.land/oras-go/v2"
)

// GraphTargeter provides a method for building an oras.GraphTarget.
type GraphTargeter interface {
	GraphTarget(ctx context.Context, reference string) (oras.GraphTarget, error)
}

// ReadOnlyGraphTargeter provides a method for building an oras.GraphTarget.
type ReadOnlyGraphTargeter interface {
	ReadOnlyGraphTarget(ctx context.Context, reference string) (oras.ReadOnlyGraphTarget, error)
}
