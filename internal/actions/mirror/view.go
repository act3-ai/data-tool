package mirror

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
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
}

// Run accepts an OCI-layout directory, OCI-layout tarball, or remote reference.
// It writes one complete reference per line, without a header.
func (action *View) Run(ctx context.Context, out, _ io.Writer, input string) error {

	// Extracts the tarball.
	dir, err := os.MkdirTemp("", "ace-dt-mirror-view-*")
	if err != nil {
		return fmt.Errorf("creating extraction directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	// Determines the input type is and opens it.
	index, fetcher, err := openViewInput(ctx, input, dir)
	if err != nil {
		return err
	}
	// prints the references to stdout.
	return printReferences(ctx, out, index, fetcher, action.Platforms)
}

// Iterates through the OCI index and writes the references as a csv.
func printReferences(ctx context.Context, out io.Writer, index ocispec.Index, fetcher content.Fetcher, includePlatforms bool) error {

	// Uses a standard csv writer.
	w := csv.NewWriter(out)
	// Grabs the top level names in the OCI layout.
	for _, desc := range index.Manifests {
		row := []string{desc.Annotations[ocispec.AnnotationRefName]}
		// Returns the platforms if requested.
		if includePlatforms {
			platforms, err := viewPlatforms(ctx, fetcher, desc)
			if err != nil {
				return err
			}
			row = append(row, strings.Join(platforms, ";"))
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// viewPlatforms walks through the OCI image index and grabs a list of platforms.
func viewPlatforms(ctx context.Context, fetcher content.Fetcher, root ocispec.Descriptor) ([]string, error) {
	var platforms []string
	seen := make(map[string]bool)

	// Walks through the tree of OCI descriptors.
	var walk func(ocispec.Descriptor) error
	walk = func(desc ocispec.Descriptor) error {
		switch desc.MediaType {

		// If it is an index, it needs to pulled and its children walked through.
		case ocispec.MediaTypeImageIndex, "application/vnd.docker.distribution.manifest.list.v2+json":
			var index ocispec.Index
			if err := fetchJSON(ctx, fetcher, desc, &index); err != nil {
				return err
			}
			// Walks through the children.
			for _, child := range index.Manifests {
				if err := walk(child); err != nil {
					return err
				}
			}

		// If it is an image, it needs to be pulled and its platform grabbed.
		case ocispec.MediaTypeImageManifest, "application/vnd.docker.distribution.manifest.v2+json":
			platform, err := manifestPlatform(ctx, fetcher, desc)
			if err != nil {
				return err
			}
			// If its a new platform, add it to the list.
			if platform != "" && !seen[platform] {
				seen[platform] = true
				platforms = append(platforms, platform)
			}
		}
		return nil
	}
	if err := walk(root); err != nil {
		return nil, err
	}
	return platforms, nil
}

// manifestPlatform returns an image manifest's platform, or "" for artifacts
// and manifests without a known platform.
func manifestPlatform(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) (string, error) {
	if desc.ArtifactType != "" || desc.Annotations["vnd.docker.reference.type"] != "" {
		return "", nil // Skip explicitly marked artifacts and BuildKit attestations.
	}
	// grab the platform from the metadata.
	if platform := formatPlatform(desc.Platform); platform != "" {
		return platform, nil
	}
	var manifest ocispec.Manifest
	if err := fetchJSON(ctx, fetcher, desc, &manifest); err != nil {
		return "", err
	}
	if manifest.ArtifactType != "" || (manifest.Config.MediaType != ocispec.MediaTypeImageConfig &&
		manifest.Config.MediaType != "application/vnd.docker.container.image.v1+json") {
		return "", nil // Artifact manifests are not image platforms.
	}
	// Fetch the image configuration to find the OS.
	var image ocispec.Image
	if err := fetchJSON(ctx, fetcher, manifest.Config, &image); err != nil {
		return "", err
	}
	return formatPlatform(&image.Platform), nil
}

// fetchJSON caps the size before fetching because content.FetchAll allocates desc.Size bytes up front, and that size comes from untrusted input.
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

// openViewInput opens an OCI-layout directory, an OCI-layout tarball, or a tagged/digest-qualified registry reference.
func openViewInput(ctx context.Context, input, dir string) (ocispec.Index, content.Fetcher, error) {
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
	repo, err := remote.NewRepository(input)
	if err != nil || repo.Reference.Reference == "" {
		return ocispec.Index{}, nil, fmt.Errorf("%q is not an existing local path or a tagged/digest-qualified registry reference", input)
	}

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

// openViewLayout re-reads index.json because the store does not expose descriptor order.
func openViewLayout(ctx context.Context, dir string) (ocispec.Index, content.Fetcher, error) {
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
	var extractedBytes int64
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading tar header: %w", err)
		}

		target := filepath.Join(dir, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			// If the tar entry is a directory, create it on the filesystem.
			if err := os.MkdirAll(target, 0700); err != nil {
				return fmt.Errorf("creating directory %q: %w", hdr.Name, err)
			}
		case tar.TypeReg:
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
