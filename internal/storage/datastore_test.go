package storage

import (
	"io"
	"sync"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"

	"gitlab.com/act3-ai/asce/data/tool/internal/cache"
)

type mockMoteWriter struct{}

func (m mockMoteWriter) Write(p []byte) (n int, err error) { return 0, nil }
func (m mockMoteWriter) Close() error                      { return nil }

type mockMoteReader struct{}

func (m mockMoteReader) Read(p []byte) (n int, err error) { return 0, nil }
func (m mockMoteReader) Close() error                     { return nil }

type mockMoteReaderAt struct{}

func (m mockMoteReaderAt) Size() int64 {
	return 0
}

func (m mockMoteReaderAt) ReadAt(p []byte, off int64) (n int, err error) { return 0, nil }
func (m mockMoteReaderAt) Close() error                                  { return nil }

type mockMoteInfo struct{}

func (m mockMoteInfo) GetDigest() digest.Digest { return "" }
func (m mockMoteInfo) Size() int64              { return 0 }
func (m mockMoteInfo) GetMediaType() string     { return "" }

type mockMoteCache struct{}

func (m mockMoteCache) Initialize() error                  { return nil }
func (m mockMoteCache) IsDirty() bool                      { return false }
func (m mockMoteCache) Refresh() (bool, error)             { return false, nil }
func (m mockMoteCache) MoteExists(dgst digest.Digest) bool { return true }
func (m mockMoteCache) MoteWriter(dgst digest.Digest) (io.WriteCloser, error) {
	return mockMoteWriter{}, nil
}

func (m mockMoteCache) MoteReader(dgst digest.Digest) (io.ReadCloser, error) {
	return mockMoteReader{}, nil
}

func (m mockMoteCache) MoteReaderAt(dgst digest.Digest) (cache.ReaderAt, error) {
	return mockMoteReaderAt{}, nil
}
func (m mockMoteCache) Find(dgst digest.Digest) (cache.MoteInfo, bool)               { return mockMoteInfo{}, false }
func (m mockMoteCache) MoteRef(dgst digest.Digest) string                            { return "" }
func (m mockMoteCache) CommitMote(ref string, mote cache.Mote, commitMode int) error { return nil }
func (m mockMoteCache) CreateMote(dgst digest.Digest, mediaType string, size int64) (cache.Mote, error) {
	return testMote, nil
}
func (m mockMoteCache) RemoveMote(dgst digest.Digest) error { return nil }
func (m mockMoteCache) Prune(maxSize int64) error           { return nil }

type mockCacheProvider struct{}

func (m mockCacheProvider) GetCache() cache.MoteCache {
	return mockMoteCache{}
}
func (m mockCacheProvider) GetPath() string { return "" }

var (
	testDataStore = DataStore{
		DisableOverwrite:          false,
		AllowPathTraversalOnWrite: false,
		root:                      "/tmp/",
		descriptor:                &sync.Map{},
		configData:                []byte{},
		dloc:                      &mockCacheProvider{},
		DisableDigestCalc:         false,
	}

	testMote = cache.Mote{
		Digest:    "",
		MediaType: "",
		DataSize:  0,
	}
)

func Test_NewDataStore(t *testing.T) {
	tests := []struct {
		name          string
		cacheProvider CacheProvider
	}{
		{"test 1", &mockCacheProvider{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			datastore := NewDataStore(test.cacheProvider)
			assert.NotNil(t, datastore)
		})
	}
}

func Test_DataStoreClose(t *testing.T) {
	assert.NoError(t, testDataStore.Close())
}

/*
// TODO fix this.  We actually need files to copy now that HandlePartMedia actually copies stuff
func Test_HandlePartMedia(t *testing.T) {
	partName := "foo"
	tests := []struct {
		name        string
		descriptor  ocispec.Descriptor
		expectedErr string
	}{
		{"fail test", testDescriptorError, "unknown layer media type"},
		{"uncompressed test", testDescriptor, ""},
		{"compressed tar+zstd test", testDescriptorCompressed, ""},
		{"compressed zstd test", testDescriptorCompressedZstd, ""},
		{"compressed tar+gzip test", testDescriptorCompressedTarGzip, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := testDataStore.HandlePartMedia(context.Background(), test.descriptor, partName)
			if len(test.expectedErr) > 0 {
				assert.ErrorContains(t, err, test.expectedErr)
			}
		})
	}
}
*/
