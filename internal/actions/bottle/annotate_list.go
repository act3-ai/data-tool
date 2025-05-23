package bottle

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/act3-ai/data-tool/internal/actions"
	"github.com/act3-ai/data-tool/internal/actions/internal/format"
	"github.com/act3-ai/data-tool/internal/bottle"
	"github.com/act3-ai/go-common/pkg/logger"
)

// AnnotateList represents the bottle annotate list action.
type AnnotateList struct {
	*Action

	Telemetry actions.TelemetryOptions

	Selectors []string // Selectors for parts to retrieve
	Parts     []string // Titles of parts to retrieve
	Artifacts []string // Filter by public artifact type
	Insecure  bool     // Enables insecure communication with a registry if https fails with http response
	Empty     bool     // Only annotatelist metadata from the bottle
}

// Run runs the bottle annotate list action.
func (action *AnnotateList) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "annotation list command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "listing annotations of the bottle", "path", action.Dir)

	if err := listBottleAnnotations(ctx, btl, out); err != nil {
		return fmt.Errorf("failed while trying to list annotation on bottle: %w", err)
	}

	log.InfoContext(ctx, "annotate command completed")
	return nil
}

func listBottleAnnotations(ctx context.Context, btl *bottle.Bottle, out io.Writer) error {
	log := logger.FromContext(ctx)
	if len(btl.Definition.Annotations) == 0 {
		log.InfoContext(ctx, "bottle has no annotations")
		return nil
	}

	t := format.NewTable()
	t.AddRow(strings.ToTitle("bottle annotation(s)"))
	for key, val := range btl.Definition.Annotations {
		t.AddRow(fmt.Sprintf(" %v: %v", key, val))
	}

	_, err := fmt.Fprintln(out, t.String())
	if err != nil {
		return err
	}

	return nil
}
