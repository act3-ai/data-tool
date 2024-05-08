package bottle

import (
	"context"
	"fmt"

	"git.act3-ace.com/ace/data/schema/pkg/mediatype"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"

	"git.act3-ace.com/ace/data/tool/internal/bottle"
	"git.act3-ace.com/ace/data/tool/internal/cache"
	"git.act3-ace.com/ace/data/tool/internal/oci"
	"git.act3-ace.com/ace/data/tool/internal/ref"
	"git.act3-ace.com/ace/data/tool/internal/storage"
	"git.act3-ace.com/ace/data/tool/internal/ui"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// FetchBottleMetadata retrieves a bottle's config and manifest from a remote source.
func FetchBottleMetadata(ctx context.Context, opts TransferConfig) ([]byte, []byte, error) {
	log := logger.FromContext(ctx).With("ref", opts.Reference)

	target, err := opts.NewGraphTargetFn(ctx, opts.Reference)
	if err != nil {
		return nil, nil, err
	}

	log.InfoContext(ctx, "Configuring local bottle")
	btl, err := bottle.NewBottle(
		bottle.WithLocalPath(opts.PullPath),
		bottle.WithCachePath(opts.cachePath),
		bottle.DisableDestinationCreate(true),
		bottle.DisableCache(true),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("bottle initialization failed: %w", err)
	}

	log.InfoContext(ctx, "Initializing bottle with remote data")
	err = setupBottleFromTransfer(ctx, btl, target, opts) // pull manifest and config, setting up while we go
	if err != nil {
		return nil, nil, err
	}

	cfgBytes, err := btl.GetConfiguration()
	if err != nil {
		return nil, nil, fmt.Errorf("getting bottle configuration: %w", err)
	}

	manBytes, err := btl.Manifest.GetManifestRaw()
	if err != nil {
		return nil, nil, fmt.Errorf("getting bottle manifest: %w", err)
	}

	return cfgBytes, manBytes, nil
}

// pull fetches a bottle from the remote target, caches its parts in their compressed OCI forms, and
// populates the pull directory appropriately.
func pull(ctx context.Context, target oras.GraphTarget, opts TransferConfig) (*bottle.Bottle, error) {
	log := logger.FromContext(ctx).With("ref", opts.Reference)

	// add reference info to local logging functionality
	log.InfoContext(ctx, "Pulling bottle")
	progress := ui.FromContextOrNoop(ctx).SubTaskWithProgress("Pulling Bottle")
	defer progress.Complete()

	log.InfoContext(ctx, "Configuring local bottle")
	btl, err := bottle.NewBottle(
		bottle.WithLocalPath(opts.PullPath),
		bottle.WithCachePath(opts.cachePath),
		bottle.WithBlobInfoCache(opts.cachePath),
		bottle.WithVirtualParts,
	)
	if err != nil {
		return nil, fmt.Errorf("bottle initialization failed: %w", err)
	}

	log.InfoContext(ctx, "Initializing bottle with remote data")
	err = setupBottleFromTransfer(ctx, btl, target, opts) // pull manifest and config, setting up while we go
	if err != nil {
		return nil, err
	}
	btlManDesc := btl.Manifest.GetManifestDescriptor()

	partSelector, err := opts.partSelector.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("initializing part selector func: %w", err)
	}

	dataCache := storage.NewDataStore(btl)
	refSpec := ref.RepoFromString(opts.Reference)
	copyOptions := oras.CopyGraphOptions{
		Concurrency:    opts.Concurrency,
		PreCopy:        prePullParts(progress, btl),
		PostCopy:       postPullParts(progress, btl, dataCache),
		OnCopySkipped:  onPullSkipped(progress, btl, dataCache),
		FindSuccessors: selectPartSuccessors(btl, partSelector, refSpec),
	}

	log.InfoContext(ctx, "Copying bottle layers from remote", "layers", len(btl.Manifest.GetLayerDescriptors()))
	err = oras.CopyGraph(ctx, target, dataCache, btlManDesc, copyOptions)
	if err != nil {
		return nil, fmt.Errorf("failure to copygraph for bottle: %w", err)
	}

	log.InfoContext(ctx, "Writing bottle metadata")
	err = btl.Save()
	if err != nil {
		return nil, err
	}

	// Verify Bottle ID if option is set
	if opts.matchBottleID != "" {
		matchDigest, err := digest.Parse(opts.matchBottleID)
		if err != nil {
			return nil, fmt.Errorf("match bottle ID parse error: %w", err)
		}

		if err := btl.VerifyBottleID(matchDigest); err != nil {
			return nil, err
		}
	}

	// TODO: bottle id saving requires an output file name, currently this output is done as part of the config handler,
	// but it would make more sense here.
	log.InfoContext(ctx, "writing bottleID")
	if err := bottle.SaveExtraBottleInfo(ctx, btl, ""); err != nil {
		return nil, err
	}

	log.InfoContext(ctx, "pull complete")
	if err := dataCache.Close(); err != nil {
		return btl, fmt.Errorf("closing datastore: %w", err)
	}

	return btl, nil
}

