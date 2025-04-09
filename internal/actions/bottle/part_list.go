package bottle

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/act3-ai/data-tool/internal/actions/internal/format"
	"github.com/act3-ai/data-tool/internal/bottle"
	"github.com/act3-ai/data-tool/internal/print"
	"github.com/act3-ai/go-common/pkg/logger"
)

// PartList represents the bottle part list action.
type PartList struct {
	*Action

	WithDigest bool // Show parts information with digest
}

// Run runs the bottle part list action.
func (action *PartList) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "part list command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	if btl.NumParts() == 0 {
		log.InfoContext(ctx, "bottle has no parts to show", "path", action.Dir)
		return nil
	}

	if action.WithDigest {
		_, err := fmt.Fprintln(out, formatPartsWithDigest(btl))
		if err != nil {
			return err
		}
	} else {
		_, err := fmt.Fprintln(out, formatPartsWithLabels(btl))
		if err != nil {
			return err
		}
	}

	log.InfoContext(ctx, "part list command completed")

	return nil
}

func formatPartsWithDigest(btl *bottle.Bottle) string {
	t := format.NewTable()
	t.AddRow("NAME", "SIZE", "DIGEST")

	for _, part := range btl.GetParts() {
		t.AddRow(part.GetName(), print.Bytes(part.GetContentSize()), part.GetContentDigest())
	}

	return t.String()
}

func formatPartsWithLabels(btl *bottle.Bottle) string {
	t := format.NewTable()
	t.AddRow("PART", "SIZE", "LABELS")

	for _, part := range btl.GetParts() {
		// create and format label string
		labelStr := ""
		for k, v := range part.GetLabels() {
			labelStr += fmt.Sprintf("%s=%s, ", k, v)
		}
		labelStr = strings.TrimSuffix(labelStr, ", ")

		t.AddRow(part.GetName(), print.Bytes(part.GetContentSize()), labelStr)
	}

	return t.String()
}
