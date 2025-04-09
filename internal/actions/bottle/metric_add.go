package bottle

import (
	"context"
	"io"

	"github.com/act3-ai/go-common/pkg/logger"

	latest "github.com/act3-ai/bottle-schema/pkg/apis/data.act3-ace.io/v1"
)

// MetricAdd represents the bottle metric add action.
type MetricAdd struct {
	*Action

	Description string // Description of metric
}

// Run runs the bottle metric add action.
func (action *MetricAdd) Run(ctx context.Context, metricName, metricValue string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "metric add command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	addedMetric := latest.Metric{
		Name:        metricName,
		Description: action.Description,
		Value:       metricValue,
	}

	log.InfoContext(ctx, "Adding specified metric to bottle", "name", addedMetric.Name,
		"value", addedMetric.Value)

	// Add metric to bottle definition
	err = btl.AddMetricInfo(addedMetric)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "Saving bottle with updated metric")

	return saveMetaChanges(ctx, btl)
}
