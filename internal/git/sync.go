package git

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"

	"git.act3-ace.com/ace/data/tool/internal/git/cache"
	"git.act3-ace.com/ace/data/tool/internal/git/cmd"
	"git.act3-ace.com/ace/data/tool/internal/git/oci"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// sync contains data and helpers used in git to OCI and OCI to git transfers.
type sync struct {
	ociHelper *oci.Helper
	cmdHelper *cmd.Helper

	base  syncBase
	lfs   syncLFS
	cache cache.ObjectCache

	syncOpts SyncOptions
}

// syncBase represents a Commit Manifest and its Config.
type syncBase struct {
	manifest ocispec.Manifest
	config   Config
}

// syncLFS represents a LFS Manifest and its Config.
type syncLFS struct {
	manifest ocispec.Manifest
	config   LFSConfig // not used, remains for plumbing
}

// SyncOptions modify git to OCI and OCI to git processes.
type SyncOptions struct {
	Clean     bool
	DTVersion string
	TmpDir    string
	CacheDir  string
}

// FetchBaseManifestConfig fetches the base sync manifest and config, populating the Original with the
// results or initializing it with empty fields. Returns a descriptor to the commit manifest if it exists.
func (s *sync) FetchBaseManifestConfig(ctx context.Context) (ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)

	var err error
	var manifestDesc ocispec.Descriptor
	var manifestBytes []byte

	if !s.syncOpts.Clean {
		log.InfoContext(ctx, "Fetching manifest", "tag", s.ociHelper.Tag)
		manifestDesc, manifestBytes, err = oras.FetchBytes(ctx, s.ociHelper.Target, s.ociHelper.Tag, oras.FetchBytesOptions{})
	} else {
		log.InfoContext(ctx, "Starting with a fresh base manifest")
	}

	switch {
	case s.syncOpts.Clean || errors.Is(err, errdef.ErrNotFound):
		s.base.config.Refs.Tags = make(map[string]ReferenceInfo, 0)
		s.base.config.Refs.Heads = make(map[string]ReferenceInfo, 0)
		return manifestDesc, errdef.ErrNotFound // propagate the error and handle accordingly, this can be ignored in ToOCI, but not in FromOCI
	case err != nil:
		return manifestDesc, fmt.Errorf("fetching base manifest: %w", err)
	}

	log.InfoContext(ctx, "Using existing manifest")
	err = json.Unmarshal(manifestBytes, &s.base.manifest)
	if err != nil {
		return manifestDesc, fmt.Errorf("unmarshaling current base manifest: %w", err)
	}

	// check types
	if s.base.manifest.ArtifactType != ArtifactTypeSyncManifest { // likely error if we check artifact type in descriptor and not manifest itself
		return manifestDesc, fmt.Errorf("expected base manifest artifact type %s, got %s", ArtifactTypeSyncManifest, manifestDesc.ArtifactType)
	}
	if s.base.manifest.Config.MediaType != MediaTypeSyncConfig {
		return manifestDesc, fmt.Errorf("expected base config media type %s, got %s", MediaTypeSyncConfig, s.base.manifest.Config.MediaType)
	}

	log.InfoContext(ctx, "Fetching manifest config", "configDigest", s.base.manifest.Config.Digest.String())
	currentConfigBytes, err := content.FetchAll(ctx, s.ociHelper.Target, s.base.manifest.Config)
	if err != nil {
		return manifestDesc, fmt.Errorf("fetching manifest config: %w", err)
	}

	err = json.Unmarshal(currentConfigBytes, &s.base.config)
	if err != nil {
		return manifestDesc, fmt.Errorf("unmarshaling current manifest configuration: %w", err)
	}

	return manifestDesc, nil
}

// FetchLFSManifestConfig copies all predecessor manifests with the LFS manifest media type, returning a pointer to an OCI CAS storage
// containing the result of the copy.
func (s *sync) FetchLFSManifestConfig(ctx context.Context, root ocispec.Descriptor, clean bool) error {
	log := logger.FromContext(ctx)

	// TODO: This feels like a hacky way to test that the root did not exist before
	if clean || (root.Digest == "") {
		log.InfoContext(ctx, "Starting with a fresh LFS manifest")
		// if adding config fields, initialize them here
		return fmt.Errorf("clean option or subject not found: %w", errdef.ErrNotFound)
	}

	log.InfoContext(ctx, "Resolving commit manifest referrers", "root", root)
	referrers, err := registry.Referrers(ctx, s.ociHelper.Target, root, ArtifactTypeLFSManifest)
	log.InfoContext(ctx, "Found commit manifest referrers", "referrers", referrers)

	// we expect one LFS manifest referrer
	switch {
	case len(referrers) < 1:
		return errdef.ErrNotFound
	case len(referrers) > 1:
		return fmt.Errorf("expected 1 LFS referrer, got %d", len(referrers)) // should never hit
	case err != nil:
		return fmt.Errorf("resolving commit manifest predecessors: %w", err)
	}
	lfsManifestDesc := referrers[0]

	log.InfoContext(ctx, "Fetching LFS manifest", "desc", lfsManifestDesc)
	lfsManifestBytes, err := content.FetchAll(ctx, s.ociHelper.Target, lfsManifestDesc)
	if err != nil {
		return fmt.Errorf("fetching LFS manifest: %w", err)
	}

	err = json.Unmarshal(lfsManifestBytes, &s.lfs.manifest)
	if err != nil {
		return fmt.Errorf("decoding LFS manifest descriptor: %w", err)
	}

	// check types
	// if config returns, check artifact type here
	if s.lfs.manifest.ArtifactType != ArtifactTypeLFSManifest { // likely error if we check artifact type in descriptor and not manifest itself
		return fmt.Errorf("expected LFS manifest artifact type %s, got %s", ArtifactTypeLFSManifest, lfsManifestDesc.ArtifactType)
	}

	return nil
}

// cleanup closes and cleans up any temporary files created during the sync process.
func (s *sync) cleanup() error {
	if err := s.ociHelper.FStore.Close(); err != nil {
		return fmt.Errorf("closing filestore: %w", err)
	}

	if err := os.RemoveAll(s.syncOpts.TmpDir); err != nil {
		return fmt.Errorf("cleaning up intermediate repository: %w", err)
	}

	return nil
}
