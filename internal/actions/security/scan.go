package security

import (
	"context"
	"fmt"
	"io"

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

// Run executes the security scan Run() action.
func (action *Scan) Run(ctx context.Context, out io.Writer) error {
	cfg := action.Config.Get(ctx)
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "Scanning Artifacts...")
	// iterate through artifactDetails in sourceFile or in a gathered object
	results, err := security.ScanArtifacts(ctx, action.SourceFile, action.GatherArtifactReference, action.Config.Repository, cfg.ConcurrentHTTP, action.DryRun)
	if err != nil {
		return err
	}

	// return some nicely formatted data, vulnerabilities, total (deduplicated) size, platforms, oci compliance
	switch action.Output {
	case "json":
		if err := security.PrintJSON(out, results); err != nil {
			return err
		}
	case "markdown":
		if err := security.PrintMarkdown(out, results); err != nil {
			return err
		}
	case "csv":
		if err := security.PrintCSV(out, results); err != nil {
			return err
		}
	case "table":
		if err := security.PrintTable(out, results); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown printing directive: %s", action.Output)
	}

	return nil
}
