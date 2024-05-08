package cache

import (
	"context"
	"testing"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"git.act3-ace.com/ace/data/tool/internal/cache/bicbackend"
	"git.act3-ace.com/ace/data/tool/internal/ref"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"git.act3-ace.com/ace/go-common/pkg/test"
)

const (
	digestA = digest.Digest("sha256:2222222222222222222222222222222222222222222222222222222222222222")
	digestB = digest.Digest("sha256:3333333333333333333333333333333333333333333333333333333333333333")
	digestC = digest.Digest("sha256:4444444444444444444444444444444444444444444444444444444444444444")
	digestD = digest.Digest("sha256:5555555555555555555555555555555555555555555555555555555555555555")
)

func TestBlobInfoCache(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T, cache bicbackend.BlobInfoCache)
	}{
		{"testFallback", testBICFallback},
		{"record single source", testRecordLayerSourceSingle},
		{"record multiple sources", testRecordLayerSourceMultiple},
		{"locate layer not found", testLocateLayerNotFound},
		{"locate layer single", testLocateLayerSingle},
		{"locate layer match reg", testLocateLayerRegMatch},
		{"locate layer with multiple sources", testLocateLayerMultiple},
		{"locate all sources for list of layers", testGetSources},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			bic := NewCache("")
			bic.Open()
			defer bic.Close()
			s.fn(t, bic)
		})
	}
}

func testBICFallback(t *testing.T, bic bicbackend.BlobInfoCache) {
	t.Helper()
	if !assert.IsType(t, &bicbackend.MemoryBlobInfoCache{}, bic) {
		t.Fatalf("BlobInfoCache fallback to memory BIC on empty string path failed")
	}
}

func makeSources() ([]ocispec.Descriptor, []ref.Ref) {
	descs := []ocispec.Descriptor{
		{Digest: digestA},
		{Digest: digestB},
		{Digest: digestC},
		{Digest: digestA},
		{Digest: digestC},
	}
	srcs := []ref.Ref{
		ref.RepoFromString("https://any.host/store/a"),
		ref.RepoFromString("https://another.host/with/b"),
		ref.RepoFromString("https://final.host/having/c"),
		ref.RepoFromString("https://duplicate.host/with/a"),
		ref.RepoFromString("https://duplicate.host/with/c"),
	}
	return descs, srcs
}

// Add a single layer source and make sure it shows up
func testRecordLayerSourceSingle(t *testing.T, bic bicbackend.BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	descs, srcs := makeSources()

	RecordLayerSource(ctx, bic, descs[0], srcs[0])

	candidates := bic.CandidateLocations(ctx, "oci", bicbackend.BICContentScope{Opaque: bicbackend.LayerContent}, descs[0].Digest, false)

	assert.Equal(t, []bicbackend.BICReplacementCandidate{
		{Digest: descs[0].Digest, TransformerName: bicbackend.UnknownTransformer, Location: bicbackend.BICLocationReference{Opaque: srcs[0].URL()}},
	}, candidates)
}

// Add multiple layer sources, and query one.  More than one result is returned, with the location sources presented
// in reverse order (newest first)
func testRecordLayerSourceMultiple(t *testing.T, bic bicbackend.BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	descs, srcs := makeSources()

	for i := range descs {
		RecordLayerSource(ctx, bic, descs[i], srcs[i])
	}

	candidates := bic.CandidateLocations(ctx, "oci", bicbackend.BICContentScope{Opaque: bicbackend.LayerContent}, descs[0].Digest, false)

	// important testing note: the order of the found locations should be in reverse order from how they were added, so
	// the srcs index used for the expected value is decreasing
	assert.Equal(t, []bicbackend.BICReplacementCandidate{
		{Digest: descs[0].Digest, TransformerName: bicbackend.UnknownTransformer, Location: bicbackend.BICLocationReference{Opaque: srcs[3].URL()}},
		{Digest: descs[0].Digest, TransformerName: bicbackend.UnknownTransformer, Location: bicbackend.BICLocationReference{Opaque: srcs[0].URL()}},
	}, candidates)
}

// Look for a layer that does not exist
func testLocateLayerNotFound(t *testing.T, bic bicbackend.BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	descs, srcs := makeSources()

	for i := range descs {
		RecordLayerSource(ctx, bic, descs[i], srcs[i])
	}

	refs := LocateLayer(ctx, bic, ocispec.Descriptor{Digest: digestD}, ref.RepoFromString("https://any.host/store/a"), false)

	assert.Equal(t, 0, len(refs))

}

