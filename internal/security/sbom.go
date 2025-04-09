package security

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	notationreg "github.com/notaryproject/notation-go/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"

	"github.com/act3-ai/go-common/pkg/logger"

	"github.com/act3-ai/data-tool/internal/mirror/encoding"
)

func extractAndGrypeSBOMs(ctx context.Context, subjectDescriptor ocispec.Descriptor, target oras.GraphTarget, digestSBOM, grypeDBChecksum string, pushReport bool) ([]Results, error) {
	log := logger.FromContext(ctx)
	results := []Results{}
	// try and extract sbom
	_, manifestBytesSBOM, err := oras.FetchBytes(ctx, target, digestSBOM, oras.DefaultFetchBytesOptions)
	if err != nil {
		return nil, fmt.Errorf("fetching SBOM manifest bytes for %s: %w", digestSBOM, err)
	}
	var man ocispec.Manifest
	if err := json.Unmarshal(manifestBytesSBOM, &man); err != nil {
		return nil, fmt.Errorf("extracting SBOM manifest bytes for %s: %w", digestSBOM, err)
	}
	for _, l := range man.Layers {
		rc, err := target.Fetch(ctx, l)
		if err != nil {
			return nil, fmt.Errorf("fetching layer for %s: %w", digestSBOM, err)
		}
		res, err := grypeSBOM(ctx, rc)
		if err != nil {
			return nil, err
		}
		if err = rc.Close(); err != nil {
			return nil, fmt.Errorf("closing the reader: %w", err)
		}
		descCfg, err := generateEmptyBlobDescriptor(ocispec.MediaTypeImageConfig, digest.Canonical)
		if err != nil {
			return nil, err
		}
		calculatedResults, err := calculateResults(res)
		if err != nil {
			return nil, err
		}
		if pushReport {
			reports, err := attachCVEResults(ctx, subjectDescriptor, descCfg, *calculatedResults, target, grypeDBChecksum)
			if err != nil {
				return nil, fmt.Errorf("failed to attach results: %w\n Do you have push permissions to the repository?", err)
			}
			for _, report := range reports {
				log.InfoContext(ctx, "PUSHED REPORT", "digest", report.Digest.String())
			}
		}
		results = append(results, *res)
	}
	return results, nil
}

// GenerateSBOM will generate and attach an SBOM for a given artifact.
// It will grype the SBOM inline and return a map of the SBOM descriptor and results.
func GenerateSBOM( //nolint:gocognit
	ctx context.Context,
	reference,
	grypeDBChecksum string,
	repository oras.GraphTarget,
	pushReport bool) (map[*ocispec.Descriptor]*Results, error) {
	results := map[*ocispec.Descriptor]*Results{}
	log := logger.FromContext(ctx)
	// fetch the main descriptor
	desc, err := repository.Resolve(ctx, reference)
	if err != nil {
		return nil, fmt.Errorf("error resolving descriptor for %s: %w", reference, err)
	}

	descCfg, err := generateEmptyBlobDescriptor(ocispec.MediaTypeImageConfig, digest.Canonical)
	if err != nil {
		return nil, err
	}
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

	if encoding.IsIndex(desc.MediaType) {
		var idx ocispec.Index
		b, err := content.FetchAll(ctx, repository, desc)
		if err != nil {
			return nil, fmt.Errorf("fetching index: %w", err)
		}
		if err := json.Unmarshal(b, &idx); err != nil {
			return nil, fmt.Errorf("decoding index: %w", err)
		}
		parsedRef, err := registry.ParseReference(reference)
		if err != nil {
			return nil, fmt.Errorf("parsing index reference: %w", err)
		}
		// get the manifests
		for _, man := range idx.Manifests {
			if man.ArtifactType == notationreg.ArtifactTypeNotation || man.ArtifactType == ArtifactTypeSPDX || man.ArtifactType == "application/vnd.in-toto+json" {
				continue
			}
			refMan := registry.Reference{
				Registry:   parsedRef.Registry,
				Repository: parsedRef.Repository,
				Reference:  man.Digest.String(),
			}

			res, err := GenerateSBOM(ctx, refMan.String(), grypeDBChecksum, repository, pushReport)
			if err != nil {
				return nil, err
			}
			for k, v := range res {
				results[k] = v
			}
		}
	} else {
		// exec out to syft to generate the SBOM
		res, err := syftReference(ctx, reference)
		if err != nil {
			// syft regularly fails if passed a signature, helmfile, or bottle manifest. Ignore but log the error.
			log.InfoContext(ctx, "failed SBOM generation", "reference", reference, "error", err)
			return nil, nil
		}

		// Upload the SBOM... create a manifest and encode SBOM into a layer.
		// The SBOM manifest Subject field must point to the digest of the main reference passed.
		descSBOM, err := oras.PushBytes(ctx, repository, ocispec.MediaTypeImageLayer, res)
		if err != nil {
			return nil, fmt.Errorf("pushing SBOM: %w", err)
		}

		packOpts := oras.PackManifestOptions{
			Subject:          &desc,
			Layers:           []ocispec.Descriptor{descSBOM},
			ConfigDescriptor: &descCfg,
		}

		log.InfoContext(ctx, "pushing SBOM", "reference", reference)
		maniDesc, err := oras.PackManifest(ctx, repository, oras.PackManifestVersion1_1, ArtifactTypeSPDX, packOpts)
		if err != nil {
			return nil, fmt.Errorf("pushing SBOM manifest: %w", err)
		}
		log.InfoContext(ctx, "SBOM pushed", "reference", reference, "manifest", maniDesc.Digest.String())

		// grype the SBOM and attach the results
		grypeResults, err := grypeSBOM(ctx, io.NopCloser(bytes.NewReader(res)))
		if err != nil {
			return nil, err
		}
		log.InfoContext(ctx, "successfully pushed SBOM", "reference", reference)
		results[&desc] = grypeResults
		calculatedResults, err := calculateResults(grypeResults)
		if err != nil {
			return nil, err
		}
		if pushReport {
			reports, err := attachCVEResults(ctx, desc, descCfg, *calculatedResults, repository, grypeDBChecksum)
			if err != nil {
				return nil, err
			}
			for _, report := range reports {
				log.InfoContext(ctx, "CVE results report", "reference", reference, "digest", report.Digest.String())
			}
		}
	}
	return results, nil
}

// IsSBOM is a helper function that identifies whether the given artifact type belongs to an ASCE or Harbor SBOM.
func IsSBOM(artifactType string) bool {
	if artifactType == ArtifactTypeSPDX || artifactType == ArtifactTypeHarborSBOM {
		return true
	}
	return false
}
