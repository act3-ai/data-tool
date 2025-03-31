package mirror

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"

	"gitlab.com/act3-ai/asce/data/tool/internal/mirror/encoding"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
	dtreg "gitlab.com/act3-ai/asce/data/tool/internal/registry"
	reg "gitlab.com/act3-ai/asce/data/tool/pkg/registry"
)

// DiffOptions represent the necessary options for the mirror ls command.
type DiffOptions struct {
	ExistingImages        []string
	RootArtifactReference string
	Targeter              reg.GraphTargeter
	Expanded              bool
}

// ListArtifacts handles the logic to list artifacts, expand artifacts, and filter artifacts in the mirror ls command.
func ListArtifacts(ctx context.Context, opts DiffOptions) ([][]string, error) {

	repository, err := opts.Targeter.GraphTarget(ctx, opts.RootArtifactReference)
	// repository, err := opts.RepoFunc(ctx, opts.RootArtifactReference)
	if err != nil {
		return nil, fmt.Errorf("fetching repository %s: %w", opts.RootArtifactReference, err)
	}

	desc, err := repository.Resolve(ctx, opts.RootArtifactReference)
	if err != nil {
		return nil, fmt.Errorf("resolving the artifact %s: %w", opts.RootArtifactReference, err)
	}

	filteredImages, err := generateFilteredList(ctx, opts.ExistingImages, opts.Targeter)
	if err != nil {
		return nil, err
	}

	isArtifact, idx, err := DescIsMirrorArtifact(ctx, desc, repository)
	if err != nil {
		return nil, err
	}
	if !isArtifact {
		return nil, fmt.Errorf("%s is not a mirror artifact", opts.RootArtifactReference)
	}
	// this is the table for printing
	manifestOriginalReferences := [][]string{}
	indexDesciptors := []ocispec.Descriptor{}
	// add in any indexes that are in the annotations if the artifact was created with --index-fallback
	v, ok := idx.Annotations[encoding.AnnotationExtraManifests]
	if ok {
		if err := json.Unmarshal([]byte(v), &indexDesciptors); err != nil {
			return nil, fmt.Errorf("unmarshalling the extra manifest annotations for manifest %s: %w", opts.RootArtifactReference, err)
		}
	}
	// add the manifests to the index descriptors
	indexDesciptors = append(indexDesciptors, idx.Manifests...)
	// iterate over the images
	for _, manifest := range indexDesciptors {
		if len(filteredImages) != 0 {
			_, ok := filteredImages[manifest.Digest]
			if ok {
				// filter it out and do not append
				continue
			}
		}
		if opts.Expanded {
			expandedImages, err := getExpandedImages(ctx, manifest, repository)
			if err != nil {
				return nil, err
			}
			if len(expandedImages) != 0 {
				for digest, srcRef := range expandedImages {
					_, ok := filteredImages[digest]
					if !ok {
						// we want to add the digest
						filteredImages[digest] = srcRef
						manifestOriginalReferences = append(manifestOriginalReferences, []string{srcRef, digest.String()})
					}
				}
				filteredImages[manifest.Digest] = ""
				// we don't want to add the top-level gather source reference to the table so we continue
				continue
			}
		}
		filteredImages[manifest.Digest] = ""
		manifestOriginalReferences = append(manifestOriginalReferences, []string{manifest.Annotations[ref.AnnotationSrcRef], manifest.Digest.String()})
	}
	// sort them in alphabetical order
	sort.Slice(manifestOriginalReferences[1:], func(i, j int) bool {
		return manifestOriginalReferences[i+1][0] < manifestOriginalReferences[j+1][0]
	})
	return manifestOriginalReferences, nil
}

// getExpandedImages asserts whether the given descriptor is a mirror artifact and if so, returns a map of its images.
func getExpandedImages(ctx context.Context, desc ocispec.Descriptor, repository oras.GraphTarget) (map[digest.Digest]string, error) {
	if encoding.IsIndex(desc.MediaType) {
		isMirrorArtifact, nestedIdx, err := DescIsMirrorArtifact(ctx, desc, repository)
		if err != nil {
			return nil, err
		}
		if isMirrorArtifact {
			// all nested images in Digest:original reference
			return getNestedArtifactImages(ctx, *nestedIdx, repository)
		}
	}
	return nil, nil
}

