package mirror

import (
	"context"
	"fmt"
	"io"
	"os"

	"git.act3-ace.com/ace/data/tool/internal/mirror"
	"git.act3-ace.com/ace/data/tool/internal/security"
)

// Diff represents the mirror ls action.
type Diff struct {
	*Action
	Expanded bool
	Output   string
}

// Run executes the mirror ls command.
func (action *Diff) Run(ctx context.Context, artifactReference string, existingImages []string) error {

	options := mirror.DiffOptions{
		ExistingImages:        existingImages,
		RootArtifactReference: artifactReference,
		Targeter:              action.Config,
		Expanded:              action.Expanded,
	}
	manifestOriginalReferences, err := mirror.ListArtifacts(ctx, options)
	if err != nil {
		return err
	}

	if len(manifestOriginalReferences) == 1 {
		_, err := fmt.Fprintf(os.Stdout, "No artifacts found for %s\n", artifactReference)
		if err != nil {
			return fmt.Errorf("printing to stdout: %w", err)
		}
		return nil
	}
	var outfile io.Writer
	// print them out nicely to out
	if action.Output == "-" {
		outfile = os.Stdout
	} else {
		// create/open the file
		file, err := os.OpenFile(action.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return fmt.Errorf("creating the destination file %s: %w", action.Output, err)
		}
		outfile = file
		defer file.Close()
	}

	return security.PrintCustomTable(outfile, manifestOriginalReferences)
}
