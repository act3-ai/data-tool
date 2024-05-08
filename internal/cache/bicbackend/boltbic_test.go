package bicbackend

import (
	"path/filepath"
	"testing"
)

var _ BlobInfoCache = &BoltBlobInfoCache{}

func newBoltTestCache(t *testing.T) BlobInfoCache {
	t.Helper()
	dir := t.TempDir()
	return BoltCache(filepath.Join(dir, "db"))
}

func TestBoltNew(t *testing.T) {
	t.Helper()
	testGenericCache(t, newBoltTestCache)
}
