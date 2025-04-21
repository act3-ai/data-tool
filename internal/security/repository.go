package security

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	notationreg "github.com/notaryproject/notation-go/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	gitoci "github.com/act3-ai/data-tool/internal/git/oci"
	"github.com/act3-ai/data-tool/internal/mirror"
	"github.com/act3-ai/data-tool/internal/mirror/encoding"
	"github.com/act3-ai/go-common/pkg/logger"
)

// ArtifactDetails contains all of the details needed to scan for vulnerabilities of a given artifact.
type ArtifactDetails struct {
	repository           oras.GraphTarget
	desc                 ocispec.Descriptor
	source               mirror.Source
	platforms            []string
	size                 int64
	isOCICompliant       bool
	isNotScanSupported   bool
	predecessors         []ocispec.Descriptor
	signatureDesc        ocispec.Descriptor
	manifestSBOMDesc     ocispec.Descriptor
	SBOMs                map[string][]*ocispec.Descriptor
	resultsReport        *ocispec.Descriptor
	virusScanReport      *ocispec.Descriptor
	originatingReference string                     // only needed for gather artifacts
	shortenedName        string                     // needed for graphing in mermaid
	CalculatedResults    ArtifactScanReport         `json:"results"`
	MalwareResults       []*VirusScanManifestReport `json:"malware-results"`
}

// GetArtifactDetails fetches the ArtifactDetails for a given reference.
func GetArtifactDetails( //nolint:gocognit
	ctx context.Context,
	reference string,
	repo oras.GraphTarget) (*ArtifactDetails, error) {

	maniDetails := &ArtifactDetails{}
	maniDetails.source.Name = reference
	var platforms []string
	var size int64
	sboms := map[string][]*ocispec.Descriptor{}
	var predecessors []ocispec.Descriptor
	maniDetails.repository = repo

	desc, err := repo.Resolve(ctx, reference)
	if err != nil {
		return nil, fmt.Errorf("error resolving reference %s: %w", reference, err)
	}

	data, err := repo.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("fetching the manifest bytes for %s: %w", reference, err)
	}

	maniDetails.desc = desc
	maniDetails.isOCICompliant = encoding.IsOCICompliant(desc.MediaType)
	p, err := repo.Predecessors(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("fetching predecessors: %w", err)
	}

	// Fetch predecessors/SBOMs/Signatures
	predecessors = append(predecessors, p...)

	if encoding.IsIndex(desc.MediaType) {
		var idx ocispec.Index
		decoder := json.NewDecoder(data)
		if err := decoder.Decode(&idx); err != nil {
			return nil, fmt.Errorf("decoding the index bytes: %w", err)
		}
		size += desc.Size
		for _, man := range idx.Manifests {
			var img ocispec.Manifest
			manData, err := repo.Fetch(ctx, man)
			if err != nil {
				return nil, fmt.Errorf("fetching the manifest bytes: %w", err)
			}
			// all of this below should be put in a manifest function, needs to return size, predecessors, and platforms
			decoder := json.NewDecoder(manData)
			if err := decoder.Decode(&img); err != nil {
				return nil, fmt.Errorf("decoding the manifest data: %w", err)
			}
			size += man.Size
			for _, blob := range img.Layers {
				size += blob.Size
			}
			platformString := formatPlatformString(man.Platform)
			platforms = append(platforms, platformString)
			p, err := repo.Predecessors(ctx, man)
			if err != nil {
				return maniDetails, fmt.Errorf("fetching predecessors: %w", err)
			}
			// Fetch predecessors/SBOMs/Signatures
			predecessors = append(predecessors, p...)
			for _, pred := range p {
				if IsSBOM(pred.ArtifactType) {
					if platformString != "" {
						sboms[platformString] = append(sboms[platformString], &pred)
					} else {
						sboms["unknown/unknown"] = append(sboms["unknown/unknown"], &pred)
					}
				}
			}
		}
		maniDetails.SBOMs = sboms
	} else {
		size += desc.Size
		if desc.Platform != nil {
			platformString := formatPlatformString(desc.Platform)
			platforms = append(platforms, platformString)
		} else {
			// pull the config and see if we can decode from there
			var img ocispec.Manifest
			decoder := json.NewDecoder(data)
			if err := decoder.Decode(&img); err != nil {
				return nil, fmt.Errorf("decoding the manifest: %w", err)
			}
			// use the config descriptor to get the bytes
			data, err := content.FetchAll(ctx, repo, img.Config)
			if err != nil {
				return nil, fmt.Errorf("getting the config descriptor for %s: %w", reference, err)
			}
			// check if this an unsupported manifest for vulnerability scanning
			if img.Config.MediaType == MediaTypeHelmChartConfig || img.Config.MediaType == gitoci.MediaTypeSyncConfig {
				maniDetails.isNotScanSupported = true
			}
			var platform ocispec.Platform
			if err := json.Unmarshal(data, &platform); err != nil {
				return nil, fmt.Errorf("decoding platform %+v: %w", &platform, err)
			}
			platformString := formatPlatformString(&platform)
			platforms = append(platforms, platformString)
		}
	}
	maniDetails.platforms = platforms
	maniDetails.size = size
	maniDetails.predecessors = predecessors

	return maniDetails, nil
}

