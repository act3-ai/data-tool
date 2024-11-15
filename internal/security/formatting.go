package security

import (
	"context"
	"fmt"
	"path"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote"

	"git.act3-ace.com/ace/data/schema/pkg/selectors"
	"git.act3-ace.com/ace/data/tool/internal/mirror"
)

func formatPlatformString(platform *ocispec.Platform) string {
	if platform != nil {
		return path.Join(platform.OS, platform.Architecture, platform.Variant)
	}
	return ""
}

// FormatSources is a formatting helper function that parses the sources in a sourcefile or gather artifact and returns the source and originating reference.
func FormatSources(ctx context.Context, sourceFile, gatherArtifact string, repo *remote.Repository, concurrency int) ([][]string, error) {
	sourceReference := [][]string{}
	if sourceFile != "" {
		sources, err := mirror.ProcessSourcesFile(ctx, sourceFile, selectors.LabelSelectorSet{}, concurrency)
		if err != nil {
			return nil, err
		}
		for _, source := range sources {
			sourceReference = append(sourceReference, []string{source.Name, source.Name})
			// this seems redundant but it saves a lot of code
		}
	}
	if gatherArtifact != "" {
		m, err := extractSourcesFromMirrorArtifact(ctx, gatherArtifact, repo)
		if err != nil {
			return nil, fmt.Errorf("extracting sources from artifact: %w", err)
		}
		sourceReference = m
	}
	return sourceReference, nil
}
