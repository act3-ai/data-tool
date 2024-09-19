// Package cache implements local cached storage of part data.
//
// Currently only cache pruning is supported.
package cache

import (
	"context"

	"gitlab.com/act3-ai/asce/data/tool/internal/cache"
)

// CachePruner removes files from a cache until the total size is less than
// or equal to maxSize.
type CachePruner func(ctx context.Context, root string, maxSize int64) error

// Prune implements CachePruner.
func Prune(ctx context.Context, root string, maxSize int64) error {
	return cache.Prune(ctx, root, maxSize)
}
