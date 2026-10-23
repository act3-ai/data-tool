package mirror

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ViewGatherIndex reads an index.json file and displays the contained image references and digests.
func ViewGatherIndex(indexPath string, detailed bool) error {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("reading gather index at %s: %w", indexPath, err)
	}

	var index ocispec.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("unmarshaling index.json: %w", err)
	}

	if len(index.Manifests) == 0 {
		fmt.Println("No images or manifests found in gather index.")
		return nil
	}

	if detailed {
		printDetailedView(index)
	} else {
		printCompactView(index)
	}

	return nil
}

func printCompactView(index ocispec.Index) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, desc := range index.Manifests {
		ref := desc.Annotations[ocispec.AnnotationRefName]
		if ref == "" {
			ref = "<unnamed/untagged>"
		}
		fmt.Fprintf(w, "%s\t%s\n", desc.Digest, ref)
	}
	w.Flush()
}

func printDetailedView(index ocispec.Index) {
	for i, desc := range index.Manifests {
		ref := desc.Annotations[ocispec.AnnotationRefName]
		if ref == "" {
			ref = "<unnamed/untagged>"
		}

		fmt.Printf("commit/manifest %s\n", desc.Digest)
		fmt.Printf("Ref:         %s\n", ref)
		fmt.Printf("Media Type:  %s\n", desc.MediaType)
		fmt.Printf("Size:        %d bytes\n", desc.Size)

		if len(desc.Annotations) > 1 {
			fmt.Println("Annotations:")
			for k, v := range desc.Annotations {
				if k != ocispec.AnnotationRefName {
					fmt.Printf("  - %s: %s\n", k, v)
				}
			}
		}

		if i < len(index.Manifests)-1 {
			fmt.Println()
		}
	}
}