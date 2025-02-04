package bottle

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"

	"git.act3-ace.com/ace/data/schema/pkg/mediatype"

	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/cache"
	"gitlab.com/act3-ai/asce/data/tool/internal/oci"
	"gitlab.com/act3-ai/asce/data/tool/internal/orasutil"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
	sigcustom "gitlab.com/act3-ai/asce/data/tool/internal/sign"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	reg "gitlab.com/act3-ai/asce/data/tool/pkg/registry"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// Resolve uses the source ReadOnlyGraphTargeter to resolve an OCI reference to a manifest descriptor.
// At minimum the reference must include the "<registry>/<repository>" section of an OCI reference.
func Resolve(ctx context.Context, reference string, src reg.ReadOnlyGraphTargeter, transferOpts TransferOptions) (oras.ReadOnlyGraphTarget, ocispec.Descriptor, error) {
	// if the ReadOnlyGraphTargeter is an ReadOnlyEndpointGraphTargeter we'll perform
	// endpoint resolution implicitly. However, reference remains the same as the original.
	// If an oras registry.ParseReference is wanted for the resolved endpoint, use the
	// ReadOnlyEndpointGraphTargeter interface instead.
	target, err := src.ReadOnlyGraphTarget(ctx, reference)
	if err != nil {
		return nil, ocispec.Descriptor{}, fmt.Errorf("creating graph target for ref '%s': %w", reference, err)
	}

	desc, err := target.Resolve(ctx, reference)
	if err != nil {
		return nil, ocispec.Descriptor{}, fmt.Errorf("resolving descriptor for tag at ref '%s': %w", reference, err)
	}

	// populate blobinfocache
	var bic cache.BIC
	if transferOpts.CachePath != "" {
		bic = cache.NewCache(filepath.Join(transferOpts.CachePath, "blobinfocache.boltdb"))
	} else {
		bic = cache.NewCache("")
	}

	// record the original source reference, not the potentially resolved endpoint
	err = recordSource(ctx, bic, target, reference, desc)
	if err != nil {
		return nil, ocispec.Descriptor{}, fmt.Errorf("recoding bottle source: %w", err)
	}

	return target, desc, nil
}

func recordSource(ctx context.Context, bic cache.BIC, src content.ReadOnlyGraphStorage,
	reference string, desc ocispec.Descriptor,
) error {
	log := logger.FromContext(ctx)

	successors, err := content.Successors(ctx, src, desc)
	if err != nil {
		return fmt.Errorf("error finding successors for %s: %w", desc.Digest.String(), err)
	}
	log.InfoContext(ctx, "found successors", "successors", len(successors))

	r, err := ref.FromString(reference, ref.DefaultRefValidator)
	if err != nil {
		return fmt.Errorf("parsing virtual part source reference: %w", err)
	}

	for _, s := range successors {
		cache.RecordLayerSource(ctx, bic, s, r) // always record blob sources
	}
	return nil
}

// FetchBottleMetadata retrieves a bottle's config and manifest from a remote source.
func FetchBottleMetadata(ctx context.Context, src content.ReadOnlyGraphStorage, desc ocispec.Descriptor,
	pullOpts PullOptions,
) ([]byte, []byte, error) {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "Configuring local bottle")
	btl, err := bottle.NewBottle(
		// bottle.WithLocalPath(opts.BottleDir),
		bottle.WithCachePath(pullOpts.CachePath),
		bottle.DisableDestinationCreate(true),
		bottle.DisableCache(true),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("bottle initialization failed: %w", err)
	}

	log.InfoContext(ctx, "Initializing bottle with remote data")
	err = fetchBottleMetadata(ctx, btl, src, desc) // pull manifest and config, setting up while we go
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

// Pull facilitates the copying of bottles, and signatures, from a remote registry to a local directory.
func Pull(ctx context.Context, src content.ReadOnlyStorage, desc ocispec.Descriptor, pullPath string,
	pullOpts PullOptions,
) error {
	// This function is called by the CSI bottle driver so do not change it needlessly.
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "verifying pull directory", "pullPath", pullPath)
	if err := bottle.VerifyPullDir(pullPath); err != nil {
		return fmt.Errorf("invalid pull directory: %w", err)
	}

	_, err := pull(ctx, src, desc, pullPath, pullOpts)
	if err != nil {
		return fmt.Errorf("pulling bottle: %w", err)
	}

	// type assert src to see if we can pull sigs via referrers
	if p := src.(content.ReadOnlyGraphStorage); p != nil {
		if err := sigcustom.Pull(ctx, pullPath, p, desc); err != nil {
			return fmt.Errorf("pulling bottle signatures: %w", err) // TODO: is a signature pull failure fatal here? Perhaps a good transfer option?
		}
	}

	return nil
}

