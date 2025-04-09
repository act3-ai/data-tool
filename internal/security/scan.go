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
	"oras.land/oras-go/v2/registry/remote"

	"github.com/act3-ai/data-tool/internal/ref"
	"github.com/act3-ai/go-common/pkg/logger"
)

// ScanOptions defines the options needed to run the scan operation.
type ScanOptions struct {
	SourceFile              string
	GatherArtifactReference string
	Output                  []string
	SaveReport              string
	VulnerabilityLevel      string
	DryRun                  bool
	PushReport              bool
}

// ScanArtifacts will fetch the artifact details for each image in a source file or a mirror (gather) artifact.
// It will then generate SBOMs for the reference if dryRun is false, upload them to the target repository, and use them for scanning.
// If dryRun is set to true, the artifacts will be scanned by reference.
// It returns a slice of results (derived from grype's json results) for the artifacts.
func ScanArtifacts(ctx context.Context,
	opts ScanOptions,
	repoFunction func(context.Context, string) (*remote.Repository, error),
	concurrency int) ([]*ArtifactDetails, error) {

	if opts.SourceFile == "" && opts.GatherArtifactReference == "" {
		return nil, fmt.Errorf("either sourcefile or gather artifact must be chosen but not both")
	}
	return scan(ctx, opts, repoFunction, concurrency)
}

func scan(ctx context.Context, //nolint:gocognit
	opts ScanOptions,
	repoFunction func(context.Context, string) (*remote.Repository, error),
	concurrency int) ([]*ArtifactDetails, error) {

	log := logger.FromContext(ctx)
	mu := sync.Mutex{}
	var repository *remote.Repository

	if opts.GatherArtifactReference != "" {
		repo, err := repoFunction(ctx, opts.GatherArtifactReference)
		if err != nil {
			return nil, err
		}
		repository = repo
	}
	m, err := FormatSources(ctx, opts.SourceFile, opts.GatherArtifactReference, repository, concurrency)
	if err != nil {
		return nil, fmt.Errorf("extracting sources from artifact: %w", err)
	}
	results := make([]*ArtifactDetails, len(m))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	// get the grype db checksum
	checksum, err := getGrypeDBChecksum(ctx)
	if err != nil {
		return nil, err
	}

	for i, source := range m {
		g.Go(func() error {

			res := Results{
				Matches: []Matches{},
			}

			if repository == nil || opts.SourceFile != "" {
				repo, err := repoFunction(ctx, source[1])
				if err != nil {
					return err
				}
				repository = repo
			}
			// this is for filtering
			matches := make(map[string]Matches)
			log.InfoContext(ctx, "fetching manifest details", "artifact", source[1], "originatingReference", source[0])
			artifactDetails, err := GetArtifactDetails(gctx, source[1], repository)
			if err != nil {
				return fmt.Errorf("getting artifact details for %s: %w", source[0], err)
			}

			// skip helm charts and git artifacts
			if artifactDetails.isNotScanSupported {
				return nil
			}

			artifactDetails.originatingReference = source[0]
			// load the predecessor digests
			artifactDetails.handlePredecessors(checksum)
			if artifactDetails.resultsReport != nil {
				// there is an existing report, we should pull it and compare checksums
				existingReportManifest, err := artifactDetails.FetchExistingResultsReportManifest(ctx)
				if err != nil {
					return err
				}
				// we can skip scanning and just return those results
				// fetch the blob and unmarshal it into an ArtifactScanReport
				// there should only be one blob
				log.InfoContext(ctx, "Found an existing and current results report", "reference", artifactDetails.originatingReference)
				blob := existingReportManifest.Layers[0]

				var scanReport ArtifactScanReport
				rc, err := artifactDetails.repository.Fetch(ctx, blob)
				if err != nil {
					return fmt.Errorf("fetching the results for %s: %w", artifactDetails.originatingReference, err)
				}
				decoder := json.NewDecoder(rc)
				if err := decoder.Decode(&scanReport); err != nil {
					return fmt.Errorf("decoding the scan report for %s: %w", artifactDetails.originatingReference, err)
				}
				// return scan report
				artifactDetails.CalculatedResults = scanReport
				mu.Lock()
				results[i] = artifactDetails
				mu.Unlock()
				return nil

			}

			switch {
			case !opts.DryRun && artifactDetails.manifestDigestSBOM == "":
				log.InfoContext(ctx, "Generating SBOM(s)...", "reference", artifactDetails.originatingReference)
				grypeResults, err := GenerateSBOM(ctx, source[1], checksum, artifactDetails.repository, opts.PushReport)
				if err != nil {
					return err
				}

				for _, r := range grypeResults {
					freshMatches, err := filterResults(r, opts.VulnerabilityLevel)
					if err != nil {
						return err
					}
					for k, v := range freshMatches {
						matches[k] = v
					}
				}

			case artifactDetails.manifestDigestSBOM != "":
				log.Info("SBOM Manifest found", "reference", artifactDetails.originatingReference, "digest", artifactDetails.manifestDigestSBOM)
				grypeRes, err := extractAndGrypeSBOMs(gctx, artifactDetails.desc, artifactDetails.repository, artifactDetails.manifestDigestSBOM, checksum, opts.PushReport)
				if err != nil {
					return err
				}
				// filter the matches (there will be duplicates for multi-architecture images)
				for _, r := range grypeRes {
					freshMatches, err := filterResults(&r, opts.VulnerabilityLevel)
					if err != nil {
						return err
					}
					for k, v := range freshMatches {
						matches[k] = v
					}
				}

			default:
				// use the reference from the *remote.Repository created by getManifestDetails, ensuring our reference
				// contains the correct endpoint if it was changed
				result, err := grypeReference(gctx, source[1])
				if err != nil {
					return fmt.Errorf("gryping reference %s: %w", source[0], err)
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
			mu.Lock()
			results[i] = artifactDetails
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// filter out any nil values (caused by artifacts that are unsupported like helm charts and git bundle artifacts)
	filteredResults := []*ArtifactDetails{}
	for _, v := range results {
		if v != nil {
			filteredResults = append(filteredResults, v)
		}
	}
	return filteredResults, nil

}

func extractSourcesFromMirrorArtifact(ctx context.Context, reference string, repo *remote.Repository) ([][]string, error) {
	sources := [][]string{}
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
		sources = append(sources, []string{manifest.Annotations[ref.AnnotationSrcRef], strings.Join([]string{reference, manifest.Digest.String()}, "@")})
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
