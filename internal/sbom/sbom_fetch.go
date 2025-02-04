// Package sbom contains the logic for fetching and listing SBOMs needed for ace-dt sbom subcommands.
package sbom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"

	security "git.act3-ace.com/ace/data/tool/internal/security"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// FetchImageSBOM accepts a reference, repository, and optional platforms and returns a map of platform-->SBOM ReadCloser for printing to stdout or saving to file.
func FetchImageSBOM(ctx context.Context, ref string, repository oras.GraphTarget, platforms []string) (map[string]*io.ReadCloser, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "fetching SBOM", "reference", ref)
	ad, err := security.GetArtifactDetails(ctx, ref, repository)
	if err != nil {
		return nil, err
	}

	data := map[string]*io.ReadCloser{}

	for platform, sbomManifests := range ad.SBOMs {
		if len(platforms) != 0 && !slices.Contains(platforms, platform) {
			continue
		}
		for _, sbomManifest := range sbomManifests {
			var m ocispec.Manifest
			rc, err := repository.Fetch(ctx, *sbomManifest)
			if err != nil {
				return nil, fmt.Errorf("fetching the SBOM manifest: %w", err)
			}
			d := json.NewDecoder(rc)
			if err := d.Decode(&m); err != nil {
				return nil, fmt.Errorf("decoding SBOM manifest %w", err)
			}
			// get the sbom layer
			for _, layer := range m.Layers {
				brc, err := repository.Fetch(ctx, layer)
				if err != nil {
					return nil, fmt.Errorf("fetching SBOM: %w", err)
				}
				data[platform] = &brc
			}
		}

	}
	return data, nil
}
