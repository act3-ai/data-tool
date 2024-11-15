package sbom

import (
	"context"
	"fmt"
	"slices"

	"oras.land/oras-go/v2/registry/remote"

	"git.act3-ace.com/ace/data/tool/internal/security"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// GetListofSBOMS accepts a mirror gather artifact and/or image string reference and creates a slice of string slices
// of references and SBOMS that are compatible with the table printer functions.
func GetListofSBOMS(ctx context.Context,
	artifact, image string,
	repository *remote.Repository,
	concurrency int,
	platforms []string) ([][]string, error) {

	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "fetching artifact information", "reference", artifact)
	m, err := security.FormatSources(ctx, "", artifact, repository, concurrency)
	if err != nil {
		return nil, fmt.Errorf("extracting sources from artifact: %w", err)
	}
	sbomList := [][]string{{"reference", "platform", "sbom manifest digest", "artifact-type"}}
	for _, entry := range m {
		reference := entry[0]
		source := entry[1]
		if image != "" && reference != image {
			continue
		}
		ad, err := security.GetArtifactDetails(ctx, source, repository)
		if err != nil {
			return nil, err
		}
		for platform, sboms := range ad.SBOMs {
			if len(platforms) != 0 && !slices.Contains(platforms, platform) {
				continue
			}
			for _, sbom := range sboms {
				sbomList = append(sbomList, []string{reference, platform, sbom.Digest.String(), sbom.ArtifactType})
			}
		}
	}
	return sbomList, nil
}
