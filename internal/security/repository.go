package security

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash"
	"log/slog"

	notationreg "github.com/notaryproject/notation-go/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"

	"github.com/act3-ai/bottle-schema/pkg/mediatype"
	gitoci "github.com/act3-ai/data-tool/internal/git/oci"
	"github.com/act3-ai/data-tool/internal/mirror"
	"github.com/act3-ai/data-tool/internal/mirror/encoding"
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
			// Is this an unsupported manifest for vulnerability scanning?
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

func (ad *ArtifactDetails) handlePredecessors(grypeChecksumDB string, clamavChecksums []ClamavDatabase) error {
	for _, p := range ad.predecessors {
		switch p.ArtifactType {
		case notationreg.ArtifactTypeNotation:
			ad.signatureDigest = p.Digest.String()
		case ArtifactTypeSPDX, ArtifactTypeHarborSBOM:
			ad.manifestDigestSBOM = p.Digest.String()
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

func attachResultsReport(ctx context.Context, subjectDescriptor, configDescriptor ocispec.Descriptor,
	scanReport any,
	repository oras.GraphTarget,
	annotations map[string]string,
	artifactType string) ([]*ocispec.Descriptor, error) {

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

		maniDesc, err := oras.PackManifest(ctx, repository, oras.PackManifestVersion1_1, artifactType, packOpts)
		if err != nil {
			return nil, fmt.Errorf("pushing vulnerability results manifest: %w", err)
		}
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

func VirusScan(ctx context.Context,
	desc ocispec.Descriptor,
	repository oras.GraphTarget,
	clamavChecksums []ClamavDatabase,
	pushResults bool,
	cachePath string) ([]*VirusScanManifestReport, error) {

	// do reports exist with matching clamav checksums?

	var reports []*VirusScanManifestReport

	descCfg, err := generateEmptyBlobDescriptor(ocispec.MediaTypeImageConfig, digest.Canonical)
	if err != nil {
		return nil, err
	}

	// get the config if it's a image manifest
	// if it's a bottle, we can pull filenames
	// if it's an image, we need to derive the image source/architecture
	// if it's a helm chart??? is it supported?
	// output any errors but do not fail until the end
	if encoding.IsIndex(desc.MediaType) {
		var idx ocispec.Index
		b, err := content.FetchAll(ctx, repository, desc)
		if err != nil {
			return nil, fmt.Errorf("fetching index manifest: %w", err)
		}
		if err := json.Unmarshal(b, &idx); err != nil {
			return nil, fmt.Errorf("parsing the image manifest: %w", err)
		}
		for _, man := range idx.Manifests {
			rl, err := VirusScan(ctx, man, repository, clamavChecksums, pushResults, cachePath)
			if err != nil {
				return nil, err
			}
			reports = append(reports, rl...)
		}
	} else {
		report := VirusScanManifestReport{}
		rl, err := scanManifestForViruses(ctx, desc, repository, cachePath)
		if err != nil {
			return nil, err
		}
		report.Results = rl
		report.ManifestDigest = desc.Digest.String()
		reports = append(reports, &report)
		if pushResults {
			// is this really the best spot to push the empty descriptor?
			cfgExists, err := repository.Exists(ctx, descCfg)
			if err != nil {
				return nil, fmt.Errorf("checking existence of config: %w", err)
			}
			if !cfgExists {
				imgcfg := ocispec.ImageConfig{}
				cfg, err := json.Marshal(imgcfg)
				if err != nil {
					return nil, fmt.Errorf("marshalling empty config: %w", err)
				}
				if err := repository.Push(ctx, descCfg, bytes.NewReader(cfg)); err != nil {
					return nil, fmt.Errorf("pushing empty config. Do you have push permissions? If not, use --check: %w", err)
				}
			}

			data, err := json.Marshal(clamavChecksums)
			if err != nil {
				return nil, fmt.Errorf("marshalling the virus database checksums: %w", err)
			}
			// attach the results report
			rd, err := attachResultsReport(ctx, desc, descCfg, reports, repository, map[string]string{AnnotationVirusDatabaseChecksum: string(data)}, ArtifactTypeVirusScanReport)
			if err != nil {
				return nil, err
			}
			for _, result := range rd {
				slog.InfoContext(ctx, "pushed results", "reference", result.Digest.String())
			}
		}
	}
	return reports, nil

}

// get the config if it's a image manifest
// if it's a bottle, we can pull filenames
// if it's an image, we need to derive the image source/architecture
// if it's a helm chart??? is it supported?
// git, pypi
// output any errors but do not fail until the end
func scanManifestForViruses(ctx context.Context,
	desc ocispec.Descriptor,
	repository oras.GraphTarget,
	cachePath string) ([]*VirusScanResults, error) {
	var clamavResults []*VirusScanResults

	// pull the image manifest
	rc, err := repository.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("fetching manifest: %w", err)
	}
	var manifest ocispec.Manifest
	decoder := json.NewDecoder(rc)
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("parsing the image manifest: %w", err)
	}

	// var config ocispec.ImageConfig
	// pull the config
	rcCfg, err := repository.Fetch(ctx, manifest.Config)
	if err != nil {
		return nil, fmt.Errorf("fetching manifest config: %w", err)
	}
	// decoder = json.NewDecoder(rcCfg)
	// if err := decoder.Decode(&config); err != nil {
	// 	return nil, fmt.Errorf("parsing the image manifest config: %w", err)
	// }

	switch manifest.Config.MediaType {
	case gitoci.MediaTypeSyncConfig:
		return clamavGitArtifact(ctx, rcCfg, cachePath, manifest.Layers, repository)
	case mediatype.MediaTypeBottleConfig:
		return clamavBottle(ctx, rcCfg, manifest.Layers, repository)
	default:
		for _, layer := range manifest.Layers {
			lrc, err := repository.Fetch(ctx, layer)
			if err != nil {
				return nil, fmt.Errorf("fetching the image layer: %w", err)
			}
			r, err := clamavBytes(ctx, lrc, "")
			if err != nil {
				return nil, err
			}

			if r != nil {
				r.LayerDigest = layer.Digest.String()
				clamavResults = append(clamavResults, r)
			}
		}
	}

	// pull each layer and pass through the virus scanner
	// add the config to the layers slice so that it also gets scanned
	// manifest.Layers = append(manifest.Layers, manifest.Config)
	// for _, layer := range manifest.Layers {
	// 	lrc, err := repository.Fetch(ctx, layer)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("fetching the image layer: %w", err)
	// 	}
	// 	// TODO CLEANUP
	// 	if layer.MediaType == gitoci.MediaTypeBundleLayer {
	// 		// download to cache
	// 		if err := store.Push(ctx, layer, lrc); err != nil {
	// 			if !strings.Contains(err.Error(), "already exists") {
	// 				return nil, fmt.Errorf("caching the git layer: %w", err)
	// 			}
	// 		}
	// 		r, err := clamavGitBundle(ctx, cachePath, filepath.Join(cachePath, "blobs", string(layer.Digest.Algorithm()), layer.Digest.Encoded()))
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		if r != nil {
	// 			r.LayerDigest = layer.Digest.String()
	// 			clamavResults = append(clamavResults, r)
	// 		}
	// 		continue
	// 	} else if layer.MediaType == mediatype.MediaTypeBottleConfig || layer.MediaType == mediatype.MediaTypeLayer {

	// 		return clamavBottle(ctx, lrc, manifest.Layers[:len(manifest.Layers)-1], repository)
	// 	} else {
	// 		r, err := clamavBytes(ctx, lrc, "")
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		if r != nil {
	// 			r.LayerDigest = layer.Digest.String()
	// 			clamavResults = append(clamavResults, r)
	// 		}
	// 	}
	// }

	return clamavResults, nil
}
