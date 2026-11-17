package oci

import (
	"archive/tar"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Tests local archive reference parsing (bare files and registry references).
func Test_LocalArchiveReference(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "archive.tar")
	if err := os.WriteFile(filePath, []byte("dummy data"), 0600); err != nil {
		t.Fatalf("writing dummy data: %v", err)
	}
	tests := []struct {
		raw      string
		wantPath string
		wantRef  string
		wantFnd  bool
		wantErr  bool
	}{
		{raw: filePath, wantPath: filePath, wantFnd: true},
		{raw: "docker.io/library/alpine:latest"},
		{raw: filePath + ":latest", wantErr: true},
		{raw: filePath + "@sha256:12345", wantErr: true},
	}

	for _, tt := range tests {
		p, r, f, err := localArchiveReference(tt.raw)
		if (err != nil) != tt.wantErr || p != tt.wantPath || r != tt.wantRef || f != tt.wantFnd {
			t.Errorf("localArchiveReference(%q) = %v, %v, %v, err=%v; want error=%v", tt.raw, p, r, f, err, tt.wantErr)
		}
	}
}

// Tests parsing the root index.json of an OCI archive
func Test_ArchiveRoots(t *testing.T) {
	dir := t.TempDir()
	index := `{"manifests": [{"digest": "sha256:12345", "size": 100}]}`
	if err := os.WriteFile(filepath.Join(dir, "index.json"), []byte(index), 0600); err != nil {
		t.Fatalf("writing index.json: %v", err)
	}
	roots, err := archiveRoots(dir)
	if err != nil || len(roots) != 1 || roots[0].Digest != "sha256:12345" {
		t.Errorf("archiveRoots() failed: roots=%v, err=%v", roots, err)
	}
}

// Tests tarball extraction by writing a minimal valid OCI layout tarball
func Test_ExtractOCILayoutTar_Success(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "test.tar")
	f, _ := os.Create(archivePath)
	tw := tar.NewWriter(f)

	// Add required OCI files to the tarball
	files := []string{"oci-layout", "index.json", "blobs/"}
	for _, name := range files {
		hdr := &tar.Header{Name: name, Mode: 0600}
		if name == "blobs/" {
			hdr.Typeflag = tar.TypeDir
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("writing tar header: %v", err)
		}
	}
	tw.Close()
	f.Close()

	extractedDir := t.TempDir()
	if err := extractOCILayoutTar(context.Background(), archivePath, extractedDir); err != nil {
		t.Fatalf("extractOCILayoutTar failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(extractedDir, "oci-layout")); err != nil {
		t.Errorf("checking extracted oci-layout: %v", err)
	}
}

// createTreeLayout builds an OCI layout with one child index for Tree.Run tests.
func createTreeLayout(t *testing.T) (string, digest.Digest) {
	t.Helper()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`), 0600); err != nil {
		t.Fatalf("writing oci-layout: %v", err)
	}
	childIndex := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[]}`)
	childDigest := digest.FromBytes(childIndex)
	// Use struct marshaling to ensure the JSON strictly matches OCI specs and ORAS expectations
	index := ocispec.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		Manifests: []ocispec.Descriptor{{
			MediaType: ocispec.MediaTypeImageIndex,
			Digest:    childDigest,
			Size:      int64(len(childIndex)),
		}},
	}
	indexBytes, err := json.Marshal(index)
	if err != nil {
		t.Fatalf("marshaling index.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.json"), indexBytes, 0600); err != nil {
		t.Fatalf("writing index.json: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "blobs", "sha256"), 0700); err != nil {
		t.Fatalf("creating blobs directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "blobs", childDigest.Algorithm().String(), childDigest.Encoded()), childIndex, 0600); err != nil {
		t.Fatalf("writing child index blob: %v", err)
	}
	return dir, childDigest
}
