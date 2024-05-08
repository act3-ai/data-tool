// Package cache implements local cached storage of part data.
//
// Currently only cache pruning is supported.
package cache

import "gitlab.com/act3-ai/asce/data/tool/internal/cache"

// BottleCachePruner removes bottle items until the total size of the cache is less than or
// equal to maxSize.
type BottleCachePruner interface {
	Prune(maxSize int64) error
}

// NewBottleCachePruner accesses a BottleFileCache strictly for pruning purposes.
func NewBottleCachePruner(cacheDir string) BottleCachePruner {
	return cache.NewBottleFileCache(cacheDir)
}
