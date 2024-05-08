package bottle

import (
	"context"
	"fmt"
	"io"
	"strings"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// PartLabel represents the bottle part label action.
type PartLabel struct {
	*Action
}

// Run runs the bottle part label action.
func (action *PartLabel) Run(ctx context.Context, partPaths []string, labels []string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "part label command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	for _, lbl := range labels {
		if isLabelDelete(lbl) {
			for _, p := range partPaths {
				if err := btl.RemovePartLabel(ctx, strings.TrimSuffix(lbl, "-"), p); err != nil {
					return err
				}
			}
		} else {
			key, value, err := parseLabel(lbl)
			if err != nil {
				return err
			}
			for _, partPath := range partPaths {
				if err = btl.AddPartLabel(ctx, key, value, partPath); err != nil {
					return fmt.Errorf("failed to add label to part: %w", err)
				}
			}
		}
	}

	log.InfoContext(ctx, "part label command completed")
	return saveMetaChanges(ctx, btl)
}

func isLabelDelete(lbl string) bool {
	return strings.HasSuffix(lbl, "-")
}
