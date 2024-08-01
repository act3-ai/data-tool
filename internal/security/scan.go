// Package security contains the logic for fetching artifact details, attaching SBOM's, etc. for ace-dt security scan.
package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"

	notationreg "github.com/notaryproject/notation-go/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	"git.act3-ace.com/ace/data/schema/pkg/selectors"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror/encoding"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
)

const (
	// ArtifactTypeSPDX is the standard artifact type for SPDX formatted SBOMs.
	ArtifactTypeSPDX = "application/spdx+json"
	// ArtifactTypeHarborSBOM is the artifact type given to SBOMs generated via the Harbor UI.
	ArtifactTypeHarborSBOM = "application/vnd.goharbor.harbor.sbom.v1"
)

// ArtifactDetails contains all of the details needed for a given artifact.
type ArtifactDetails struct {
	repository           *remote.Repository
	Source               mirror.Source
	Platforms            []string
	Size                 int64
	IsOCICompliant       bool
	Predecessors         []ocispec.Descriptor
	SignatureDigest      string
	SBOMDigest           string
	OriginatingReference string              // only needed for gather artifacts
	shortenedName        string              // needed for graphing in mermaid
	CalculatedResults    ArtifactScanResults `json:"results"`
}

// Results holds the vulnerability data for all given artifacts.
type Results struct {
	Matches []Matches `json:"matches"`
}

// Matches represents the vulnerability matches and details for a given artifact.
type Matches struct {
	Vulnerabilities Vulnerability `json:"vulnerability"`
	Artifact        Artifact      `json:"artifact"`
}

