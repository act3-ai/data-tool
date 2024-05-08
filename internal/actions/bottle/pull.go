package bottle

import (
	"context"
	"strings"

	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	tbtl "gitlab.com/act3-ai/asce/data/tool/pkg/transfer/bottle"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// Pull represents the bottle pull action.
type Pull struct {
	*Action

	Write        WriteBottleOptions
	Telemetry    actions.TelemetryOptions
	PartSelector bottle.PartSelectorOptions

	CheckBottleID string // bottle id in string format (e.g. sha256:abcdef01234...). If not empty, this value is checked against an incoming bottle to ensure they match.
}

// Run runs the bottle pull action.
func (action *Pull) Run(ctx context.Context, bottleRef string) error {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	// build transfer options
	opts := []tbtl.TransferOption{
		tbtl.WithBottleIDFile(action.Write.BottleIDFile),
	}

	// part selection options
	if action.PartSelector.Empty {
		opts = append(opts, tbtl.WithNoParts())
	} else {
		opts = append(opts, tbtl.WithPartSelection(action.PartSelector.Selectors,
			action.PartSelector.Parts,
			action.PartSelector.Artifacts))
	}

	transferCfg := tbtl.NewTransferConfig(ctx, bottleRef, action.Dir, action.Config, opts...)

	log.InfoContext(ctx, "pulling with telemetry", "reference", bottleRef, "pullPath", action.Dir)
	telemURLs, err := tbtl.Pull(ctx, *transferCfg)
	if err != nil {
		return err
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
