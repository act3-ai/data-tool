package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orascontent "oras.land/oras-go/v2/content"

	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// PredecessorCacher wraps an oras content.Storage to cache referrers included in manifests
// during Fetch, Push, and Exists operations. Implements oras.GraphStorage.
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
	default:
		if target.MediaType != ocispec.MediaTypeImageManifest &&
			target.MediaType != ocispec.MediaTypeImageIndex {
			return true, nil
		}

		rc, err := pc.Storage.Fetch(ctx, target)
		if err != nil {
			return true, fmt.Errorf("fetching manifest from embedded storage: %w", err)
		}
		defer rc.Close()

		manifestBytes, err := io.ReadAll(rc)
		if err != nil {
			return true, fmt.Errorf("reading manifest into memory: %w", err)
		}

		err = pc.addAsPredecessor(ctx, manifestBytes, target)
		if err != nil {
			return true, fmt.Errorf("adding potential predecessors: %w", err)
		}
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
	case desc.MediaType != ocispec.MediaTypeImageManifest &&
		desc.MediaType != ocispec.MediaTypeImageIndex:
		return rc, nil
	default:
		manifestBytes, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("reading manifest into memory: %w", err)
		}

		err = pc.addAsPredecessor(ctx, manifestBytes, desc)
		if err != nil {
			return nil, fmt.Errorf("adding potential predecessors: %w", err)
		}
		return io.NopCloser(bytes.NewReader(manifestBytes)), nil
	}
}

// Push will read the content into memory if the expected descriptor has an image manifest
// or index mediatype, establish a predecessor if applicable, finally propagating the push
// to the underlying content.Storage.
func (pc *PredecessorCacher) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	if expected.MediaType != ocispec.MediaTypeImageManifest &&
		expected.MediaType != ocispec.MediaTypeImageIndex {
		return pc.Storage.Push(ctx, expected, content) //nolint
	}

	manifestBytes, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("reading manifest into memory: %w", err)
	}

	err = pc.addAsPredecessor(ctx, manifestBytes, expected)
	if err != nil {
		return fmt.Errorf("adding potential predecessors: %w", err)
	}
	return pc.Storage.Push(ctx, expected, bytes.NewReader(manifestBytes)) //nolint
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

// addAsPredecessors attempts to add a blob as a predecessor of its subject, if set, in the
// in-memory predecessor map. It is safe to call for all mediatypes, but only establishes
// predecessors for mediatypes: ocispec.MediaTypeImageManifest and ocispec.MediaTypeImageIndex.
func (pc *PredecessorCacher) addAsPredecessor(ctx context.Context, blob []byte, desc ocispec.Descriptor) error {
	// addAsPredecessors could take a reader instead, allowing us to once again tee our reader stream.
	// Since we end up storing the manifest in memery anyhow, we can slightly reduce complexity by reading
	// it into memory earlier.
	log := logger.FromContext(ctx).With("blobDigest", desc.Digest)

	var subjectDigest digest.Digest
	var subjectMediaType string
	switch {
	case desc.MediaType == ocispec.MediaTypeImageManifest:
		var manifest ocispec.Manifest
		err := json.Unmarshal(blob, &manifest)
		if err != nil {
			return fmt.Errorf("failed to decode manifest blob: %w", err)
		}

		if manifest.Subject != nil {
			subjectDigest = manifest.Subject.Digest
			subjectMediaType = manifest.Subject.MediaType
		}
	case desc.MediaType == ocispec.MediaTypeImageIndex:
		var index ocispec.Index
		err := json.Unmarshal(blob, &index)
		if err != nil {
			return fmt.Errorf("failed to decode index manifest blob: %w", err)
		}

		if index.Subject != nil {
			subjectDigest = index.Subject.Digest
			subjectMediaType = index.Subject.MediaType
		}
	default:
		log.InfoContext(ctx, "unknown mediatype, skipping evaluation of subject status", "mediatype", desc.MediaType)
		return nil
	}

	if subjectDigest != "" {
		pc.pMux.Lock()
		existingList, ok := pc.predecessors[subjectDigest]
		if ok {
			for _, desc := range existingList {
				if desc.Digest == subjectDigest && desc.MediaType == subjectMediaType {
					pc.pMux.Unlock()
					return nil // blob is already known to be a predecessor
				}
			}
			pc.predecessors[subjectDigest] = append(pc.predecessors[subjectDigest], desc)

		} else {
			pc.predecessors[subjectDigest] = []ocispec.Descriptor{desc}
		}
		pc.pMux.Unlock()
		log.InfoContext(ctx, "adding blob manifest to subject's predecessors", "subjectDigest", subjectDigest)
	}

	return nil
}
