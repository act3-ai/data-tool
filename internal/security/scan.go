// Package security contains the logic for fetching artifact details, attaching SBOM's, etc. for ace-dt security scan.
package security

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"

	"github.com/act3-ai/data-tool/internal/ref"
	reg "github.com/act3-ai/data-tool/pkg/registry"
	"github.com/act3-ai/go-common/pkg/logger"
)

// ScanOptions defines the options needed to run the scan operation.
type ScanOptions struct {
	CachePath               string
	GatherArtifactReference string
	SaveReport              string
	SourceFile              string
	VulnerabilityLevel      string
	Output                  []string
	DryRun                  bool
	PushReport              bool
	ScanVirus               bool
	Targeter                reg.GraphTargeter
}

// ScanArtifacts will fetch the artifact details for each image in a source file or a mirror (gather) artifact.
// It will then generate SBOMs for the reference if dryRun is false, upload them to the target repository, and use them for scanning.
// If dryRun is set to true, the artifacts will be scanned by reference.
// It returns a slice of results (derived from grype's json results) for the artifacts.
func ScanArtifacts(ctx context.Context, opts ScanOptions, concurrency int) ([]*ArtifactDetails, int, error) {
	if opts.SourceFile == "" && opts.GatherArtifactReference == "" {
		return nil, 3, fmt.Errorf("either sourcefile or gather artifact must be chosen but not both")
	}
	return scan(ctx, opts, concurrency)
}

