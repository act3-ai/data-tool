package orasutil

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	orascontent "oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/data/tool/internal/cache"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// BlobCacher wraps oras interfaces ensuring they share a persistent blob cache.
type BlobCacher struct {
	fcache *cache.FileCache
}

// NewBlobCacher returns a BlobCacher that utilizes a shared peristant blob storage
// for all oras.GraphTargets that it creates.
func NewBlobCacher(ctx context.Context, root string, opts ...cache.FileCacheOpt) (*BlobCacher, error) {
	fcache, err := cache.NewFileCache(root, opts...)
	if err != nil {
		return nil, fmt.Errorf("accessing cache storage: %w", err)
	}

	return &BlobCacher{
		fcache: fcache,
	}, nil
}

// GraphTarget wraps an oras.GraphTarget with a shared peristant blob storage.
func (bc *BlobCacher) GraphTarget(gt oras.GraphTarget) oras.GraphTarget {
	return &CachedGraphTarget{
		GraphTarget: gt,
		Cache:       bc.fcache,
	}
}

// CachedGraphTarget implements oras.GraphTarget, preferring to fetch blobs
// from a persistent cache and falling back to the remote if necessary.
type CachedGraphTarget struct {
	oras.GraphTarget

	Cache orascontent.GraphStorage
}

// Fetch prefers to read from the local cache, falling back to the remote on misses.
func (c *CachedGraphTarget) Fetch(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error) {
	return fetch(ctx, c.Cache, target, c.GraphTarget.Fetch)
}

// Push caches blobs while pushing to the remote.
func (c *CachedGraphTarget) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	return push(ctx, c.Cache, expected, content, c.GraphTarget.Push)
}

// Predecessors returns the union of locally known predecessors and remote known predecessors.
func (c *CachedGraphTarget) Predecessors(ctx context.Context, node ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	return predecessors(ctx, c.Cache, node, c.GraphTarget.Predecessors)
}

// fetch prefers to read from a local cache (storage), falling back to the
// remote with the fetcherFn on misses.
func fetch(ctx context.Context, storage orascontent.Storage, target ocispec.Descriptor, remoteFetcherFn func(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error)) (io.ReadCloser, error) {
	log := logger.FromContext(ctx)

	rc, err := storage.Fetch(ctx, target)
	if err != nil {
		log.DebugContext(ctx, "unable to fetch blob from cache, falling back to remote", "error", err)

		remoterc, err := remoteFetcherFn(ctx, target)
		if err != nil {
			return nil, fmt.Errorf("fetching blob from remote: %w", err)
		}
		defer remoterc.Close()

		// since we don't have control over when the returned io.ReadCloser is read,
		// we must first fully cache the blob, then return a reader; i.e. we cannot
		// fully stream here.
		err = storage.Push(ctx, target, remoterc)
		if err != nil {
			return nil, fmt.Errorf("caching fetched blob: %w", err)
		}

		err = remoterc.Close()
		if err != nil {
			return nil, fmt.Errorf("closing remote fetcher: %w", err)
		}

		localrc, err := storage.Fetch(ctx, target)
		if err != nil {
			return nil, fmt.Errorf("fetching reader for cached blob: %w", err)
		}
		return localrc, nil
	}
	return rc, nil
}

// push caches blobs (to storage) while pushing with the provided pusherFn.
func push(ctx context.Context, storage orascontent.Storage, expected ocispec.Descriptor, content io.Reader, pusherFn func(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error) error {
	log := logger.FromContext(ctx)

	cached, err := storage.Exists(ctx, expected)
	switch {
	case err != nil:
		log.DebugContext(ctx, "unable to check blob existence in cache", "error", err)
		fallthrough
	case cached:
		log.InfoContext(ctx, "skipping caching of blob")
		err = pusherFn(ctx, expected, content)
		if err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
			return fmt.Errorf("pushing blob to remote: %w", err)
		}
	default:
		// TODO: if caching the blob fails, it may block the push to the remotes?
		pr, pw := io.Pipe()
		tr := io.TeeReader(content, pw)
		defer pw.Close()

		done := make(chan struct{})
		var cacheErr error
		go func() {
			cacheErr = storage.Push(ctx, expected, pr)
			close(done)
		}()

		err = pusherFn(ctx, expected, tr)
		switch {
		case err != nil && !errors.Is(err, errdef.ErrAlreadyExists):
			return fmt.Errorf("pushing blob to remote: %w", err)
		case errors.Is(err, errdef.ErrAlreadyExists):
			log.InfoContext(ctx, "blob already exists in remote, discarding remote push, continuing caching")
			// ensure we complete writes to pipe, so the cache push may read them
			_, err := io.Copy(io.Discard, tr)
			if err != nil {
				return fmt.Errorf("discarding duplicate remote blob: %w", err)
			}
			fallthrough
		default:
			err = pw.Close()
			if err != nil {
				// goroutine will not return
				return fmt.Errorf("closing pipe writer: %w", err)
			}

			<-done
			if cacheErr != nil {
				return fmt.Errorf("caching blob: %w", err)
			}
		}
	}

	return nil
}

// predecessors returns the union of predecessors known by the cache (storage) and the predecessors known by the remote.
func predecessors(ctx context.Context, storage orascontent.GraphStorage, node ocispec.Descriptor,
	remotePredecessors func(ctx context.Context, node ocispec.Descriptor) ([]ocispec.Descriptor, error)) ([]ocispec.Descriptor, error) {

	rp, err := remotePredecessors(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("retrieving predecessors from remote: %w", err)
	}

	lp, err := storage.Predecessors(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("retrieving predecessors from local storage: %w", err)
	}

	resolver := make(map[digest.Digest]ocispec.Descriptor, len(lp))
	for _, desc := range rp {
		resolver[desc.Digest] = desc
	}

	// take the union, including duplicate digests if the mediatypes are different
	// TODO: Do we want to include dups with different mediatypes?
	for _, desc := range lp {
		existingDesc, ok := resolver[desc.Digest]
		if !ok || existingDesc.MediaType != desc.MediaType {
			rp = append(rp, desc)
			continue
		}
	}

	return rp, nil
}
