package encoding

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

/*
Below is the Tracker logic for knowing when to send a taggable to the registry (i.e., that all of its dependencies are satisfied).

NOTE a Taggable is either an image or image index.  It is something you can tag in a registry.  This excludes blobs and image configs since they are not taggable.

This tracker does not directly interact with the registry (for testability).

List of all the events that cause a taggable to be completed:
- New blob (may complete many images or indexes)
- New image (may complete itself but also can complete indexes that point to it)
- Index designates a blob as a image (may complete many images or indexes)
- Index designates a blob as a index (may complete many images or indexes)

It is possible to have the same file digest be used for more than one taggable.  For instance a Manifest might be is in an Index by itself and in a sub index.
Therefore the tracker must keep the data (keyed on digest) and allow it to be used for creating multiple taggables.

NOTE the current blob tells us nothing about the type.
We only know its name/digest and the contents, but we should NEVER inspect the contents until we know the type.

NOTE that the digest (i.e., algorithm selection) of a blob is ONLY determined by the referrer which is encoded in the name of the file in the archive.  This is also true for other fields in the Descriptor (e.g., media type).

We do not want to assume that all the manifests will be included in the data stream every time.  We are going to store manifests (before we know they are a manifest) as blobs in the registry (and cached in-memory locally after we know it is a manifest)

QUESTION What does a sane registry do when you send it a manifest (same digest) twice but with two different Content-Types?
ANSWER This is not possible.  The OCI spec says that if the media type is in the manifest then that must be set to the content-type in the request.  In other words the registry does not separately store the media type of a manifest.

RISK We are assuming that you can upload a blob and then get it back without including it in a manifest first.
Registries allow blobs to exist for at least some amount of time before GC deletes them.  They must due to the way images are uploaded (blobs first then manifest).
*/

// TODO Some of this functionality might be able to be cast in terms of calls to Predecessors on the manifest cache.
// We would need to modify oras-go to alow a custom Successors() but that might be it.

// incompleteTaggable represents an incomplete manifest (image or index).
type incompleteTaggable struct {
	ocispec.Descriptor

	// number of missing (from the repo) referenced descriptors (layers, config, or manifests).
	// when this becomes zero it is safe to send the taggable to the repository.
	missing int
}

// TaggableTracker is used to tack taggables as we become aware of them through the deserialize process.
type TaggableTracker struct {
	// missingBlobs are missing blobs (layer, config) indexed by digest.
	// The value is the list of all taggables that are missing this digest.
	// The number of times a taggable shows up in missingBlobs is always equal to "missing" of the incompleteTaggable.
	missingBlobs map[digest.Digest][]*incompleteTaggable

	// missingTaggables are missing taggables (index and image) indexed by digest.
	// The value is the list of all taggables that are missing this digest.
	// The number of times a taggable shows up in missingTaggables is always equal to "missing" of the incompleteTaggable.
	missingTaggables map[digest.Digest][]*incompleteTaggable

	// taggableDescriptors keeps a record of the descriptors we have seen, indexed by digest.
	// Only keeps descriptors of taggables (index and image but not layers or config).
	taggableDescriptors map[digest.Digest][]ocispec.Descriptor

	// target is where completed manifests are pushed.
	target content.Storage

	// cache of taggables for consumption by the tracker.
	cache content.Storage

	FindSuccessors func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error)
}

// NewTaggableTracker creates a new tracker.
func NewTaggableTracker(target content.Storage, cache content.Storage) *TaggableTracker {
	return &TaggableTracker{
		missingBlobs:        make(map[digest.Digest][]*incompleteTaggable),
		missingTaggables:    make(map[digest.Digest][]*incompleteTaggable),
		taggableDescriptors: make(map[digest.Digest][]ocispec.Descriptor),
		target:              target,
		cache:               cache,
	}
}

