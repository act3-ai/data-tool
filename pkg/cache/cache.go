// Package cache implements local cached storage of part data.
//
// Currently only cache pruning is supported.
package cache

import (
	"context"

	"github.com/act3-ai/data-tool/internal/cache"
)

// Pruner removes files from a cache until the total size is less than
// or equal to maxSize.
type Pruner func(ctx context.Context, root string, maxSize int64) error

// Prune implements CachePruner.
func Prune(ctx context.Context, root string, maxSize int64) error {
	return cache.Prune(ctx, root, maxSize)
}