func scan(ctx context.Context, opts ScanOptions, concurrency int) ([]*ArtifactDetails, int, error) {

	log := logger.FromContext(ctx)
	mu := sync.Mutex{}
	var repository oras.GraphTarget

	// exitCode is set to 0 if clear, 2 if there are viruses found in the artifacts, and 3 for program errors.
	exitCode := 0
	repository, err := initializeRepository(ctx, opts)
	if err != nil {
		return nil, 3, err
	}

	grypeChecksumDB, clamavDBChecksums, err := initializeChecksums(ctx, opts)
	if err != nil {
		return nil, 3, err
	}

	m, err := FormatSources(ctx, opts.SourceFile, opts.GatherArtifactReference, repository, opts.Targeter, concurrency)
	if err != nil {
		return nil, 3, fmt.Errorf("extracting sources from artifact: %w", err)
	}
	results := make([]*ArtifactDetails, len(m))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for i, source := range m {
		g.Go(func() error {
			log.InfoContext(ctx, "Processing artifact", "reference", source.ArtifactReference, "originatingReference", source.OriginalReference)
			artifactDetails, err := processArtifact(gctx, source, opts, opts.Targeter, grypeChecksumDB, clamavDBChecksums, repository)
			if err != nil {
				return err
			}
			mu.Lock()
			results[i] = artifactDetails
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, 3, err
	}

	// filter out any nil values (caused by artifacts that are unsupported like helm charts and git bundle artifacts)
	filteredResults := []*ArtifactDetails{}
	for _, v := range results {
		if v != nil {
			filteredResults = append(filteredResults, v)
		}
	}
	return filteredResults, exitCode, nil

}

func initializeRepository(ctx context.Context, opts ScanOptions) (oras.GraphTarget, error) {
	if opts.GatherArtifactReference == "" {
		return nil, nil
	}
	return opts.Targeter.GraphTarget(ctx, opts.GatherArtifactReference)
}

func initializeChecksums(ctx context.Context, opts ScanOptions) (string, []ClamavDatabase, error) {
	grypeChecksumDB, err := getGrypeDBChecksum(ctx)
	if err != nil {
		return "", nil, err
	}

	var clamavDBChecksums []ClamavDatabase
	if opts.ScanVirus {
		cs, err := getClamAVChecksum(ctx)
		if err != nil {
			return "", nil, err
		}
		clamavDBChecksums = cs
	}
	return grypeChecksumDB, clamavDBChecksums, nil
}

func processArtifact(ctx context.Context,
	source Source, opts ScanOptions,
	targeter reg.GraphTargeter,
	grypeChecksumDB string,
	clamavDBChecksums []ClamavDatabase,
	repository oras.GraphTarget) (*ArtifactDetails, error) {
	log := logger.FromContext(ctx)

	if repository == nil || opts.SourceFile != "" {
		// repo, err := repoFunction(ctx, source.ArtifactReference)
		repo, err := targeter.GraphTarget(ctx, source.ArtifactReference)
		if err != nil {
			return nil, err
		}
		repository = repo
	}

	log.InfoContext(ctx, "fetching manifest details", "artifact", source.ArtifactReference, "originatingReference", source.OriginalReference)
	artifactDetails, err := GetArtifactDetails(ctx, source.ArtifactReference, repository)
	if err != nil {
		return nil, fmt.Errorf("getting artifact details for %s: %w", source.OriginalReference, err)
	}
	artifactDetails.originatingReference = source.OriginalReference
	artifactDetails.handlePredecessors(grypeChecksumDB, clamavDBChecksums)

	if opts.ScanVirus {
		if err := processVirusScanning(ctx, artifactDetails, source.OriginalReference, repository, clamavDBChecksums, opts.PushReport, opts.CachePath); err != nil {
			return nil, err
		}
	}

	res, err := processVulnerabilityScanning(ctx, artifactDetails, source, grypeChecksumDB, opts)
	if err != nil {
		return nil, err
	}

	calculatedResults, err := calculateResults(res)
	if err != nil {
		return nil, err
	}

	artifactDetails.CalculatedResults = *calculatedResults
	return artifactDetails, nil
}

func processVulnerabilityScanning(ctx context.Context, artifactDetails *ArtifactDetails, source Source,
	grypeChecksumDB string, opts ScanOptions) (*VulnerabilityScanResults, error) {
	log := logger.FromContext(ctx)

	res := VulnerabilityScanResults{
		Matches: []Matches{},
	}

	// this is for filtering
	matches := make(map[string]Matches)

	switch {
	case !opts.DryRun && artifactDetails.manifestDigestSBOM == "":
		log.InfoContext(ctx, "Generating SBOM(s)...", "reference", artifactDetails.originatingReference)
		grypeResults, err := GenerateSBOM(ctx, source.ArtifactReference, grypeChecksumDB, artifactDetails.repository, opts.PushReport)
		if err != nil {
			return nil, err
		}

		for _, r := range grypeResults {
			freshMatches, err := filterResults(r, opts.VulnerabilityLevel)
			if err != nil {
				return nil, err
			}
			for k, v := range freshMatches {
				matches[k] = v
			}
		}

	case artifactDetails.manifestDigestSBOM != "":
		log.Info("SBOM Manifest found", "reference", artifactDetails.originatingReference, "digest", artifactDetails.manifestDigestSBOM)
		grypeRes, err := extractAndGrypeSBOMs(ctx, artifactDetails.desc, artifactDetails.repository, artifactDetails.manifestDigestSBOM, grypeChecksumDB, opts.PushReport)
		if err != nil {
			return nil, err
		}
		// filter the matches (there will be duplicates for multi-architecture images)
		for _, r := range grypeRes {
			freshMatches, err := filterResults(&r, opts.VulnerabilityLevel)
			if err != nil {
				return nil, err
			}
			for k, v := range freshMatches {
				matches[k] = v
			}
		}

	default:
		// use the reference from the *remote.Repository created by getManifestDetails, ensuring our reference
		// contains the correct endpoint if it was changed
		result, err := grypeReference(ctx, source.ArtifactReference)
		if err != nil {
			return nil, fmt.Errorf("gryping reference %s: %w", source.OriginalReference, err)
		}
		res = *result
	}

	for _, v := range matches {
		res.Matches = append(res.Matches, v)
	}

	return &res, nil
}

func processVirusScanning(ctx context.Context,
	artifactDetails *ArtifactDetails,
	reference string,
	repository oras.GraphTarget,
	clamavDBChecksums []ClamavDatabase,
	pushReport bool,
	cachePath string) error {
	log := logger.FromContext(ctx)

	if artifactDetails.virusScanReport != nil {
		existingVirusScanReportManifest, err := artifactDetails.FetchExistingVirusScanningReportManifest(ctx)
		if err != nil {
			return err
		}
		log.InfoContext(ctx, "Found an existing and current virus scanning report", "reference", reference)
		blob := existingVirusScanReportManifest.Layers[0]

		var vr []*VirusScanManifestReport
		rc, err := artifactDetails.repository.Fetch(ctx, blob)
		if err != nil {
			return fmt.Errorf("fetching the results for %s: %w", reference, err)
		}
		decoder := json.NewDecoder(rc)
		if err := decoder.Decode(&vr); err != nil {
			return fmt.Errorf("decoding the scan report for %s: %w", reference, err)
		}
		artifactDetails.MalwareResults = vr
	} else {
		virusResults, err := VirusScan(ctx, artifactDetails.desc, repository, clamavDBChecksums, pushReport, cachePath)
		if err != nil {
			return fmt.Errorf("virus scanning for reference %s: %w", reference, err)
		}
		artifactDetails.MalwareResults = virusResults
	}
	return nil
}

func extractSourcesFromMirrorArtifact(ctx context.Context, reference registry.Reference, repo oras.GraphTarget) ([]Source, error) {
	sources := []Source{}
	// fetch the reference index
	_, data, err := oras.FetchBytes(ctx, repo, reference.String(), oras.DefaultFetchBytesOptions)
	if err != nil {
		return nil, fmt.Errorf("error fetching the artifact index: %w", err)
	}
	var idx ocispec.Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("unmarshalling artifact index: %w", err)
	}
	for _, manifest := range idx.Manifests {
		// create a source
		sources = append(sources, Source{manifest.Annotations[ref.AnnotationSrcRef], strings.Join([]string{reference.String(), manifest.Digest.String()}, "@")})
	}
	return sources, nil
}

// FetchExistingResultsReportManifest fetches the artifact's existing scan report.
func (ad *ArtifactDetails) FetchExistingResultsReportManifest(ctx context.Context) (*ocispec.Manifest, error) {
	report := ocispec.Manifest{}
	rc, err := ad.repository.Fetch(ctx, *ad.resultsReport)
	if err != nil {
		return nil, fmt.Errorf("fetching existing results for %s: %w", ad.originatingReference, err)
	}
	decoder := json.NewDecoder(rc)
	if err := decoder.Decode(&report); err != nil {
		return nil, fmt.Errorf("decoding results report for %s: %w", ad.originatingReference, err)
	}
	return &report, nil
}

// FetchExistingVirusScanningReportManifest fetches the artifact's existing scan report.
func (ad *ArtifactDetails) FetchExistingVirusScanningReportManifest(ctx context.Context) (*ocispec.Manifest, error) {
	report := ocispec.Manifest{}
	rc, err := ad.repository.Fetch(ctx, *ad.virusScanReport)
	if err != nil {
		return nil, fmt.Errorf("fetching existing results for %s: %w", ad.originatingReference, err)
	}
	decoder := json.NewDecoder(rc)
	if err := decoder.Decode(&report); err != nil {
		return nil, fmt.Errorf("decoding results report for %s: %w", ad.originatingReference, err)
	}
	return &report, nil
}
