package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orascontent "oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// PredecessorCacher wraps an oras content.Storage to cache referrers included in manifests
// during Fetch, Push, and Exists operations. It is not efficient for remote
// storages, and is ideal for a local file-based implementation. Implements oras.GraphStorage.
type PredecessorCacher struct {
	// Storage is where fetches/existence checks originate from and pushes are forwarded to.
	orascontent.Storage

	// predecessors is a map of subject to referring descriptors
	pMux         sync.RWMutex
	predecessors map[digest.Digest][]ocispec.Descriptor
}

// NewPredecessorCacher extends and oras content.Storage to a content.GraphStorage
// by storing predecessors in-memory.
func NewPredecessorCacher(storage orascontent.Storage) orascontent.GraphStorage {
	return &PredecessorCacher{
		Storage:      storage,
		predecessors: make(map[digest.Digest][]ocispec.Descriptor),
	}
}

// Exists ultimately returns the same value as a call to Exists to the underlying
// content.Storage. If the target descriptor exists and has an image manifest or index
// mediatype it will fetch from the underlying storage and add establish a predecessor
// if applicable.
func (pc *PredecessorCacher) Exists(ctx context.Context, target ocispec.Descriptor) (bool, error) {
	exists, err := pc.Storage.Exists(ctx, target)
	switch {
	case err != nil:
		return exists, err //nolint
	case !exists:
		return false, nil
	case target.MediaType == ocispec.MediaTypeImageManifest ||
		target.MediaType == ocispec.MediaTypeImageIndex:
		if err := pc.addAsPredecessor(ctx, target); err != nil {
			return true, fmt.Errorf("adding potential predecessors: %w", err)
		}
		return true, nil
	default:
		return true, nil
	}
}

// Fetch first fetches the blob from the underlying content.Storage. If the descriptor has an
// image manifest or index mediatype it will read the manifest into memory, establish a predecessor
// if applicable, finally returning a io.ReadCloser for the in-memory manifest.
func (pc *PredecessorCacher) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	rc, err := pc.Storage.Fetch(ctx, desc)
	switch {
	case err != nil:
		return nil, err //nolint
	case desc.MediaType == ocispec.MediaTypeImageManifest ||
		desc.MediaType == ocispec.MediaTypeImageIndex:
		err = pc.addAsPredecessor(ctx, desc)
		if err != nil {
			return nil, fmt.Errorf("adding potential predecessors: %w", err)
		}
		return rc, nil
	default:
		// blobs
		return rc, nil
	}
}

// Push will read the content into memory if the expected descriptor has an image manifest
// or index mediatype, establish a predecessor if applicable, finally propagating the push
// to the underlying content.Storage.
func (pc *PredecessorCacher) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	// push to the cache first, to avoid data race if pc.Predecessors() is called and we haven't completed
	// the push to the underlying storage yet.
	err := pc.Storage.Push(ctx, expected, content)
	if err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
		return fmt.Errorf("pushing to storage: %w", err)
	}

	if expected.MediaType == ocispec.MediaTypeImageManifest ||
		expected.MediaType == ocispec.MediaTypeImageIndex {
		if err := pc.addAsPredecessor(ctx, expected); err != nil {
			return fmt.Errorf("adding potential predecessors: %w", err)
		}
	}
	// propagate potential ErrAlreadyExists
	return err //nolint
}

// Predecessors finds the nodes directly pointing to a given node of a directed acyclic graph. In other
// words, returns the "parents" of the current descriptor. Predecessors implements oras content.PredecessorFinder, and
// is an extension of oras conent.Storage.
//
// Predecessors returns an error if the FileCache was not initialized with the WithPredecessors Option.
func (pc *PredecessorCacher) Predecessors(ctx context.Context, node ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	pc.pMux.RLock()
	predecessors, ok := pc.predecessors[node.Digest]
	pc.pMux.RUnlock()
	if !ok {
		return []ocispec.Descriptor{}, nil
	}
	return predecessors, nil

}

// addAsPredecessors attempts to add a manifest as a predecessor of its subject, if set, in the
// in-memory predecessor map.
// It MUST only be used for mediatypes ocispec.MediaTypeImageManifest
// and ocispec.MediaTypeImageIndex.
// It expects the descriptor to already exist in the underlying storage. Such an expectation
// helps to avoid potential data races.
func (pc *PredecessorCacher) addAsPredecessor(ctx context.Context, desc ocispec.Descriptor) error {
	log := logger.FromContext(ctx)
	var subjectDigest digest.Digest
	var subjectMediaType string
	switch desc.MediaType {
	case ocispec.MediaTypeImageManifest:
		manBytes, err := orascontent.FetchAll(ctx, pc.Storage, desc)
		if err != nil {
			return fmt.Errorf("fetching manifest from storage: %w", err)
		}

		var manifest ocispec.Manifest
		err = json.Unmarshal(manBytes, &manifest)
		if err != nil {
			return fmt.Errorf("failed to decode manifest: %w", err)
		}

		if manifest.Subject != nil {
			subjectDigest = manifest.Subject.Digest
			subjectMediaType = manifest.Subject.MediaType
		}
	case ocispec.MediaTypeImageIndex:
		manBytes, err := orascontent.FetchAll(ctx, pc.Storage, desc)
		if err != nil {
			return fmt.Errorf("fetching manifest from storage: %w", err)
		}

		var index ocispec.Index
		err = json.Unmarshal(manBytes, &index)
		if err != nil {
			return fmt.Errorf("failed to decode index manifest: %w", err)
		}

		if index.Subject != nil {
			subjectDigest = index.Subject.Digest
			subjectMediaType = index.Subject.MediaType
		}
	default:
		return fmt.Errorf("unknown mediatype '%s', skipping evaluation of subject status for digest '%s'", desc.MediaType, desc.Digest)
	}

	if subjectDigest != "" {
		pc.pMux.Lock()
		defer pc.pMux.Unlock()
		existingList, ok := pc.predecessors[subjectDigest]
		if ok {
			for _, desc := range existingList {
				if desc.Digest == subjectDigest && desc.MediaType == subjectMediaType {
					return nil // blob is already known to be a predecessor
				}
			}
			pc.predecessors[subjectDigest] = append(pc.predecessors[subjectDigest], desc)

		} else {
			pc.predecessors[subjectDigest] = []ocispec.Descriptor{desc}
		}
		log.InfoContext(ctx, "adding blob manifest to subject's predecessors", "subjectDigest", subjectDigest)
	}

	return nil
}
