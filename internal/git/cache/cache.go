// Package cache facilitates caching git and git-lfs objects.
package cache

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2/content/file"

	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
)

// Cache represents a cache of git and git-lfs objects.
type Cache struct {
	fstore     *file.Store
	fstorePath string

	cachePath string
}

// NewCache is a constructor for a Cache, expects cachePath to exist. The cmdOpts are only used
// for initialization.
// TODO: Should we have a default cache path? If so, this should be added to the flag not here.
func NewCache(ctx context.Context, cachePath, fstorePath string, fstore *file.Store, cmdOpts *cmd.Options) (*Cache, error) {
	// these variables are shared among all cache connections
	c := &Cache{
		fstore,
		fstorePath,
		cachePath,
	}

	ch, err := cmd.NewHelper(ctx, c.CachePath(), cmdOpts)
	if err != nil {
		return nil, fmt.Errorf("creating command helper: %w", err)
	}

	err = ch.InitializeRepo()
	if err != nil {
		return nil, fmt.Errorf("initializing shared object repository: %w", err)
	}

	return c, nil
}

// NewLink builds a new link to an existing cache, allowing for concurrent access.
func (c *Cache) NewLink(ctx context.Context, tag string, cmdOpts cmd.Options) (ObjectCacher, error) {
	// links use the cache shared variables and allow for custom command options on a per-repo basis,
	// e.g. we can use different LFS servers for different repos
	link := &Link{
		Cache: c,
	}

	var err error
	link.CmdHelper, err = cmd.NewHelper(ctx, link.CachePath(), &cmdOpts)
	if err != nil {
		return nil, fmt.Errorf("creating command helper for git cache link: %w", err)
	}

	return link, nil
}

// CachePath returns the path to the cache directory.
func (c *Cache) CachePath() string {
	return c.cachePath
}
