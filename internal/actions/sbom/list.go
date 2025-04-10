package sbom

import (
	"context"
	"fmt"
	"os"

	sbom "github.com/act3-ai/data-tool/internal/sbom"
	"github.com/act3-ai/data-tool/internal/security"
)

// List represents the sbom list action.
type List struct {
	*Action
	GatherArtifactReference string
	SourceImage             string
	Platforms               []string
}

// Run executes the sbom list command.
func (action *List) Run(ctx context.Context) error {
	cfg := action.Config.Get(ctx)
	if action.GatherArtifactReference == "" {
		return fmt.Errorf("use --gathered-image to define your artifact")
	}
	repository, err := action.Config.Repository(ctx, action.GatherArtifactReference)
	if err != nil {
		return err
	}
	list, err := sbom.GetListofSBOMS(ctx, action.GatherArtifactReference, action.SourceImage, repository, cfg.ConcurrentHTTP, action.Platforms)
	if err != nil {
		return err
	}

	return security.PrintCustomTable(os.Stdout, list)
}
