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
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/cache"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
	sigcustom "gitlab.com/act3-ai/asce/data/tool/internal/sign"
	reg "gitlab.com/act3-ai/asce/data/tool/pkg/registry"
)

// PushBottle copies a bottle to a remote location via oras.ExtendedCopyGraph. ReferrerOptions are used
// to include refferers of the bottle in the copy.
func PushBottle(ctx context.Context, btl *bottle.Bottle, gt reg.EndpointGraphTargeter, reference string, pushCfg PushOptions, rOpts ...ReferrerOption) error {
	log := logger.FromContext(ctx)

	// prepare referrers
	log.InfoContext(ctx, "preparing bottle referrers")
	rOpts = append(rOpts, withSignatures()) // always push with signatures
	for _, o := range rOpts {
		if err := o(ctx, btl); err != nil {
			return fmt.Errorf("preparing bottle referrers: %w", err)
		}
	}

	// prep bottle, parts should have already been prepped via commit
	log.InfoContext(ctx, "preparing bottle metadata")
	if err := AddBottleMetadataToStore(ctx, btl); err != nil {
		return fmt.Errorf("preparing bottle metadata: %w", err)
	}

	destRef, err := ref.FromString(reference)
	if err != nil {
		return fmt.Errorf("parsing destination repository reference: %w", err)
	}

	// copy bottle
	extCopyOpts := oras.ExtendedCopyGraphOptions{
		CopyGraphOptions: oras.CopyGraphOptions{
			Concurrency: pushCfg.Concurrency,         // TODO: this should be a method, which already exists, but we're in a different pkg
			PreCopy:     prePush(btl, destRef, gt),   // cross-registry virtual part handling
			MountFrom:   pushMountFrom(btl, destRef), // cross-repo virtual part mounting (same registry)
		},
	}

	repo, err := gt.GraphTarget(ctx, destRef.String())
	if err != nil {
		return fmt.Errorf("creating repository reference: %w", err)
	}

	manDesc := btl.Manifest.GetManifestDescriptor()
	log.InfoContext(ctx, "pushing bottle", "bottleID", btl.GetBottleID(), "manDescDigest", manDesc.Digest) //nolint
	if err := oras.ExtendedCopyGraph(ctx, btl.GetCache(), repo, manDesc, extCopyOpts); err != nil {
		return fmt.Errorf("pushing bottle: %w", err)
	}

	log.InfoContext(ctx, "tagging bottle manifest")
	rr, err := gt.ParseEndpointReference(destRef.String())
	if err != nil {
		return fmt.Errorf("parsing endpoint reference '%s': %w", destRef.String(), err)
	}
	if err := repo.Tag(ctx, manDesc, rr.String()); err != nil {
		return fmt.Errorf("tagging bottle manifest: %w", err)
	}

	return nil
}

// ReferrerOption prepares a bottle's referrers for transfer via oras.ExtendedCopyGraph.
type ReferrerOption func(ctx context.Context, btl *bottle.Bottle) error

// withSignatures prepares a bottle's signatures for transfer via oras.ExtendedCopyGraph.
func withSignatures() ReferrerOption {
	return func(ctx context.Context, btl *bottle.Bottle) error {
		manDesc := btl.Manifest.GetManifestDescriptor()
		if err := sigcustom.PrepareSigsGraph(ctx, btl.GetPath(), btl.GetCache(), manDesc); err != nil {
			return fmt.Errorf("preparing bottle signatures: %w", err)
		}
		return nil
	}
}

