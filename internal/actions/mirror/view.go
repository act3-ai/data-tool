package mirror

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
)

// ViewGatherIndex opens an OCI store (from a directory, direct index file, or tarball),
// wraps it in an ORAS ReadOnlyGraphStorage (oci.Store), and traverses its content-addressable blobs.
func ViewGatherIndex(ctx context.Context, inputPath string, detailed bool) error {
	// Prepare the local OCI store directory (unpacking tar archives if necessary)
	storeDir, cleanup, err := prepareOCIStoreDir(inputPath)
	if err != nil {
		return fmt.Errorf("preparing OCI store: %w", err)
	}
	defer cleanup()

	// Initialize ORAS OCI Content Store (satisfies content.ReadOnlyGraphStorage and content.Fetcher)
	store, err := oci.New(storeDir)
	if err != nil {
		return fmt.Errorf("initializing ORAS OCI store at %s: %w", storeDir, err)
	}

	// Read root index.json from the OCI store directory
	indexData, err := os.ReadFile(filepath.Join(storeDir, "index.json"))
	if err != nil {
		return fmt.Errorf("reading index.json from OCI store: %w", err)
	}

	var index ocispec.Index
	if err := json.Unmarshal(indexData, &index); err != nil {
		return fmt.Errorf("unmarshaling index.json: %w", err)
	}

	if len(index.Manifests) == 0 {
		fmt.Println("No images or manifests found in gather index.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	for _, desc := range index.Manifests {
		ref := desc.Annotations[ocispec.AnnotationRefName]
		if ref == "" {
			ref = "<unnamed/untagged>"
		}

		// 1. Fetch Manifest blob via ORAS Content Storage
		manifestBytes, err := content.FetchAll(ctx, store, desc)
		if err != nil {
			fmt.Fprintf(w, "%s\t%s\t(manifest blob missing in CAS store)\n", desc.Digest, ref)
			continue
		}

		var manifest ocispec.Manifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			fmt.Fprintf(w, "%s\t%s\t(invalid manifest JSON)\n", desc.Digest, ref)
			continue
		}

		// 2. Fetch Config blob via ORAS Content Storage
		var imgConfig ocispec.Image
		if manifest.Config.Digest != "" {
			configBytes, err := content.FetchAll(ctx, store, manifest.Config)
			if err == nil {
				_ = json.Unmarshal(configBytes, &imgConfig)
			}
		}

		// Format output
		if detailed {
			fmt.Printf("Digest:       %s\n", desc.Digest)
			fmt.Printf("Ref:          %s\n", ref)
			if !imgConfig.Created.IsZero() {
				fmt.Printf("Created:      %s\n", imgConfig.Created.Format("2006-01-02 15:04:05 MST"))
			}
			if imgConfig.Architecture != "" || imgConfig.OS != "" {
				fmt.Printf("Platform:     %s / %s\n", imgConfig.Architecture, imgConfig.OS)
			}
			fmt.Printf("Media Type:   %s\n", desc.MediaType)
			fmt.Printf("Layers:       %d\n", len(manifest.Layers))
			fmt.Printf("Config SHA:   %s\n", manifest.Config.Digest)

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
			archOS := "unknown"
			if imgConfig.Architecture != "" || imgConfig.OS != "" {
				archOS = fmt.Sprintf("%s/%s", imgConfig.Architecture, imgConfig.OS)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%d layers\n",
				desc.Digest, ref, archOS, len(manifest.Layers))
		}
	}

	if !detailed {
		w.Flush()
	}

	return nil
}

// prepareOCIStoreDir manages extracting tar archives or pointing to existing OCI directories.
func prepareOCIStoreDir(path string) (string, func(), error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", func() {}, err
	}

	noopCleanup := func() {}

	// Directory input
	if info.IsDir() {
		return path, noopCleanup, nil
	}

	// Tarball input (.tar, .tar.gz, .tgz) -> Unpack to temp directory
	if isTarball(path) {
		tmpDir, err := os.MkdirTemp("", "oci-store-*")
		if err != nil {
			return "", noopCleanup, fmt.Errorf("creating temp directory for OCI store: %w", err)
		}
		cleanup := func() { _ = os.RemoveAll(tmpDir) }

		if err := extractTarToDir(path, tmpDir); err != nil {
			cleanup()
			return "", noopCleanup, fmt.Errorf("extracting OCI tar archive: %w", err)
		}
		return tmpDir, cleanup, nil
	}

	// Direct index.json input -> Use containing directory
	if strings.HasSuffix(path, "index.json") {
		return filepath.Dir(path), noopCleanup, nil
	}

	return "", noopCleanup, fmt.Errorf("unsupported input format at %s", path)
}

func isTarball(path string) bool {
	ext := strings.ToLower(path)
	return strings.HasSuffix(ext, ".tar") || strings.HasSuffix(ext, ".tar.gz") || strings.HasSuffix(ext, ".tgz")
}

func extractTarToDir(tarPath, targetDir string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var reader io.Reader = f
	if gzR, err := gzip.NewReader(f); err == nil {
		defer gzR.Close()
		reader = gzR
	} else {
		_, _ = f.Seek(0, io.SeekStart)
	}

	tr := tar.NewReader(reader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		relPath := filepath.Clean(strings.TrimPrefix(hdr.Name, "./"))
		outPath := filepath.Join(targetDir, relPath)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}