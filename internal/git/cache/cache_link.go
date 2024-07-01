package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/oci"
)

// ObjectCacher proivdes methods for adding git or git-lfs objects to a cache.
type ObjectCacher interface {
	CachePath() string
	UpdateFromGit(ctx context.Context, gitRemote string, argRevList ...string) error
	UpdateFromOCI(ctx context.Context, src content.ReadOnlyGraphStorage, manDesc ocispec.Descriptor) error
	UpdateLFSFromGit(ctx context.Context, gitRemote string, argRevList ...string) error
	UpdateLFSFromOCI(ctx context.Context, src content.ReadOnlyGraphStorage, lfsManDesc ocispec.Descriptor) ([]ocispec.Descriptor, error)
}

// Link provides concurrency safe access to a git object cache.
// Link implements ObjectCacher.
type Link struct {
	*Cache

	CmdHelper *cmd.Helper
}

// UpdateFromGit updates the cache with objects from a remote git repository that
// are reachable from argRevList, or all objects if argRevList is empty.
func (c *Link) UpdateFromGit(ctx context.Context, gitRemote string, argRevList ...string) error {
	args := make([]string, 0, len(argRevList)+1)
	args = append(args, gitRemote)
	args = append(args, argRevList...)
	if err := c.CmdHelper.Git.Fetch(ctx, args...); err != nil {
		return fmt.Errorf("fetching from remote: %w", err)
	}
	return nil
}

// UpdateFromOCI updates the cache with the specified contents from an OCI repository.
func (c *Link) UpdateFromOCI(ctx context.Context, src content.ReadOnlyGraphStorage, manDesc ocispec.Descriptor) error {
	log := logger.V(logger.FromContext(ctx), 1)

	// determine min set of bundles and copy
	manBytes, err := content.FetchAll(ctx, src, manDesc)
	if err != nil {
		return fmt.Errorf("fetching manifest: %w", err)
	}

	var manifest ocispec.Manifest
	err = json.Unmarshal(manBytes, &manifest)
	if err != nil {
		return fmt.Errorf("decoding manifest: %w", err)
	}

	configBytes, err := content.FetchAll(ctx, src, manifest.Config)
	if err != nil {
		return fmt.Errorf("fetching config: %w", err)
	}

	var config oci.Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return fmt.Errorf("decoding config: %w", err)
	}

	bundleDescs, err := c.resolveUncachedBundles(ctx, manifest.Layers, config)
	if err != nil {
		return fmt.Errorf("resovling uncached git bundles: %w", err)
	}

	copyOpts := oras.CopyGraphOptions{
		// PreCopy func simply logs
		PreCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
			if desc.MediaType == oci.MediaTypeBundleLayer {
				log.InfoContext(ctx, "Retrieving bundle layer", "bundle", desc.Annotations[ocispec.AnnotationTitle])
			}
			return nil
		},
		FindSuccessors: oci.FindSuccessorsBundles(manDesc, bundleDescs),
	}

	err = oras.CopyGraph(ctx, src, c.fstore, manDesc, copyOpts)
	if err != nil {
		return fmt.Errorf("copying bundles: %w", err)
	}

	// fetch from copied bundles
	log.InfoContext(ctx, "adding bundles as remotes")
	shortnames := make([]string, 0, len(bundleDescs))
	for _, desc := range bundleDescs {
		// resolve bundle path
		bundleName := desc.Annotations[ocispec.AnnotationTitle]
		bundlePath := filepath.Join(c.fstorePath, bundleName)

		// add bundle as a remote
		// TODO: we don't have unique bundle names, so this is not concurrency safe
		shortname := strings.TrimSuffix(bundleName, ".bundle")
		err := c.CmdHelper.RemoteAdd(ctx, shortname, bundlePath)
		if err != nil {
			return fmt.Errorf("adding bundle '%s' as remote: %w", bundlePath, err)
		}
		shortnames = append(shortnames, shortname)
	}

	args := append(shortnames, "--tags", "--multiple") //nolint
	if c.CmdHelper.Force {
		args = append(args, "--force")
		log.InfoContext(ctx, "force fetching from bundles")
	} else {
		log.InfoContext(ctx, "fetching from bundles")
	}
	if err := c.CmdHelper.Git.Fetch(ctx, args...); err != nil {
		return fmt.Errorf("fetching from bundles: %w", err)
	}

	// remove remotes
	// TODO: is this necessary?
	log.InfoContext(ctx, "removing bundles from remotes")
	for _, shortname := range shortnames {
		err := c.CmdHelper.RemoteRemove(ctx, shortname)
		if err != nil {
			return fmt.Errorf("removing remote bundle: %w", err)
		}
	}

	return nil
}

// UpdateLFSFromGit updates the cache with LFS objects from a remote git repository that
// are reachable from argRevList, or all objects if argRevList is empty.
func (c *Link) UpdateLFSFromGit(ctx context.Context, gitRemote string, commits ...string) error {
	if err := c.CmdHelper.ConfigureLFS(ctx); err != nil {
		return fmt.Errorf("configuring LFS: %w", err)
	}

	args := []string{"--all"}
	args = append(args, commits...)
	err := c.CmdHelper.LFS.Fetch(ctx, gitRemote, args...)
	if err != nil {
		return fmt.Errorf("fetching remote repository lfs files to cache: %w", err)
	}

	return nil
}

