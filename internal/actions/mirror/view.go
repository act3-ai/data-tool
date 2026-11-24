package mirror

import (
	"archive/tar"
	"bufio"
	"bytes"
	"cmp"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
)

// View lists references from OCI layouts, archives, or remote registries.
type View struct {
	*Action
	Platforms bool
	Stats     bool // Per-reference cost, unique/shared data, missing blobs, and dedup savings.
}

// Run accepts an OCI-layout directory, OCI-layout tarball, or remote reference.
// By default it writes one complete reference per line, without a header; with
// Stats it writes a human-readable table instead.
func (action *View) Run(ctx context.Context, out, _ io.Writer, input string) error {
	// Extracts the tarball.
	dir, err := os.MkdirTemp("", "ace-dt-mirror-view-*")
	if err != nil {
		return fmt.Errorf("creating extraction directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	// Determines the input type is and opens it.
	index, store, err := openViewInput(ctx, input, dir)
	if err != nil {
		return err
	}
	if action.Stats {
		return printStats(ctx, out, index, store, action.Platforms)
	}
	// prints the references to stdout.
	return printReferences(ctx, out, index, store, action.Platforms)
}

// Iterates through the OCI index and writes the references as a csv.
// The OCI layout specification does not guarantee full registry references in
// ref-name annotations; an absent annotation produces an empty reference field.
func printReferences(ctx context.Context, out io.Writer, index ocispec.Index, store content.ReadOnlyStorage, includePlatforms bool) error {
	// Uses a standard csv writer.
	w := csv.NewWriter(out)
	// Grabs the top level names in the OCI layout.
	for _, desc := range index.Manifests {
		row := []string{desc.Annotations[ocispec.AnnotationRefName]}
		// Returns the platforms if requested.
		if includePlatforms {
			ref, err := walkReference(ctx, store, desc)
			if err != nil {
				return err
			}
			row = append(row, strings.Join(ref.platforms, ";"))
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// printStats writes references sorted by cost, largest first, then totals. Cost
// is the size of a reference's distinct blobs; unique is the part no other
// reference uses (what deleting it would free); shared is the rest.
func printStats(ctx context.Context, out io.Writer, index ocispec.Index, store content.ReadOnlyStorage, includePlatforms bool) error {
	var refs []*refWalk
	uses := map[digest.Digest]int{}
	stored, missing := map[digest.Digest]int64{}, map[digest.Digest]int64{}
	for _, desc := range index.Manifests {
		ref, err := walkReference(ctx, store, desc)
		if err != nil {
			return err
		}
		refs = append(refs, ref)
		for d := range ref.blobs {
			uses[d]++
		}
		maps.Copy(stored, ref.blobs)
		maps.Copy(missing, ref.missing)
	}
	slices.SortStableFunc(refs, func(a, b *refWalk) int { return cmp.Compare(b.cost, a.cost) })

	header := "REFERENCE\tCOST\tUNIQUE\tSHARED\tMISSING"
	if includePlatforms {
		header += "\tPLATFORMS"
	}
	lines := []string{header}
	var logical int64
	for _, ref := range refs {
		var unique int64
		for d, size := range ref.blobs {
			if uses[d] == 1 {
				unique += size
			}
		}
		logical += ref.cost
		name := ref.root.Annotations[ocispec.AnnotationRefName]
		if name == "" {
			name = ref.root.Digest.String() + " (untagged)"
		}
		line := strings.Join([]string{name, formatBytes(ref.cost), formatBytes(unique),
			formatBytes(ref.cost - unique), formatMissing(ref.missing)}, "\t")
		if includePlatforms {
			line += "\t" + strings.Join(ref.platforms, ";")
		}
		lines = append(lines, line)
	}
	storedBytes := sumSizes(stored)
	lines = append(lines, "",
		fmt.Sprintf("references: %d   blobs: %d   stored: %s   logical: %s   saved by dedup: %s", len(refs),
			len(stored), formatBytes(storedBytes), formatBytes(logical), formatBytes(logical-storedBytes)),
		"missing blobs: "+formatMissing(missing))

	tw := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	if _, err := fmt.Fprintln(tw, strings.Join(lines, "\n")); err != nil {
		return err
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("writing stats: %w", err)
	}
	return nil
}

// refWalk holds what one root reference reaches. Sizes come from descriptors,
// so they are compressed sizes and are known even for missing blobs.
type refWalk struct {
	store     content.ReadOnlyStorage
	root      ocispec.Descriptor
	blobs     map[digest.Digest]int64 // Every distinct blob reached, present or missing.
	missing   map[digest.Digest]int64
	cost      int64    // Sum of blobs.
	platforms []string // Distinct image platforms in descriptor order.
}

// walkReference walks root's index and manifest graph. Missing manifests and
// indexes are recorded and not descended into, rather than failing the walk.
// Artifacts count toward blobs but not platforms. Subjects are not followed:
// they point back at an image that is counted under its own reference.
func walkReference(ctx context.Context, store content.ReadOnlyStorage, root ocispec.Descriptor) (*refWalk, error) {
	ref := &refWalk{store: store, root: root, blobs: map[digest.Digest]int64{}, missing: map[digest.Digest]int64{}}
	if err := ref.walk(ctx, root); err != nil {
		return nil, err
	}
	ref.cost = sumSizes(ref.blobs)
	return ref, nil
}

// Walks through the tree of OCI descriptors.
func (ref *refWalk) walk(ctx context.Context, desc ocispec.Descriptor) error {
	if _, done := ref.blobs[desc.Digest]; done {
		return nil // Repeated subtrees are walked once.
	}
	if present, err := ref.visit(ctx, desc); err != nil || !present {
		return err
	}
	switch desc.MediaType {
	// If it is an index, it needs to pulled and its children walked through.
	case ocispec.MediaTypeImageIndex, "application/vnd.docker.distribution.manifest.list.v2+json":
		var index ocispec.Index
		if err := fetchJSON(ctx, ref.store, desc, &index); err != nil {
			return err
		}
		// Walks through the children.
		for _, child := range index.Manifests {
			if err := ref.walk(ctx, child); err != nil {
				return err
			}
		}
	// If it is an image, it needs to be pulled and its platform grabbed.
	case ocispec.MediaTypeImageManifest, "application/vnd.docker.distribution.manifest.v2+json":
		return ref.walkManifest(ctx, desc)
	}
	return nil
}

// walkManifest records the config and layers, then the platform of an image
// manifest: from its descriptor when complete, otherwise from the image config.
func (ref *refWalk) walkManifest(ctx context.Context, desc ocispec.Descriptor) error {
	var manifest ocispec.Manifest
	if err := fetchJSON(ctx, ref.store, desc, &manifest); err != nil {
		return err
	}
	configPresent, err := ref.visit(ctx, manifest.Config)
	if err != nil {
		return err
	}
	for _, layer := range manifest.Layers {
		if _, err := ref.visit(ctx, layer); err != nil {
			return err
		}
	}
	if desc.ArtifactType != "" || desc.Annotations["vnd.docker.reference.type"] != "" ||
		manifest.ArtifactType != "" || (manifest.Config.MediaType != ocispec.MediaTypeImageConfig &&
		manifest.Config.MediaType != "application/vnd.docker.container.image.v1+json") {
		return nil // Artifacts and BuildKit attestations are data, not image platforms.
	}
	// grab the platform from the metadata.
	platform := formatPlatform(desc.Platform)
	if platform == "" && configPresent {
		// Fetch the image configuration to find the OS.
		var image ocispec.Image
		if err := fetchJSON(ctx, ref.store, manifest.Config, &image); err != nil {
			return err
		}
		platform = formatPlatform(&image.Platform)
	}
	// If its a new platform, add it to the list.
	if platform != "" && !slices.Contains(ref.platforms, platform) {
		ref.platforms = append(ref.platforms, platform)
	}
	return nil
}

// visit records desc and reports whether its content is in the store.
func (ref *refWalk) visit(ctx context.Context, desc ocispec.Descriptor) (bool, error) {
	if _, seen := ref.blobs[desc.Digest]; seen {
		_, missing := ref.missing[desc.Digest]
		return !missing, nil
	}
	ref.blobs[desc.Digest] = desc.Size
	exists, err := ref.store.Exists(ctx, desc)
	if err != nil {
		return false, fmt.Errorf("checking %s: %w", desc.Digest, err)
	}
	if !exists {
		ref.missing[desc.Digest] = desc.Size
	}
	return exists, nil
}

// fetchJSON caps the size before fetching because content.FetchAll allocates
// desc.Size bytes up front, and that size comes from untrusted input.
func fetchJSON(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor, v any) error {
	if desc.Size > 4<<20 {
		return fmt.Errorf("%s: size %d exceeds 4 MiB metadata limit", desc.Digest, desc.Size)
	}
	data, err := content.FetchAll(ctx, fetcher, desc)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", desc.Digest, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("decoding %s: %w", desc.Digest, err)
	}
	return nil
}

// formatPlatform returns "" for missing or unknown platforms.
func formatPlatform(p *ocispec.Platform) string {
	if p == nil || p.OS == "" || p.Architecture == "" || p.OS == "unknown" || p.Architecture == "unknown" {
		return ""
	}
	return path.Join(p.OS, p.Architecture, p.Variant)
}

func sumSizes(blobs map[digest.Digest]int64) int64 {
	var total int64
	for _, size := range blobs {
		total += size
	}
	return total
}

// formatMissing returns a blob count, with the declared bytes when nonzero.
func formatMissing(blobs map[digest.Digest]int64) string {
	if len(blobs) == 0 {
		return "0"
	}
	return fmt.Sprintf("%d (%s)", len(blobs), formatBytes(sumSizes(blobs)))
}

// formatBytes formats n in binary units, e.g. "1.5 MiB".
func formatBytes(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	value, i := float64(n)/1024, 0
	for value >= 1024 && i < 5 {
		value /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %ciB", value, "KMGTPE"[i])
}

// openViewInput opens an OCI-layout directory, an OCI-layout tarball (extracted
// into the caller-owned dir), or a tagged/digest-qualified registry reference.
func openViewInput(ctx context.Context, input, dir string) (ocispec.Index, content.ReadOnlyStorage, error) {
	// Checks if the input is a local path.
	info, err := os.Stat(input)
	switch {
	// Case 1: its a local directory.
	case err == nil && info.IsDir():
		return openViewLayout(ctx, input)
	// Case 2: its a local tarball.
	case err == nil && info.Mode().IsRegular():
		if err := extractViewTar(ctx, input, dir); err != nil {
			return ocispec.Index{}, nil, fmt.Errorf("extracting %q: %w", input, err)
		}
		return openViewLayout(ctx, dir)
	// If its something else, throw an error
	case err == nil:
		return ocispec.Index{}, nil, fmt.Errorf("%q is not a regular OCI archive or OCI-layout directory", input)
	case !errors.Is(err, os.ErrNotExist):
		return ocispec.Index{}, nil, fmt.Errorf("checking %q: %w", input, err)
	}

	// If it is not a local file nor throw an error, assume its a remote reference.
	// Missing local paths fail here: they don't parse as references, or they lack a tag/digest.
	repo, err := remote.NewRepository(input)
	if err != nil || repo.Reference.Reference == "" {
		return ocispec.Index{}, nil, fmt.Errorf("%q is not an existing local path or a tagged/digest-qualified registry reference", input)
	}

	// TODO: replace temporary Docker credentials with the application's
	// registry-auth helper when the Action/client wiring is provided. That helper
	// must also map docker.io to registry-1.docker.io.
	// load the user's Docker credentials
	store, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
	if err != nil {
		return ocispec.Index{}, nil, fmt.Errorf("loading Docker credentials: %w", err)
	}
	repo.Client = &auth.Client{Credential: credentials.Credential(store)}

	// resolve the reference to a manifest descriptor.
	desc, err := repo.Resolve(ctx, repo.Reference.Reference)
	if err != nil {
		return ocispec.Index{}, nil, fmt.Errorf("resolving %q: %w", input, err)
	}
	// Only the requested root has this reference; children do not inherit its tag.
	desc.Annotations = map[string]string{ocispec.AnnotationRefName: input}
	return ocispec.Index{Manifests: []ocispec.Descriptor{desc}}, repo, nil
}

// openViewLayout relies on oci.NewFromFS to validate oci-layout and index.json,
// then re-reads index.json because the store does not expose descriptor order.
func openViewLayout(ctx context.Context, dir string) (ocispec.Index, content.ReadOnlyStorage, error) {
	store, err := oci.NewFromFS(ctx, os.DirFS(dir))
	if err != nil {
		return ocispec.Index{}, nil, fmt.Errorf("%q is not an OCI layout: %w", dir, err)
	}

	// read index.json to preserve the order of descriptors.
	data, err := os.ReadFile(filepath.Join(dir, ocispec.ImageIndexFile))
	if err != nil {
		return ocispec.Index{}, nil, fmt.Errorf("reading index.json: %w", err)
	}

	var index ocispec.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return ocispec.Index{}, nil, fmt.Errorf("decoding index.json: %w", err)
	}
	return index, store, nil
}

// extractViewTar accepts plain tar, gzip tar, and zstd tar. The caller owns dir.
func extractViewTar(ctx context.Context, archivePath, dir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("opening archive: %w", err)
	}
	defer f.Close()

	// checks the magic bytes to determine if the tarball is compressed.
	buffered := bufio.NewReader(f)
	magic, _ := buffered.Peek(4) // Short or unreadable input fails in tr.Next below.
	var reader io.Reader = buffered
	switch {
	// Case for gzip.
	case bytes.HasPrefix(magic, []byte{0x1f, 0x8b}):
		gz, err := gzip.NewReader(buffered)
		if err != nil {
			return fmt.Errorf("opening gzip stream: %w", err)
		}
		defer gz.Close()
		reader = gz
	// Case for Zstandard.
	case bytes.HasPrefix(magic, []byte{0x28, 0xb5, 0x2f, 0xfd}):
		dec, err := zstd.NewReader(buffered)
		if err != nil {
			return fmt.Errorf("opening zstd stream: %w", err)
		}
		defer dec.Close()
		reader = dec
	}

	tr := tar.NewReader(&viewContextReader{ctx: ctx, reader: reader})
	const maxExtractedBytes int64 = 100 << 30
	var extractedBytes int64
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading tar header: %w", err)
		}
		if !filepath.IsLocal(hdr.Name) {
			return fmt.Errorf("unsafe tar path %q", hdr.Name)
		}
		target := filepath.Join(dir, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			// If the tar entry is a directory, create it on the filesystem.
			if err := os.MkdirAll(target, 0700); err != nil {
				return fmt.Errorf("creating directory %q: %w", hdr.Name, err)
			}
		case tar.TypeReg:
			if hdr.Size > maxExtractedBytes-extractedBytes {
				return fmt.Errorf("tarball exceeds %d extracted bytes; extract it yourself", maxExtractedBytes)
			}
			extractedBytes += hdr.Size
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
				return fmt.Errorf("creating parent of %q: %w", hdr.Name, err)
			}
			flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
			if target == filepath.Join(dir, ocispec.ImageIndexFile) {
				// The last root index.json wins; other duplicate files are errors.
				flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			}
			out, err := os.OpenFile(target, flags, 0600)
			if err != nil {
				return fmt.Errorf("creating file %q: %w", hdr.Name, err)
			}
			_, copyErr := io.CopyN(out, tr, hdr.Size)
			if err := errors.Join(copyErr, out.Close()); err != nil {
				return fmt.Errorf("extracting file %q: %w", hdr.Name, err)
			}
		default:
			return fmt.Errorf("unsupported tar entry type %q for %q", hdr.Typeflag, hdr.Name)
		}
	}
}

// viewContextReader wraps an io.Reader and checks the context for cancellation before each read.
type viewContextReader struct {
	// io.Reader cannot be modified to accept a context.
	ctx    context.Context //nolint:containedctx
	reader io.Reader
}

func (r *viewContextReader) Read(p []byte) (int, error) {
	// checks for cancellation.
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p) //nolint:wrapcheck
}
