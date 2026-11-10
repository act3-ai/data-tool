package print

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/oci"
)

// TestCheckBlobs builds an OCI layout on disk, then removes a referenced blob.
// It tests the printer's graph traversal and existence checks without the CLI or action.
func TestCheckBlobs(t *testing.T) {
	for _, tc := range []struct {
		name        string
		remove      string
		wantErr     bool
		wantMissing bool
	}{
		// missing nodes shouldn't break the traversal and be makred as missing.
		{name: "all blobs present"},
		{name: "missing layer", remove: "layer", wantMissing: true},
		{name: "missing config", remove: "config", wantMissing: true},
		{name: "missing referenced manifest", remove: "manifest", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runCheckBlobsCase(t, tc.remove, tc.wantErr, tc.wantMissing)
		})
	}
}

func runCheckBlobsCase(t *testing.T, remove string, wantErr, wantMissing bool) {
	t.Helper()
	ctx := context.Background()
	dir, root, blobs := buildCheckBlobsLayout(t)

	// initialize the ORAS file store.
	store, err := oci.NewFromFS(ctx, os.DirFS(dir))
	if err != nil {
		t.Fatal(err)
	}

	// delete a blob to simulate a missing blob.
	if remove != "" {
		d := blobs[remove].Digest
		if err := os.Remove(filepath.Join(dir, "blobs", d.Algorithm().String(), d.Encoded())); err != nil {
			t.Fatal(err)
		}
	}
	// test the function.
	var out bytes.Buffer
	err = Successors(ctx, &out, store, root, Options{CheckBlobs: true})
	// check the results.
	if wantErr {
		if err == nil || !strings.Contains(err.Error(), "fetching manifest") {
			t.Fatalf("expected missing-manifest fetch error, got %v; output: %s", err, out.String())
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	assertCheckBlobsOutput(t, out.String(), blobs[remove], wantMissing)
}

func buildCheckBlobsLayout(t *testing.T) (string, ocispec.Descriptor, map[string]ocispec.Descriptor) {
	t.Helper()
	dir := t.TempDir()
	blobs := make(map[string]ocispec.Descriptor)

	// writes mock OCI blob data.
	writeBlob := func(name, mediaType string, data []byte) ocispec.Descriptor {
		t.Helper()
		d := digest.FromBytes(data)
		path := filepath.Join(dir, "blobs", d.Algorithm().String(), d.Encoded())
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		desc := ocispec.Descriptor{MediaType: mediaType, Digest: d, Size: int64(len(data))}
		blobs[name] = desc
		return desc
	}
	// serializes the OCI structs into JSON bytes.
	marshal := func(v any) []byte {
		t.Helper()
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}

	// creates a valid OCI layout
	config := writeBlob("config", ocispec.MediaTypeImageConfig, []byte(`{}`))
	layer := writeBlob("layer", ocispec.MediaTypeImageLayer, []byte("test layer contents"))
	// writes the manifest that points to the config and layer.
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageManifest,
		Config:    config,
		Layers:    []ocispec.Descriptor{layer},
	}
	// writes the index that poitns to the manifest.
	manifestDesc := writeBlob("manifest", ocispec.MediaTypeImageManifest, marshal(manifest))
	index := ocispec.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: ocispec.MediaTypeImageIndex,
		Manifests: []ocispec.Descriptor{manifestDesc},
	}
	indexBytes := marshal(index)
	// writes root structures for the OCI layout.
	root := writeBlob("index", ocispec.MediaTypeImageIndex, indexBytes)
	if err := os.WriteFile(filepath.Join(dir, "index.json"), indexBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`), 0600); err != nil {
		t.Fatal(err)
	}
	return dir, root, blobs
}

func assertCheckBlobsOutput(t *testing.T, output string, removed ocispec.Descriptor, wantMissing bool) {
	t.Helper()
	if got := strings.Contains(output, "MISSING"); got != wantMissing {
		t.Fatalf("MISSING in output = %t, want %t:\n%s", got, wantMissing, output)
	}
	// checks that the missing blob is associated with the correct hash of the file, ensuring its the correct missing blob.
	if wantMissing {
		var found bool
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, removed.Digest.String()) && strings.Contains(line, "MISSING") {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing blob %s not marked on its descriptor line:\n%s", removed.Digest, output)
		}
	}
}
