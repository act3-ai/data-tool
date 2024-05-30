// Package cache facilitates caching git and git-lfs objects.
package cache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"

	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// ObjectCache proivdes methods for managing a cache of git objects.
type ObjectCache interface {
	CachePath() string
	UpdateFromGit(gitRemote string, opts cmd.Options, argRevList ...string) error
	UpdateLFSFromGit(gitRemote string, opts cmd.Options, argRevList ...string) error
	FetchFromBundle(bundlePath string, opts cmd.Options) error
	UpdateLFSFromOCI(ctx context.Context, target oras.GraphTarget, opts cmd.Options, layers []ocispec.Descriptor) error
}

// Cache implements ObjectCache.
type Cache struct {
	store content.Storage
	*cmd.Helper
}

// NewCache is a constructor for a Cache, expects cachePath to exist.
// TODO: Should we have a default cache path? If so, this should be added to the flag not here.
func NewCache(ctx context.Context, cachePath string, opts *cmd.Options) (ObjectCache, error) {
	ch, err := cmd.NewHelper(ctx, cachePath, opts)
	if err != nil {
		return nil, fmt.Errorf("creating git command helper for cache: %w", err)
	}

	c := &Cache{
		Helper: ch,
	}

	// safe, even if it already exists
	if err := c.initCache(); err != nil {
		return nil, fmt.Errorf("initializing cache: %w", err)
	}

	return c, nil
}

// CachePath returns the path to the cache directory.
func (c *Cache) CachePath() string {
	return c.Dir()
}

// UpdateFromGit updates the cache with objects from a remote git repository that
// are reachable from argRevList, or all objects if argRevList is empty.
func (c *Cache) UpdateFromGit(gitRemote string, opts cmd.Options, argRevList ...string) error {
	c.Options = opts
	if err := c.Git.Fetch(gitRemote, argRevList...); err != nil {
		return fmt.Errorf("fetching from remote: %w", err)
	}
	return nil
}

// UpdateLFSFromGit updates the cache with LFS objects from a remote git repository that
// are reachable from argRevList, or all objects if argRevList is empty.
func (c *Cache) UpdateLFSFromGit(gitRemote string, opts cmd.Options, commits ...string) error {
	c.Options = opts

	if err := c.ConfigureLFS(); err != nil {
		return fmt.Errorf("configuring LFS: %w", err)
	}

	args := []string{"--all"}
	args = append(args, commits...)
	err := c.LFS.Fetch(gitRemote, args...)
	if err != nil {
		return fmt.Errorf("fetching remote repository lfs files to cache: %w", err)
	}

	return nil
}

// FetchFromBundle updates the cache with all objects in a git bundle.
func (c *Cache) FetchFromBundle(bundlePath string, opts cmd.Options) error {
	c.Options = opts
	return c.Helper.FetchFromBundle(bundlePath)
}

// UpdateLFSFromOCI updates the cache with LFS objects from a remote git repository stored in
// OCI format that are reachable from argRevList, or all objects if argRevList is empty.
func (c *Cache) UpdateLFSFromOCI(ctx context.Context, target oras.GraphTarget, opts cmd.Options, layers []ocispec.Descriptor) error {
	uncached, err := c.resolveUncachedOCILFSFiles(ctx, layers)
	if err != nil {
		return fmt.Errorf("resolving uncached git-lfs files: %w", err)
	}

	for _, layerDesc := range uncached {
		oid := layerDesc.Digest.Hex() // TODO: This is not friendly for cryptographic agility
		oidDest := filepath.Join(c.CachePath(), c.ResolveLFSOIDPath(oid))
		if err := copyLFSFromOCI(ctx, target, oidDest, layerDesc); err != nil {
			return fmt.Errorf("copying lfs layer from OCI: %w", err)
		}
	}

	return nil
}

// initCache initializes the cache repository, expects cachePath to exist.
// It is always safe to run on an existing cache.
func (c *Cache) initCache() error {
	err := c.InitializeRepo()
	if err != nil {
		return fmt.Errorf("initializing shared object repository: %w", err)
	}

	c.store, err = file.New(c.CachePath())
	if err != nil {
		return fmt.Errorf("initializing oras file store: %w", err)
	}
	return nil
}

// resolveUncachedLFSFiles excludes cached git-lfs files from a slice of
// git-lfs OCI layers in-place.
func (c *Cache) resolveUncachedOCILFSFiles(ctx context.Context, layers []ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)

	excludeIdx := 0
	for i, layerDesc := range layers {
		oid := layerDesc.Digest.Hex() // TODO: This is not friendly for cryptographic agility
		relativeOIDPath := c.ResolveLFSOIDPath(oid)
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

// copyLFSFromOCI copies a git-lfs file stored as an OCI layer, written to objDest.
//
// TODO: This is inefficient and not a good interface. See issue https://git.act3-ace.com/ace/data/tool/-/issues/504.
func copyLFSFromOCI(ctx context.Context, target oras.GraphTarget, objDest string, layerDesc ocispec.Descriptor) error {
	log := logger.FromContext(ctx)

	// prepare destination
	oidDir := filepath.Dir(objDest)

	log.InfoContext(ctx, "initializing path to oid", "oidDir", oidDir)
	err := os.MkdirAll(oidDir, 0o777)
	if err != nil {
		return fmt.Errorf("creating path to oid file: %w", err)
	}

	log.InfoContext(ctx, "creating oid file", "objDest", objDest)
	oidFile, err := os.Create(objDest)
	if err != nil {
		return fmt.Errorf("creating oid file: %w", err)
	}

	// download
	r, err := target.Fetch(ctx, layerDesc)
	if err != nil {
		return fmt.Errorf("fetching LFS layer: %w", err)
	}

	n, err := io.Copy(oidFile, r)
	if err != nil {
		return fmt.Errorf("copying LFS layer to file: %w", err)
	}
	if n < layerDesc.Size {
		return fmt.Errorf("total bytes copied from LFS layer does not equal LFS layer size, layerSize: %d, copied: %d", layerDesc.Size, n)
	}

	if err := oidFile.Close(); err != nil {
		return fmt.Errorf("closing oid file: %w", err)
	}

	return nil
}
