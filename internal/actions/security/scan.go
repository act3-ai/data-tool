package security

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	security "github.com/act3-ai/data-tool/internal/security"
	"github.com/act3-ai/go-common/pkg/logger"
)

// Scan represents the scan action.
type Scan struct {
	*Action
	SourceFile              string
	GatherArtifactReference string
	Output                  []string
	VulnerabilityLevel      string
	DryRun                  bool
	DisplayCVE              bool
	DisplayPlatforms        bool
	PushReport              bool
	ScanVirus               bool
}

// Run executes the security scan Run() action.
func (action *Scan) Run(ctx context.Context) (int, error) {
	cfg := action.Config.Get(ctx)
	log := logger.FromContext(ctx)

	// Build the scan options
	opts := security.ScanOptions{
		CachePath:               cfg.CachePath,
		SourceFile:              action.SourceFile,
		GatherArtifactReference: action.GatherArtifactReference,
		Output:                  action.Output,
		VulnerabilityLevel:      action.VulnerabilityLevel,
		DryRun:                  action.DryRun,
		PushReport:              action.PushReport,
		ScanVirus:               action.ScanVirus,
		Targeter:                action.Config,
	}

	log.InfoContext(ctx, "Scanning Artifacts...")
	// iterate through artifactDetails in sourceFile or in a gathered object
	results, exitCode, err := security.ScanArtifacts(ctx, opts, cfg.ConcurrentHTTP)
	if err != nil {
		return 3, err
	}

	if len(results) == 0 {
		_, err := fmt.Fprintf(os.Stdout, "No supported images were found to be scanned.\n")
		if err != nil {
			return 3, fmt.Errorf("printing to standard out that no scan-supported images were found: %w", err)
		}
		return 3, nil
	}

	outputMethods := map[string][]io.Writer{}
	// parse the output
	for _, o := range action.Output {
		var outfile io.Writer
		output := strings.Split(o, "=")
		if len(output) < 2 {
			// default to std out
			outfile = os.Stdout
		} else {
			outfile, err = os.OpenFile(output[1], os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				return 3, fmt.Errorf("creating/opening output file: %w", err)
			}
		}
		outputMethods[output[0]] = append(outputMethods[output[0]], outfile)
	}
	for method, writers := range outputMethods {
		// for each writer match to proper
		for _, writer := range writers {
			// return some nicely formatted data, vulnerabilities, total (deduplicated) size, platforms, oci compliance
			switch method {
			case "json":
				if err := security.PrintJSON(writer, results); err != nil {
					return 3, err
				}
			case "markdown":
				if err := security.PrintMarkdown(writer, results, action.VulnerabilityLevel); err != nil {
					return 3, err
				}
			case "csv":
				if err := security.PrintCSV(writer, results, action.VulnerabilityLevel); err != nil {
					return 3, err
				}
			case "table":
				if err := security.PrintTable(writer, results, action.VulnerabilityLevel, action.DisplayCVE, action.DisplayPlatforms, action.ScanVirus); err != nil {
					return 3, err
				}
			default:
				return 3, fmt.Errorf("unknown printing directive: %s", action.Output)
			}
		}
	}
	// if the virus scanning results are not empty, we need to send a non-0/ non-1 exit code.
	return exitCode, nil
}
