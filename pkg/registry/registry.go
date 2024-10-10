// Package registry provides options for remote OCI registry config and caching.
package registry

import (
	"context"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
)

// GraphTargeter provides a method for building an oras.GraphTarget.
type GraphTargeter interface {
	GraphTarget(ctx context.Context, reference string) (oras.GraphTarget, error)
}

// ReadOnlyGraphTargeter provides a method for building an oras.GraphTarget.
type ReadOnlyGraphTargeter interface {
	ReadOnlyGraphTarget(ctx context.Context, reference string) (oras.ReadOnlyGraphTarget, error)
}

// EndpointReferenceParser provides a method for parsing OCI references
// with alternate endpoint replacements.
type EndpointReferenceParser interface {
	ParseEndpointReference(reference string) (registry.Reference, error)
}

// EndpointGraphTargeter is a GraphTargeter that supports alternative reference endpoints.
type EndpointGraphTargeter interface {
	GraphTargeter
	EndpointReferenceParser
}

// ReadOnlyEndpointGraphTargeter is a ReadOnlyGraphTargeter that supports alternative reference endpoints.
type ReadOnlyEndpointGraphTargeter interface {
	ReadOnlyGraphTargeter
	EndpointReferenceParser
}
