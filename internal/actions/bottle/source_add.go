package bottle

import (
	"context"
	"fmt"
	"io"

	"github.com/opencontainers/go-digest"

	latest "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io/v1"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	telem "gitlab.com/act3-ai/asce/data/tool/pkg/telemetry"
	tbottle "gitlab.com/act3-ai/asce/data/tool/pkg/transfer/bottle"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// SourceAdd represents the bottle source add action.
type SourceAdd struct {
	*Action

	// BottleForURI is true when the srcURI should be interpreted as a bottle directory (to extract the bottle ID from)
	BottleForURI bool
	// ReferenceForURI is true when the srcURI should be interpreted as a bottle reference
	ReferenceForURI bool
}

// Run runs the bottle source add action.
func (action *SourceAdd) Run(ctx context.Context, srcName, srcURI string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "source add command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	// check for uri = bottle path. Then replace given path with bottleID at given bottle path
	// alternatively we could parse the URL and inspect the "scheme"
	if action.BottleForURI {
		log.InfoContext(ctx, "bottle path found for URI, checking for bottleID")
		srcBottlePath := srcURI
		// we do not want to "Load and upgrade" the bottle here because if the source bottle has been modified on disk (a file added) that would provide a bottle ID that is not in the telemetry server.
		// we only want the last "pushed/pulled" bottle ID to reference as the source
		// alternatively we could warn the user if the source bottle changed
		bottleID, err := bottle.ReadBottleIDFile(srcBottlePath)
		if err != nil {
			return fmt.Errorf("no bottle found in given source bottle %s: %w", srcURI, err)
		}
		srcURI = "bottle:" + bottleID.String()
	} else if action.ReferenceForURI {

		// if the reference follows the bottle scheme, then resolving with telemetry is not necessary as we should
		// be able to directly add the source. However, we resolve it anyways as if telemetry cannot find it, adding
		// the source doesn't achieve anything; perhaps an indication of telemetry data loss or an invalid bottle source reference
		cfg := action.Config.Get(ctx)
		telemAdapt := telem.NewAdapter(ctx, cfg.Telemetry, cfg.TelemetryUserName, telem.WithCredStore(action.Config.CredStore()))

		log.InfoContext(ctx, "resolving reference with telemetry", "ref", srcURI)
		transferOpts := tbottle.TransferOptions{
			Concurrency: cfg.ConcurrentHTTP,
			CachePath:   cfg.CachePath,
		}
		src, desc, _, err := telemAdapt.ResolveWithTelemetry(ctx, srcURI, action.Config, transferOpts)
		if err != nil {
			return fmt.Errorf("resolving bottle reference: %w", err)
		}

		pullOpts := tbottle.PullOptions{
			PartSelectorOptions: tbottle.PartSelectorOptions{
				Empty: true,
			},
		}
		cfgBytes, _, err := tbottle.FetchBottleMetadata(ctx, src, desc, pullOpts)
		if err != nil {
			return fmt.Errorf("fetching bottleID from registry reference: %w", err)
		}
		// we don't notify telemetry, in favor of notifying on push

		srcURI = "bottle:" + digest.FromBytes(cfgBytes).String()

	}

	// Create source info
	inputSrc := latest.Source{
		Name: srcName,
		URI:  srcURI,
	}

	// add sources info
	if err = btl.AddSourceInfo(inputSrc); err != nil {
		return err
	}

	log.InfoContext(ctx, "Saving bottle with added source", "name", inputSrc.Name, "uri",
		inputSrc.URI)

	return saveMetaChanges(ctx, btl)
}
