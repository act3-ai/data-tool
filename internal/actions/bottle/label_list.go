package bottle

import (
	"context"
	"fmt"
	"io"

	"gitlab.com/act3-ai/asce/data/tool/internal/actions/internal/format"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// LabelList represents the bottle label list action.
type LabelList struct {
	*Action
}

// Run runs the bottle label list action.
func (action *LabelList) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "bottle label command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "listing labels of bottle", "path", btl.GetPath())
	return listBottleLabels(ctx, btl, out)
}

func listBottleLabels(ctx context.Context, btl *bottle.Bottle, out io.Writer) error {
	log := logger.FromContext(ctx)
	if len(btl.Definition.Labels) == 0 {
		log.InfoContext(ctx, "bottle has no labels")
		return nil
	}

	t := format.NewTable()
	t.AddRow("BOTTLE LABEL(S)")

	for key, val := range btl.Definition.Labels {
		t.AddRow(fmt.Sprintf(" %v=%v", key, val))
	}

	_, err := fmt.Fprintln(out, t.String())
	if err != nil {
		return err
	}

	return nil
}