// KnownTaggable returns the descriptor if the digest refers to a taggable that is already known to the tracker.
// Nil otherwise.
// This is conservative so at a later time this might return a non-nil descriptor when it previously returned nil.
// As such, it can only be used for optimizations.
func (mt *TaggableTracker) KnownTaggable(h digest.Digest) (desc *ocispec.Descriptor) {
	// return the small descriptor (mediaType, digest, size) since that MUST match the
	for _, d := range mt.taggableDescriptors[h] {
		if desc != nil {
			// check uniqueness
			if desc.MediaType != d.MediaType ||
				desc.Digest != d.Digest ||
				desc.Size != d.Size {
				panic(fmt.Sprintf("expected MediaType, Digest, and Size to match for all but found %v != %v", desc, d))
			}
		}

		// ignore other fields because they may not be the same for all descriptors
		desc = &ocispec.Descriptor{
			MediaType: d.MediaType,
			Digest:    d.Digest,
			Size:      d.Size,
		}
	}
	return
}

// MissingBlobs returns the digests of the blobs that are known to be missing given the manifests we know about.
// The missing digests do not contain duplicates and are sorted.
func (mt *TaggableTracker) MissingBlobs() []digest.Digest {
	missing := make([]digest.Digest, 0, len(mt.missingBlobs)+len(mt.missingTaggables))
	for h := range mt.missingBlobs {
		missing = append(missing, h)
	}
	for h := range mt.missingTaggables {
		missing = append(missing, h)
	}
	sort.Slice(missing, func(i, j int) bool {
		if missing[i].Algorithm() < missing[j].Algorithm() {
			return true
		}
		return missing[i].Encoded() < missing[j].Encoded()
	})
	return missing
}

// complete will push the completed taggable and any parent taggables as a result of its completion.
// desc MUST be in the cache.
func (mt *TaggableTracker) complete(ctx context.Context, desc ocispec.Descriptor) error {
	// get the manifest data from the cache
	manifestData, err := content.FetchAll(ctx, mt.cache, desc)
	if err != nil {
		return fmt.Errorf("fetching completed manifest from cache: %w", err)
	}

	// Push the first Taggable
	if err := tryPushBytes(ctx, mt.target, desc, manifestData); err != nil {
		return fmt.Errorf("pushing complete manifest: %w", err)
	}

	// we recurse to handle "index to index" because an index could now be completed due to an index being completed and so on...
	h := desc.Digest
	for _, ic := range mt.missingTaggables[h] {
		ic.missing--
		if ic.missing == 0 {
			// this taggable is now complete
			if err := mt.complete(ctx, ic.Descriptor); err != nil {
				return err
			}
		}
	}
	// this manifest is no longer missing

	delete(mt.missingTaggables, h)
	return nil
}

// NotifyManifest notifies the tracker that the descriptor exists.
// That data need not exist yet.
func (mt *TaggableTracker) NotifyManifest(ctx context.Context, desc ocispec.Descriptor) error {
	exists, err := mt.notifyManifest(ctx, desc)
	if err != nil {
		return err
	}

	if exists {
		if err := mt.complete(ctx, desc); err != nil {
			return err
		}
	}

	return nil
}

// NotifyManifest notifies the tracker that the descriptor exists.
// That data need not exist yet.  Returns true if the data exists.
func (mt *TaggableTracker) notifyManifest(ctx context.Context, desc ocispec.Descriptor) (bool, error) {
	log := logger.FromContext(ctx).With("digest", desc.Digest)

	log.InfoContext(ctx, "Notify manifest")

	if !IsManifest(desc.MediaType) {
		panic(fmt.Sprintf("%q is not a media type of a manifest", desc.MediaType))
	}

	mt.taggableDescriptors[desc.Digest] = append(mt.taggableDescriptors[desc.Digest], desc)

	manifestData, err := content.FetchAll(ctx, mt.target, desc)
	switch {
	case errors.Is(err, errdef.ErrNotFound):
		// manifest is missing
	case err != nil:
		return false, fmt.Errorf("checked existence of manifest: %w", err)
	default:
		log.InfoContext(ctx, "Manifest already in repository")

		// push to the cache
		if err := tryPushBytes(ctx, mt.cache, desc, manifestData); err != nil {
			return false, fmt.Errorf("populating cache with manifest: %w", err)
		}

		// nothing else to do with this descriptor.  It is already complete.
		return true, nil
	}

	// try promoting a blob to a manifest
	// We know we will need the contents of the blob (really a manifest) in addTaggable -> FindSuccessors so we fetch it now
	target := desc                                // copy
	target.MediaType = "application/octet-stream" // not a manifest media type
	manifestData, err = content.FetchAll(ctx, mt.target, target)
	if errors.Is(err, errdef.ErrNotFound) {
		// blob is missing
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("fetch manifest as a blob: %w", err)
	}

	log.InfoContext(ctx, "Manifest is stored as a blob")
	// store the manifest data in the cache for later
	if err := tryPushBytes(ctx, mt.cache, desc, manifestData); err != nil {
		return false, fmt.Errorf("populating cache with manifest (from blob): %w", err)
	}

	// we have the manifest data in the cache so process the taggable
	if err := mt.addTaggable(ctx, desc); err != nil {
		return false, err
	}

	// the addTaggable() above might complete this taggable so we have to
	// check to see if it is now in the repo as a manifest
	ic, err := mt.target.Exists(ctx, desc)
	if err != nil {
		return false, fmt.Errorf("checking existence of manifest again: %w", err)
	}
	if ic {
		log.InfoContext(ctx, "Manifest is in the repository")
		return true, nil
	}
	return false, nil
}

