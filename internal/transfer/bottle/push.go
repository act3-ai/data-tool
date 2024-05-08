// Package bottle provides functions for managing transfer of bottle objects to and from an OCI registry, including
// configuring a pulled bottle, and establishing local metadata and file structure.
package bottle

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"

	"git.act3-ace.com/ace/data/schema/pkg/mediatype"
	"git.act3-ace.com/ace/data/tool/internal/bottle"
	"git.act3-ace.com/ace/data/tool/internal/cache"
	"git.act3-ace.com/ace/data/tool/internal/ref"
	sigcustom "git.act3-ace.com/ace/data/tool/internal/sign"
	"git.act3-ace.com/ace/data/tool/internal/storage"
	reg "git.act3-ace.com/ace/data/tool/pkg/registry"
	tbottle "git.act3-ace.com/ace/data/tool/pkg/transfer/bottle"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// PushBottle copies a bottle to a remote location via oras.ExtendedCopyGraph. ReferrerOptions are used
// to include refferers of the bottle in the copy.
func PushBottle(ctx context.Context, btl *bottle.Bottle, tOpts tbottle.TransferConfig, rOpts ...ReferrerOption) error {
	log := logger.FromContext(ctx)
	store := storage.NewDataStore(btl)

	// prepare referrers
	log.InfoContext(ctx, "preparing bottle referrers")
	for _, o := range rOpts {
		if err := o(ctx, btl, store); err != nil {
			return fmt.Errorf("preparing bottle referrers: %w", err)
		}
	}

	// prep bottle, parts should have already been prepped via commit
	log.InfoContext(ctx, "preparing bottle metadata")
	if err := addBottleMetadataToStore(ctx, btl, store); err != nil {
		return fmt.Errorf("preparing bottle metadata: %w", err)
	}

	destRef, err := ref.FromString(tOpts.Reference)
	if err != nil {
		return fmt.Errorf("parsing destination repository reference: %w", err)
	}

	// copy bottle
	extCopyOpts := oras.ExtendedCopyGraphOptions{
		CopyGraphOptions: oras.CopyGraphOptions{
			Concurrency: tOpts.Concurrency,
			PreCopy:     prePush(btl, store, destRef, tOpts.NewGraphTargetFn), // cross-registry virtual part handling
			MountFrom:   pushMountFrom(btl, destRef),                          // cross-repo virtual part mounting (same registry)
		},
	}

	repo, err := tOpts.NewGraphTargetFn(ctx, destRef.String())
	if err != nil {
		return fmt.Errorf("creating repository reference: %w", err)
	}

	manDesc := btl.Manifest.GetManifestDescriptor()
	log.InfoContext(ctx, "pushing bottle", "bottleID", btl.GetBottleID(), "manDescDigest", manDesc.Digest) //nolint
	if err := oras.ExtendedCopyGraph(ctx, store, repo, manDesc, extCopyOpts); err != nil {
		return fmt.Errorf("pushing bottle: %w", err)
	}

	log.InfoContext(ctx, "tagging bottle manifest")
	if err := repo.Tag(ctx, manDesc, destRef.String()); err != nil {
		return fmt.Errorf("tagging bottle manifest: %w", err)
	}

	return nil
}

// ReferrerOption prepares a bottle's referrers for transfer via oras.ExtendedCopyGraph.
type ReferrerOption func(ctx context.Context, btl *bottle.Bottle, store *storage.DataStore) error

// WithSignatures prepares a bottle's signatures for transfer via oras.ExtendedCopyGraph.
func WithSignatures() ReferrerOption {
	return func(ctx context.Context, btl *bottle.Bottle, store *storage.DataStore) error {
		manDesc := btl.Manifest.GetManifestDescriptor()
		if err := sigcustom.PrepareSigsGraph(ctx, btl.GetPath(), store, manDesc); err != nil {
			return fmt.Errorf("preparing bottle signatures: %w", err)
		}
		return nil
	}
}

// addBottleMetadataToStore adds config and manifest data to the DataStore as loose files for oras to find.  Another
// option would be to cache these.
func addBottleMetadataToStore(ctx context.Context, btl *bottle.Bottle, store *storage.DataStore) error {
	manData, err := btl.Manifest.GetManifestRaw()
	if err != nil {
		return fmt.Errorf("bottle manifest not configured before push")
	}
	cfgData, err := btl.GetConfiguration()
	if err != nil {
		return fmt.Errorf("bottle manifest not configured before push")
	}
	_, err = store.AddLooseData(ctx, bytes.NewReader(manData), btl.Manifest.GetManifestDescriptor().MediaType, nil)
	if err != nil {
		return fmt.Errorf("unable to add manifest data to data store")
	}
	_, err = store.AddLooseData(ctx, bytes.NewReader(cfgData), mediatype.MediaTypeBottleConfig, nil)
	if err != nil {
		return fmt.Errorf("unable to add manifest data to data store")
	}
	return nil
}

