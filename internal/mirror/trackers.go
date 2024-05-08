package mirror

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"

	"gitlab.com/act3-ai/asce/data/tool/internal/descriptor"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror/encoding"
)

// ByteTracker is an object for tracking digests seen, total bytes seen, and deduplication size.
type ByteTracker struct {
	mutex        sync.Mutex
	mapper       map[digest.Digest]int // mapping from digest to count (number of times this digest has occurred)
	Total        int64
	Deduplicated int64
}

// AddDescriptor adds a digest to the tracker.
func (bt *ByteTracker) AddDescriptor(desc ocispec.Descriptor) {
	bt.mutex.Lock()
	defer bt.mutex.Unlock()

	if bt.mapper == nil {
		bt.mapper = map[digest.Digest]int{}
	}

	bt.Total += desc.Size

	if _, ok := bt.mapper[desc.Digest]; !ok {
		bt.Deduplicated += desc.Size
	}

	// increment the map's reference counter regardless of its existence
	bt.mapper[desc.Digest]++
}

// WorkTracker is an object for tracking the number of blobs and bytes actually pushed.
type WorkTracker struct {
	blobs       atomic.Int64
	transferred atomic.Int64
}

// Add adds the digest and blob to the work tracker count.
func (wt *WorkTracker) Add(desc ocispec.Descriptor) {
	wt.blobs.Add(1)
	wt.transferred.Add(desc.Size)
}

type manifestTracker struct {
	manifests    map[descriptor.Descriptor]struct{}
	allManifests []ocispec.Descriptor
}

func newManifestTracker() *manifestTracker {
	return &manifestTracker{
		manifests: make(map[descriptor.Descriptor]struct{}),
	}
}

func (mt *manifestTracker) Exists(desc ocispec.Descriptor) bool {
	_, ok := mt.manifests[descriptor.FromOCI(desc)]
	return ok
}

func (mt *manifestTracker) Add(desc ocispec.Descriptor) {
	mt.manifests[descriptor.FromOCI(desc)] = struct{}{}
	mt.allManifests = append(mt.allManifests, desc)
}

func (mt *manifestTracker) Manifests() []ocispec.Descriptor {
	return mt.allManifests
}

func extractBlobs(ctx context.Context, exists func(ocispec.Descriptor), fetcher content.Fetcher, desc ocispec.Descriptor) error {
	successors, err := encoding.Successors(ctx, fetcher, desc)
	if err != nil {
		return fmt.Errorf("getting successors for existing images: %w", err)
	}

	for _, d := range successors {
		if encoding.IsManifest(d.MediaType) {
			// recurse
			if err := extractBlobs(ctx, exists, fetcher, d); err != nil {
				return err
			}
			// we do not record manifests (for now)
			continue
		}
		// blob
		exists(d)
	}
	return nil
}
