package bottle

import (
	"context"
	"fmt"
	"io"

	"git.act3-ace.com/ace/data/tool/internal/actions/internal/format"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// MetricList represents the bottle metric list action.
type MetricList struct {
	*Action
}

// Run runs the bottle metric list action.
func (action *MetricList) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	metrics := btl.Definition.Metrics
	if len(metrics) == 0 {
		log.InfoContext(ctx, "bottle has no metrics to show", "path", action.Dir)
		return nil
	}

	t := format.NewTable()
	t.AddRow("METRIC", "DESCRIPTION")

	for _, m := range btl.Definition.Metrics {
		if m.Name != "" && m.Value != "" {
			t.AddRow(fmt.Sprintf("%v=%v", m.Name, m.Value), m.Description)
		}
	}

	_, err = fmt.Fprintln(out, t.String())
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "metric list command completed")
	return nil
}