// pull fetches a bottle from the remote target, caches its parts in their compressed OCI forms, and
// populates the pull directory appropriately.
func pull(ctx context.Context, target content.ReadOnlyStorage, desc ocispec.Descriptor, pullPath string,
	pullOpts PullOptions,
) (*bottle.Bottle, error) {
	log := logger.FromContext(ctx)

	// add reference info to local logging functionality
	log.InfoContext(ctx, "Pulling bottle")
	progress := ui.FromContextOrNoop(ctx).SubTaskWithProgress("Pulling Bottle")
	defer progress.Complete()

	log.InfoContext(ctx, "Configuring local bottle")
	btl, err := bottle.NewBottle(
		bottle.WithLocalPath(pullPath),
		bottle.WithCachePath(pullOpts.CachePath),
		bottle.WithBlobInfoCache(pullOpts.CachePath),
		bottle.WithVirtualParts,
	)
	if err != nil {
		return nil, fmt.Errorf("bottle initialization failed: %w", err)
	}

	log.InfoContext(ctx, "Initializing bottle with remote data")
	err = fetchBottleMetadata(ctx, btl, target, desc) // pull manifest and config, setting up while we go
	if err != nil {
		return nil, err
	}
	// btlManDesc := btl.Manifest.GetManifestDescriptor()

	err = bottle.CreateBottle(btl.GetPath(), true)
	if err != nil {
		return nil, err
	}

	partSelector, err := pullOpts.PartSelectorOptions.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("initializing part selector func: %w", err)
	}

	// protects btl.Parts, which is updated with the part modification times when finalized.
	var btlPartMutex sync.Mutex
	copyOptions := oras.CopyGraphOptions{
		Concurrency: pullOpts.concurrency(),
		PreCopy:     prePullParts(progress),
		// whether or not we copy/skip the part is irrelevant, in both cases
		// we need to populate the bottle directory with the parts.
		PostCopy:       postPull(progress, btl, &btlPartMutex),
		OnCopySkipped:  postPull(progress, btl, &btlPartMutex),
		FindSuccessors: selectPartSuccessors(btl, partSelector),
	}

	// ensure parts are not skipped, since CopyGraph will skip now that the manifest exists
	dest := &orasutil.UnreliableStorage{
		Storage: btl.GetCache(),
	}

	log.InfoContext(ctx, "copying bottle layers from remote", "layers", len(btl.Manifest.GetLayerDescriptors()))
	err = oras.CopyGraph(ctx, target, dest, desc, copyOptions)
	if err != nil {
		return nil, fmt.Errorf("failure to copygraph for bottle: %w", err)
	}

	log.InfoContext(ctx, "writing bottle metadata")
	err = btl.Save()
	if err != nil {
		return nil, err
	}

	log.InfoContext(ctx, "saving bottle OCI info")
	if err := bottle.SaveExtraBottleInfo(ctx, btl); err != nil {
		return nil, err
	}

	log.InfoContext(ctx, "pull complete")
	return btl, nil
}

// fetchBottleMetadata configures the provided bottle with data retrieved from a configured transfer.  This performs
// manifest and configuration retrieval, and applies the retrieved data to the bottle metadata.
func fetchBottleMetadata(ctx context.Context, btl *bottle.Bottle, target content.Fetcher,
	desc ocispec.Descriptor,
) error {
	// get manifest data
	if err := handleBottleManifest(ctx, btl, target, desc); err != nil {
		return fmt.Errorf("fetching bottle manifest: %w", err)
	}

	// get config data and apply pre- and post-config functions
	if err := handleBottleConfig(ctx, btl, target, btl.Manifest.GetConfigDescriptor()); err != nil {
		return fmt.Errorf("fetching bottle config: %w", err)
	}

	if err := verifyPull(btl); err != nil {
		return fmt.Errorf("validating bottle manifest against configuration: %w", err)
	}

	return nil
}

