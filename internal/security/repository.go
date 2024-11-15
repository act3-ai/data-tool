package security

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash"

	notationreg "github.com/notaryproject/notation-go/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"

	gitoci "git.act3-ace.com/ace/data/tool/internal/git/oci"
	"git.act3-ace.com/ace/data/tool/internal/mirror"
	"git.act3-ace.com/ace/data/tool/internal/mirror/encoding"
	"git.act3-ace.com/ace/go-common/pkg/logger"
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
	signatureDigest      string
	manifestDigestSBOM   string
	SBOMs                map[string][]*ocispec.Descriptor
	resultsReport        *ocispec.Descriptor
	originatingReference string             // only needed for gather artifacts
	shortenedName        string             // needed for graphing in mermaid
	CalculatedResults    ArtifactScanReport `json:"results"`
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
	r, err := registry.ParseReference(reference)
	if err != nil {
		return nil, fmt.Errorf("error parsing ref %s: %w", reference, err)
	}

	maniDetails.repository = repo
	d, data, err := oras.FetchBytes(ctx, repo, r.ReferenceOrDefault(), oras.DefaultFetchBytesOptions)
	if err != nil {
		return nil, fmt.Errorf("fetching the manifest bytes: %w", err)
	}
	maniDetails.desc = d
	maniDetails.isOCICompliant = encoding.IsOCICompliant(d.MediaType)
	p, err := repo.Predecessors(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("fetching predecessors: %w", err)
	}

	// Fetch predecessors/SBOMs/Signatures
	predecessors = append(predecessors, p...)

	if encoding.IsIndex(d.MediaType) {
		var idx ocispec.Index
		err = json.Unmarshal(data, &idx)
		if err != nil {
			return nil, err
		}
		size += d.Size
		for _, man := range idx.Manifests {
			var img ocispec.Manifest
			manDesc, manData, err := oras.FetchBytes(ctx, repo, man.Digest.String(), oras.DefaultFetchBytesOptions)
			if err != nil {
				return nil, fmt.Errorf("fetching the manifest bytes: %w", err)
			}
			// all of this below should be put in a manifest function, needs to return size, predecessors, and platforms
			if err := json.Unmarshal(manData, &img); err != nil {
				return nil, err
			}
			size += manDesc.Size
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
		size += d.Size
		if d.Platform != nil {
			platformString := formatPlatformString(d.Platform)
			platforms = append(platforms, platformString)
		} else {
			// pull the config and see if we can decode from there
			var img ocispec.Manifest
			if err := json.Unmarshal(data, &img); err != nil {
				return nil, err
			}
			// use the config descriptor to get the bytes
			data, err := content.FetchAll(ctx, repo, img.Config)
			if err != nil {
				return nil, fmt.Errorf("getting the config descriptor for %s: %w", reference, err)
			}
			// Is this an unsupported manifest for scan?
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

func (ad *ArtifactDetails) handlePredecessors(checksum string) {
	for _, p := range ad.predecessors {
		switch p.ArtifactType {
		case notationreg.ArtifactTypeNotation:
			ad.signatureDigest = p.Digest.String()
		case ArtifactTypeSPDX, ArtifactTypeHarborSBOM:
			ad.manifestDigestSBOM = p.Digest.String()
		case ArtifactTypeVulnerabilityReport:
			if p.Annotations[AnnotationGrypeDatabaseChecksum] == checksum {
				ad.resultsReport = &p
			}
		default:
			continue
		}
	}
}

func attachCVEResults(ctx context.Context, subjectDescriptor, configDescriptor ocispec.Descriptor,
	scanReport ArtifactScanReport,
	repository oras.GraphTarget,
	grypeChecksum string) ([]*ocispec.Descriptor, error) {

	var reports []*ocispec.Descriptor
	slog := logger.FromContext(ctx)
	// create the json results document
	data, err := json.Marshal(scanReport)
	if err != nil {
		return nil, fmt.Errorf("marshalling CVE results: %w", err)
	}

	blobDesc, err := oras.PushBytes(ctx, repository, ocispec.MediaTypeImageLayer, data)
	if err != nil {
		return nil, fmt.Errorf("pushing the CVE results: %w", err)
	}
	if encoding.IsIndex(subjectDescriptor.MediaType) {
		// fetch the reference index
		rc, err := repository.Fetch(ctx, subjectDescriptor)
		if err != nil {
			return nil, fmt.Errorf("error fetching the artifact index: %w", err)
		}
		var idx ocispec.Index
		decoder := json.NewDecoder(rc)
		if err := decoder.Decode(&idx); err != nil {
			return nil, fmt.Errorf("decoding subject index: %w", err)
		}
		for _, manifest := range idx.Manifests {
			r, err := attachCVEResults(ctx, manifest, configDescriptor, scanReport, repository, grypeChecksum)
			if err != nil {
				return nil, err
			}
			reports = append(reports, r...)
		}
	} else {
		packOpts := oras.PackManifestOptions{
			Subject:          &subjectDescriptor,
			Layers:           []ocispec.Descriptor{blobDesc},
			ConfigDescriptor: &configDescriptor,
			ManifestAnnotations: map[string]string{
				AnnotationGrypeDatabaseChecksum: grypeChecksum,
			},
		}

		maniDesc, err := oras.PackManifest(ctx, repository, oras.PackManifestVersion1_1, ArtifactTypeVulnerabilityReport, packOpts)
		if err != nil {
			return nil, fmt.Errorf("pushing vulnerability results manifest: %w", err)
		}
		slog.InfoContext(ctx, "pushed results", "reference", subjectDescriptor.Digest.String(), "resultsDigest", maniDesc.Digest.String())
		reports = append(reports, &maniDesc)
	}

	return reports, nil
}

func generateEmptyBlobDescriptor(mediatype string, algo digest.Algorithm) (ocispec.Descriptor, error) {
	emptyCfg := ocispec.ImageConfig{}
	emptyData, err := json.Marshal(emptyCfg)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("marshalling the empty config")
	}
	var h hash.Hash
	switch algo {
	case "sha256":
		h = sha256.New()
	case "sha384":
		h = sha512.New384()
	case "sha512":
		h = sha512.New()
	default:
		return ocispec.Descriptor{}, fmt.Errorf("invalid algorithm %s. Expected sha256, sha384, or sha512", algo)
	}

	// Compute the digest of the empty blob using the algorithm.
	if _, err := h.Write(emptyData); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("failed to compute hash: %w", err)
	}
	emptyDigest := digest.NewDigest(algo, h)
	// Create a descriptor using the computed digest.
	emptyBlobDescriptor := ocispec.Descriptor{
		MediaType: mediatype,
		Digest:    emptyDigest,
		Size:      int64(len(emptyData)), // should be 2
	}
	return emptyBlobDescriptor, nil
}
