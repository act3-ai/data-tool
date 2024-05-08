package pypi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/data/tool/internal/python"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// increment these if the data model changes so we are forced to re-upload data.
const syncVersionIndex = "5"
const syncVersionManifest = "7"

type pythonDistributionIndex struct {
	target oras.ReadOnlyTarget

	// index that we started with (index.Manifests is invalid/old)
	index ocispec.Index

	// maps distribution filenames to their descriptor
	lookup map[string]ocispec.Descriptor
}

// newPythonDistributionIndex downloads and parses a OCI index of python distributions.
func newPythonDistributionIndex(ctx context.Context, target oras.ReadOnlyTarget, fresh bool, project string) (*pythonDistributionIndex, error) {
	log := logger.FromContext(ctx).With("project", project)

	// project is the normalized name so it is lowercase alpha-numeric with dashes.  This is compatible with tags so no special encoding is needed.
	tag := project

	index := ocispec.Index{}
	desc, indexData, err := oras.FetchBytes(ctx, target, tag, oras.FetchBytesOptions{})
	switch {
	case fresh:
		log.InfoContext(ctx, "Starting a \"fresh\" with a new index")
	case err == nil:
		if desc.MediaType == ocispec.MediaTypeImageIndex {
			if err := json.Unmarshal(indexData, &index); err != nil {
				return nil, fmt.Errorf("invalid index: %w", err)
			}
			log.InfoContext(ctx, "Using existing index")
		} else {
			log.InfoContext(ctx, "Got a manifest but it was not an OCI index, ignoring the manifest")
		}
	case errors.Is(err, errdef.ErrNotFound):
		// not found, just start from scratch
		log.InfoContext(ctx, "Index not found, starting from scratch")
	default:
		return nil, fmt.Errorf("error getting old image index: %w", err)
	}

	lookup := make(map[string]ocispec.Descriptor, len(index.Manifests))
	if index.Annotations[PythonSyncVersion] == syncVersionIndex {
		// invalidate manifests that are not the same sync version sync version
		for _, d := range index.Manifests {
			if d.Annotations[PythonSyncVersion] == syncVersionManifest {
				// we can keep it
				filename := d.Annotations[PythonDistributionFilename]
				lookup[filename] = d
			}
		}
	} else {
		// invalidate the old index if the sync version does not match
		index = ocispec.Index{}
	}

	// set standard annotations
	index.Annotations = map[string]string{
		PythonDistributionProject: project,
		PythonSyncVersion:         syncVersionIndex,
	}

	// set the media type for idx (Image Index) for Gitlab compatibility
	// either of these will work types.DockerManifestList or types.OCIImageIndex
	index.MediaType = ocispec.MediaTypeImageIndex

	index.ArtifactType = MediaTypePythonDistributionIndex

	// required as per the OCI image spec
	index.SchemaVersion = 2

	// TODO make an map of filename to descriptor index so we can quickly run Update() and FindLayer()

	return &pythonDistributionIndex{
		target: target,
		index:  index,
		lookup: lookup,
	}, nil
}

// IndexData is the OCI image index data.  It is the updated manifest (modified by Update()).
func (p *pythonDistributionIndex) IndexData() ([]byte, error) {
	// convert from the lookup map to the slice of manifests
	p.index.Manifests = make([]ocispec.Descriptor, 0, len(p.lookup))
	for _, d := range p.lookup {
		p.index.Manifests = append(p.index.Manifests, d)
	}

	// sort the names so the index is deterministic
	slices.SortFunc(p.index.Manifests, func(a, b ocispec.Descriptor) int {
		// TODO sort based on version (then alphabetically)
		return strings.Compare(a.Annotations[PythonDistributionFilename], b.Annotations[PythonDistributionFilename])
	})

	return json.Marshal(p.index)
}

func (p *pythonDistributionIndex) Update(manifest ocispec.Descriptor) {
	// Amend the index by doing effectively a replacement keyed on filename
	filename := manifest.Annotations[PythonDistributionFilename]
	p.lookup[filename] = manifest
}

func (p *pythonDistributionIndex) LookupDistribution(filename string) (ocispec.Descriptor, bool) {
	d, ok := p.lookup[filename]
	return d, ok
}

// FetchDistribution downloads the manifest and extracts the descriptors for the python components.
// Returns the descriptors for the distribution file and the python metadata.
func (p *pythonDistributionIndex) FetchDistribution(ctx context.Context, img ocispec.Descriptor) (ocispec.Descriptor, *ocispec.Descriptor, error) {
	// fetch the manifest for the image
	manifestData, err := content.FetchAll(ctx, p.target, img)
	if err != nil {
		return ocispec.Descriptor{}, nil, fmt.Errorf("getting python image artifact: %w", err)
	}

	// decode the bytes
	manifest := ocispec.Manifest{}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return ocispec.Descriptor{}, nil, err
	}

	// ensure we only have one layer as expected
	n := len(manifest.Layers)
	if n < 1 {
		return ocispec.Descriptor{}, nil, fmt.Errorf("expected at least one layer in manifest but found %d", n)
	}

	var metadata *ocispec.Descriptor
	if n >= 2 && manifest.Layers[1].MediaType == MediaTypePythonDistributionMetadata {
		metadata = &manifest.Layers[1]
	}

	return manifest.Layers[0], metadata, nil
}

// GetDistributions returns distributions (filtered on yanked status) from the original OCI index.
func (p *pythonDistributionIndex) GetDistributions(yanked bool) ([]python.DistributionEntry, error) {
	distributions := make([]python.DistributionEntry, 0, len(p.index.Manifests))
	for _, m := range p.index.Manifests {
		dist := distributionFromDescriptor(m)
		if !yanked && dist.Yanked != nil {
			// This distribution has been yanked
			continue
		}
		distributions = append(distributions, dist)
	}
	return distributions, nil
}

func distributionFromDescriptor(desc ocispec.Descriptor) python.DistributionEntry {
	entry := python.DistributionEntry{
		URL:            desc.Annotations[PythonDistributionFilename], // relative URL
		Digest:         digest.Digest(desc.Annotations[PythonDistributionDigest]),
		Filename:       desc.Annotations[PythonDistributionFilename],
		RequiresPython: desc.Annotations[PythonDistributionRequiresPython],
		MetadataDigest: digest.Digest(desc.Annotations[PythonDistributionMetadataDigest]),
	}

	if reason, yanked := desc.Annotations[PythonDistributionYanked]; yanked {
		entry.Yanked = &reason
	}
	return entry
}

func annotationsFromDistribution(entry python.DistributionEntry, reproducible bool) map[string]string {
	annotations := map[string]string{
		PythonDistributionFilename:       entry.Filename,
		PythonDistributionRequiresPython: entry.RequiresPython,
		PythonSyncVersion:                syncVersionManifest,
	}

	if d := entry.Digest; d != "" {
		// this might not be the same digest algorithm used in the descriptor in the layer, so it is good to keep here
		// we aso need it to render the project.html page
		annotations[PythonDistributionDigest] = d.String()
	}

	if d := entry.MetadataDigest; d != "" {
		// This is needed to render the project.html page (data-core-metadata attribute)
		annotations[PythonDistributionMetadataDigest] = d.String()
	}

	if entry.Yanked != nil {
		annotations[PythonDistributionYanked] = *entry.Yanked
	}

	if reproducible {
		// this timestamp will be automatically generated by oras.PackManifest() if not specified
		// use a fixed value here in order to have reproducible images
		annotations[ocispec.AnnotationCreated] = "1970-01-01T00:00:00Z" // POSIX epoch
	}

	return annotations
}
