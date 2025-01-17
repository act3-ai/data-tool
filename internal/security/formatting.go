package security

import (
	"context"
	"fmt"
	"path"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"

	"github.com/act3-ai/bottle-schema/pkg/selectors"
	"github.com/act3-ai/data-tool/internal/mirror"
	dtreg "github.com/act3-ai/data-tool/internal/registry"
	reg "github.com/act3-ai/data-tool/pkg/registry"
)

func formatPlatformString(platform *ocispec.Platform) string {
	if platform != nil {
		return path.Join(platform.OS, platform.Architecture, platform.Variant)
	}
	return ""
}

// FormatSources is a formatting helper function that parses the sources in a sourcefile or gather artifact and returns the source and originating reference.
func FormatSources(ctx context.Context, sourceFile, gatherArtifact string, repository oras.GraphTarget, targeter reg.GraphTargeter, concurrency int) ([]Source, error) {
	sourceReference := []Source{}
	if sourceFile != "" {
		sources, err := mirror.ProcessSourcesFile(ctx, sourceFile, selectors.LabelSelectorSet{}, concurrency)
		if err != nil {
			return nil, err
		}
		for _, source := range sources {
			// get the true reference
			artifactReference, err := dtreg.ParseEndpointOrDefault(targeter, source.Name)
			if err != nil {
				return nil, err
			}
			sourceReference = append(sourceReference, Source{source.Name, artifactReference.String()})

		}
	}
	if gatherArtifact != "" {
		reference, err := dtreg.ParseEndpointOrDefault(targeter, gatherArtifact)
		if err != nil {
			return nil, err
		}
		m, err := extractSourcesFromMirrorArtifact(ctx, reference, repository)
		if err != nil {
			return nil, fmt.Errorf("extracting sources from artifact: %w", err)
		}
		sourceReference = m
	}
	return sourceReference, nil
}
