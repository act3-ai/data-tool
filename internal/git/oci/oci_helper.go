// Package oci facilitates transferring of git and git-lfs OCI artifacts.
package oci

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/git/cmd"
)

// Helper assists in pushing to or fetching from an OCI compliant registry.
type Helper struct {
	Target     oras.GraphTarget
	Tag        string
	FStore     *file.Store
	FStorePath string
}

// PostCopyLFS returns a func for the oras.CopyGraphOptions option PostCopy func. It adds a hardlink from the
// LFS file in fstorePath to the destRepoPath.
func PostCopyLFS(fstorePath, destRepoPath string) func(ctx context.Context, desc ocispec.Descriptor) error {
	return func(ctx context.Context, desc ocispec.Descriptor) error {
		if desc.MediaType == MediaTypeLFSLayer {
			oid := desc.Annotations[ocispec.AnnotationTitle] // oid filename
			destPath := filepath.Join(destRepoPath, cmd.ResolveLFSOIDPath(oid))
			logger.V(logger.FromContext(ctx), 1).InfoContext(ctx, "linking LFS file to intermediate dir", "objectID", oid, "destPath", destPath) //nolint:sloglint

			// init nested dirs
			err := os.MkdirAll(filepath.Dir(destPath), 0777)
			if err != nil {
				return fmt.Errorf("creating path to oid file: %w", err)
			}

			// link
			srcPath := filepath.Join(fstorePath, oid)
			if err := os.Link(srcPath, destPath); err != nil {
				return fmt.Errorf("adding LFS file to cache: %w", err)
			}
		}
		return nil
	}
}

// FindSuccessorsLFS limits the LFS layers copied to the provided set.
// Returns the default oras FindSuccessors result if no layers are provided.
// Assumes any image manifest encountered is an LFS manifest.
func FindSuccessorsLFS(lfsLayers []ocispec.Descriptor) func(ctx context.Context, fetcher content.Fetcher,
	desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	return func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		log := logger.FromContext(ctx)

		var successors []ocispec.Descriptor
		switch {
		case ocispec.MediaTypeImageManifest == desc.MediaType && (lfsLayers != nil || len(lfsLayers) > 0): // only filter if layers are provided
			// do we already have the artifact type?
			at := desc.ArtifactType
			if ArtifactTypeLFSManifest == at {
				successors = lfsLayers
				break
			}

			// resolve artifact type
			manBytes, err := content.FetchAll(ctx, fetcher, desc)
			if err != nil {
				return nil, fmt.Errorf("fetching manifest: %w", err)
			}

			var manifest ocispec.Manifest
			err = json.Unmarshal(manBytes, &manifest)
			if err != nil {
				return nil, fmt.Errorf("decoding manifest: %w", err)
			}

			at = manifest.ArtifactType
			if ArtifactTypeLFSManifest == at {
				successors = lfsLayers
			}
		default:
			var err error
			successors, err = content.Successors(ctx, fetcher, desc)
			if err != nil {
				return nil, fmt.Errorf("error finding successors for %s: %w", desc.Digest.String(), err)
			}
			log.InfoContext(ctx, "found successors of descriptor", "descriptor", desc, "successors", len(successors))

		}
		return successors, nil
	}
}

// FindSuccessorsBundles limits the bundle layers copied to the provided set.
// Returns the default oras FindSuccessors result if no layers are provided.
// Assumes any image manifest encounters is a base git manifest.
func FindSuccessorsBundles(manDesc ocispec.Descriptor, bundleLayers []ocispec.Descriptor) func(ctx context.Context, fetcher content.Fetcher,
	desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	return func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		if desc.MediaType == ocispec.MediaTypeImageManifest {
			return bundleLayers, nil
		}
		return content.Successors(ctx, fetcher, desc)
	}
}
