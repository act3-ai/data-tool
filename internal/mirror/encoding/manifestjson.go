package encoding

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"

	"git.act3-ace.com/ace/data/tool/internal/ref"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// ManifestJSON is the top level structure of the manifest.json file.
type ManifestJSON struct {
	Manifests []ManifestInfo

	// dedup, ensuring unique references but allowing multiple to the same digest
	processedRefs    map[string]struct{}
	processedDigests map[digest.Digest]struct{}
}

// ManifestInfo is an entry in the manifest.json file.
type ManifestInfo struct {
	dgst     digest.Digest // why isn't standard?
	Config   string        `json:"Config"`
	RepoTags []string      `json:"RepoTags"`
	Layers   []string      `json:"Layers"`
}

// BuildManifestJSON iterates through a list of manifests adding them to the manifest.json appropriately. If a manifest is an index,
// it selects one manifest from it prioritized by platform: user's current platform, linux/amd64, the first manifest in
// the index. If a user wants a specific platform they should use an image gathered with the appropriate platform option.
func BuildManifestJSON(ctx context.Context, fetcher content.Fetcher, manifests []ocispec.Descriptor) (ManifestJSON, error) {
	log := logger.FromContext(ctx)

	mj := ManifestJSON{
		// may malloc more than necessary, in the unlikely event there are duplicates
		Manifests:        make([]ManifestInfo, 0, len(manifests)),
		processedRefs:    make(map[string]struct{}, len(manifests)),
		processedDigests: make(map[digest.Digest]struct{}, len(manifests)),
	}

	for _, man := range manifests {
		r, ok := man.Annotations[ref.AnnotationSrcRef]
		if !ok {
			continue
		}
		if _, ok := mj.processedRefs[r]; ok {
			// case hits if index was gathered with multiple platforms specified explicitly, we choose the first encountered
			log.InfoContext(ctx, "skipping duplicate entry for the same reference", "ref", r, "digest", man.Digest)
			continue
		}

		// inefficient if manifests are not cached by fetcher
		mBytes, err := content.FetchAll(ctx, fetcher, man)
		if err != nil {
			return ManifestJSON{}, fmt.Errorf("fetching manifest from source: %w", err)
		}

		// manifests are added directly, while we take only one entry from an index
		switch {
		case IsImage(man.MediaType):
			if err := mj.addManifest(ctx, r, man.Digest, mBytes); err != nil {
				return ManifestJSON{}, fmt.Errorf("adding manifest to manifest.json: %w", err)
			}
		case IsIndex(man.MediaType):
			if err := mj.addManifestFromIndex(ctx, fetcher, r, mBytes); err != nil {
				return ManifestJSON{}, fmt.Errorf("adding manifest from index to manifest.json: %w", err)
			}
		default:
			log.InfoContext(ctx, "skipping evaluation of unknown manifest type for addition to manifest.json", "mediatype", man.MediaType)
			continue
		}
	}

	return mj, nil
}

// addManifest parses an OCI manifest, adding the metadata it contains to the manifest.json.
func (mj *ManifestJSON) addManifest(ctx context.Context, srcRef string, dgst digest.Digest, man []byte) error {
	log := logger.V(logger.FromContext(ctx), 1).With("ref", srcRef, "digest", dgst)

	// if we already have the entry, just add the ref
	if _, ok := mj.processedDigests[dgst]; ok {
		for _, entry := range mj.Manifests {
			if entry.dgst == dgst {
				entry.RepoTags = append(entry.RepoTags, srcRef)
				mj.processedRefs[srcRef] = struct{}{}
				log.InfoContext(ctx, "appended RepoTag to exisiting entry in manifest.json")
				return nil
			}
		}
		// how did we end up here?
		log.ErrorContext(ctx, "failed to find already processed digest entry")
	}

	var manifest ocispec.Manifest
	err := json.Unmarshal(man, &manifest)
	if err != nil {
		return fmt.Errorf("parsing manifest data: %w", err)
	}

	info := ManifestInfo{
		Config:   addPrefix(manifest.Config.Digest),
		RepoTags: []string{srcRef},
		Layers:   make([]string, 0, len(manifest.Layers)),
	}
	for _, desc := range manifest.Layers {
		info.Layers = append(info.Layers, addPrefix(desc.Digest))
	}

	mj.Manifests = append(mj.Manifests, info)
	mj.processedRefs[srcRef] = struct{}{}
	mj.processedDigests[dgst] = struct{}{}
	log.InfoContext(ctx, "added manifest to manifest.json")
	return nil
}

// addManifestFromIndex parses an OCI index, adding the metadata from only one of its manifests to the manifest.json.
func (mj *ManifestJSON) addManifestFromIndex(ctx context.Context, fetcher content.Fetcher, srcRef string, man []byte) error {
	var idx ocispec.Index
	err := json.Unmarshal(man, &idx)
	if err != nil {
		return fmt.Errorf("decoding index manifest: %w", err)
	}

	if len(idx.Manifests) < 1 {
		// how did we end up here?
		logger.FromContext(ctx).ErrorContext(ctx, "skipping evaluation of index manifest with no images", "ref", srcRef)
		return nil
	}

	// the following behavior is similar to that of `ctr image export` and `docker image save`.
	// select one, prioritized by: user's platform, linux/amd64, the first in the list
	// if the user really wants a certain platform, they should have used gather with
	// the appropriate flags; as such, this is really just a fallback and we expect to hit
	// the previous case instead
	var chosen ocispec.Descriptor = idx.Manifests[0]
	for _, img := range idx.Manifests {
		if img.Platform.OS == "linux" && img.Platform.Architecture == "amd64" {
			chosen = img
		}

		if img.Platform.OS == runtime.GOOS && img.Platform.Architecture == runtime.GOARCH {
			chosen = img
			break
		}
	}

	mBytes, err := content.FetchAll(ctx, fetcher, chosen)
	if err != nil {
		return fmt.Errorf("fetching manifest from source: %w", err)
	}

	err = mj.addManifest(ctx, srcRef, chosen.Digest, mBytes)
	if err != nil {
		return fmt.Errorf("adding manifest to manifest.json: %w", err)
	}

	return nil
}

// addPrefix resolves and adds the appropriate prefix to a blob digest as
// it exists in a tar archive, e.g. "blobs/sha256/a1b2c3...".
func addPrefix(dgst digest.Digest) string {
	return filepath.Join("blobs", dgst.Algorithm().String(), dgst.Hex())
}