// Vulnerability represents a specific vulnerability for a given artifact.
type Vulnerability struct {
	ID          string `json:"id"`
	Source      string `json:"dataSource"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// Artifact represents the identifying details for a given artifact.
type Artifact struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ScanArtifacts will fetch the artifact details for each image in a source file or a mirror (gather) artifact.
// It will then generate SBOMs for the reference if dryRun is false, upload them to the target repository, and use them for scanning.
// If dryRun is set to true, the artifacts will be scanned by reference.
// It returns a slice of results (derived from grype's json results) for the artifacts.
func ScanArtifacts(ctx context.Context,
	sourceFile, artifactReference string,
	repoFunction func(context.Context, string) (*remote.Repository, error),
	concurrency int,
	dryRun bool) ([]*ArtifactDetails, error) {

	var results []*ArtifactDetails
	switch {
	case sourceFile != "":
		scannedResults, err := scanFromSourceFile(ctx, sourceFile, concurrency, repoFunction)
		if err != nil {
			return nil, err
		}
		results = scannedResults.results

	case artifactReference != "":
		scannedResults, err := scanFromMirrorArtifact(ctx, artifactReference, concurrency, repoFunction, dryRun)
		if err != nil {
			return nil, err
		}
		results = scannedResults.results

	default:
		return nil, fmt.Errorf("must pass either a source file or a remote reference to a Gather artifact")

	}
	return results, nil
}

func scanFromSourceFile(ctx context.Context, sourceFile string, concurrency int, repoFunction func(context.Context, string) (*remote.Repository, error)) (*ScanningResults, error) {
	scanned := ScanningResults{
		results: []*ArtifactDetails{},
		mu:      sync.Mutex{},
	}
	sources, err := mirror.ProcessSourcesFile(ctx, sourceFile, selectors.LabelSelectorSet{}, concurrency)
	if err != nil {
		return nil, err
	}
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)
	for _, source := range sources {
		g.Go(func() error {
			// dryRun is manually overridden (set to true) for source-file input so we don't try to send SBOM's to directories in which we cannot push.
			sourceDetails, err := getManifestDetails(gctx, source.Name, repoFunction, true)
			if err != nil {
				return err
			}
			// the originating reference is the source name for a sourcefile reference, but we still assign it for simplicity.
			sourceDetails.OriginatingReference = source.Name
			sourceDetails.handlePredecessors()
			res, err := grypeReference(ctx, source.Name)
			if err != nil {
				return fmt.Errorf("gryping by reference for %s: %w", source.Name, err)
			}
			// analyze the results
			calculatedResults, err := calculateResults(res)
			if err != nil {
				return fmt.Errorf("counting vulnerabilities for %s: %w", source.Name, err)
			}
			// add artifact detail to the formatted results
			sourceDetails.CalculatedResults = *calculatedResults
			scanned.mu.Lock()
			scanned.results = append(scanned.results, sourceDetails)
			scanned.mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return &scanned, nil
}

func scanFromMirrorArtifact(ctx context.Context, //nolint:gocognit
	artifactReference string,
	concurrency int,
	repoFunction func(context.Context, string) (*remote.Repository, error),
	dryRun bool) (*ScanningResults, error) {

	scanned := ScanningResults{
		results: []*ArtifactDetails{},
		mu:      sync.Mutex{},
	}

	m, err := extractSourcesFromMirrorArtifact(ctx, artifactReference, repoFunction)
	if err != nil {
		return nil, fmt.Errorf("extracting sources from artifact: %w", err)
	}
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)
	for s, source := range m {
		// var results []Results
		g.Go(func() error {
			res := Results{
				Matches: []Matches{},
			}
			matches := make(map[string]Matches)
			artifactDetails, err := getManifestDetails(gctx, source.Name, repoFunction, dryRun)
			if err != nil {
				return fmt.Errorf("getting artifact details for %s: %w", s, err)
			}
			artifactDetails.OriginatingReference = s
			// load the predecessor digests
			artifactDetails.handlePredecessors()
			// if this is not a dry run
			switch {
			case !dryRun && artifactDetails.SBOMDigest == "":
				sboms, err := GenerateSBOM(ctx, source.Name, artifactDetails.repository)
				if err != nil {
					return err
				}
				for _, digest := range sboms {
					grypeResult, err := extractAndGrypeSBOMs(gctx, artifactDetails.repository, digest)
					if err != nil {
						return err
					}
					// filter the matches (there will be duplicates for multi-architecture images)
					for _, result := range grypeResult {
						for _, match := range result.Matches {
							matches[match.Vulnerabilities.ID] = match
						}
					}
				}

			case !dryRun && artifactDetails.SBOMDigest != "":
				grypeRes, err := extractAndGrypeSBOMs(gctx, artifactDetails.repository, artifactDetails.SBOMDigest)
				if err != nil {
					return err
				}
				// filter the matches (there will be duplicates for multi-architecture images)
				for _, result := range grypeRes {
					for _, match := range result.Matches {
						matches[match.Vulnerabilities.ID] = match
					}
				}

			default:
				result, err := grypeReference(gctx, source.Name)
				if err != nil {
					return fmt.Errorf("gryping reference %s: %w", source.Name, err)
				}
				res = *result
			}

			for _, v := range matches {
				res.Matches = append(res.Matches, v)
			}

			calculatedResults, err := calculateResults(&res)
			if err != nil {
				return err
			}
			// add results to struct
			artifactDetails.CalculatedResults = *calculatedResults
			scanned.mu.Lock()
			scanned.results = append(scanned.results, artifactDetails)
			scanned.mu.Unlock()
			return nil
		})
	}
	return &scanned, nil
}

func extractSourcesFromMirrorArtifact(ctx context.Context, reference string, repoFunction func(context.Context, string) (*remote.Repository, error)) (map[string]mirror.Source, error) {
	sources := make(map[string]mirror.Source)
	repo, err := repoFunction(ctx, reference)
	if err != nil {
		return nil, err
	}
	// fetch the reference index
	_, data, err := oras.FetchBytes(ctx, repo, repo.Reference.ReferenceOrDefault(), oras.DefaultFetchBytesOptions)
	if err != nil {
		return nil, fmt.Errorf("error fetching the artifact index: %w", err)
	}
	var idx ocispec.Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("unmarshalling artifact index: %w", err)
	}

	for _, manifest := range idx.Manifests {
		// create a source
		sources[manifest.Annotations[ref.AnnotationSrcRef]] = mirror.Source{
			Name:   strings.Join([]string{reference, manifest.Digest.String()}, "@"),
			Labels: nil,
		}
	}
	return sources, nil
}

func getManifestDetails(ctx context.Context, reference string, repoFunction func(context.Context, string) (*remote.Repository, error), dryRun bool) (*ArtifactDetails, error) {
	maniDetails := &ArtifactDetails{}
	maniDetails.Source.Name = reference
	var platforms []string
	var size int64
	var predecessors []ocispec.Descriptor
	r, err := registry.ParseReference(reference)
	if err != nil {
		return maniDetails, fmt.Errorf("error parsing ref %s: %w", reference, err)
	}
	repo, err := repoFunction(ctx, reference)
	if err != nil {
		return maniDetails, err
	}
	maniDetails.repository = repo
	d, data, err := oras.FetchBytes(ctx, repo, r.ReferenceOrDefault(), oras.DefaultFetchBytesOptions)
	if err != nil {
		return maniDetails, fmt.Errorf("fetching the manifest bytes: %w", err)
	}
	maniDetails.IsOCICompliant = encoding.IsOCICompliant(d.MediaType)
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
			if err := json.Unmarshal(manData, &img); err != nil {
				return nil, err
			}
			size += manDesc.Size
			for _, blob := range img.Layers {
				size += blob.Size
			}
			p, err := repo.Predecessors(ctx, man)
			if err != nil {
				return maniDetails, fmt.Errorf("fetching predecessors: %w", err)
			}
			// Fetch predecessors/SBOMs/Signatures
			predecessors = append(predecessors, p...)
			platformString := formatPlatformString(man.Platform)
			platforms = append(platforms, platformString)
		}
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
			_, rc, err := repo.Blobs().FetchReference(ctx, img.Config.Digest.String())
			if err != nil {
				return nil, fmt.Errorf("fetching reference %s: %w", img.Config.Digest.String(), err)
			}
			var platform ocispec.Platform
			if err = json.NewDecoder(rc).Decode(&platform); err != nil && errors.Is(io.EOF, err) {
				return nil, fmt.Errorf("decoding platform %+v: %w", &platform, err)
			}
			platformString := formatPlatformString(&platform)
			platforms = append(platforms, platformString)
		}
	}
	maniDetails.Platforms = platforms
	maniDetails.Size = size
	maniDetails.Predecessors = predecessors

	// if err := maniDetails.handlePredecessors(ctx, dryRun); err != nil {
	// 	return nil, err
	// }
	return maniDetails, nil
}

func formatPlatformString(platform *ocispec.Platform) string {
	if platform != nil {
		return path.Join(platform.OS, platform.Architecture, platform.Variant)
	}
	return ""
}

func (ad *ArtifactDetails) handlePredecessors() {
	for _, p := range ad.Predecessors {
		switch p.ArtifactType {
		case notationreg.ArtifactTypeNotation:
			ad.SignatureDigest = p.Digest.String()
		case ArtifactTypeSPDX:
			ad.SBOMDigest = p.Digest.String()
		default:
			continue
		}
	}
}

func extractAndGrypeSBOMs(ctx context.Context, target oras.Target, digestSBOM string) ([]Results, error) {
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
		results = append(results, *res)
	}
	return results, nil
}

// GenerateSBOM will generate and attach an SBOM for a given artifact.
// Returns the digest of the SBOM and the bytes or an error.
func GenerateSBOM(ctx context.Context, reference string, repository oras.GraphTarget) ([]string, error) {
	sboms := []string{}
	log := logger.FromContext(ctx)
	// fetch the main descriptor
	desc, err := repository.Resolve(ctx, reference)
	if err != nil {
		return nil, fmt.Errorf("error resolving descriptor for %s: %w", reference, err)
	}

	// TODO does an empty blob exist? If so we can skip this entirely.
	// create an empty config and marshal it
	config := ocispec.ImageConfig{}
	cfg, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshalling empty config: %w", err)
	}

	// push the empty config
	descCfg, err := oras.PushBytes(ctx, repository, ocispec.MediaTypeEmptyJSON, cfg)
	if err != nil {
		return nil, fmt.Errorf("pushing empty config: %w", err)
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

			entry, err := GenerateSBOM(ctx, refMan.String(), repository)
			if err != nil {
				// ignore the error for now but log it out
				// TODO: handle this better- usually fails on in-toto signature manifests and bottles, find a way to filter without having to download and interate through those manifests.
				log.ErrorContext(ctx, "failed SBOM generation", "reference", refMan.String(), "error", err)
				return nil, nil
			}
			sboms = append(sboms, entry...)
		}
	} else {
		// exec out to syft to generate the SBOM
		res, err := syftReference(ctx, reference)
		if err != nil {
			return nil, err
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
		log.InfoContext(ctx, "successfully pushed SBOM", "reference", reference)
		sboms = append(sboms, maniDesc.Digest.String())
	}
	return sboms, nil
}
