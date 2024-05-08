package bottle

import (
	"context"
	"fmt"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// Delete represents the bottle delete action.
type Delete struct {
	*Action
	Ref string
}

// Run runs the bottle delete action.
func (action *Delete) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)

	repo, err := action.Config.ConfigureRepository(ctx, action.Ref)
	if err != nil {
		return err
	}

	target, err := repo.Resolve(ctx, repo.Reference.ReferenceOrDefault())
	if err != nil {
		return fmt.Errorf("resolving image reference failed: %w", err)
	}

	if err := repo.Delete(ctx, target); err != nil {
		return fmt.Errorf("deleting remote content: %w", err)
	}

	log.InfoContext(ctx, "bottle delete command completed")
	return nil
}
