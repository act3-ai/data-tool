package bottle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"oras.land/oras-go/v2/errdef"

	latest "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io/v1"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
	sigcustom "gitlab.com/act3-ai/asce/data/tool/internal/sign"

	"gitlab.com/act3-ai/asce/data/telemetry/pkg/types"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// Pull facilitates bottle transfers from a remote registry to a local directory. The
// transfer configuration provides options for modifying the pull process, such as
// discovering bottle locations from telemetry hosts.
func Pull(ctx context.Context, opts TransferConfig) ([]string, error) {
	// This function is called by the CSI bottle driver so do not change it needlessly.
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "verifying pull directory", "pullPath", opts.PullPath)
	if err := bottle.VerifyPullDir(opts.PullPath); err != nil {
		return nil, fmt.Errorf("invalid pull directory: %w", err)
	}

	// Check that there is a valid oci reference specified in the options
	r, err := ref.FromString(opts.Reference)
	if err != nil {
		return nil, fmt.Errorf("invalid input argument %s: %w", opts.Reference, err)
	}

	telem := bottle.NewTelemetryAdapter(opts.telemHosts, opts.telemUserName)

	opts.postConfigHandler = saveBottleIDFromConfig(ctx, defaultConfigHandler, opts.bottleIDFile)

	refList := []ref.Ref{r}
	if r.Scheme == ref.SchemeBottle {
		log.InfoContext(ctx, "Discovering bottle locations")
		tempRefList, err := telem.FindBottle(ctx, r)
		if err != nil {
			return nil, fmt.Errorf("unable to find bottle using telemetry server(s): %w", err)
		}
		opts.matchBottleID = r.Digest
		refList = tempRefList
		opts.preConfigHandler = ensureBottleVersion
	}

	if len(refList) == 0 {
		return nil, errdef.ErrNotFound
	}

	var btl *bottle.Bottle
	var ignorableErrors []error

	logger.V(log, 1).InfoContext(ctx, "Trying references", "references", refList)
	for _, r := range refList {
		log.InfoContext(ctx, "Trying to pull bottle", "ref", r)
		opts.Reference = r.String() // drop the scheme for oras

		// attempt to connect to the repository using options loaded from ace-dt config
		repo, err := opts.NewGraphTargetFn(ctx, opts.Reference)
		if err != nil {
			ignorableErrors = append(ignorableErrors, fmt.Errorf("creating repository reference '%s': %w", opts.Reference, err))
			continue
		}

		if btl, err = pull(ctx, repo, opts); err != nil {
			ignorableErrors = append(ignorableErrors, fmt.Errorf("pulling bottle: %w", err))
			continue
		}

		// pull signatures
		btlManDesc := btl.Manifest.GetManifestDescriptor()
		if err := sigcustom.Pull(ctx, btl.GetPath(), repo, btlManDesc); err != nil {
			return nil, fmt.Errorf("pulling bottle signatures: %w", err) // TODO: is a signature pull failure fatal here? Perhaps a good transfer option?
		}

		// Increase telemetry log verbosity level
		telemURLs, err := telem.SendTelemetry(logger.NewContext(ctx, logger.V(log, 1)), btl, r, types.EventPull)
		if err != nil {
			log.ErrorContext(ctx, "Failed to send telemetry", "error", err.Error())
			return nil, ErrTelemetrySend
		}

		summary, err := sigcustom.NewSummaryFromBottle(ctx, btl)
		if err != nil {
			log.ErrorContext(ctx, "Failed to generate signature summary message", "error", err.Error())
			return nil, fmt.Errorf("generating signature detail message: %w", err)
		}
		if summary != nil {
			err = telem.SendSignatures(logger.NewContext(ctx, logger.V(log, 1)), summary)
			if err != nil {
				return nil, err
			}
		}

		// whichever ref works first is sufficient
		return telemURLs, nil
	}

	// return error if no pull was successful
	return nil, errors.Join(ignorableErrors...)
}

// ensureBottleVersion checks the bottle version.
// This is called when the bottle ID is used to pull the bottle (i.e., we don't trust the manifest).
func ensureBottleVersion(btl *bottle.Bottle, rawConfig []byte) error {
	// pulled by bottle ID
	typeMeta := metav1.TypeMeta{}
	if err := json.Unmarshal(rawConfig, &typeMeta); err != nil {
		// returning an insecure archive error here since we don't know for sure if the version is sufficient.  Note
		// this error is unlikely to occur since a decode error will likely have already occurred for this config
		return fmt.Errorf("json decode error, unable to verify bottle version. %w", bottle.ErrPartInsecureArchive)
	}

	if typeMeta.GroupVersionKind() != latest.GroupVersion.WithKind("Bottle") {
		return fmt.Errorf("cannot pull by bottle id, use manifest digest or tag directly.  There is a security concern with pulling by this version of the bottle by bottle ID: %w", bottle.ErrPartInsecureArchive)
	}

	return nil
}

// SaveBottleIDFromConfig intercepts the config handle callback during a query, enabling the bottle id to be saved, and
// forwarding to a specified config handler function.
func saveBottleIDFromConfig(ctx context.Context, cfgHandler configHandlerFn, writeBottleIDFile string) configHandlerFn {
	return func(btl *bottle.Bottle, rawConfig []byte) error {
		if writeBottleIDFile != "" {
			if err := bottle.SaveBottleIDFile(btl, writeBottleIDFile); err != nil {
				return err
			}
		}
		return cfgHandler(btl, rawConfig)
	}
}

// ErrTelemetrySend indicates that a telemetry send event was not properly
// reported.
var ErrTelemetrySend = errors.New("telemetry reporting delegate failed")
