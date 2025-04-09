package sbom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/act3-ai/go-common/pkg/logger"

	"github.com/act3-ai/data-tool/internal/ref"
	sbom "github.com/act3-ai/data-tool/internal/sbom"
)

// FetchSBOM represents the scan action.
type FetchSBOM struct {
	*Action
	SourceImage             string
	GatherArtifactReference string
	Output                  string
	Platforms               []string
}

// Run executes the sbom fetch command.
func (action *FetchSBOM) Run(ctx context.Context) error { //nolint:gocognit
	// load the config
	_ = action.Config.Get(ctx)
	log := logger.FromContext(ctx)
	var outfile io.WriteCloser
	// is Output stdout or a directory path?
	if action.Output == "-" {
		outfile = os.Stdout
	}
	var sbomDirectory string
	switch {
	case action.GatherArtifactReference != "":
		if action.Output != "-" {
			gatherSBOMDirName, tag, err := ParseOriginalReference(action.GatherArtifactReference)
			if err != nil {
				return err
			}
			directorySBOMs := filepath.Join(action.Output, gatherSBOMDirName, tag)
			if err := os.MkdirAll(directorySBOMs, 0775); err != nil {
				return fmt.Errorf("creating directory structure: %w", err)
			}
			sbomDirectory = directorySBOMs
		}
		endpointReference, err := action.Config.ParseEndpointReference(action.GatherArtifactReference)
		if err != nil {
			return fmt.Errorf("parsing endpoint reference: %w", err)
		}

		repository, err := action.Config.Repository(ctx, endpointReference.String())
		if err != nil {
			return err
		}

		var idx ocispec.Index
		_, idxRC, err := repository.FetchReference(ctx, endpointReference.String())
		if err != nil {
			return fmt.Errorf("fetching gather index: %w", err)
		}
		decoder := json.NewDecoder(idxRC)
		if err := decoder.Decode(&idx); err != nil {
			return fmt.Errorf("decoding the index reader: %w", err)
		}

		for _, manifest := range idx.Manifests {
			originalReference := manifest.Annotations[ref.AnnotationSrcRef]
			if action.SourceImage != "" && originalReference != action.SourceImage {
				continue
			}
			parsedRef, err := repository.ParseReference(manifest.Digest.String())
			if err != nil {
				return fmt.Errorf("parsing the manifest reference: %w", err)
			}
			sboms, err := sbom.FetchImageSBOM(ctx, parsedRef.String(), repository, action.Platforms)
			if err != nil {
				return err
			}
			if len(sboms) == 0 {
				log.InfoContext(ctx, "No SBOMs found", "reference", parsedRef.String())
				continue
			}
			var currentPath string
			if action.Output != "-" {
				baseImage, tagOrReference, err := ParseOriginalReference(manifest.Annotations[ref.AnnotationSrcRef])
				if err != nil {
					return err
				}
				log.InfoContext(ctx, "creating subdirectory", "name", baseImage)
				imageDir := filepath.Join(sbomDirectory, baseImage, tagOrReference)
				if err := os.MkdirAll(imageDir, 0775); err != nil {
					return fmt.Errorf("creating directory structure: %w", err)
				}
				currentPath = imageDir
			}
			for platform, manifest := range sboms {
				if platform == "" {
					platform = "unknown/unknown"
				}
				if action.Output != "-" {
					var architectureDirectory string
					ps := strings.Split(platform, "/")
					if len(ps) == 3 {
						architectureDirectory = strings.Join(ps[1:], "-")
					} else if len(ps) == 2 {
						architectureDirectory = ps[1]
					}
					platformDirectory := filepath.Join(currentPath, ps[0], architectureDirectory)
					if err := os.MkdirAll(platformDirectory, 0775); err != nil {
						return fmt.Errorf("creating directory structure: %w", err)
					}
					filename := filepath.Join(platformDirectory, "sbom.json")
					file, err := os.Create(filename)
					if err != nil {
						return fmt.Errorf("creating SBOM %s for platform %s: %w", originalReference, platform, err)
					}
					n, err := io.Copy(file, *manifest)
					if err != nil {
						return fmt.Errorf("writing sbom: %w", err)
					}
					if err := file.Close(); err != nil {
						return fmt.Errorf("closing the file: %w", err)
					}
					log.InfoContext(ctx, "Copied SBOM bytes", "image", originalReference, "platform", platform, "copiedBytes", n)
				}
				n, err := io.Copy(outfile, *manifest)
				if err != nil {
					return fmt.Errorf("writing sbom: %w", err)
				}
				log.InfoContext(ctx, "Copied SBOM bytes", "image", originalReference, "platform", platform, "copiedBytes", n)
			}
		}

	case action.SourceImage != "":
		endpointReference, err := action.Config.ParseEndpointReference(action.SourceImage)
		if err != nil {
			return fmt.Errorf("parsing endpoint reference: %w", err)
		}

		repository, err := action.Config.Repository(ctx, endpointReference.String())
		if err != nil {
			return err
		}
		image, tag, err := ParseOriginalReference(action.SourceImage)
		if err != nil {
			return err
		}

		platformSBOMs, err := sbom.FetchImageSBOM(ctx, endpointReference.String(), repository, action.Platforms)
		if err != nil {
			return err
		}

		if len(platformSBOMs) == 0 {
			_, err := fmt.Fprintf(os.Stdout, "No SBOMS were found for reference %s\n", action.SourceImage)
			if err != nil {
				return fmt.Errorf("printing to standard out that no SBOMS were found: %w", err)
			}
			return nil
		}

		if action.Output != "-" {
			directorySBOMs := filepath.Join(action.Output, image, tag)
			// create the directory for the SBOMs to go in
			if err := os.MkdirAll(directorySBOMs, 0775); err != nil {
				return fmt.Errorf("creating subdirectory %s for SBOMs: %w", directorySBOMs, err)
			}

			sbomDirectory = directorySBOMs
		}

		for platform, data := range platformSBOMs {
			var architectureDirectory string
			if action.Output != "-" {
				ps := strings.SplitAfterN(platform, "/", 1)
				if len(ps) == 3 {
					architectureDirectory = strings.Join(ps[1:], "-")
				} else if len(ps) == 2 {
					architectureDirectory = ps[1]
				}
				platformDirectory := filepath.Join(sbomDirectory, ps[0], architectureDirectory)
				if err := os.MkdirAll(platformDirectory, 0775); err != nil {
					return fmt.Errorf("creating directory structure: %w", err)
				}
				filename := filepath.Join(platformDirectory, "sbom.json")
				file, err := os.Create(filename)
				if err != nil {
					return fmt.Errorf("creating SBOM %s for platform %s: %w", action.SourceImage, platform, err)
				}
				outfile = file
			}
			_, err = io.Copy(outfile, *data)
			if err != nil {
				return fmt.Errorf("writing sbom: %w", err)
			}
		}
	}
	if action.Output != "-" {
		log.InfoContext(ctx, "SBOM's saved to directory", "name", sbomDirectory)
		_, err := fmt.Printf("SBOM's saved to directory %s\n", sbomDirectory)
		if err != nil {
			return fmt.Errorf("outputting sbom directory information: %w", err)
		}
	}

	if outfile != nil {
		if err := outfile.Close(); err != nil {
			return fmt.Errorf("closing the outfile: %w", err)
		}
	}

	return nil
}

// ParseOriginalReference splits a string reference into a repository and a tag or digest reference.
func ParseOriginalReference(reference string) (string, string, error) {
	var image string
	var tagOrDigest string

	s := strings.Split(reference, "@")
	if len(s) != 2 {
		d := strings.Split(reference, ":")
		if len(d) != 2 {
			return "", "", fmt.Errorf("invalid reference, expecting tag or digest: %s", reference)
		}
		image = d[0]
		tagOrDigest = d[1]
	} else {
		image = s[0]
		tagOrDigest = s[1]
	}
	return image, tagOrDigest, nil
}
