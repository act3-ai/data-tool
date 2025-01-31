package bottle

import (
	"context"
	"errors"
	"fmt"

	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/data/telemetry/v3/pkg/types"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
	tbtl "gitlab.com/act3-ai/asce/data/tool/internal/transfer/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	telem "gitlab.com/act3-ai/asce/data/tool/pkg/telemetry"
	tbottle "gitlab.com/act3-ai/asce/data/tool/pkg/transfer/bottle"
)

// Push represents the bottle push action.
type Push struct {
	*Action

	Telemetry   actions.TelemetryOptions
	Compression CompressionLevelOptions

	NoOverwrite bool // Only push data if if doesn't already exist
	NoDeprecate bool // Don't deprecate existing bottle

	Ref string
}

// Run runs the bottle push action.
func (action *Push) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	log.InfoContext(ctx, "bottle push command activated")

	cfg, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	// first we must commit, this saves everything: manifest, config, archived parts, etc.
	log.InfoContext(ctx, "committing bottle")
	if err := commit(ctx, cfg, btl, action.NoDeprecate); err != nil {
		return err
	}

	if action.NoOverwrite {
		log.InfoContext(ctx, "checking for existing bottles")
		parsedref := ref.RepoFromString(action.Ref)
		repo, err := action.Config.Repository(ctx, parsedref.String())
		if err != nil {
			return fmt.Errorf("creating repository reference: %w", err)
		}

		_, err = repo.Resolve(ctx, repo.Reference.Reference)
		switch {
		case errors.Is(err, errdef.ErrNotFound):
			// continue
		case err != nil:
			// unwanted error
			return fmt.Errorf("checking if reference exists: %w", err)
		default:
			// reference already exists
			return fmt.Errorf("bottle reference %s already exists. Please choose another repository or tag before pushing. %w",
				action.Ref, err)
		}

	}

	log.InfoContext(ctx, "pushing bottle with signatures")
	pushOpts := tbtl.PushOptions{
		TransferOptions: tbottle.TransferOptions{
			Concurrency: cfg.ConcurrentHTTP,
			CachePath:   cfg.CachePath,
		},
	}
	if err := tbtl.PushBottle(ctx, btl, action.Config, action.Ref, pushOpts); err != nil {
		return fmt.Errorf("pushing bottle and signatures: %w", err)
	}

	r, err := ref.FromString(action.Ref)
	if err != nil {
		return fmt.Errorf("invalid bottle reference %s: %w", action.Ref, err)
	}

	// Handle telemetry
	rawManifest, err := btl.Manifest.GetManifestRaw()
	if err != nil {
		return fmt.Errorf("getting bottle manifest: %w", err)
	}

	telemAdapt := telem.NewAdapter(ctx, cfg.Telemetry, cfg.TelemetryUserName, telem.WithCredStore(action.Config.CredStore()))
	event := telemAdapt.NewEvent(r.String(), rawManifest, types.EventPush)

	telemUrls, err := telemAdapt.NotifyTelemetry(ctx, btl.GetCache(), btl.Manifest.GetManifestDescriptor(), action.Dir, event)
	if err != nil {
		return fmt.Errorf("notifying telemetry: %w", err)
	}

	rootUI.Info(formatBottleURLs(telemUrls))
	rootUI.Infof("Bottle push complete.  BottleID: %s\n", btl.GetBottleID())
	return nil
}
