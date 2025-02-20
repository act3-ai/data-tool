package mirror

import (
	"cmp"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"gitlab.com/act3-ai/asce/data/schema/pkg/selectors"
	"gitlab.com/act3-ai/asce/data/tool/internal/actions/oci"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror/encoding"
)

// Source represents a single source line in the `sources.list` file. It includes the source reference (name) and any user-defined labels.
type Source struct {
	Name   string
	Labels map[string]string
}

// ProcessSourcesFile processes the `sources.list` file and returns a slice of Source objects
// that include the source reference, remote repository, descriptor, and the user-defined labels.
// If selectors or platforms filters are passed, then the source list will be modified to only include entries that follow those filters.
func ProcessSourcesFile(ctx context.Context,
	sourceFile string,
	sels selectors.LabelSelectorSet,
	concurrency int,
) ([]Source, error) {
	if sourceFile == "" {
		return []Source{}, nil
	}

	var srcMutex sync.Mutex
	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	file, err := os.Open(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open sources file %s: %w", sourceFile, err)
	}
	defer file.Close()

	scanner := csv.NewReader(file)
	// each record may not have the same number of fields so this is set to -1 (see https://stackoverflow.com/questions/61336787/how-do-i-fix-the-wrong-number-of-fields-with-the-missing-commas-in-csv-file-in)
	scanner.FieldsPerRecord = -1
	scanner.Comment = '#'
	scanner.TrimLeadingSpace = true

	records, err := scanner.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse csv file %q: %w", sourceFile, err)
	}
	if err = file.Close(); err != nil {
		return nil, fmt.Errorf("error parsing sources file: %w", err)
	}
	// put everything above in its own function

	// because of platform filtering, there may be more sources returned than the length of records (indexes having multiple platform-matching manifests)
	var sourceList []Source
	for _, t := range records {
		g.Go(func() error {
			// ignore commented out lines and blank lines
			if len(t) == 0 {
				return nil
			}

			src, lbls, err := processSourceLabels(t)
			if err != nil {
				return err
			}

			if !matchFilter(sels, lbls) {
				return nil
			}

			srcMutex.Lock()
			sourceList = append(sourceList, Source{
				Name:   src,
				Labels: lbls,
			})
			srcMutex.Unlock()
			//}

			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return nil, err
	}

	slices.SortFunc(sourceList, func(a, b Source) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return sourceList, nil
}

func processSourceLabels(t []string) (string, map[string]string, error) {
	var source string
	lbls := make(map[string]string)

	if len(t) == 0 {
		return source, lbls, fmt.Errorf("error: empty source")
	}

	source = t[0]
	t = t[1:]
	if len(t) > 0 {
		for _, text := range t {
			pair := strings.Split(text, "=")
			if len(pair) != 2 {
				return "", nil, fmt.Errorf("error parsing the source labels: should be in key=value format but received %s", text)
			}
			lbls[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
		}
	}

	err := validation.ValidateLabels(lbls, field.NewPath("")).ToAggregate()
	if err != nil {
		return "", nil, fmt.Errorf("error validating labels: %w", err)
	}
	return source, lbls, nil
}

func parsePlatforms(p []string) ([]*ocispec.Platform, error) {
	// accept an os/arch/variant formatted string
	platforms := make([]*ocispec.Platform, len(p))

	for i, platform := range p {
		parsed, err := oci.ParsePlatform(platform)
		if err != nil {
			return nil, err
		}
		platforms[i] = parsed
	}

	return platforms, nil
}

func parseFilters(sel []string) (selectors.LabelSelectorSet, error) {
	// get the selectors
	if len(sel) != 0 {
		filters, err := selectors.Parse(sel)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		return filters, nil
	}
	return selectors.LabelSelectorSet{}, nil
}

func matchFilter(filters selectors.LabelSelectorSet, annotations map[string]string) bool {
	if len(filters) != 0 {
		var lb labels.Set = annotations

		return filters.Matches(lb)
	}
	// if there are no filters then we should not filter out the ref.
	return true
}

func createSourcesMap(sources []Source) map[string]map[string]string {
	subsetMap := make(map[string]map[string]string, len(sources))
	for _, s := range sources {
		subsetMap[s.Name] = s.Labels
	}
	return subsetMap
}

func filterRefByLabelAndGenerateDestinations(desc ocispec.Descriptor,
	filters selectors.LabelSelectorSet,
	mapper mapperFunc,
) ([]string, error) {
	lb := labels.Set{}
	// process the labels in the annotations if we have any
	if l, ok := desc.Annotations[encoding.AnnotationLabels]; ok {
		if err := json.Unmarshal([]byte(l), &lb); err != nil {
			return nil, fmt.Errorf("error decoding labels of %q: %w", l, err)
		}
	}

	if !matchFilter(filters, lb) {
		return nil, nil
	}

	destinations, err := mapper(desc)
	if err != nil {
		return nil, err
	}
	return destinations, nil
}
