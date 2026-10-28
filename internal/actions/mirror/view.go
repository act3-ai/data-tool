package mirror

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
)

// Verify at compile time that the store supports read-only graph operations.
var _ content.ReadOnlyGraphStorage = (*oci.ReadOnlyStore)(nil)

// ViewGatherIndex renders references or image metadata from an extracted OCI layout.
// When diffPath is set, it compares this layout's top-level index entries against
// diffPath and writes tab-separated additions and removals.
func ViewGatherIndex(ctx context.Context, inputPath string, detailed, showRefs, showImages bool, diffPath string) error {
	if diffPath != "" {
		if detailed || showRefs || showImages {
			return fmt.Errorf("--diff cannot be combined with --detailed, --refs, or --images")
		}
		return diffIndexes(inputPath, diffPath)
	}

	storeDir, err := prepareOCIStoreDir(inputPath)
	if err != nil {
		return fmt.Errorf("preparing OCI store: %w", err)
	}

	index, err := readOCIIndex(storeDir)
	if err != nil {
		return err
	}
	if len(index.Manifests) == 0 {
		if !showRefs && !showImages {
			fmt.Println("No images or manifests found in gather index.")
		}
		return nil
	}

	if showRefs {
		for _, desc := range index.Manifests {
			if _, err := parseBlobDigest(desc.Digest); err != nil {
				return err
			}
			fmt.Printf("%s\t%s\t%s\n", desc.Digest, referenceName(desc), getKindFromMediaType(desc.MediaType))
		}
		return nil
	}

	// Open the existing layout read-only: no temporary extraction and no writes.
	store, err := oci.NewFromFS(ctx, os.DirFS(storeDir))
	if err != nil {
		return fmt.Errorf("opening read-only OCI store at %s: %w", storeDir, err)
	}
	var graph content.ReadOnlyGraphStorage = store

	if showImages {
		for _, desc := range index.Manifests {
			if err := printImages(ctx, graph, os.Stdout, desc, referenceName(desc), make(map[digest.Digest]bool)); err != nil {
				return err
			}
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, desc := range index.Manifests {
		ref := referenceName(desc)
		manifestBytes, err := fetchMetadata(ctx, graph, desc)
		if err != nil {
			fmt.Fprintf(w, "%s\t%s\t(manifest missing)\n", desc.Digest, ref)
			continue
		}
		var manifest ocispec.Manifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			fmt.Fprintf(w, "%s\t%s\t(invalid manifest)\n", desc.Digest, ref)
			continue
		}
		var imgConfig ocispec.Image
		if manifest.Config.Digest != "" {
			configBytes, err := fetchMetadata(ctx, graph, manifest.Config)
			if err == nil {
				_ = json.Unmarshal(configBytes, &imgConfig)
			}
		}
		platform := "unknown"
		if imgConfig.OS != "" || imgConfig.Architecture != "" {
			platform = fmt.Sprintf("%s/%s", imgConfig.OS, imgConfig.Architecture)
		}
		if detailed {
			fmt.Printf("Digest:       %s\n", desc.Digest)
			fmt.Printf("Ref:          %s\n", ref)
			if imgConfig.Created != nil && !imgConfig.Created.IsZero() {
				fmt.Printf("Created:      %s\n", imgConfig.Created.Format("2006-01-02 15:04:05 MST"))
			}
			fmt.Printf("Platform:     %s\n", platform)
			fmt.Printf("Media Type:   %s\n", desc.MediaType)
			fmt.Printf("Layers:       %d\n", len(manifest.Layers))
			fmt.Printf("Config Digest: %s\n", manifest.Config.Digest)
			if len(desc.Annotations) > 1 {
				fmt.Println("Annotations:")
				for k, v := range desc.Annotations {
					if k != ocispec.AnnotationRefName {
						fmt.Printf("  - %s: %s\n", k, v)
					}
				}
			}
			fmt.Println("--------------------------------------------------------------------------------")
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d layers\n", desc.Digest, ref, platform, len(manifest.Layers))
		}
	}
	return w.Flush()
}

// diffIndexes compares top-level descriptors only. A row is identified by its
// digest, OCI reference name, and normalized kind. This makes a retag or a
// change from image to index visible as one removal and one addition.
// Output is stable and tab-separated: +/-<TAB>digest<TAB>ref<TAB>kind.
func diffIndexes(inputPath, againstPath string) error {
	inputDir, err := prepareOCIStoreDir(inputPath)
	if err != nil {
		return fmt.Errorf("preparing input OCI store: %w", err)
	}
	againstDir, err := prepareOCIStoreDir(againstPath)
	if err != nil {
		return fmt.Errorf("preparing comparison OCI store: %w", err)
	}

	inputIndex, err := readOCIIndex(inputDir)
	if err != nil {
		return err
	}
	againstIndex, err := readOCIIndex(againstDir)
	if err != nil {
		return err
	}

	inputEntries, err := indexEntries(inputIndex)
	if err != nil {
		return err
	}
	againstEntries, err := indexEntries(againstIndex)
	if err != nil {
		return err
	}

	// Entries present only in --diff's argument were removed from the input.
	printIndexDiff("-", againstEntries, inputEntries)
	// Entries present only in the input were added relative to --diff's argument.
	printIndexDiff("+", inputEntries, againstEntries)
	return nil
}

type indexEntry struct {
	digest digest.Digest
	ref    string
	kind   string
}

func readOCIIndex(storeDir string) (ocispec.Index, error) {
	indexData, err := os.ReadFile(filepath.Join(storeDir, "index.json"))
	if err != nil {
		return ocispec.Index{}, fmt.Errorf("reading index.json from OCI store: %w", err)
	}
	var index ocispec.Index
	if err := json.Unmarshal(indexData, &index); err != nil {
		return ocispec.Index{}, fmt.Errorf("unmarshaling index.json: %w", err)
	}
	return index, nil
}

func indexEntries(index ocispec.Index) (map[string]indexEntry, error) {
	entries := make(map[string]indexEntry, len(index.Manifests))
	for _, desc := range index.Manifests {
		parsed, err := parseBlobDigest(desc.Digest)
		if err != nil {
			return nil, err
		}
		entry := indexEntry{digest: parsed, ref: referenceName(desc), kind: getKindFromMediaType(desc.MediaType)}
		entries[indexEntryKey(entry)] = entry
	}
	return entries, nil
}

func indexEntryKey(entry indexEntry) string {
	return entry.digest.String() + "\x00" + entry.ref + "\x00" + entry.kind
}

func printIndexDiff(prefix string, entries, other map[string]indexEntry) {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		if _, found := other[key]; !found {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		entry := entries[key]
		fmt.Printf("%s\t%s\t%s\t%s\n", prefix, entry.digest, entry.ref, entry.kind)
	}
}

func referenceName(desc ocispec.Descriptor) string {
	if ref := desc.Annotations[ocispec.AnnotationRefName]; ref != "" {
		return ref
	}
	return "<unnamed>"
}

func getKindFromMediaType(mediaType string) string {
	switch mediaType {
	case ocispec.MediaTypeImageIndex, "application/vnd.docker.distribution.manifest.list.v2+json":
		return "index"
	case ocispec.MediaTypeImageManifest, "application/vnd.docker.distribution.manifest.v2+json":
		return "image"
	default:
		return "artifact"
	}
}

// Parse before reading: the digest selects blobs/<algorithm>/<encoded>, not blobs/sha256.
// The ORAS OCI store performs that path lookup using the validated descriptor.
func parseBlobDigest(value digest.Digest) (digest.Digest, error) {
	parsed, err := digest.Parse(value.String())
	if err != nil {
		return "", fmt.Errorf("invalid OCI blob digest %q: %w", value, err)
	}
	return parsed, nil
}

func fetchMetadata(ctx context.Context, graph content.ReadOnlyGraphStorage, desc ocispec.Descriptor) ([]byte, error) {
	parsed, err := parseBlobDigest(desc.Digest)
	if err != nil {
		return nil, err
	}
	data, err := content.FetchAll(ctx, graph, desc)
	if err != nil {
		return nil, fmt.Errorf("reading blobs/%s/%s: %w", parsed.Algorithm(), parsed.Encoded(), err)
	}
	return data, nil
}

// --images has exactly three tab-separated fields on every row:
// digest<TAB>index<TAB>ref or digest<TAB>image<TAB>platform.
// A child manifest is listed beneath its parent index; it does not inherit its tag.
func printImages(ctx context.Context, graph content.ReadOnlyGraphStorage, w io.Writer, desc ocispec.Descriptor, ref string, ancestors map[digest.Digest]bool) error {
	parsed, err := parseBlobDigest(desc.Digest)
	if err != nil {
		return err
	}
	if ancestors[parsed] {
		return fmt.Errorf("cycle in OCI indexes at %s", parsed)
	}
	switch getKindFromMediaType(desc.MediaType) {
	case "index":
		indexBytes, err := fetchMetadata(ctx, graph, desc)
		if err != nil {
			return err
		}
		var childIndex ocispec.Index
		if err := json.Unmarshal(indexBytes, &childIndex); err != nil {
			return fmt.Errorf("decoding OCI index %s: %w", parsed, err)
		}
		fmt.Fprintf(w, "%s\tindex\t%s\n", parsed, ref)
		ancestors[parsed] = true
		defer delete(ancestors, parsed)
		for _, child := range childIndex.Manifests {
			if err := printImages(ctx, graph, w, child, referenceName(child), ancestors); err != nil {
				return err
			}
		}
	case "image":
		platform, err := imagePlatform(ctx, graph, desc)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "%s\timage\t%s\n", parsed, platform)
	}
	return nil
}

