// Package registry provides options for remote OCI registry config and caching.
package registry

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"

	reg "git.act3-ace.com/ace/data/tool/internal/registry"
	regcache "git.act3-ace.com/ace/data/tool/internal/registry/cache"
	"git.act3-ace.com/ace/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// NewGraphTargetFn defines a func used to create oras.GraphTarget references.
type NewGraphTargetFn func(ctx context.Context, ref string) (oras.GraphTarget, error)

// RemoteTargeter contains shared information used for creating new ORAS GraphTargets.
type RemoteTargeter struct {
	regCache  *regcache.RegistryCache
	regConfig *v1alpha1.RegistryConfig
	userAgent string
}

// NewRemoteTarget creates a ORAS remote.Repository. Implements RemoteTargeter.
func (t *RemoteTargeter) NewRemoteTarget(ctx context.Context, ref string, credOpts ...CredentialOption) (*remote.Repository, error) {
	log := logger.V(logger.FromContext(ctx), 1)

	// add fallback credentials
	credStore := credentials.NewMemoryStore()
	for _, opt := range credOpts {
		opt(credStore)
	}

	log.InfoContext(ctx, "Creating repository", "registryConfig", ref)
	regRepo, err := reg.CreateRepoWithCustomConfig(ctx, t.regConfig, ref, t.regCache, t.userAgent, credStore)
	if err != nil {
		return nil, fmt.Errorf("creating repository %q: %w", ref, err)
	}
	repo, ok := regRepo.(*remote.Repository)
	if !ok {
		return nil, fmt.Errorf("error creating registry repository: %s", ref)
	}

	return repo, nil
}

// NewRemoteTargeter creates a RemoteTargeter uitilizing the given registry configuration
// as well as a shared registry cache.
func NewRemoteTargeter(rcfg *v1alpha1.RegistryConfig, userAgent string) *RemoteTargeter {
	return &RemoteTargeter{
		regCache:  regcache.NewRegistryCache(),
		regConfig: rcfg,
		userAgent: userAgent,
	}
}

// CredentialOption applies optional credential configuration.
type CredentialOption func(credentials.Store)

// WithCredential adds credentials for a registry to an existing credential store.
func WithCredential(ctx context.Context, hostname string, cred auth.Credential) CredentialOption {
	return func(cs credentials.Store) {
		_ = cs.Put(ctx, credentials.ServerAddressFromHostname(hostname), cred)
	}
}