// handlePredecessors populates ArtifactDetails with discovered predecessors.
func (ad *ArtifactDetails) handlePredecessors(grypeChecksumDB string, clamavChecksums []ClamavDatabase) error {
	for _, p := range ad.predecessors {
		switch p.ArtifactType {
		case notationreg.ArtifactTypeNotation:
			ad.signatureDesc = p
		case ArtifactTypeSPDX, ArtifactTypeHarborSBOM:
			ad.manifestSBOMDesc = p
		case ArtifactTypeVulnerabilityReport:
			if p.Annotations[AnnotationGrypeDatabaseChecksum] == grypeChecksumDB {
				ad.resultsReport = &p
			}
		case ArtifactTypeVirusScanReport:
			b, err := json.Marshal(clamavChecksums)
			if err != nil {
				return fmt.Errorf("marshalling the clamav checksums: %w", err)
			}
			if p.Annotations[AnnotationVirusDatabaseChecksum] == string(b) {
				ad.virusScanReport = &p
			}
		default:
			continue
		}
	}
	return nil
}

// attachResultsReport pushes a referrer manifest containing vulnerabiliy or
// virus scanning results.
func attachResultsReport(ctx context.Context, subjectDescriptor, configDescriptor ocispec.Descriptor,
	scanReport any,
	repository oras.GraphTarget,
	annotations map[string]string,
	artifactType string) ([]*ocispec.Descriptor, error) {

	log := logger.FromContext(ctx)
	var reports []*ocispec.Descriptor
	// create the json results document
	data, err := json.Marshal(scanReport)
	if err != nil {
		return nil, fmt.Errorf("marshalling results report: %w", err)
	}

	blobDesc, err := oras.PushBytes(ctx, repository, ocispec.MediaTypeImageLayer, data)
	if err != nil {
		return nil, fmt.Errorf("pushing the results report: %w", err)
	}
	if encoding.IsIndex(subjectDescriptor.MediaType) {
		// fetch the reference index
		rc, err := repository.Fetch(ctx, subjectDescriptor)
		if err != nil {
			return nil, fmt.Errorf("error fetching the subject artifact index: %w", err)
		}
		var idx ocispec.Index
		decoder := json.NewDecoder(rc)
		if err := decoder.Decode(&idx); err != nil {
			return nil, fmt.Errorf("decoding subject index: %w", err)
		}
		for _, manifest := range idx.Manifests {
			r, err := attachResultsReport(ctx, manifest, configDescriptor, scanReport, repository, annotations, artifactType)
			if err != nil {
				return nil, err
			}
			reports = append(reports, r...)
		}
	} else {
		packOpts := oras.PackManifestOptions{
			Subject:             &subjectDescriptor,
			Layers:              []ocispec.Descriptor{blobDesc},
			ConfigDescriptor:    &configDescriptor,
			ManifestAnnotations: annotations,
		}

		manDesc, err := oras.PackManifest(ctx, repository, oras.PackManifestVersion1_1, artifactType, packOpts)
		if err != nil {
			return nil, fmt.Errorf("pushing results manifest: %w", err)
		}
		reports = append(reports, &manDesc)
		log.InfoContext(ctx, "pushed results report", "digest", manDesc.Digest.String())
	}

	return reports, nil
}

// pushEmptyDesc pushes an empty blob, i.e. `{}`, exiting gracefully if it already exists.
// https://github.com/opencontainers/image-spec/blob/main/manifest.md#guidance-for-an-empty-descriptor.
func pushEmptyDesc(ctx context.Context, storage content.Pusher, alg digest.Algorithm) error {
	switch alg {
	case digest.SHA256:
		err := storage.Push(ctx, ocispec.DescriptorEmptyJSON, bytes.NewReader([]byte(`{}`)))
		if err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
			return fmt.Errorf("pushing empty config with sha256 algorithm: %w", err)
		}
	// case digest.SHA384:
	// TODO
	// sha384 isn't registered with the OCI spec, but oras is looking to support it, so we might as well
	// https://github.com/oras-project/oras-go/issues/898
	// case digest.SHA512:
	// TODO
	default:
		return fmt.Errorf("unsupported algorithm %s, accepted algorithms sha256", alg)
	}

	return nil
}
