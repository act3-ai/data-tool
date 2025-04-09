package mirror

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"

	"github.com/act3-ai/data-tool/internal/mirror/encoding"
)

// Copier represents a copy object for mirror operations.
type Copier struct {
	log              *slog.Logger
	src              content.ReadOnlyGraphStorage
	dest             content.Storage
	root             ocispec.Descriptor
	options          oras.CopyOptions
	platforms        []*ocispec.Platform
	referrers        bool
	originatingIndex []byte
}

// NewCopier creates a new Copier.
func NewCopier(ctx context.Context,
	log *slog.Logger,
	src content.ReadOnlyGraphStorage,
	dest content.Storage,
	root ocispec.Descriptor,
	referrers bool,
	platforms []*ocispec.Platform,
	opts oras.CopyGraphOptions) (*Copier, error) {

	copier := Copier{
		log:  log,
		src:  src,
		dest: dest,
		root: root,
		options: oras.CopyOptions{
			CopyGraphOptions: opts,
		},
		platforms: platforms,
		referrers: referrers,
	}

	if opts.FindSuccessors == nil {
		opts.FindSuccessors = encoding.Successors
	}

	// originating index is needed for multi-architecture images, we only want to pull the data if the descriptor is an index and platforms are defined.
	if encoding.IsIndex(copier.root.MediaType) && platforms != nil {
		// need descriptor to assess the mediaType otherwise I would just use oras.FetchBytes instead of resolving above
		b, err := content.FetchAll(ctx, src, copier.root)
		if err != nil {
			return nil, fmt.Errorf("fetching the originating index: %w", err)
		}
		copier.originatingIndex = b
	}
	return &copier, nil
}

// Copy copies a source target to a destination target while handling cross-repo mounting. It uses the
// ExtendedCopyGraph function in oras to capture predecessor references.
func Copy(ctx context.Context, c *Copier) error {
	// load cross-repo mounting and findSuccessors functions
	// options := modifyOpts(c)

	if c.referrers {
		// we want to use ExtendedCopyGraph so we copy for SBOMs and signatures as well
		options := oras.ExtendedCopyGraphOptions{
			CopyGraphOptions: c.options.CopyGraphOptions,
		}
		if err := oras.ExtendedCopyGraph(ctx, c.src, c.dest, c.root, options); err != nil {
			return fmt.Errorf("copying image (and predecessors): %w", err)
		}
	} else {
		if err := oras.CopyGraph(ctx, c.src, c.dest, c.root, c.options.CopyGraphOptions); err != nil {
			return fmt.Errorf("copying image: %w", err)
		}
	}
	// we do not use oras.ExtendedCopy() because we copy from a descriptor (not a ref)

	return nil
}

// CopyFilterOnPlatform will copy from the root descriptor only the manifests that match the platforms defined.
func CopyFilterOnPlatform(ctx context.Context, c *Copier) ([]ocispec.Descriptor, error) {
	var platformDescriptors []ocispec.Descriptor
	switch {
	case encoding.IsImage(c.root.MediaType):
		var err error
		platformDescriptors, err = copyFilterOnPlatformManifest(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("copying image manifest: %w", err)
		}
	case encoding.IsIndex(c.root.MediaType):
		var err error
		platformDescriptors, err = copyFilterOnPlatformIndex(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("copying image index: %w", err)
		}
	default:
		return nil, fmt.Errorf("mediatype unsupported for copying by platform, got '%s', want '%s' or '%s'", c.root.MediaType, ocispec.MediaTypeImageManifest, ocispec.MediaTypeImageIndex)
	}

	return platformDescriptors, nil
}

func copyFilterOnPlatformIndex(ctx context.Context, c *Copier) ([]ocispec.Descriptor, error) {
	var platformDescriptors []ocispec.Descriptor
	var idx ocispec.Index
	b, err := content.FetchAll(ctx, c.src, c.root)
	if err != nil {
		return nil, fmt.Errorf("error fetching index manifest: %w", err)
	}
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, fmt.Errorf("error unmarshalling the index manifest: %w", err)
	}

	copyErrs := make([]error, 0)
	for _, manDesc := range idx.Manifests {
		platform := manDesc.Platform
		for i, wantPlatform := range c.platforms {
			if match(platform, wantPlatform) {
				err := oras.CopyGraph(ctx, c.src, c.dest, manDesc, c.options.CopyGraphOptions)
				if err != nil && !errors.Is(err, errdef.ErrNotFound) {
					copyErrs = append(copyErrs, fmt.Errorf("copying sub-DAG for platform '%s': %w", platform.OS+"/"+platform.Architecture, err))
					continue
				}
				platformDescriptors = append(platformDescriptors, manDesc)
				break
			} else if i == len(c.platforms) {
				// no match
				c.log.InfoContext(ctx, "platform not found in index", "indexDesc", c.root, "platform", platform.OS+"/"+platform.Architecture)
			}
		}

	}
	if len(copyErrs) > 0 {
		return nil, errors.Join(copyErrs...)
	}
	return platformDescriptors, nil
}

