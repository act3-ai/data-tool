package oci

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/act3-ai/data-tool/internal/mirror/encoding"
	"github.com/act3-ai/go-common/pkg/logger"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
)

// IdxOfIdx represents the idx-of-idx action.
type IdxOfIdx struct {
	*Action

	File string
}

// Run runs the prune action.
func (action *IdxOfIdx) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	records, err := action.readSourceFile()
	if err != nil {
		return err
	}

	var results []string
	for _, record := range records {
		if record[0] == "" {
			return fmt.Errorf("missing ref field in record: %v", record)
		}
		ref := record[0]
		log.InfoContext(ctx, "evaluating record", "reference", ref)

		gt, err := action.Config.ReadOnlyGraphTarget(ctx, ref)
		if err != nil {
			return fmt.Errorf("initializing target: %w", err)
		}

		desc, err := gt.Resolve(ctx, ref)
		if err != nil {
			return fmt.Errorf("resolving reference descriptor: %w", err)
		}

		if encoding.IsIndex(desc.MediaType) {
			log.InfoContext(ctx, "identified record as index, evaluating")
			manBytes, err := content.FetchAll(ctx, gt, desc)
			if err != nil {
				return fmt.Errorf("fetching manifest: %w", err)
			}

			var idx ocispec.Index
			if err := json.Unmarshal(manBytes, &idx); err != nil {
				return fmt.Errorf("decoding index manifest: %w", err)
			}

			for _, desc := range idx.Manifests {
				if encoding.IsIndex(desc.MediaType) {
					results = append(results, fmt.Sprintf("%s : %s\n", ref, desc.Digest))
				}
			}
		} else {
			log.InfoContext(ctx, "record is not an index, skipping")
		}
	}

	if len(results) > 0 {
		fmt.Fprintln(out, "Discovered index-of-indexes:")

		for _, result := range results {
			if _, err := out.Write([]byte(result)); err != nil {
				return fmt.Errorf("writing results: %w", err)
			}
		}
	} else {
		fmt.Fprintln(out, "No index-of-indexes found.")
	}

	return nil
}

// compatible with existing gather source files
func (action *IdxOfIdx) readSourceFile() ([][]string, error) {
	if action.File == "" {
		return nil, fmt.Errorf("no source file provided")
	}

	file, err := os.Open(action.File)
	if err != nil {
		return nil, fmt.Errorf("unable to open sources file %s: %w", action.File, err)
	}
	defer file.Close()

	scanner := csv.NewReader(file)
	// each record may not have the same number of fields so this is set to -1 (see https://stackoverflow.com/questions/61336787/how-do-i-fix-the-wrong-number-of-fields-with-the-missing-commas-in-csv-file-in)
	scanner.FieldsPerRecord = -1
	scanner.Comment = '#'
	scanner.TrimLeadingSpace = true

	records, err := scanner.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse csv file %q: %w", action.File, err)
	}
	if err = file.Close(); err != nil {
		return nil, fmt.Errorf("error parsing sources file: %w", err)
	}
	// put everything above in its own function

	return records, nil
}
