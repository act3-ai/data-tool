package oci

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fortytw2/leaktest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
)

func Test_PushDirOp(t *testing.T) {
	defer leaktest.Check(t)() //nolint

	dir := t.TempDir()

	// file 1
	f1 := filepath.Join(dir, "f1")
	if err := os.WriteFile(f1, []byte("this is for testing"), 0666); err != nil {
		t.Fatalf("writing to temp file1: err = %v", err)
	}

	// file 2
	f2 := filepath.Join(dir, "f2")
	if err := os.WriteFile(f2, []byte("this is for testing again"), 0777); err != nil {
		t.Fatalf("writing to temp file2: err = %v", err)
	}

	// nested file
	nestDir := filepath.Join(dir, "nest")
	if err := os.Mkdir(nestDir, 0777); err != nil {
		t.Fatal(err)
	}
	nestFile := filepath.Join(nestDir, "f3")
	if err := os.WriteFile(nestFile, []byte("this is the nested file for testing"), 0666); err != nil {
		t.Fatalf("writing to nested file: err = %v", err)
	}

	ctx := context.Background()
	tag := "v10"
	target := memory.New()
	platform1 := &ocispec.Platform{
		OS:           "linux",
		Architecture: "arm64",
	}

	desc1, err := PushDirOp(ctx, dir, target, tag, platform1, false, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", desc1)

	successors1, err := content.Successors(ctx, target, desc1)
	if err != nil {
		t.Fatal(err)
	}
	if len(successors1) != 1 {
		t.Error("expected one manifest")
	}

	platform2 := platform1
	platform2.Variant = "v8"
	desc2, err := PushDirOp(ctx, dir, target, tag, platform2, false, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", desc2)

	successors2, err := content.Successors(ctx, target, desc2)
	if err != nil {
		t.Fatal(err)
	}
	if len(successors2) != 2 {
		t.Error("expected two manifests")
	}
}

func Test_transferLayer(t *testing.T) {
	defer leaktest.Check(t)() //nolint

	dir := t.TempDir()

	// file 1
	f1 := filepath.Join(dir, "f1")
	if err := os.WriteFile(f1, []byte("this is for testing"), 0666); err != nil {
		t.Fatalf("writing to temp file1: err = %v", err)
	}

	// file 2
	f2 := filepath.Join(dir, "f2")
	if err := os.WriteFile(f2, []byte("this is for testing again"), 0777); err != nil {
		t.Fatalf("writing to temp file2: err = %v", err)
	}

	// nested file
	nestDir := filepath.Join(dir, "nest")
	if err := os.Mkdir(nestDir, 0777); err != nil {
		t.Fatal(err)
	}
	nestFile := filepath.Join(nestDir, "f3")
	if err := os.WriteFile(nestFile, []byte("this is the nested file for testing"), 0666); err != nil {
		t.Fatalf("writing to nested file: err = %v", err)
	}

	ctx := context.Background()
	target := memory.New()

	desc, err := transferLayer(ctx, dir, target)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", desc)

	rc, err := target.Fetch(ctx, desc)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	gzr, err := gzip.NewReader(rc)
	if err != nil {
		t.Fatal(err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	expectedFiles := map[string]struct{}{
		".":       {},
		"f1":      {},
		"f2":      {},
		"nest":    {},
		"nest/f3": {},
	}
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		t.Log("Found in layer:", hdr.Name)
		if _, ok := expectedFiles[hdr.Name]; !ok {
			t.Errorf("unexpected file %s", hdr.Name)
		} else {
			delete(expectedFiles, hdr.Name)
		}
	}

	if len(expectedFiles) != 0 {
		t.Errorf("missing files %v", expectedFiles)
	}
}