func copyFilterOnPlatformManifest(ctx context.Context, c *Copier) ([]ocispec.Descriptor, error) {
	var platformDescriptors []ocispec.Descriptor
	var img ocispec.Manifest
	b, err := content.FetchAll(ctx, c.src, c.root)
	if err != nil {
		return nil, fmt.Errorf("error fetching manifest: %w", err)
	}
	if err := json.Unmarshal(b, &img); err != nil {
		return nil, fmt.Errorf("error unmarshalling the manifest: %w", err)
	}

	// this allows signatures, helm charts, bottles, etc to still be pushed because they have special config media types.
	if img.Config.MediaType != ocispec.MediaTypeImageConfig {
		platformDescriptors = append(platformDescriptors, c.root)
		// return the root descriptor and copy it normally.
		return platformDescriptors, Copy(ctx, c)
	}

	configBytes, err := content.FetchAll(ctx, c.src, img.Config)
	if err != nil {
		return nil, fmt.Errorf("fetching manifest config: %w", err)
	}

	var platform ocispec.Platform
	err = json.Unmarshal(configBytes, &platform)
	if err != nil {
		return nil, fmt.Errorf("decoding manifest config: %w", err)
	}

	for i, wantPlatform := range c.platforms {
		if match(&platform, wantPlatform) {
			err := oras.CopyGraph(ctx, c.src, c.dest, c.root, c.options.CopyGraphOptions)
			if err != nil && !errors.Is(err, errdef.ErrNotFound) {
				return nil, fmt.Errorf("copying sub-DAG for platform '%s': %w", platform.OS+"/"+platform.Architecture, err)
			}
			platformDescriptors = append(platformDescriptors, c.root)
			break
		} else if i == len(c.platforms) {
			// no match
			c.log.InfoContext(ctx, "manifest does not match any wanted platform", "desc", c.root, "platform", platform.OS+"/"+platform.Architecture)
		}
	}
	return platformDescriptors, nil
}

func mountFrom(srcRef, destRef registry.Reference) func(ctx context.Context, desc ocispec.Descriptor) ([]string, error) {
	if srcRef.Registry == destRef.Registry {
		return func(ctx context.Context, desc ocispec.Descriptor) ([]string, error) {
			return []string{srcRef.Repository}, nil
		}
	}
	return func(ctx context.Context, desc ocispec.Descriptor) ([]string, error) {
		return nil, nil
	}

}

func onMounted(log *slog.Logger) func(ctx context.Context, desc ocispec.Descriptor) error {
	return func(ctx context.Context, desc ocispec.Descriptor) error {
		log.InfoContext(ctx, "Mounted", "descriptor", desc)
		return nil
	}
}

// match is sourced from oras-go/internal/platform/platform.go.
//
// Match checks whether the current platform matches the target platform.
// Match will return true if all of the following conditions are met.
//   - Architecture and OS exactly match.
//   - Variant and OSVersion exactly match if target platform provided.
//   - OSFeatures of the target platform are the subsets of the OSFeatures
//     array of the current platform.
//
// Note: Variant, OSVersion and OSFeatures are optional fields, will skip
// the comparison if the target platform does not provide specific value.
func match(got *ocispec.Platform, want *ocispec.Platform) bool {
	if got == nil && want == nil {
		return true
	}

	if got == nil || want == nil {
		return false
	}

	if got.Architecture != want.Architecture || got.OS != want.OS {
		return false
	}

	if want.OSVersion != "" && got.OSVersion != want.OSVersion {
		return false
	}

	if want.Variant != "" && got.Variant != want.Variant {
		return false
	}

	if len(want.OSFeatures) != 0 && !isSubset(want.OSFeatures, got.OSFeatures) {
		return false
	}

	return true
}

// isSubset is sourced from oras-go/internal/platform/platform.go.
// isSubset returns true if all items in slice A are present in slice B.
func isSubset(a, b []string) bool {
	set := make(map[string]bool, len(b))
	for _, v := range b {
		set[v] = true
	}
	for _, v := range a {
		if _, ok := set[v]; !ok {
			return false
		}
	}

	return true
}
