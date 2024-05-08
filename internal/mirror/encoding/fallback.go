package encoding

import (
	"context"
	"encoding/json"
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
)

// IndexFallback HACK for JFrog and other registries that do not support Index of Index.
func IndexFallback(index *ocispec.Index) {
	// split out indexes into annotations
	indexDesciptors := make([]ocispec.Descriptor, 0, len(index.Manifests))
	imageDescriptors := make([]ocispec.Descriptor, 0, len(index.Manifests))
	for _, desc := range index.Manifests {
		if IsIndex(desc.MediaType) {
			indexDesciptors = append(indexDesciptors, desc)
		} else {
			imageDescriptors = append(imageDescriptors, desc)
		}
	}
	index.Manifests = imageDescriptors

	if len(indexDesciptors) > 0 {
		data, err := json.Marshal(indexDesciptors)
		if err != nil {
			panic(err) // this should never happen
		}
		// add the annotations to the main index
		index.Annotations[AnnotationExtraManifests] = string(data)
		// TODO The manifests are in the repo but an aggressive GC might remove them if not tagged
		// so we could tag them with something like sha256-digest-of-new-index.extra-manifest.0, .1, .2, etc.
	}
}

// ExtraManifests extracts extra (shadow) manifests from the image index.
// This is to support registries not properly support index of index.
func ExtraManifests(index *ocispec.Index) ([]ocispec.Descriptor, error) {
	encData, ok := index.Annotations[AnnotationExtraManifests]
	if !ok {
		return nil, nil
	}

	var list []ocispec.Descriptor
	if err := json.Unmarshal([]byte(encData), &list); err != nil {
		return nil, err
	}

	return list, nil
}

// Successors implements the oras.CopyGraphOptions.FindSuccessors callback function.
// Successors finds the successors of the current node.
// fetcher provides cached access to the source storage, and is suitable
// for fetching non-leaf nodes like manifests. Since anything fetched from
// fetcher will be cached in the memory, it is recommended to use original
// source storage to fetch large blobs.
func Successors(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	successors, err := content.Successors(ctx, fetcher, desc)
	if err != nil {
		return nil, fmt.Errorf("error finding successors for %s: %w", desc.Digest.String(), err)
	}

	if !IsIndex(desc.MediaType) {
		return successors, nil
	}

	// add the "unsupported" / shadow manifests
	data, err := content.FetchAll(ctx, fetcher, desc)
	if err != nil {
		return nil, fmt.Errorf("error fetching the content for %s: %w", desc.Digest.String(), err)
	}

	// OCI manifest index schema can be used to marshal docker manifest list
	var index ocispec.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	list, err := ExtraManifests(&index)
	if err != nil {
		return nil, err
	}
	successors = append(successors, list...)

	return successors, nil
}
