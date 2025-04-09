package registry

import (
	"context"
	"fmt"
	"net/url"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"

	"github.com/act3-ai/data-tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
	reg "github.com/act3-ai/data-tool/pkg/registry"
)

// EndpointReferenceParser provides a method for parsing OCI references
// with alternate endpoint replacements.
type EndpointReferenceParser interface {
	ParseEndpointReference(reference string) (registry.Reference, error)
}

// EndpointGraphTargeter is a GraphTargeter that supports alternative reference endpoints.
type EndpointGraphTargeter interface {
	reg.GraphTargeter
	EndpointReferenceParser
}

// ReadOnlyEndpointGraphTargeter is a ReadOnlyGraphTargeter that supports alternative reference endpoints.
type ReadOnlyEndpointGraphTargeter interface {
	reg.ReadOnlyGraphTargeter
	EndpointReferenceParser
}

// ParseEndpointOrDefault checks a GraphTargeter to see if parsing with endpoint resolution is necessary,
// defaulting to the oras registry.ParseReference.
func ParseEndpointOrDefault(targeter reg.GraphTargeter, reference string) (registry.Reference, error) {
	var ref registry.Reference
	var err error
	if endpointTargeter, ok := targeter.(EndpointGraphTargeter); ok {
		ref, err = endpointTargeter.ParseEndpointReference(reference)
		if err != nil {
			return registry.Reference{}, fmt.Errorf("parsing endpoint reference '%s': %w", reference, err)
		}
	} else {
		ref, err = registry.ParseReference(reference)
		if err != nil {
			return registry.Reference{}, fmt.Errorf("parsing reference '%s': %w", reference, err)
		}
	}

	ref.Reference = ref.ReferenceOrDefault()
	return ref, nil
}

// endpointResolver wraps an oras Resolve func to override references with
// an alternative endpoint.
type endpointResolver struct {
	oras.GraphTarget

	host string
}

// NewEndpointResolver provides capabilities to overwrite the Registry
// portion of references passed to the Resolve method. Use ResolveEndpoint to
// determine the correct host.
func NewEndpointResolver(resolver oras.GraphTarget, host string) oras.GraphTarget {
	return &endpointResolver{
		GraphTarget: resolver,
		host:        host,
	}
}

// Resolve overwrites the Registry portion of references passed to the Resolve method.
// At a minimum it requires a registry and repository in the reference, which is
// notably the opposite of what oras-go can handle.
// A returned Descriptor remains unchanged.
func (er *endpointResolver) Resolve(ctx context.Context, reference string) (ocispec.Descriptor, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return ocispec.Descriptor{}, err //nolint
	}

	ref.Registry = er.host // new endpoint
	ref.Reference = ref.ReferenceOrDefault()
	return er.GraphTarget.Resolve(ctx, ref.String()) //nolint
}

// ResolveEndpoint checks for alternative registry endpoints in an RegistryConfig.
// It returns the original endpoint if one was not found.
// Currently only supports handling the first endpoint in the config.
func ResolveEndpoint(rc *v1alpha1.RegistryConfig, ref registry.Reference) (*url.URL, error) {
	defaultEndpoint := "https://" + ref.Registry
	endpoints := make([]string, 0, 2) // typically default + {0,1}

	// configured endpoints take precedence
	r := rc.Configs[ref.Registry]
	if len(r.Endpoints) != 0 {
		endpoints = append(endpoints, r.Endpoints...)
	}
	endpoints = append(endpoints, defaultEndpoint)

	// handle endpoints
	// TODO: support more than first endpoint when registry.Ping is better supported
	// by registries (nvcr.io and quay.io)
	endpoint := endpoints[0]

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parsing endpoint URL: %w", err)
	}

	return endpointURL, nil
}
