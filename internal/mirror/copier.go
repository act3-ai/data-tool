package mirror

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	"gitlab.com/act3-ai/asce/data/tool/internal/mirror/encoding"
)

// Copier represents a copy object for mirror operations.
type Copier struct {
	log              *slog.Logger
	src              oras.ReadOnlyGraphTarget
	dest             oras.Target
	root             ocispec.Descriptor
	srcRef           registry.Reference
	destRef          registry.Reference
	options          oras.CopyOptions
	platforms        []*ocispec.Platform
	referrers        bool
	originatingIndex []byte
}

// NewCopier creates a new Copier.
func NewCopier(ctx context.Context,
	log *slog.Logger,
	source string,
	dest string,
	srcTarget oras.ReadOnlyGraphTarget,
	srcRef registry.Reference,
	root ocispec.Descriptor,
	destTarget oras.Target,
	destRef registry.Reference,
	referrers bool,
	platforms []*ocispec.Platform,
	repoFunc func(ctx context.Context, ref string) (*remote.Repository, error),
) (*Copier, error) {
	// These CopyGraphOptions will be nested within CopyOptions for the case of platform
	options := oras.CopyGraphOptions{}
	if srcTarget == nil {
		if source == "" {
			return nil, fmt.Errorf("NewCopier requires either a non-nil source target or a source string")
		}
		// get the source ref and the source target repo from the Source object
		r, err := repoFunc(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("error getting the source repository: %w", err)
		}
		srcTarget = r
		srcRef = r.Reference
		// resolve the source reference
		desc, err := srcTarget.Resolve(ctx, r.Reference.ReferenceOrDefault())
		if err != nil {
			return nil, fmt.Errorf("error resolving the source repository: %w", err)
		}
		root = desc
	}

	if destTarget == nil {
		if dest == "" {
			return nil, fmt.Errorf("NewCopier requires either a non-nil destination target or a destination string")
		}
		// get the source ref and the source target repo from the Source object
		r, err := repoFunc(ctx, dest)
		if err != nil {
			return nil, fmt.Errorf("error getting the destination repository: %w", err)
		}
		destTarget = r
		destRef = r.Reference
	}

	copier := Copier{
		log:     log,
		src:     srcTarget,
		dest:    destTarget,
		root:    root,
		srcRef:  srcRef,
		destRef: destRef,
		options: oras.CopyOptions{
			CopyGraphOptions: options,
		},
		platforms: platforms,
		referrers: referrers,
	}
	// originating index is needed for multi-architecture images, we only want to pull the data if the descriptor is an index and platforms are defined.
	if encoding.IsIndex(copier.root.MediaType) && platforms != nil {
		// need descriptor to assess the mediaType otherwise I would just use oras.FetchBytes instead of resolving above
		rc, err := srcTarget.Fetch(ctx, copier.root)
		if err != nil {
			return nil, fmt.Errorf("fetching the originating index: %w", err)
		}
		b, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("reading the index bytes: %w", err)
		}
		copier.originatingIndex = b
	}
	return &copier, nil
}

// Copy copies a source target to a destination target while handling cross-repo mounting. It uses the
// ExtendedCopyGraph function in oras to capture predecessor references.
func Copy(ctx context.Context, c *Copier) error {
	// load cross-repo mounting and findSuccessors functions
	options := modifyOpts(c)

	if c.referrers {
		// we want to use ExtendedCopyGraph so we copy for SBOMs and signatures as well
		options := oras.ExtendedCopyGraphOptions{
			CopyGraphOptions: options,
		}
		if err := oras.ExtendedCopyGraph(ctx, c.src, c.dest, c.root, options); err != nil {
			return fmt.Errorf("copying image (and predecessors): %w", err)
		}
	} else {
		if err := oras.CopyGraph(ctx, c.src, c.dest, c.root, options); err != nil {
			return fmt.Errorf("copying image: %w", err)
		}
	}
	// we do not use oras.ExtendedCopy() because we copy from a descriptor (not a ref)

	return nil
}

// CopyFilterOnPlatform will copy from the root descriptor only the manifests that match the platforms defined.
func CopyFilterOnPlatform(ctx context.Context, c *Copier) ([]ocispec.Descriptor, error) {
	platformDescriptors := make([]ocispec.Descriptor, 0, len(c.platforms))
	// We need to evaluate the config media type to allow helm charts, sboms, and data bottles.
	if encoding.IsImage(c.root.MediaType) {
		var img ocispec.Manifest
		b, err := content.FetchAll(ctx, c.src, c.root)
		if err != nil {
			return nil, fmt.Errorf("error fetching manifest: %w", err)
		}
		if err := json.Unmarshal(b, &img); err != nil {
			return nil, fmt.Errorf("error unmarshalling the manifest")
		}
		// this allows signatures, helm charts, bottles, etc to still be pushed because they have special config media types.
		if !(img.Config.MediaType == ocispec.MediaTypeImageConfig) {
			platformDescriptors = append(platformDescriptors, c.root)
			// return the root descriptor and copy it normally.
			return platformDescriptors, Copy(ctx, c)
		}
	}
	// we should evaluate platforms before referrers, referrers TODO
	for _, platform := range c.platforms {
		c.options.WithTargetPlatform(platform)
		desc, err := oras.Copy(ctx, c.src, c.srcRef.Reference, c.dest, c.destRef.Reference, c.options)
		if err != nil {
			// we don't want to throw an error if the platform is not found in a multi-arch index in the case of supporting SBOMs and helm charts
			if errors.Is(err, errdef.ErrNotFound) {
				continue
			}
			return nil, fmt.Errorf("error copying specific platform: %w", err)
		}
		platformDescriptors = append(platformDescriptors, desc)
	}
	return platformDescriptors, nil
}

// modifyOpts will modify the nested CopyGraphOptions to add custom successors and cross-repo blob mounting functions.
func modifyOpts(c *Copier) oras.CopyGraphOptions {
	options := c.options.CopyGraphOptions
	if options.FindSuccessors == nil {
		options.FindSuccessors = encoding.Successors
	}
	// Try mounting if source and destination are on the same registry
	// TODO use a blob info cache to find more source repositories
	if c.srcRef.Registry == c.destRef.Registry {
		// see https://github.com/oras-project/oras-go/issues/580
		options.MountFrom = func(ctx context.Context, desc ocispec.Descriptor) ([]string, error) {
			return []string{c.srcRef.Repository}, nil
		}
		options.OnMounted = func(ctx context.Context, desc ocispec.Descriptor) error {
			c.log.InfoContext(ctx, "Mounted", "descriptor", desc)
			return nil
		}
	}
	return options
}
