package oci

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Tests local archive reference parsing (tags, digests, and bare files)
func TestLocalArchiveReference(t *testing.T) {
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
	}{
		{filePath, filePath, "", true},
		{filePath + ":latest", filePath, "latest", true},
		{filePath + "@sha256:12345", filePath, "sha256:12345", true},
		{"nonexistent.tar:latest", "", "", false},
	}

	for _, tt := range tests {
		p, r, f, err := localArchiveReference(tt.raw)
		if err != nil || p != tt.wantPath || r != tt.wantRef || f != tt.wantFnd {
			t.Errorf("localArchiveReference(%q) = %v, %v, %v, err=%v", tt.raw, p, r, f, err)
		}
	}
}

// Tests parsing the root index.json of an OCI archive
func TestArchiveRoots(t *testing.T) {
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
func TestExtractOCILayoutTar_Success(t *testing.T) {
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

	extractedDir, err := extractOCILayoutTar(context.Background(), archivePath)
	if err != nil {
		t.Fatalf("extractOCILayoutTar failed: %v", err)
	}
	defer os.RemoveAll(extractedDir)

	if _, err := os.Stat(filepath.Join(extractedDir, "oci-layout")); os.IsNotExist(err) {
		t.Errorf("Failed to extract files properly")
	}
}

// Tests a successful execution of Tree.Run using a local OCI directory
func TestTree_Run_Success(t *testing.T) {
	dir := t.TempDir()

	// Setup a minimal valid OCI layout on disk
	if err := os.WriteFile(filepath.Join(dir, "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`), 0600); err != nil {
		t.Fatalf("writing oci-layout: %v", err)
	}
	childIndex := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[]}`)
	childDigest := digest.FromBytes(childIndex)
	// Use struct marshaling to ensure the JSON strictly matches OCI specs and ORAS expectations
	index := ocispec.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		Manifests: []ocispec.Descriptor{
			{
				MediaType: ocispec.MediaTypeImageIndex,
				Digest:    childDigest,
				Size:      int64(len(childIndex)),
			},
		},
	}
	indexBytes, _ := json.Marshal(index)
	if err := os.WriteFile(filepath.Join(dir, "index.json"), indexBytes, 0600); err != nil {
		t.Fatalf("writing index.json: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "blobs", "sha256"), 0700); err != nil {
		t.Fatalf("creating blobs directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "blobs", childDigest.Algorithm().String(), childDigest.Encoded()), childIndex, 0600); err != nil {
		t.Fatal(err)
	}
	tree := &Tree{
		OCILayout: true, // Bypasses remote registry resolution
		Depth:     1,
	}

	var buf bytes.Buffer
	// Append the digest so parseOCILayoutReference successfully splits the directory from the reference
	err := tree.Run(context.Background(), &buf, dir+"@"+childDigest.String())
	if err != nil {
		t.Fatalf("Tree.Run expected success, got error: %v", err)
	}

	if buf.Len() == 0 {
		t.Errorf("Expected Tree.Run to print output, got empty buffer")
	}
}
