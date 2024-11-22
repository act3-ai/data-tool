package mirror

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"

	"git.act3-ace.com/ace/data/tool/internal/mirror/encoding"
	"git.act3-ace.com/ace/data/tool/internal/ref"
	dtreg "git.act3-ace.com/ace/data/tool/internal/registry"
	reg "git.act3-ace.com/ace/data/tool/pkg/registry"
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

	// iterate over the images
	manifestOriginalReferences := [][]string{{"reference", "digest"}}
	for _, manifest := range idx.Manifests {
		if len(filteredImages) != 0 {
			_, ok := filteredImages[manifest.Digest]
			if ok {
				// filter it out and do not append
				continue
			}
		}
		if opts.Expanded {
			// if it's not an index we don't need to expand to get the artifact type
			expandedImages, err := getExpandedImages(ctx, manifest, filteredImages, repository)
			if err != nil {
				return nil, err
			}
			if len(expandedImages) != 0 {
				manifestOriginalReferences = append(manifestOriginalReferences, expandedImages...)
				filteredImages[manifest.Digest] = ""
				// we don't want to add the top-level gather source reference to the table
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

func getExpandedImages(ctx context.Context, desc ocispec.Descriptor, filteredImages map[digest.Digest]string, repository oras.GraphTarget) ([][]string, error) {
	manifestOriginalReferences := [][]string{}
	if encoding.IsIndex(desc.MediaType) {
		isMirrorArtifact, nestedIdx, err := DescIsMirrorArtifact(ctx, desc, repository)
		if err != nil {
			return nil, err
		}
		if isMirrorArtifact {
			// all nested images in Digest:original reference
			nestedImages, err := getNestedArtifactImages(ctx, *nestedIdx, repository)
			if err != nil {
				return nil, err
			}
			for digest, srcRef := range nestedImages {
				_, ok := filteredImages[digest]
				if !ok {
					// we want to add the digest
					filteredImages[digest] = srcRef
					manifestOriginalReferences = append(manifestOriginalReferences, []string{srcRef, digest.String()})
				}
			}
		}
	}
	return manifestOriginalReferences, nil
}

func getNestedArtifactImages(ctx context.Context, gatheredIndex ocispec.Index, repository oras.GraphTarget) (map[digest.Digest]string, error) {
	// create a map to hold existing images from nested gather artifacts
	diffImages := map[digest.Digest]string{}

	for _, manifest := range gatheredIndex.Manifests {
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
			return diffImages, nil
		}
		diffImages[manifest.Digest] = manifest.Annotations[ref.AnnotationSrcRef]
	}
	return diffImages, nil
}

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
		e, err := existingManifests(ctx, reference.String(), repo)
		if err != nil {
			return nil, err
		}
		for k := range e {
			filteredImages[k] = ""
		}
	}
	return filteredImages, nil
}

func existingManifests(ctx context.Context, reference string, repo oras.GraphTarget) (map[digest.Digest]string, error) {
	existing := map[digest.Digest]string{}
	// get the digest
	desc, err := repo.Resolve(ctx, reference)
	if err != nil {
		return nil, fmt.Errorf("resolving existing image %s: %w", reference, err)
	}
	// add the root to the existing map
	existing[desc.Digest] = ""
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
		for k, v := range nestedImages {
			existing[k] = v
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