// verifyPull validates a bottle's part configuration against the manifest layers.
func verifyPull(btl *bottle.Bottle) error {
	numParts := btl.NumParts()
	numLayers := len(btl.Manifest.GetLayerDescriptors())
	if numParts != numLayers {
		return fmt.Errorf("layer and part count mismatch: layers=%d, parts=%d", numLayers, numParts)
	}

	return nil
}

// handleBottleManifest fetches a bottle's manifest and populates the appropriate manifest
// related fields.
func handleBottleManifest(ctx context.Context, btl *bottle.Bottle, storage content.Fetcher,
	desc ocispec.Descriptor) error {

	// fetch from cache storage
	manBytes, err := content.FetchAll(ctx, storage, desc)
	if err != nil {
		return fmt.Errorf("fetching bottle manifest from storage: %w", err)
	}

	manifestHandler := oci.ManifestFromData(ocispec.MediaTypeImageManifest, manBytes)
	if manifestHandler.GetStatus().Error != nil {
		return fmt.Errorf("constructing manifest handler from raw manifest: %w", manifestHandler.GetStatus().Error)
	}
	btl.SetManifest(manifestHandler)

	raw, err := manifestHandler.GetManifestRaw() // raw should equal manBytes, but let's be safe incase ManifestFromData alters something
	if err != nil {
		return fmt.Errorf("getting original bottle manifest: %w", err)
	}
	btl.OriginalManifest = raw

	return nil
}

// handleBottleConfig fetches a bottle's config and populates the appropriate config
// related fields.
func handleBottleConfig(ctx context.Context, btl *bottle.Bottle, src content.Fetcher,
	desc ocispec.Descriptor) error {
	cfgBytes, err := content.FetchAll(ctx, src, desc)
	if err != nil {
		return fmt.Errorf("fetching from remote: %w", err)
	}

	btl.OriginalConfig = cfgBytes
	originalConfigDigest := digest.FromBytes(cfgBytes) // This is the correct bottleID for the pulled bottle, before any config changes are made

	// Perform bottle metadata configuration from the received config
	err = btl.Configure(cfgBytes)
	if err != nil {
		return err
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
func prePullParts(progress *ui.Progress) func(ctx context.Context, desc ocispec.Descriptor) error {
	return func(ctx context.Context, desc ocispec.Descriptor) error {
		switch {
		case desc.MediaType == ocispec.MediaTypeImageManifest:
			return oras.SkipNode // manifest already handled and we don't want to cache it
		case mediatype.IsBottleConfig(desc.MediaType):
			return oras.SkipNode // config already handled and shouldn't be in the successor list, i.e. reaching here should be impossible
		case mediatype.IsLayer(desc.MediaType):
			progress.Update(0, desc.Size)
		default:
			logger.FromContext(ctx).DebugContext(ctx, "unsupported mediatype encountered pre copy", "mediatype",
				desc.MediaType, "digest", desc.Digest)
		}
		return nil
	}
}

// postPull returns a func for the oras.CopyGraphOptions option PostCopy/OnCopySkipped func. It extracts a recently
// cached part to its final destination or appropriately handles the manifest or config.
func postPull(progress *ui.Progress, btl *bottle.Bottle, btlPartMutex *sync.Mutex) func(ctx context.Context, desc ocispec.Descriptor) error {
	return func(ctx context.Context, desc ocispec.Descriptor) error {
		switch {
		case desc.MediaType == ocispec.MediaTypeImageManifest:
			// noop
		case mediatype.IsBottleConfig(desc.MediaType):
			// noop
		case mediatype.IsLayer(desc.MediaType):
			btlPartMutex.Lock()
			name := btl.GetPartByLayerDescriptor(desc).GetName()
			btlPartMutex.Unlock()
			handled, err := bottle.CopyFromCache(ctx, btl, desc, name, btlPartMutex)
			// update the progress after copy, even if the copy failed.
			progress.Update(desc.Size, 0)
			// now check for copy errors/failures
			if err != nil {
				return fmt.Errorf("failed to finalize part with digest %s: %w", desc.Digest, err)
			}
			if !handled {
				return fmt.Errorf("part not found in cache after copy %s", desc.Digest)
			}
		default:
			logger.FromContext(ctx).DebugContext(ctx, "unsupported mediatype encountered post copy", "mediatype",
				desc.MediaType, "digest", desc.Digest)
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
func selectPartSuccessors(btl *bottle.Bottle, selector bottle.PartSelectorFunc) func(ctx context.Context,
	fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
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
						btl.VirtualPartTracker.Add(s.Digest, partInfo.GetContentDigest())
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
