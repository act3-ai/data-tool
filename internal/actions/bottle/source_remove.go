package bottle

import (
	"context"
	"fmt"
	"io"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// SourceRemove represents the bottle source remove action.
type SourceRemove struct {
	*Action
}

// Run runs the bottle source remove action.
func (action *SourceRemove) Run(ctx context.Context, srcName string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "source remove command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	err = btl.RemoveSourceInfo(srcName)
	if err != nil {
		return fmt.Errorf("failed while trying to remove a source: %w", err)
	}

	log.InfoContext(ctx, "removed specified source from bottle", "name", srcName, "bottlePath", btl.GetPath())

	return saveMetaChanges(ctx, btl)
}