// Look for a layer with only one source
func testLocateLayerSingle(t *testing.T, bic bicbackend.BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	descs, srcs := makeSources()

	for i := range descs {
		RecordLayerSource(ctx, bic, descs[i], srcs[i])
	}

	refs := LocateLayerDigest(ctx, bic, digestB, ref.RepoFromString("https://any.host/store/a"), false)

	assert.Equal(t, len(refs), 1)
	assert.Equal(t, ref.Ref{Scheme: "https", Reg: "another.host", Repo: "with", Name: "b", Tag: "latest"}, refs[0])

}

// Look for a layer that has multiple sources, but pick one matching the registry
func testLocateLayerRegMatch(t *testing.T, bic bicbackend.BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	descs, srcs := makeSources()

	for i := range descs {
		RecordLayerSource(ctx, bic, descs[i], srcs[i])
	}

	refs := LocateLayer(ctx, bic, ocispec.Descriptor{Digest: digestA}, ref.RepoFromString("https://duplicate.host/store/a"), true)

	assert.Equal(t, len(refs), 1)
	assert.Equal(t, ref.Ref{Scheme: "https", Reg: "duplicate.host", Repo: "with", Name: "a", Tag: "latest"}, refs[0])

}

// Look for a layer that has multiple sources, and get all of them. Note, the most recent addition should appear first in
// list
func testLocateLayerMultiple(t *testing.T, bic bicbackend.BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	descs, srcs := makeSources()

	for i := range descs {
		RecordLayerSource(ctx, bic, descs[i], srcs[i])
	}

	refs := LocateLayer(ctx, bic, ocispec.Descriptor{Digest: digestA}, ref.RepoFromString("https://duplicate.host/store/a"), false)

	assert.Equal(t, len(refs), 2)
	// important testing note: the order of the found locations should be in reverse order from how they were added, so
	// the srcs index used for the expected value is decreasing
	assert.Equal(t,
		[]ref.Ref{
			{Scheme: "https", Reg: "duplicate.host", Repo: "with", Name: "a", Tag: "latest"},
			{Scheme: "https", Reg: "any.host", Repo: "store", Name: "a", Tag: "latest"},
		}, refs)

}

// Match all sources in a list of given digests
func testGetSources(t *testing.T, bic bicbackend.BlobInfoCache) {
	t.Helper()
	ctx := logger.NewContext(context.Background(), test.Logger(t, 0))
	descs, srcs := makeSources()

	for i := range descs {
		RecordLayerSource(ctx, bic, descs[i], srcs[i])
	}

	lbi := LayerBlobInfo{layerIDs: []digest.Digest{digestA, digestB}}
	dgstSources := lbi.GetSources(ctx, bic)
	assert.Equal(t, 2, len(dgstSources))
	assert.Equal(t, 1, len(dgstSources[digestB]))
	assert.Equal(t, 2, len(dgstSources[digestA]))

	assert.Equal(t,
		[]string{
			"another.host/with/b:latest",
		}, dgstSources[digestB])

	assert.Equal(t,
		[]string{"duplicate.host/with/a:latest", "any.host/store/a:latest"},
		dgstSources[digestA])

}

func TestTransportFromImageName(t *testing.T) {
	cases := []struct {
		name     string
		refStr   string
		expected string
	}{
		{
			name:     "docker transport",
			refStr:   "docker:image",
			expected: "oci",
		},
		{
			name:     "oci alone",
			refStr:   "oci",
			expected: "oci",
		},
		{
			name:     "oci transport",
			refStr:   "oci:https:/any.host/here",
			expected: "oci",
		},
		{
			name:     "directory transport",
			refStr:   "dir:path/to/file",
			expected: "dir",
		},
		{
			name:     "docker archive transport",
			refStr:   "docker-archive:path/to/data",
			expected: "docker-archive",
		},
		{
			name:     "oci archive transport",
			refStr:   "oci-archive:/path/to/file",
			expected: "oci-archive",
		},
		{
			name:     "tarball transport",
			refStr:   "tarball:~/bigfiles/bob.tar",
			expected: "tarball",
		},
		{
			name:     "not defined transport",
			refStr:   "something:is.maybe.bottle",
			expected: "oci",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := TransportFromImageName(tc.refStr)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}
