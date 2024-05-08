package bicbackend

import "testing"

var _ BlobInfoCache = &MemoryBlobInfoCache{}

func newMemoryTestCache(t *testing.T) BlobInfoCache {
	t.Helper()
	return MemoryCache()
}

func TestMemoryNew(t *testing.T) {
	t.Helper()
	testGenericCache(t, newMemoryTestCache)
}
