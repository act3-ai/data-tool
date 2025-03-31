package bottle

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	telem "gitlab.com/act3-ai/asce/data/tool/pkg/telemetry"
	tbottle "gitlab.com/act3-ai/asce/data/tool/pkg/transfer/bottle"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// Pull represents the bottle pull action.
type Pull struct {
	*Action

	Telemetry    actions.TelemetryOptions
	PartSelector bottle.PartSelectorOptions
}

// Run runs the bottle pull action.
func (action *Pull) Run(ctx context.Context, bottleRef string) error {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	cfg := action.Config.Get(ctx)
	telemAdapt := telem.NewAdapter(ctx, cfg.Telemetry, cfg.TelemetryUserName, telem.WithCredStore(action.Config.CredStore()))

	log.InfoContext(ctx, "resolving reference with telemetry", "ref", bottleRef)
	transferOpts := tbottle.TransferOptions{
		Concurrency: cfg.ConcurrentHTTP,
		CachePath:   cfg.CachePath,
	}
	src, desc, event, err := telemAdapt.ResolveWithTelemetry(ctx, bottleRef, action.Config, transferOpts)
	if err != nil {
		return fmt.Errorf("resolving bottle reference: %w", err)
	}

	log.InfoContext(ctx, "pulling bottle", "reference", bottleRef, "pullPath", action.Dir)
	pullOpts := tbottle.PullOptions{
		TransferOptions: tbottle.TransferOptions{
			Concurrency: cfg.ConcurrentHTTP,
			CachePath:   cfg.CachePath,
		},
		PartSelectorOptions: action.PartSelector,
	}
	err = tbottle.Pull(ctx, src, desc, action.Dir, pullOpts)
	if err != nil {
		return fmt.Errorf("pulling bottle: %w", err)
	}

	log.InfoContext(ctx, "notifying telemetry")
	telemURLs, err := telemAdapt.NotifyTelemetry(ctx, src, desc, action.Dir, event)
	if err != nil {
		return fmt.Errorf("notifying telemetry: %w", err)
	}

	rootUI.Info(formatBottleURLs(telemURLs))

	return nil
}

func formatBottleURLs(urls []string) string {
	if len(urls) == 0 {
		return "Telemetry servers not notified.  Consider adding one or more telemetry servers to your configuration file."
	}

	items := make([]string, len(urls)+1)
	items[0] = "Successfully notified the telemetry server(s).  To view the bottle, browse to any of the following locations:"

	for i, location := range urls {
		items[i+1] = location
	}

	return strings.Join(items, "\n\t")
}
