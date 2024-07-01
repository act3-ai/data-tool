package oci

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror/encoding"
)

// PushDir represents the oci pushdir action.
type PushDir struct {
	*Action

	Platform platformValue

	// Use legacy media types
	Legacy bool

	// Reproducible removes timestamps.
	Reproducible bool
}

// Run performs the pushdir operation.
func (action *PushDir) Run(ctx context.Context, dir, ref string) error {
	repo, err := action.Config.Repository(ctx, ref)
	if err != nil {
		return err
	}

	_, err = PushDirOp(ctx, dir, repo, repo.Reference.ReferenceOrDefault(),
		action.Platform.platform, action.Legacy, action.Reproducible)
	if err != nil {
		return err
	}

	return nil
}

// PushDirOp pushes a directory as the single layer in a OCI image.
func PushDirOp(ctx context.Context, dir string, target oras.Target, ref string,
	platform *ocispec.Platform, legacy, reproducible bool,
) (ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "Fetching index", "ociref", ref)
	index := ocispec.Index{}
	idxDesc, idxData, err := oras.FetchBytes(ctx, target, ref, oras.FetchBytesOptions{})
	switch {
	case err == nil:
		if encoding.IsIndex(idxDesc.MediaType) {
			if err := json.Unmarshal(idxData, &index); err != nil {
				return ocispec.Descriptor{}, fmt.Errorf("invalid index: %w", err)
			}
			log.InfoContext(ctx, "Using existing index")
		} else {
			log.InfoContext(ctx, "Got a manifest but it was not an OCI index, ignoring the manifest")
		}
	case errors.Is(err, errdef.ErrNotFound):
		// not found, just start from scratch
		log.InfoContext(ctx, "Index not found, starting from scratch")
	default:
		return ocispec.Descriptor{}, fmt.Errorf("error getting old image index: %w", err)
	}

	// these values should always be the same
	index.SchemaVersion = 2
	index.MediaType = ocispec.MediaTypeImageIndex

	// set the media type for idx (Image Index) for Gitlab compatibility
	// either of these will work types.DockerManifestList or types.OCIImageIndex
	// For compatibility see https://github.com/opencontainers/image-spec/blob/main/media-types.md
	// .annotations is not supported with the legacy media type
	if legacy {
		log.InfoContext(ctx, "Using legacy/docker media type for index")
		index.MediaType = encoding.MediaTypeDockerManifestList
	}

	// push dir tar
	log.InfoContext(ctx, "Pushing directory as a layer")
	desc, err := transferLayer(ctx, dir, target)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	// create and push config
	config := ocispec.Image{
		Platform: *platform,
		RootFS:   ocispec.RootFS{Type: "layers"}, // TODO the DiffIDs are wrong
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("encoding manifest config: %w", err)
	}

	log.InfoContext(ctx, "Pushing config")
	configDesc, err := oras.PushBytes(ctx, target, ocispec.MediaTypeImageConfig, configBytes)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("pushing config to repo: %w", err)
	}

	// pack and push manifest
	manOpts := oras.PackManifestOptions{
		Layers:           []ocispec.Descriptor{desc},
		ConfigDescriptor: &configDesc,
	}

	if reproducible {
		// this timestamp will be automatically generated by oras.PackManifest() if not specified
		// use a fixed value here in order to have reproducible images
		manOpts.ManifestAnnotations = map[string]string{ocispec.AnnotationCreated: "1970-01-01T00:00:00Z"} // POSIX epoch
	}

	log.InfoContext(ctx, "Pushing manifest")
	manDesc, err := oras.PackManifest(ctx, target, oras.PackManifestVersion1_1, ocispec.MediaTypeImageManifest, manOpts)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("packing and pushing manifest: %w", err)
	}
	log.InfoContext(ctx, "Image manifest digest", "digest", manDesc.Digest)

	// modify desc so we include the platform in it's manifest descriptor
	manDesc.Platform = platform

	// append manifest to index or replace an existing platform.
	found := false
	for i, man := range index.Manifests {
		if man.Platform.Architecture == platform.Architecture &&
			man.Platform.OS == platform.OS &&
			man.Platform.Variant == platform.Variant {
			found = true
			log.InfoContext(ctx, "Replacing platform in image index", "platform", platform)
			index.Manifests[i] = manDesc
		}
	}
	if !found {
		log.InfoContext(ctx, "Platform not found, appending new platform to index", "platform", platform)
		index.Manifests = append(index.Manifests, manDesc)
	}

	// "pack" and push index
	indexData, err := json.Marshal(index)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("encoding index for pushing: %w", err)
	}
	log.InfoContext(ctx, "Pushing index")
	indexDesc, err := oras.TagBytes(ctx, target, index.MediaType, indexData, ref)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("pushing index to repo: %w", err)
	}

	// return the index digest reference
	return indexDesc, nil
}

// transferLayer archives a local directory or file, sending it to the target repo as a single layer.
func transferLayer(ctx context.Context, dir string, target content.Storage) (ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)

	fstore, err := file.New(dir)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("initializing filestore: %w", err)
	}
	fstore.TarReproducible = true
	defer fstore.Close()

	// add tar file to filestore, the image.title annotation is set to the name, which is the dir path, by oras.
	fdesc, err := fstore.Add(ctx, ".", ocispec.MediaTypeImageLayerGzip, "")
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("adding directory %q to filestore: %w", dir, err)
	}

	log.InfoContext(ctx, "Pushing tar archive")
	if err := oras.CopyGraph(ctx, fstore, target, fdesc, oras.DefaultCopyGraphOptions); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("copying layer: %w", err)
	}

	log.InfoContext(ctx, "Layer completed", "digest", fdesc.Digest)
	return fdesc, nil
}