// setupBottleFromTransfer configures the provided bottle with data retrieved from a configured transfer.  This performs
// manifest and configuration retrieval, and applies the retrieved data to the bottle metadata.
func setupBottleFromTransfer(ctx context.Context, btl *bottle.Bottle, target oras.GraphTarget, opts TransferConfig) error {

	// resolve ref to descriptor
	btlManDesc, err := target.Resolve(ctx, opts.Reference)
	if err != nil {
		return fmt.Errorf("resolving bottle reference to descriptor: %w", err)
	}

	// get manifest data
	if err := fetchBottleManifest(ctx, btl, target, btlManDesc); err != nil {
		return fmt.Errorf("fetching bottle manifest: %w", err)
	}

	// get config data and apply pre- and post-config functions
	if err := fetchBottleConfig(ctx, btl, target, btl.Manifest.GetConfigDescriptor(), opts); err != nil {
		return fmt.Errorf("fetching bottle config: %w", err)
	}

	numParts := btl.NumParts()
	numLayers := len(btl.Manifest.GetLayerDescriptors())
	if numParts != numLayers {
		return fmt.Errorf("layer and part count mismatch: layers=%d, parts=%d", numLayers, numParts)
	}

	return nil
}

// fetchBottleManifest fetches a bottle's manifest and populates the appropriate manifest
// related fields.
func fetchBottleManifest(ctx context.Context, btl *bottle.Bottle, target oras.GraphTarget, desc ocispec.Descriptor) error {
	manBytes, err := content.FetchAll(ctx, target, desc)
	if err != nil {
		return fmt.Errorf("fetching bottle manifest: %w", err)
	}

	manifestHandler := oci.ManifestFromData(ocispec.MediaTypeImageManifest, manBytes)
	if manifestHandler.GetStatus().Error != nil {
		return fmt.Errorf("constructing manifest handler from raw manifest: %w", err)
	}
	btl.SetManifest(manifestHandler)

	raw, err := manifestHandler.GetManifestRaw() // raw should equal manBytes, but let's be safe incase ManifestFromData alters something
	if err != nil {
		return fmt.Errorf("getting original bottle manifest: %w", err)
	}
	btl.OriginalManifest = raw

	return nil
}

// fetchBottleConfig fetches a bottle's config and populates the appropriate config
// related fields.
func fetchBottleConfig(ctx context.Context, btl *bottle.Bottle, target oras.GraphTarget,
	desc ocispec.Descriptor, opts TransferConfig) error {
	cfgBytes, err := content.FetchAll(ctx, target, desc)
	if err != nil {
		return fmt.Errorf("fetching from remote: %w", err)
	}

	btl.OriginalConfig = cfgBytes
	originalConfigDigest := digest.FromBytes(cfgBytes) // This is the correct bottleID for the pulled bottle, before any config changes are made

	// apply pre-config func
	if opts.preConfigHandler != nil {
		if err := opts.preConfigHandler(btl, btl.OriginalConfig); err != nil {
			return err
		}
	}

	// Perform bottle metadata configuration from the received config
	err = btl.Configure(cfgBytes)
	if err != nil {
		return err
	}

	if opts.postConfigHandler == nil {
		if !btl.DisableCreateDestDir {
			// We only want the DefaultConfigHandler if we are creating a directory
			// For example, the show command does not create a directory
			opts.postConfigHandler = defaultConfigHandler
		}
	}

	// apply post-config func
	if opts.postConfigHandler != nil {
		if err := opts.postConfigHandler(btl, cfgBytes); err != nil {
			return err
		}
	}

	// check if bottle was upgraded (and consequently the bottleID was changed)
	// deprecate the previous bottleID to promote using the latest bottle (config) version
	if btl.GetBottleID() != originalConfigDigest {
		btl.DeprecateBottleID(originalConfigDigest)
	}

	return nil
}

// prePullParts returns a func for the oras.CopyGraphOptions option PreCopy func.
// All parts encountered by this function have been selected and were not found in the cache.
// prePullParts is used for skipping selected successors that shouldn't be cached and increasing
// progress total.
func prePullParts(progress *ui.Progress, btl *bottle.Bottle) func(ctx context.Context, desc ocispec.Descriptor) error {
	return func(ctx context.Context, desc ocispec.Descriptor) error {
		switch {
		case desc.MediaType == ocispec.MediaTypeImageManifest:
			return oras.SkipNode // manifest already handled and we don't want to cache it
		case mediatype.IsBottleConfig(desc.MediaType):
			return oras.SkipNode // config already handled and shouldn't be in the successor list, i.e. reaching here should be impossible
		case mediatype.IsLayer(desc.MediaType):
			progress.Update(0, desc.Size)
		default:
			logger.FromContext(ctx).DebugContext(ctx, "unsupported mediatype encountered pre copy", "mediatype", desc.MediaType, "digest", desc.Digest)
		}
		return nil
	}
}

