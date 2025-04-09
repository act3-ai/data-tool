package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"

	"github.com/act3-ai/data-tool/internal/git/cache"
	"github.com/act3-ai/data-tool/internal/git/cmd"
	"github.com/act3-ai/data-tool/internal/git/oci"
	"github.com/act3-ai/go-common/pkg/logger"
)

// sync contains data and helpers used in git to OCI and OCI to git transfers.
type sync struct {
	ociHelper *oci.Helper
	cmdHelper *cmd.Helper

	base syncBase
	lfs  syncLFS

	syncOpts SyncOptions
}

// syncBase represents a Commit Manifest and its Config.
type syncBase struct {
	manDesc  ocispec.Descriptor
	manifest ocispec.Manifest
	config   oci.Config
}

// syncLFS represents a LFS Manifest and its Config.
type syncLFS struct {
	manifest ocispec.Manifest
	config   oci.LFSConfig // not used, remains for plumbing
}

// SyncOptions modify git to OCI and OCI to git processes.
type SyncOptions struct {
	Clean             bool
	UserAgent         string
	IntermediateDir   string
	IntermediateStore *file.Store // TODO: This is a duplicate of what's in OCIHelper, let's remove OCIHelper
	Cache             cache.ObjectCacher
}

// FetchBaseManifestConfig fetches the base sync manifest and config, populating the Original with the
// results or initializing it with empty fields.
func (s *sync) FetchBaseManifestConfig(ctx context.Context) error {
	log := logger.FromContext(ctx)

	manifestBytes, err := content.FetchAll(ctx, s.ociHelper.Target, s.base.manDesc)
	if err != nil {
		return fmt.Errorf("fetching base manifest: %w", err)
	}

	log.InfoContext(ctx, "Using existing manifest")
	err = json.Unmarshal(manifestBytes, &s.base.manifest)
	if err != nil {
		return fmt.Errorf("unmarshaling current base manifest: %w", err)
	}

	// check types
	if s.base.manifest.ArtifactType != oci.ArtifactTypeSyncManifest { // likely error if we check artifact type in descriptor and not manifest itself
		return fmt.Errorf("expected base manifest artifact type %s, got %s", oci.ArtifactTypeSyncManifest, s.base.manDesc.ArtifactType)
	}
	if s.base.manifest.Config.MediaType != oci.MediaTypeSyncConfig {
		return fmt.Errorf("expected base config media type %s, got %s", oci.MediaTypeSyncConfig, s.base.manifest.Config.MediaType)
	}

	log.InfoContext(ctx, "Fetching manifest config", "configDigest", s.base.manifest.Config.Digest.String())
	currentConfigBytes, err := content.FetchAll(ctx, s.ociHelper.Target, s.base.manifest.Config)
	if err != nil {
		return fmt.Errorf("fetching manifest config: %w", err)
	}

	err = json.Unmarshal(currentConfigBytes, &s.base.config)
	if err != nil {
		return fmt.Errorf("unmarshaling current manifest configuration: %w", err)
	}

	return nil
}

// FetchLFSManifestConfig copies all predecessor manifests with the LFS manifest media type, returning a pointer to an OCI CAS storage
// containing the result of the copy.
func (s *sync) FetchLFSManifestConfig(ctx context.Context, root ocispec.Descriptor, clean bool) (ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)

	// TODO: This feels like a hacky way to test that the root did not exist before
	if clean || (root.Digest == "") {
		log.InfoContext(ctx, "Starting with a fresh LFS manifest")
		// if adding config fields, initialize them here
		return ocispec.Descriptor{}, fmt.Errorf("clean option or subject not found: %w", errdef.ErrNotFound)
	}

	log.InfoContext(ctx, "Resolving commit manifest referrers", "root", root)
	referrers, err := registry.Referrers(ctx, s.ociHelper.Target, root, oci.ArtifactTypeLFSManifest)
	log.InfoContext(ctx, "Found commit manifest referrers", "referrers", referrers)

	// we expect one LFS manifest referrer
	switch {
	case len(referrers) < 1:
		return ocispec.Descriptor{}, errdef.ErrNotFound
	case len(referrers) > 1:
		return ocispec.Descriptor{}, fmt.Errorf("expected 1 LFS referrer, got %d", len(referrers)) // should never hit
	case err != nil:
		return ocispec.Descriptor{}, fmt.Errorf("resolving commit manifest predecessors: %w", err)
	}
	lfsManifestDesc := referrers[0]

	log.InfoContext(ctx, "Fetching LFS manifest", "desc", lfsManifestDesc)
	lfsManifestBytes, err := content.FetchAll(ctx, s.ociHelper.Target, lfsManifestDesc)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("fetching LFS manifest: %w", err)
	}

	err = json.Unmarshal(lfsManifestBytes, &s.lfs.manifest)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("decoding LFS manifest descriptor: %w", err)
	}

	// check types
	// if config returns, check artifact type here
	if s.lfs.manifest.ArtifactType != oci.ArtifactTypeLFSManifest { // likely error if we check artifact type in descriptor and not manifest itself
		return ocispec.Descriptor{}, fmt.Errorf("expected LFS manifest artifact type %s, got %s", oci.ArtifactTypeLFSManifest, lfsManifestDesc.ArtifactType)
	}

	return lfsManifestDesc, nil
}

// removeFromConfig removes all references to a commit from the sync config.
func (s *sync) removeFromConfig(obj cmd.Commit) {
	// check every reference, as more than one can refer to the same commit
	for i, ref := range s.base.config.Refs.Heads {
		if ref.Commit == obj {
			delete(s.base.config.Refs.Heads, i)
		}
	}

	for i, ref := range s.base.config.Refs.Tags {
		if ref.Commit == obj {
			delete(s.base.config.Refs.Tags, i)
		}
	}
}

// headRefsFromCommit returns all head and tag references, respectively, to the given commit
// included in the sync config.
func (s *sync) headRefsFromCommit(obj cmd.Commit) []string {
	heads := make([]string, 0, 1) // expected >= 1
	for ref, refInfo := range s.base.config.Refs.Heads {
		if refInfo.Commit == obj {
			heads = append(heads, ref)
		}
	}

	return heads
}

// cleanup closes and cleans up any temporary files created during the sync process.
func (s *sync) cleanup() error {
	if err := os.RemoveAll(s.syncOpts.IntermediateDir); err != nil {
		return fmt.Errorf("cleaning up intermediate repository: %w", err)
	}

	return nil
}
