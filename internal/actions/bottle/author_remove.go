package bottle

import (
	"context"
	"fmt"
	"io"

	"github.com/act3-ai/go-common/pkg/logger"
)

// AuthorRemove represents the bottle author remove action.
type AuthorRemove struct {
	*Action
}

// Run runs the bottle author remove action.
func (action *AuthorRemove) Run(ctx context.Context, authorName string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "author remove command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	err = btl.RemoveAuthorInfo(authorName)
	if err != nil {
		return fmt.Errorf("failed to remove author %s: %w", authorName, err)
	}

	log.InfoContext(ctx, "removing specified author from bottle", "name", authorName, "bottlePath", btl.GetPath())
	log.InfoContext(ctx, "Saving bottle with specified author removed")

	return saveMetaChanges(ctx, btl)
}