func imagePlatform(ctx context.Context, graph content.ReadOnlyGraphStorage, desc ocispec.Descriptor) (string, error) {
	if desc.Platform != nil && desc.Platform.OS != "" && desc.Platform.Architecture != "" {
		return desc.Platform.OS + "/" + desc.Platform.Architecture, nil
	}
	manifestBytes, err := fetchMetadata(ctx, graph, desc)
	if err != nil {
		return "", err
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return "", fmt.Errorf("decoding image manifest %s: %w", desc.Digest, err)
	}
	if manifest.Config.Digest == "" {
		return "unknown", nil
	}
	configBytes, err := fetchMetadata(ctx, graph, manifest.Config)
	if err != nil {
		return "", err
	}
	var imgConfig ocispec.Image
	if err := json.Unmarshal(configBytes, &imgConfig); err != nil {
		return "", fmt.Errorf("decoding image config %s: %w", manifest.Config.Digest, err)
	}
	if imgConfig.OS == "" || imgConfig.Architecture == "" {
		return "unknown", nil
	}
	return imgConfig.OS + "/" + imgConfig.Architecture, nil
}

func prepareOCIStoreDir(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return path, nil
	}
	if filepath.Base(path) == "index.json" {
		return filepath.Dir(path), nil
	}
	return "", fmt.Errorf("%s is not an extracted OCI layout directory or index.json; extract the archive before running mirror view", path)
}
