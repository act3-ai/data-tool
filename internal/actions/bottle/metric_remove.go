package bottle

import (
	"context"
	"fmt"
	"io"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// MetricRemove represents the bottle metric remove action.
type MetricRemove struct {
	*Action
}

// Run runs the bottle metric remove action.
func (action *MetricRemove) Run(ctx context.Context, metricName string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "metric remove command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	err = btl.RemoveMetricInfo(metricName)
	if err != nil {
		return fmt.Errorf("failed to remove metric %s: %w", metricName, err)
	}

	log.InfoContext(ctx, "removed specified metric from bottle", "name", metricName, "bottlePath", btl.GetPath())
	log.InfoContext(ctx, "Saving bottle with specified source removed")

	return saveMetaChanges(ctx, btl)
}
