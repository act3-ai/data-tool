package cache

import (
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// BytesTracker is an object for tracking digests seen, total bytes seen, and deduplication size.
type BytesTracker struct {
	mapper map[digest.Digest]int // mapping from digest to count (number of times this digest has occurred)

	// Total number of bytes of bytes seen so far
	Total int64

	// Total number of bytes seen so far with duplicates removed
	Deduplicated int64
}

// Add adds a digest to the BytesTracker map and computes the total size and deduplicated size.
func (bt *BytesTracker) Add(desc ocispec.Descriptor) {
	if bt.mapper == nil {
		bt.mapper = map[digest.Digest]int{}
	}

	size := desc.Size
	digest := desc.Digest
	count, exists := bt.mapper[digest]

	bt.Total += size
	if !exists {
		bt.Deduplicated += size
	}

	// increment the map's reference counter regardless of its existence
	count++
	bt.mapper[digest] = count
}