// AddBlob notifies the tracker that a blob was added to the registry.
// If it KnownTaggable(desc) != nil then the manifest data MUST be in the cache.
func (mt *TaggableTracker) AddBlob(ctx context.Context, h digest.Digest) error {
	// loop over all known descriptors for taggables for this data and process them
	for _, descriptor := range mt.taggableDescriptors[h] {
		if err := mt.addTaggable(ctx, descriptor); err != nil {
			return err
		}
	}

	for _, ic := range mt.missingBlobs[h] {
		ic.missing--
		if ic.missing == 0 {
			if err := mt.complete(ctx, ic.Descriptor); err != nil {
				return err
			}
		}
	}

	// this blob is no longer missing
	delete(mt.missingBlobs, h)

	// completed may contain duplicates (taggables with the same data) but it is unlikely
	return nil
}

// addTaggable adds the manifest (index or image) to the tracker and processes all newly discovered completed manifests.
// The data for the manifest must be available from the cache as a manifest.
func (mt *TaggableTracker) addTaggable(ctx context.Context, desc ocispec.Descriptor) error {
	log := logger.FromContext(ctx).With("digest", desc.Digest)

	if !IsManifest(desc.MediaType) {
		panic(fmt.Sprintf("%q is not a media type of a manifest", desc.MediaType))
	}
	log.InfoContext(ctx, "Processing manifest in tracker")

	// fetch from the cache
	successors, err := mt.FindSuccessors(ctx, mt.cache, desc)
	if err != nil {
		return fmt.Errorf("fetching successors: %w", err)
	}

	taggable := &incompleteTaggable{
		Descriptor: desc,
	}

	// process the missing taggables that are referenced in the index
	for _, desc := range successors {
		if IsManifest(desc.MediaType) {
			exists, err := mt.notifyManifest(ctx, desc)
			if err != nil {
				return err
			}

			if !exists {
				log.InfoContext(ctx, "Storing a record of the missing manifest")
				// add them to "manifests" by append to the list
				taggable.missing++
				mt.missingTaggables[desc.Digest] = append(mt.missingTaggables[desc.Digest], taggable)
			}
		} else {
			exists, err := mt.target.Exists(ctx, desc)
			if err != nil {
				return fmt.Errorf("checking existence of blob: %w", err)
			}

			if !exists {
				// blob is missing
				log.InfoContext(ctx, "Storing a record of the missing blob")
				// add them to "blobs" by appending to the list
				taggable.missing++
				mt.missingBlobs[desc.Digest] = append(mt.missingBlobs[desc.Digest], taggable)
			}
		}
	}

	if taggable.missing == 0 {
		// this taggable is complete
		if err := mt.complete(ctx, taggable.Descriptor); err != nil {
			return err
		}
	}

	return nil
}

// tryPushBytes will try to push the given bytes and if already exists the error is ignored.
func tryPushBytes(ctx context.Context, pusher content.Pusher, desc ocispec.Descriptor, data []byte) error {
	if int64(len(data)) != desc.Size {
		panic("Size mismatch")
	}

	err := pusher.Push(ctx, desc, bytes.NewReader(data))
	switch {
	case errors.Is(err, errdef.ErrAlreadyExists):
	case err != nil:
		return err //nolint:wrapcheck
	}
	return nil
}