// UpdateLFSFromOCI updates the cache with LFS objects from a remote git repository stored in
// OCI format that are reachable from argRevList, or all objects if argRevList is empty.
// Returns a list of fetched LFS OIDs.
func (c *Link) UpdateLFSFromOCI(ctx context.Context, src content.ReadOnlyGraphStorage, lfsManDesc ocispec.Descriptor) ([]ocispec.Descriptor, error) {

	manBytes, err := content.FetchAll(ctx, src, lfsManDesc)
	if err != nil {
		return nil, fmt.Errorf("fetching manifest: %w", err)
	}

	var manifest ocispec.Manifest
	err = json.Unmarshal(manBytes, &manifest)
	if err != nil {
		return nil, fmt.Errorf("decoding manifest: %w", err)
	}

	uncached, err := c.resolveUncachedLFSFiles(ctx, manifest.Layers)
	if err != nil {
		return nil, fmt.Errorf("resolving uncached git-lfs files: %w", err)
	}

	copyOpts := oras.CopyGraphOptions{
		// TODO: Using oras' default concurrency (3), until plumbing arrives here...
		PostCopy:       oci.PostCopyLFS(c.fstorePath, c.cachePath),
		FindSuccessors: oci.FindSuccessorsLFS(uncached),
	}

	if err := oras.CopyGraph(ctx, src, c.fstore, lfsManDesc, copyOpts); err != nil {
		return nil, fmt.Errorf("copying LFS files as OCI layers: %w", err)
	}

	return uncached, nil
}

// resolveUncachedLFSFiles excludes cached git-lfs files from a slice of
// git-lfs OCI layers, in-place.
func (c *Link) resolveUncachedLFSFiles(ctx context.Context, layers []ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)

	excludeIdx := 0
	for i, layerDesc := range layers {
		oid := layerDesc.Digest.Hex() // TODO: This is not friendly for cryptographic agility
		relativeOIDPath := cmd.ResolveLFSOIDPath(oid)
		cacheOIDPath := filepath.Join(c.CachePath(), relativeOIDPath)
		_, err := os.Stat(cacheOIDPath) // Note: we do not check if the oid is a directory, which should never happen anyways
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("checking status of lfs OID '%s': %w", oid, err)
		}

		switch {
		case errors.Is(err, os.ErrNotExist):
			log.DebugContext(ctx, "Cache miss", "lfsOid", oid)
			continue
		case err != nil:
			return nil, fmt.Errorf("checking status of lfs OID '%s': %w", oid, err)
		default:
			log.DebugContext(ctx, "Cache hit", "lfsOid", oid)
			layers[0], layers[i] = layers[i], layers[0]
			excludeIdx++
		}
	}

	return layers[excludeIdx:], nil
}

// resolveUncachedBundles excludes bundles that the cache already contains all objects of, in-place.
func (c *Link) resolveUncachedBundles(ctx context.Context, layers []ocispec.Descriptor, config oci.Config) ([]ocispec.Descriptor, error) {
	// When resolving with the intermediate dir, cloned from the remote, we use references to derive the oids then compare the oid to
	// the manifest config. Here, we do not have access to references so we check the oid directly.
	// TODO: This approach may be, marginally, more efficient and applicable to both cache and non-cache cases, we should consider
	// using this elsewhere as well; although it may not be quite as robust.

	// layer digest to layer index resolver
	layerNumResolver := make(map[digest.Digest]int, len(layers))
	for i, layerDesc := range layers {
		layerNumResolver[layerDesc.Digest] = i
	}

	layerCutoff := len(layers) // start cutoff after the total num of layers
	for _, refInfo := range config.Refs.Tags {
		// only check for a possible update if the ref is before the cutoff
		if layerNumResolver[refInfo.Layer] < layerCutoff {
			// fullTagRef := filepath.Join(cmd.TagRefPrefix, tag)
			// refCommits, err := c.CmdHelper.ShowRefs(fullTagRef) // returned slice should be of length 1
			err := c.CmdHelper.Git.CatFile(ctx, "-e", string(refInfo.Commit))
			if err != nil {
				// oid DNE
				layerCutoff = layerNumResolver[refInfo.Layer]
			}
		}
	}

	for _, refInfo := range config.Refs.Heads {
		// only check for a possible update if the ref is before the cutoff
		if layerNumResolver[refInfo.Layer] < layerCutoff {
			// fullHeadRef := filepath.Join(cmd.HeadRefPrefix, head)
			// refCommits, err := c.CmdHelper.ShowRefs(fullHeadRef) // returned slice should be of length 1
			err := c.CmdHelper.Git.CatFile(ctx, "-e", string(refInfo.Commit))
			if err != nil {
				// oid DNE
				layerCutoff = layerNumResolver[refInfo.Layer]
			}
			// if err != nil {
			// 	// try to recover by assuming this ref DNE
			// 	layerCutoff = layerNumResolver[refInfo.Layer]
			// 	continue
			// }

			// split := strings.Fields(refCommits[0])
			// remoteCommit := oci.Commit(split[0]) // not technically the remote, but our intermediate repo should be identical at this point
			// if refInfo.Commit != remoteCommit {
			// 	layerCutoff = layerNumResolver[refInfo.Layer]
			// }
		}
	}

	return layers[layerCutoff:], nil
}