// getNestedArtifactImages returns a map of all mirror artifact manifests (inclusive of nested mirror artifacts).
func getNestedArtifactImages(ctx context.Context, gatheredIndex ocispec.Index, repository oras.GraphTarget) (map[digest.Digest]string, error) {
	// create a map to hold existing images from nested gather artifacts
	diffImages := map[digest.Digest]string{}

	// are there any extra indexes at the top level of the gatheredIndex?
	manifestList, err := getExtraIndexes(gatheredIndex)
	if err != nil {
		return nil, err
	}
	// add all of the index manifests to the list
	manifestList = append(manifestList, gatheredIndex.Manifests...)
	for _, manifest := range manifestList {
		isGatherArtifact, nestedIndex, err := DescIsMirrorArtifact(ctx, manifest, repository)
		if err != nil {
			return nil, err
		}
		if isGatherArtifact {
			images, err := getNestedArtifactImages(ctx, *nestedIndex, repository)
			if err != nil {
				return nil, err
			}
			for k, v := range images {
				diffImages[k] = v
			}
		} else {
			diffImages[manifest.Digest] = manifest.Annotations[ref.AnnotationSrcRef]
		}
	}
	return diffImages, nil
}

// getExtraIndexes returns the index descriptors in a given artifact's extra manifests annotation.
func getExtraIndexes(nestedIndex ocispec.Index) ([]ocispec.Descriptor, error) {
	indexDescriptors := []ocispec.Descriptor{}
	subExtraIndexes, ok := nestedIndex.Annotations[encoding.AnnotationExtraManifests]
	if ok {
		if err := json.Unmarshal([]byte(subExtraIndexes), &indexDescriptors); err != nil {
			return nil, fmt.Errorf("unmarshalling the extra manifest annotations for nested manifest %s: %w", nestedIndex.Annotations[ref.AnnotationSrcRef], err)
		}
	}
	return indexDescriptors, nil
}

// generateFilteredList creates a map inclusive of all manifest and sub-manifest digests within the given slice of reference strings in existingImages. It returns the map for easy searching.
func generateFilteredList(ctx context.Context, existingImages []string, graphTargeter reg.GraphTargeter) (map[digest.Digest]string, error) {
	filteredImages := map[digest.Digest]string{}
	// get the image and put its digest and all of its sub-digests into the map
	for _, image := range existingImages {
		repo, err := graphTargeter.GraphTarget(ctx, image)
		if err != nil {
			return nil, fmt.Errorf("generating target for existing image %s: %w", image, err)
		}
		reference, err := dtreg.ParseEndpointOrDefault(graphTargeter, image)
		if err != nil {
			return nil, err
		}
		// get the descriptor
		desc, err := repo.Resolve(ctx, reference.String())
		if err != nil {
			return nil, fmt.Errorf("resolving existing image %s: %w", reference, err)
		}
		// get all of the manifest and sub-manifest digests for the given descriptor.
		e, err := existingManifests(ctx, desc, repo)
		if err != nil {
			return nil, err
		}
		for _, k := range e {
			filteredImages[k] = "" // nothing needs to be stored in the value, the map is purely for key searching (by digest).
		}
	}
	return filteredImages, nil
}

// existingManifests returns a slice of all manifest and sub-manifest digests for a given descriptor.
func existingManifests(ctx context.Context, desc ocispec.Descriptor, repo oras.GraphTarget) ([]digest.Digest, error) {
	existing := []digest.Digest{}
	// add the root to the existing map
	existing = append(existing, desc.Digest)
	// is it a mirror artifact?
	isMirrorArtifact, gatheredIndex, err := DescIsMirrorArtifact(ctx, desc, repo)
	if err != nil {
		return nil, err
	}
	if isMirrorArtifact {
		// all nested images in Digest:original reference
		nestedImages, err := getNestedArtifactImages(ctx, *gatheredIndex, repo)
		if err != nil {
			return nil, err
		}
		for k := range nestedImages {
			existing = append(existing, k)
		}
	}

	return existing, nil
}

// DescIsMirrorArtifact takes a descriptor and a graph target and returns whether the descriptor is a mirror artifact, the index, and/or an error.
func DescIsMirrorArtifact(ctx context.Context, desc ocispec.Descriptor, repo oras.GraphTarget) (bool, *ocispec.Index, error) {
	rc, err := repo.Fetch(ctx, desc)
	if err != nil {
		return false, nil, fmt.Errorf("fetching artifact: %w", err)
	}
	if encoding.IsIndex(desc.MediaType) {
		var idx ocispec.Index
		decoder := json.NewDecoder(rc)
		if err := decoder.Decode(&idx); err != nil {
			return false, nil, fmt.Errorf("decoding nested index: %w", err)
		}
		if idx.ArtifactType == encoding.MediaTypeGather {
			// is a mirror artifact
			return true, &idx, nil
		}
		// is not a mirror artifact but is an index (this might help cut down on network calls)
		return false, &idx, nil
	}
	// is not a mirror artifact or an index
	return false, nil, nil
}
