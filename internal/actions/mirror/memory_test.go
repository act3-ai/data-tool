package mirror

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"

	"git.act3-ace.com/ace/data/tool/internal/orasutil"
)

func TestCAS_MissingBlob(t *testing.T) {
	rne := require.New(t).NoError
	ctx := context.Background()

	cas := memory.New()

	blob1, err := oras.PushBytes(ctx, cas, "", []byte("blob1"))
	rne(err)

	blob2, err := oras.PushBytes(ctx, cas, "", []byte("blob2"))
	rne(err)

	blob3 := content.NewDescriptorFromBytes("", []byte("blob3"))
	// but do not push to the cas

	options := oras.PackManifestOptions{
		Layers: []ocispec.Descriptor{blob1, blob2, blob3},
	}
	image1, err := oras.PackManifest(ctx, cas, oras.PackManifestVersion1_1, "application/sample+json", options)
	rne(err) // we actually want an error here
	t.Log(image1)
}

func TestCAS_MissingBlobFixed(t *testing.T) {
	rne := require.New(t).NoError
	ctx := context.Background()

	cas := &orasutil.CheckedStorage{Target: memory.New()}

	blob1, err := oras.PushBytes(ctx, cas, "", []byte("blob1"))
	rne(err)

	blob2, err := oras.PushBytes(ctx, cas, "", []byte("blob2"))
	rne(err)

	blob3 := content.NewDescriptorFromBytes("", []byte("blob3"))
	// but do not push to the cas

	options := oras.PackManifestOptions{
		Layers: []ocispec.Descriptor{blob1, blob2, blob3},
	}
	image1, err := oras.PackManifest(ctx, cas, oras.PackManifestVersion1_1, "application/sample+json", options)
	assert.Error(t, err)
	t.Log(image1)
}

func TestCAS_Predecessors(t *testing.T) {
	rne := require.New(t).NoError
	ctx := context.Background()

	cas := memory.New()

	blob1, err := oras.PushBytes(ctx, cas, "", []byte("blob1"))
	rne(err)
	t.Log("Blob1", blob1.Digest)

	blob2, err := oras.PushBytes(ctx, cas, "", []byte("blob2"))
	rne(err)
	t.Log("Blob2", blob2.Digest)

	options1 := oras.PackManifestOptions{
		Layers: []ocispec.Descriptor{blob1},
	}
	image1, err := oras.PackManifest(ctx, cas, oras.PackManifestVersion1_1, "application/sample+json", options1)
	rne(err) // we actually want an error here
	t.Log("Image1", image1.Digest)

	options2 := oras.PackManifestOptions{
		Subject: &image1,
		Layers:  []ocispec.Descriptor{blob2},
	}
	image2, err := oras.PackManifest(ctx, cas, oras.PackManifestVersion1_1, "application/sample+json", options2)
	rne(err) // we actually want an error here
	t.Log("Image2", image2.Digest)

	index := ocispec.Index{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		// Subject:   &image1,
		MediaType: ocispec.MediaTypeImageIndex,
		Manifests: []ocispec.Descriptor{image2},
	}
	indexData, err := json.Marshal(index)
	rne(err)
	index1, err := oras.PushBytes(ctx, cas, ocispec.MediaTypeImageIndex, indexData)
	rne(err)
	t.Log("Index1", index1.Digest)

	predecessorsImage1, err := cas.Predecessors(ctx, image1)
	rne(err)
	for _, p := range predecessorsImage1 {
		t.Log("Image1 predecessor", p.Digest)
	}

	predecessorsImage2, err := cas.Predecessors(ctx, image2)
	rne(err)
	for _, p := range predecessorsImage2 {
		t.Log("Image2 predecessor", p.Digest)
	}

	predecessorsBlob1, err := cas.Predecessors(ctx, blob1)
	rne(err)
	for _, p := range predecessorsBlob1 {
		t.Log("Blob1 predecessor", p.Digest)
	}

	// Conclusions
	// Predecessors() returns only the immediate ancestors
	// Predecessors() on Index returns both the subject and index manifest links
	// Predecessors() of blobs are images (this is not possible with the OCI distribution spec referrers API).
	// Referrers API calls out tat only manifests are listed in the response.
}