// postPullParts returns a func for the oras.CopyGraphOptions option PostCopy func. It extracts a recently
// cached part to its final destination.
func postPullParts(progress *ui.Progress, btl *bottle.Bottle,
	dataStore *storage.DataStore) func(ctx context.Context, desc ocispec.Descriptor) error {
	return func(ctx context.Context, desc ocispec.Descriptor) error {
		switch {
		case desc.MediaType == ocispec.MediaTypeImageManifest:
			// noop
		case mediatype.IsBottleConfig(desc.MediaType):
			// noop
		case mediatype.IsLayer(desc.MediaType):
			handled, err := dataStore.CopyFromCache(ctx, desc, btl.GetPartByLayerDescriptor(desc).GetName())
			// update the progress after copy, even if the copy failed.
			// TODO: progress doesn't really seem to be working...
			progress.Update(desc.Size, desc.Size)
			// now check for copy errors/failures
			if err != nil {
				return fmt.Errorf("failed to finalize part with digest %s: %w", desc.Digest, err)
			}
			if !handled {
				return fmt.Errorf("part not found in cache after copy %s", desc.Digest)
			}
		default:
			logger.FromContext(ctx).DebugContext(ctx, "unsupported mediatype encountered post copy", "mediatype", desc.MediaType, "digest", desc.Digest)
		}
		return nil
	}
}

// selectPartSuccessors returns a function that implements oras.CopyGraphOptions.FindSuccessors callback function.
// selectSuccessors finds all successors of a bottle, reducing the set to selected parts only. Excluded parts
// are added to the bottle's VirtualPartTracker. If no selector is provided, all successors (excluding config) are returned.
// The caching status of the returned descriptors is unknown. Not safe to use with oras.ExtendedCopyGraph.
// fetcher provides cached access to the source storage, and is suitable
// for fetching non-leaf nodes like manifests. Since anything fetched from
// fetcher will be cached in the memory, it is recommended to use original
// source storage to fetch large blobs.
func selectPartSuccessors(btl *bottle.Bottle, selector bottle.PartSelectorFunc,
	refSpec ref.Ref) func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	return func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		log := logger.FromContext(ctx)

		successors, err := content.Successors(ctx, fetcher, desc)
		if err != nil {
			return nil, fmt.Errorf("error finding successors for %s: %w", desc.Digest.String(), err)
		}
		log.InfoContext(ctx, "found successors", "successors", len(successors))

		selected := make([]ocispec.Descriptor, 0, len(successors))
		for _, s := range successors {

			// apply part selector
			switch {
			case mediatype.IsBottleConfig(s.MediaType):
				// do not select config, this should have already been handled
				log.DebugContext(ctx, "removing config from successors")
			case mediatype.IsLayer(s.MediaType):
				cache.RecordLayerSource(ctx, btl.BIC(), desc, refSpec) // always record blob sources
				if selector == nil {
					// skip selection if no selector was provided
					continue
				}

				partInfo := btl.GetPartByLayerDescriptor(s)
				if partInfo == nil {
					return successors, fmt.Errorf("part referenced in manifest does not exist in bottle config: layer digest = %s", s.Digest)
				}

				if selector(partInfo) {
					// part selected
					log.InfoContext(ctx, "selected part",
						"part", partInfo.GetName(),
						"layerDigest", s.Digest,
						"size", s.Size,
						"type", s.MediaType)

					selected = append(selected, s)
				} else {
					// part not selected, add as virtual part
					log.InfoContext(ctx, "did not select part",
						"part", partInfo.GetName(),
						"layerDigest", s.Digest,
						"size", s.Size,
						"type", s.MediaType)

					if btl.VirtualPartTracker != nil {
						btl.VirtualPartTracker.Add(s.Digest, partInfo.GetContentDigest(), refSpec)
					}
				}
			default:
				// TODO: We should add signature or other referrer mediatypes, otherwise this will cause
				// errors if this func is used with oras.ExtendedCopyGraph
				log.DebugContext(ctx, "unexpected successor type", "mediatype", s.MediaType, "digest", s.Digest)
			}
		}
		return selected, nil
	}
}

// onPullSkipped handles the extraction of cached parts to their destinations when they're skipped
// during a copy to the cache. This funcion is triggered whenever the cache hits, i.e. returns true
// on existence check.
func onPullSkipped(progress *ui.Progress, btl *bottle.Bottle, dataStore *storage.DataStore) func(ctx context.Context, desc ocispec.Descriptor) error {
	return func(ctx context.Context, desc ocispec.Descriptor) error {
		switch {
		case desc.MediaType == ocispec.MediaTypeImageManifest:
			// noop
		case mediatype.IsBottleConfig(desc.MediaType):
			// noop
		case mediatype.IsLayer(desc.MediaType):
			handled, err := dataStore.CopyFromCache(ctx, desc, btl.GetPartByLayerDescriptor(desc).GetName())
			progress.Update(desc.Size, desc.Size)
			switch {
			case err != nil:
				return err
			case !handled:
				return fmt.Errorf("part not found in cache despite passing prior existence check %s", desc.Digest) // should be impossible
			default:
				logger.V(logger.FromContext(ctx), 1).InfoContext(ctx, "copied from cache")
			}
		default:
			return fmt.Errorf("unsupported mediatype skipped '%s'", desc.MediaType)
		}
		// skip was safe
		return nil
	}
}
