package bottle

import (
	"context"
	"fmt"
	"io"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/apimachinery/pkg/labels"

	"git.act3-ace.com/ace/data/schema/pkg/mediatype"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/actions/internal/format"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/oci"
	"gitlab.com/act3-ai/asce/data/tool/internal/print"
	telem "gitlab.com/act3-ai/asce/data/tool/pkg/telemetry"
	tbottle "gitlab.com/act3-ai/asce/data/tool/pkg/transfer/bottle"
)

// Show represents the bottle show action.
type Show struct {
	*Action

	PartSelector bottle.PartSelectorOptions
	Ref          string
}

// Run runs the bottle show action.
func (action *Show) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "bottle show command activated")

	cfg := action.Config.Get(ctx)

	// check if input is a reference, else fall back to bottle directory
	var btl *bottle.Bottle
	if action.Ref != "" {

		telemAdapt := telem.NewAdapter(ctx, cfg.Telemetry, cfg.TelemetryUserName, telem.WithCredStore(action.Config.CredStore()))

		log.InfoContext(ctx, "resolving reference with telemetry", "ref", action.Ref)
		transferOpts := tbottle.TransferOptions{
			Concurrency: cfg.ConcurrentHTTP,
			CachePath:   cfg.CachePath,
		}
		src, desc, event, err := telemAdapt.ResolveWithTelemetry(ctx, action.Ref, action.Config, transferOpts)
		if err != nil {
			return fmt.Errorf("resolving bottle reference: %w", err)
		}

		log.InfoContext(ctx, "fetching bottle metdata")
		pullOpts := tbottle.PullOptions{
			TransferOptions: tbottle.TransferOptions{
				Concurrency: cfg.ConcurrentHTTP,
				CachePath:   cfg.CachePath,
			},
			PartSelectorOptions: action.PartSelector,
		}
		cfgBytes, manBytes, err := tbottle.FetchBottleMetadata(ctx, src, desc, pullOpts)
		if err != nil {
			return fmt.Errorf("fetching bottle metadata: %w", err)
		}

		log.InfoContext(ctx, "notifying telemetry")
		_, err = telemAdapt.NotifyTelemetry(ctx, src, desc, action.Dir, event)
		if err != nil {
			return fmt.Errorf("notifying telemetry: %w", err)
		}

		manifestHandler := oci.ManifestFromData(ocispec.MediaTypeImageManifest, manBytes)
		if manifestHandler.GetStatus().Error != nil {
			return fmt.Errorf("constructing manifest handler from raw manifest: %w", err)
		}

		log.InfoContext(ctx, "Configuring local bottle")
		btl, err = bottle.NewBottle(
			bottle.WithLocalPath(action.Dir),
			bottle.DisableDestinationCreate(true),
			bottle.DisableCache(true),
		)
		if err != nil {
			return fmt.Errorf("bottle initialization failed: %w", err)
		}
		btl.SetManifest(manifestHandler)

		if err := btl.Configure(cfgBytes); err != nil {
			return fmt.Errorf("configuring bottle: %w", err)
		}
	} else {
		// Check if the supplied path is nested within a bottle by finding the root bottle
		rootPath, err := bottle.FindBottleRootDir(action.Dir)
		if err != nil {
			return err
		}
		action.Dir = rootPath

		log.InfoContext(ctx, "loading bottle information from path", "path", action.Dir)
		btl, err = bottle.LoadBottle(action.Dir,
			bottle.WithCachePath(cfg.CachePath),
			bottle.WithLocalLabels(),
		)
		if err != nil {
			return err
		}
	}

	partSelector, err := action.PartSelector.New(ctx)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, prettyPrintBtlInfo(btl, partSelector)); err != nil {
		return err
	}

	log.InfoContext(ctx, "bottle show command completed")
	return nil
}

// prettyPrintBtlInfo formats and prints bottle information.
func prettyPrintBtlInfo(btl *bottle.Bottle, partSelector bottle.PartSelectorFunc) string {
	// TODO this should probably use a GO Template to render this page (that is how cobra does help for example).

	// General bottle information
	t := format.NewTable()

	// Add information to table
	t.AddRow("ID", btl.GetBottleID())

	// bottle size information
	total := int64(0)
	selectedTotal := int64(0)
	for _, p := range btl.GetParts() {
		total += p.GetLayerSize()
		if partSelector(p) {
			selectedTotal += p.GetLayerSize()
		}
	}
	t.AddRow("SIZE", fmt.Sprintf("%s selected / %s total", print.Bytes(selectedTotal), print.Bytes(total)))

	if btl.Manifest != nil {
		mDigest := btl.Manifest.GetManifestDescriptor().Digest
		if mDigest != "" {
			t.AddRow("MANIFEST ID", fmt.Sprintf("%v", mDigest))
		}
	}

	t.AddRow("SCHEMA", fmt.Sprintf("%v", btl.Definition.APIVersion))
	if btl.Definition.Description != "" {
		t.AddRow("DESCRIPTION", btl.Definition.Description+"\n")
	}

	if len(btl.Definition.Labels) != 0 {
		t.AddRow("LABELS", labels.Set(btl.Definition.Labels).String()+"\n")
	}

	// bottle sources
	hasSrc := false
	srcStrBuilder := strings.Builder{}

	for _, v := range btl.Definition.Sources {
		if v.Name != "" && v.URI != "" {

			if hasSrc {
				// on first pass, this will be false, but will add a line before the next source is added. This will
				// also prevent adding a line at the end of sources, preventing a double line
				srcStrBuilder.WriteString("\n")
			}

			line := fmt.Sprintf("%q is from %s\n", v.Name, v.URI)
			srcStrBuilder.WriteString(line)
			hasSrc = true
		}
	}

	// add source information to table
	if hasSrc {
		t.AddRow("SOURCES", srcStrBuilder.String())
	}

	// bottle authors
	authorStrBuilder := strings.Builder{}
	for _, v := range btl.Definition.Authors {
		// use RFC 5322 format
		// note that Name and email are required fields so we do not need to check for existence
		line := fmt.Sprintf("%s <%s> %s\n", v.Name, v.Email, v.URL)
		authorStrBuilder.WriteString(line)
		authorStrBuilder.WriteRune('\n')
	}
	// add authors to information table
	if len(btl.Definition.Authors) > 0 {
		t.AddRow("AUTHORS", authorStrBuilder.String())
	}

	partsEntry := strings.Builder{}
	for _, p := range btl.GetParts() {
		if partSelector(p) {
			compStr := "uncompressed"
			if mediatype.IsCompressed(p.GetMediaType()) {
				compStr = "compressed"
			}
			// print "unknown size" if a bottle part does not know the layer size, due to not having been committed.
			sizeStr := print.Bytes(p.GetLayerSize())
			if p.GetLayerSize() == 0 {
				sizeStr = "<unknown size, commit needed>"
			}

			entry := fmt.Sprintf("%s (%s, %s)\nlabels: %s\n\n", p.GetName(), sizeStr, compStr, p.GetLabels())
			partsEntry.WriteString(entry)
		}
	}

	t.AddRow("PARTS", partsEntry.String())

	return t.String()
}
