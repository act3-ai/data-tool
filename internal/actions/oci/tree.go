package oci

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"

	"github.com/act3-ai/data-tool/internal/print"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

)

// Tree represents the oci tree action.
type Tree struct {
	*Action

	OCILayout bool

	Depth int

	ShowBlobs       bool
	ShortDigests    bool
	ShowAnnotations bool
	CheckBlobs      bool

	OnlyReferrers bool
	ArtifactType  string
}

// Run performs the oci tree operation.
func (action *Tree) Run(ctx context.Context, out io.Writer, rawRef string) error {
	var storage oras.ReadOnlyGraphTarget
	var ref string

	var roots []ocispec.Descriptor

	archivePath := ""
	archiveRef := ""
	if !action.OCILayout {
		var found bool
		var err error
		archivePath, archiveRef, found, err = localArchiveReference(rawRef)
		if err != nil {
			return err
		}
		if !found {
			archivePath = ""
		}
	}

	switch {
	case action.OCILayout:
		// local directory in oci layout
		srcPath, r, err := parseOCILayoutReference(rawRef)
		if err != nil {
			return err
		}

		store, err := oci.NewFromFS(ctx, os.DirFS(srcPath))
		if err != nil {
			return fmt.Errorf("opening OCI image layout directory: %w", err)
		}

		storage = store
		ref = r

	case archivePath != "":
		tempDir, err := extractOCILayoutTar(ctx, archivePath)
		if err != nil {
			return fmt.Errorf("extracting OCI layout tarball %q: %w", archivePath, err)
		}
		defer os.RemoveAll(tempDir)

		store, err := oci.NewFromFS(ctx, os.DirFS(tempDir))
		if err != nil {
			return fmt.Errorf("opening OCI image layout tarball: %w", err)
		}
		storage = store
		ref = archiveRef
		if ref == "" {
			roots, err = archiveRoots(tempDir)
			if err != nil {
				return err
			}
		}
	default:
		// remote reference
		repo, err := action.Config.Repository(ctx, rawRef)
		if err != nil {
			return err
		}
		storage = repo
		ref = repo.Reference.ReferenceOrDefault()
	}

	var node ocispec.Descriptor
	if len(roots) == 0 {
		var err error
		node, err = storage.Resolve(ctx, ref)
		if err != nil {
			return fmt.Errorf("resolving image reference %s failed: %w", ref, err)
		}
	}

	o := print.Options{}
	o.Depth = action.Depth
	o.DisableBlobs = !action.ShowBlobs
	o.ShortDigests = action.ShortDigests
	o.ShowAnnotations = action.ShowAnnotations
	o.OnlyReferrers = action.OnlyReferrers
	o.ArtifactType = action.ArtifactType
	o.CheckBlobs = action.CheckBlobs

	if len(roots) != 0 {
    for i, root := range roots {
        if i != 0 {
            if _, err := fmt.Fprintln(out); err != nil {
                return err
            }
        }
        if err := print.All(ctx, out, storage, root, o); err != nil {
            return fmt.Errorf("printing OCI archive descriptor %s: %w", root.Digest, err)
        }
    }
    return nil
}
	return print.All(ctx, out, storage, node, o)
}

// parseOCILayoutReference parses the raw in format of <path>[:<tag>|@<digest>].
func parseOCILayoutReference(raw string) (string, string, error) {
	if idx := strings.LastIndex(raw, "@"); idx != -1 {
		// `digest` found
		return raw[:idx], raw[idx+1:], nil
	}
	// find `tag`
	if idx := strings.LastIndex(raw, ":"); idx != -1 {
		return raw[:idx], raw[idx+1:], nil
	}

	return "", "", fmt.Errorf(`directory path and reference must be separated by "@" for digests and ":" for tags in %q`, raw)
}