// AddBottleMetadataToStore adds config and manifest data to the DataStore as loose files for oras to find.  Another
// option would be to cache these.
func AddBottleMetadataToStore(ctx context.Context, btl *bottle.Bottle) error {
	log := logger.V(logger.FromContext(ctx), 1)

	storage := btl.GetCache()

	exists, err := storage.Exists(ctx, btl.Manifest.GetManifestDescriptor())
	switch {
	case err != nil:
		return fmt.Errorf("checking bottle manifest existence in storage: %w", err)
	case exists:
		log.InfoContext(ctx, "bottle manifest already exists in cache", "digest", btl.Manifest.GetManifestDescriptor().Digest)
	default:
		manData, err := btl.Manifest.GetManifestRaw()
		if err != nil {
			return fmt.Errorf("bottle manifest not configured before push: %w", err)
		}

		if err := storage.Push(ctx, btl.Manifest.GetManifestDescriptor(), bytes.NewReader(manData)); err != nil {
			return fmt.Errorf("pushing bottle manifest to storage: %w", err)
		}
	}

	exists, err = storage.Exists(ctx, btl.Manifest.GetConfigDescriptor())
	switch {
	case err != nil:
		return fmt.Errorf("checking bottle config existence in storage: %w", err)
	case exists:
		log.InfoContext(ctx, "bottle config already exists in cache", "digest", btl.Manifest.GetConfigDescriptor().Digest)
	default:
		cfgData, err := btl.GetConfiguration()
		if err != nil {
			return fmt.Errorf("bottle config not configured before push: %w", err)
		}

		if err := storage.Push(ctx, btl.Manifest.GetConfigDescriptor(), bytes.NewReader(cfgData)); err != nil {
			return fmt.Errorf("pushing bottle config to storage: %w", err)
		}
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
		if !mediatype.IsLayer(desc.MediaType) && len(bicSources) < 1 {
			// no sources available for cross-repo mounting
			return []string{}, nil
		}

		validSources := make([]string, 0)
		for _, source := range bicSources {
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
// it instead attempts to copy the part from its known source locations resolved with the blob info cache.
//
// PreCopy handles the current descriptor before it is copied. PreCopy can
// return a SkipNode to signal that desc should be skipped when it already
// exists in the target.
func prePush(btl *bottle.Bottle, dest ref.Ref, gt reg.EndpointGraphTargeter) func(ctx context.Context, desc ocispec.Descriptor) error {
	return func(ctx context.Context, desc ocispec.Descriptor) error {
		log := logger.FromContext(ctx).With("digest", desc.Digest)

		if !mediatype.IsLayer(desc.MediaType) {
			return nil
		}

		// prefer copying from cache over another registry
		exists, err := btl.GetCache().Exists(ctx, desc)
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
		bicSources := cache.LocateLayer(ctx, btl.BIC(), desc, dest, false)

		errs := make([]error, 0)
		for _, source := range bicSources { // sources is always of length 1, but let's be safe incase this changes
			if dest.Match(source, ref.RefMatchReg) {
				// virtual part is from the same registry, and should have already been handled by the MountFrom func
				continue
			}
			src := source.String()
			// ensure we've attempted to copy from another registry at least once
			log.DebugContext(ctx, "attempting cross-registry copy of virtual part", "source", src)

			// connect to source & dest
			srcRepo, err := gt.GraphTarget(ctx, src)
			if err != nil {
				errs = append(errs, fmt.Errorf("configuring source repository '%s': %w", src, err))
				continue
			}
			destRepo, err := gt.GraphTarget(ctx, dest.RepoString())
			if err != nil {
				// should be impossible, as the calling fn has already successfully connected to the desintation
				return fmt.Errorf("configuring destination repository: %w", err)
			}

			// copy from source to dest
			rc, err := srcRepo.Fetch(ctx, desc)
			if err != nil {
				errs = append(errs, fmt.Errorf("fetching part from source '%s': %w", src, err))
				continue
			}
			if err := destRepo.Push(ctx, desc, rc); err != nil {
				errs = append(errs, fmt.Errorf("pushing part to destination: %w", err))
				continue
			}

			err = rc.Close()
			if err != nil {
				return fmt.Errorf("closing part fetcher: %w", err)
			}

			log.DebugContext(ctx, "successfully completed cross-registry copy of virtual part")
			return oras.SkipNode

		}

		return errors.Join(errs...)
	}
}
