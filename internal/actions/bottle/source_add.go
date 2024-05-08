package bottle

import (
	"context"
	"fmt"
	"io"

	"github.com/opencontainers/go-digest"

	latest "git.act3-ace.com/ace/data/schema/pkg/apis/data.act3-ace.io/v1"
	"git.act3-ace.com/ace/data/tool/internal/bottle"
	tbtl "git.act3-ace.com/ace/data/tool/pkg/transfer/bottle"
	"git.act3-ace.com/ace/go-common/pkg/logger"
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
		log.InfoContext(ctx, "bottle reference found for URI, requesting bottleID")

		// build transfer options
		opts := []tbtl.TransferOption{
			tbtl.WithNoParts(),
		}

		transferCfg := tbtl.NewTransferConfig(ctx, srcURI, action.Dir, action.Config, opts...)

		cfgBytes, _, err := tbtl.FetchBottleMetadata(ctx, *transferCfg)
		if err != nil {
			return fmt.Errorf("fetching bottleID from registry reference: %w", err)
		}

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