// pushMountFrom returns an oras.CopyGraphOptions MountFrom func. It attempts to mount a part from repositories
// within the same destination registry resolved with the bottle's VirtualPartTracker.
//
// MountFrom returns the candidate repositories that desc may be mounted from.
// The OCI references will be tried in turn.  If mounting fails on all of them,
// then it falls back to a copy.
func pushMountFrom(btl *bottle.Bottle, dest ref.Ref) func(ctx context.Context, desc ocispec.Descriptor) ([]string, error) {
	return func(ctx context.Context, desc ocispec.Descriptor) ([]string, error) {
		log := logger.FromContext(ctx).With("digest", desc.Digest)

		bicSources := cache.LocateLayer(ctx, btl.BIC(), desc, dest, true)
		if !mediatype.IsLayer(desc.MediaType) || (btl.VirtualPartTracker == nil && len(bicSources) < 1) {
			// no sources available for cross-repo mounting
			return []string{}, nil
		}

		// sources ∪ bicSources
		vptSources := btl.VirtualPartTracker.Sources(desc.Digest, dest)
		sources := dedupSources(vptSources, bicSources)

		validSources := make([]string, 0)
		for _, source := range sources {
			if !dest.Match(source, ref.RefMatchReg) {
				// virtual part is from another registry and should be handled by PreCopy func
				log.DebugContext(ctx, "virtual part identified in another registry", "source", source.String())
				continue
			}
			log.DebugContext(ctx, "adding part source for mounting", "source", source.String())
			validSources = append(validSources, source.MountRef())
		}
		if len(validSources) > 0 {
			log.DebugContext(ctx, "attempting cross-repo mount", "sources", validSources)
		} else {
			log.DebugContext(ctx, "no sources available for cross-repo mounting")
		}

		return validSources, nil // source repositories, within same dest regisitry, to attempt mounting from
	}
}

// prePush returns and oras.CopyGraphOptions PreCopy func. It first attempts to locate
// the part in the cache. On a cache hit, it resumes the basic copy (from the cache). On a cache miss,
// it instead attempts to copy the part from its known source locations resolved with the bottle's
// VirtualPartTracker.
//
// PreCopy handles the current descriptor before it is copied. PreCopy can
// return a SkipNode to signal that desc should be skipped when it already
// exists in the target.
func prePush(btl *bottle.Bottle, cacheStorage *storage.DataStore, dest ref.Ref, newRepoFn reg.NewGraphTargetFn) func(ctx context.Context, desc ocispec.Descriptor) error {
	return func(ctx context.Context, desc ocispec.Descriptor) error {
		log := logger.FromContext(ctx).With("digest", desc.Digest)

		if btl.VirtualPartTracker == nil || !mediatype.IsLayer(desc.MediaType) {
			return nil
		}

		// prefer copying from cache over another registry
		exists, err := cacheStorage.Exists(ctx, desc)
		switch {
		case err != nil:
			return fmt.Errorf("checking for descriptor in bottle datastore: %w", err)
		case exists:
			log.DebugContext(ctx, "part found in cache, resuming copy from cache")
			return nil
		default:
			log.DebugContext(ctx, "part not found in cache, resolving sources for cross-registry copy")
		}

		// sources ∪ bicSources
		vptSources := btl.VirtualPartTracker.Sources(desc.Digest, dest)
		bicSources := cache.LocateLayer(ctx, btl.BIC(), desc, dest, true)
		sources := dedupSources(vptSources, bicSources)

		errs := make([]error, 0)
		for _, source := range sources { // sources is always of length 1, but let's be safe incase this changes
			if dest.Match(source, ref.RefMatchReg) {
				// virtual part is from the same registry, and should have already been handled by the MountFrom func
				continue
			}
			// ensure we've attempted to copy from another registry at least once
			log.DebugContext(ctx, "attempting cross-registry copy of virtual part", "source", source)

			// connect to source & dest
			srcRepo, err := newRepoFn(ctx, source.String())
			if err != nil {
				errs = append(errs, fmt.Errorf("configuring source repository '%s': %w", source.String(), err))
				continue
			}
			destRepo, err := newRepoFn(ctx, dest.String())
			if err != nil {
				// should be impossible, as the calling fn has already successfully connected to the desintation
				return fmt.Errorf("configuring destination repository: %w", err)
			}

			// copy from source to dest
			rc, err := srcRepo.Fetch(ctx, desc)
			if err != nil {
				errs = append(errs, fmt.Errorf("fetching part from source '%s': %w", source.String(), err))
				continue
			}
			if err := destRepo.Push(ctx, desc, rc); err != nil {
				errs = append(errs, fmt.Errorf("pushing part to destination: %w", err))
				continue
			}

			if err := rc.Close(); err != nil {
				errs = append(errs, fmt.Errorf("closing part fetcher: %w", err))
			}
			log.DebugContext(ctx, "successfully completed cross-registry copy of virtual part")
			if len(errs) < 1 {
				return oras.SkipNode
			}
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}

		return nil
	}
}

// dedupSources takes the union of two sets of source references. It deduplicates
// the set while ensuring sources within the primary set take precedence.
func dedupSources(primary, secondary []ref.Ref) []ref.Ref {
	result := make([]ref.Ref, 0, len(primary)+len(secondary))
	result = append(result, primary...) // primary takes precedence

	dedup := make(map[string]struct{}, len(primary))
	for _, src := range primary {
		dedup[src.String()] = struct{}{}
	}
	for _, src := range secondary {
		if _, exists := dedup[src.String()]; !exists {
			result = append(result, src)
		}
	}

	return result
}
