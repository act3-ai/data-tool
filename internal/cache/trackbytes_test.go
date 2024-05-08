package cache

import (
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/stretchr/testify/assert"
)

func TestByteTrackerProcessEntry(t *testing.T) {
	bt := BytesTracker{}

	bt.Add(ocispec.Descriptor{
		Size:   10,
		Digest: "1",
	})
	assert.Equal(t, int64(10), bt.Total)
	assert.Equal(t, int64(10), bt.Deduplicated)

	// duplicated
	bt.Add(ocispec.Descriptor{
		Size:   10,
		Digest: "1",
	})
	assert.Equal(t, int64(20), bt.Total)
	assert.Equal(t, int64(10), bt.Deduplicated)

	bt.Add(ocispec.Descriptor{
		Size:   11,
		Digest: "2",
	})
	assert.Equal(t, int64(31), bt.Total)
	assert.Equal(t, int64(21), bt.Deduplicated)
}
