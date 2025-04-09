package mirror

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/act3-ai/data-tool/internal/mirror"
	"github.com/act3-ai/data-tool/internal/security"
)

// Diff represents the mirror ls action.
type Diff struct {
	*Action
	Expanded bool
	Output   []string
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

	if len(manifestOriginalReferences) == 0 {
		_, err := fmt.Fprintf(os.Stdout, "No artifacts found for %s\n", artifactReference)
		if err != nil {
			return fmt.Errorf("printing to stdout: %w", err)
		}
		return nil
	}

	outputMethods := map[string][]io.Writer{}
	for _, o := range action.Output {
		var outfile io.Writer
		output := strings.Split(o, "=")
		if len(output) < 2 {
			// default to std out
			outfile = os.Stdout
		} else {
			outfile, err = os.OpenFile(output[1], os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				return fmt.Errorf("creating/opening output file: %w", err)
			}
		}
		outputMethods[output[0]] = append(outputMethods[output[0]], outfile)
	}

	for method, writers := range outputMethods {
		// for each writer match to proper
		for _, writer := range writers {
			switch method {
			case "json":
				b, err := generateJSONDiffResults(manifestOriginalReferences)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(writer, string(b))
				if err != nil {
					return fmt.Errorf("error printing JSON output: %w", err)
				}
			case "csv":
				table := [][]string{{"reference", "digest"}}
				table = append(table, manifestOriginalReferences...)
				w := csv.NewWriter(writer)
				if err := w.WriteAll(table); err != nil {
					return fmt.Errorf("writing csv table: %w", err)
				}
			case "table":
				table := [][]string{{"reference", "digest"}}
				table = append(table, manifestOriginalReferences...)
				if err := security.PrintCustomTable(writer, table); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown printing directive: %s", action.Output)
			}
		}
	}
	return nil
}

// DiffArtifact is for json encoding the diff results.
type DiffArtifact struct {
	Digest    string `json:"digest"`
	Reference string `json:"reference"`
}

func generateJSONDiffResults(manifestOriginalReferences [][]string) ([]byte, error) {
	results := make([]DiffArtifact, len(manifestOriginalReferences))
	for i, artifact := range manifestOriginalReferences {
		results[i] = DiffArtifact{
			Reference: artifact[0],
			Digest:    artifact[1],
		}
	}
	b, err := json.Marshal(&results)
	if err != nil {
		return nil, fmt.Errorf("marshalling the json data: %w", err)
	}
	return b, nil
}
