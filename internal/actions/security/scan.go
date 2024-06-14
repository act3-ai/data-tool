package security

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"

	"github.com/dustin/go-humanize"
	"golang.org/x/sync/errgroup"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	security "gitlab.com/act3-ai/asce/data/tool/internal/security"
)

// Scan represents the scan action.
type Scan struct {
	*Action
	SourceFile              string
	GatherArtifactReference string
	Output                  string
	DryRun                  bool
}

// ScanningResults represents the list of security results and contains a mutex for concurrent writes.
type ScanningResults struct {
	results []*ArtifactScanResults
	mu      sync.Mutex
}

// ArtifactScanResults formats the artifact's pertinent grype JSON results for printing.
type ArtifactScanResults struct {
	// Results           *Results `json:"results"`
	Reference          string   `json:"reference"`
	CriticalVulnCount  int      `json:"critical_vulnerabilities"`
	HighVulnCount      int      `json:"highVulnerabilites"`
	MediumVulnCount    int      `json:"mediumVulnerabilities"`
	Platforms          []string `json:"platforms"`
	OciCompliance      bool     `json:"ociCompliant"`
	Size               string   `json:"size"`
	IsSigned           bool     `json:"signed"`
	SignatureReference string   `json:"signatureReference,omitempty"`
	HasSBOM            bool     `json:"SBOM"`
	SBOMReference      string   `json:"SBOMReference,omitempty"`
	ShortenedName      string
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

// Run executes the security scan Run() action.
func (action *Scan) Run(ctx context.Context, out io.Writer) error {
	cfg := action.Config.Get(ctx)
	log := logger.FromContext(ctx)

	// iterate through artifactDetails in sourceFile or in a gathered object!
	artifactDetails, err := security.ResolveScanReferences(ctx, action.SourceFile, action.GatherArtifactReference, action.Config.Repository, cfg.ConcurrentHTTP, action.DryRun)
	if err != nil {
		return err
	}
	noArtifacts := strconv.Itoa(len(artifactDetails))
	log.InfoContext(ctx, "Resolved references", noArtifacts, "")

	results, err := scanArtifacts(ctx, artifactDetails, cfg.ConcurrentHTTP)
	if err != nil {
		return err
	}

	// return some nicely formatted data, vulnerabilities, total (deduplicated) size, platforms, oci compliance
	// TODO maybe make this work like scatter mapping functions?
	switch action.Output {
	case "json":
		if err := printJSON(out, results); err != nil {
			return err
		}
	case "markdown":
		if err := printMarkdown(out, results); err != nil {
			return err
		}
	case "csv":
		if err := printCSV(out, results); err != nil {
			return err
		}
	case "table":
		if err := printTable(out, results); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown printing directive: %s", action.Output)
	}

	return nil
}

func grypeReference(ctx context.Context, reference string) (*Results, error) {
	vulnerabilities := Results{}
	cmd := exec.CommandContext(ctx, "grype", reference, "-o", "json")
	res, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing command: %s\n%w\n output: %s", cmd, err, string(res))
	}
	if err := json.Unmarshal(res, &vulnerabilities); err != nil {
		return nil, fmt.Errorf("parsing vulnerabilities: %w", err)
	}
	return &vulnerabilities, nil
}

func grypeSBOM(ctx context.Context, sbom []byte) (*Results, error) {
	vulnerabilities := Results{}
	cmd := exec.CommandContext(ctx, "grype", "-o", "json")
	cmd.Stdin = bytes.NewReader(sbom)
	res, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing command: %s\n %w\n output: %s", cmd, err, string(res))
	}
	if err := json.Unmarshal(res, &vulnerabilities); err != nil {
		return nil, fmt.Errorf("parsing vulnerabilities: %w", err)
	}

	return &vulnerabilities, nil
}

func calculateResults(results *Results) (*ArtifactScanResults, error) {
	var securityResults ArtifactScanResults
	// count crits, high, medium and add to results
	for _, res := range results.Matches {
		switch res.Vulnerabilities.Severity {
		case "Critical":
			securityResults.CriticalVulnCount++
		case "High":
			securityResults.HighVulnCount++
		case "Medium":
			securityResults.MediumVulnCount++
		default:
			// filter out low/negligible/unknown
			continue
		}
	}

	return &securityResults, nil
}

func scanArtifacts(ctx context.Context, artifactDetails []security.ArtifactDetails, concurrentHTTP int) ([]*ArtifactScanResults, error) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrentHTTP)
	scanned := ScanningResults{
		results: []*ArtifactScanResults{},
		mu:      sync.Mutex{},
	}
	for _, detail := range artifactDetails {
		g.Go(func() error {
			var res *Results
			reference := formatReference(detail)
			log := logger.FromContext(ctx)
			// if an SBOM exists, grype that instead, need to pull the blob though
			if len(detail.SBOM) != 0 {
				log.InfoContext(ctx, "SBOM digest for reference", reference, detail.SBOMDigest)
				for _, v := range detail.SBOM {
					var results *Results
					results, err := grypeSBOM(ctx, v)
					if err != nil {
						// fallback to reference
						r, err := grypeReference(ctx, reference)
						if err != nil {
							return err
						}
						res = r
					} else {
						res = results
					}
				}
			} else {
				log.InfoContext(ctx, "Scanning by reference", reference, detail.Source.Name)
				results, err := grypeReference(ctx, detail.Source.Name)
				if err != nil {
					return err
				}
				res = results
			}
			partialResults, err := calculateResults(res)
			if err != nil {
				return fmt.Errorf("counting vulnerabilities: %w", err)
			}
			// this is the only thing that changes with a gather repo!!!
			partialResults.Reference = reference
			// add total size
			partialResults.Size = humanize.Bytes(uint64(detail.Size))

			// add platforms
			partialResults.Platforms = detail.Platforms

			// add oci compliance
			partialResults.OciCompliance = detail.IsOCICompliant

			// add sbom and signature details
			if len(detail.SBOM) != 0 {
				partialResults.SBOMReference = detail.SBOMDigest
				partialResults.HasSBOM = true
			}
			if partialResults.SignatureReference != "" {
				partialResults.SignatureReference = detail.SignatureDigest
				partialResults.IsSigned = true
			}

			scanned.mu.Lock()
			scanned.results = append(scanned.results, partialResults)
			scanned.mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return scanned.results, nil
}
