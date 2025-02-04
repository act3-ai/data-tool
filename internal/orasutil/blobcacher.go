package orasutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	orascontent "oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/errdef"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// BlobCacher wraps oras interfaces ensuring they share a persistent blob cache.
type BlobCacher struct {
	cache orascontent.Storage
}

// NewBlobCacher returns a BlobCacher that utilizes a shared peristant file storage
// for all oras.GraphTargets that it creates. Safe to use if root does not yet exist.
func NewBlobCacher(root string) (*BlobCacher, error) {
	if _, err := os.Stat(root); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(root, 0777); err != nil {
			return nil, fmt.Errorf("creating cache directory: %w", err)
		}
	}

	storage, err := oci.NewStorage(root)
	if err != nil {
		return nil, fmt.Errorf("initializing cache storage interface: %w", err)
	}

	return &BlobCacher{
		cache: storage,
	}, nil
}

// GraphTarget wraps an oras.GraphTarget with a shared peristant blob storage.
func (bc *BlobCacher) GraphTarget(gt oras.GraphTarget) oras.GraphTarget {
	return &CachedGraphTarget{
		GraphTarget: gt,
		Cache:       bc.cache,
	}
}

// CachedGraphTarget implements oras.GraphTarget, preferring to fetch blobs
// from a persistent cache and falling back to the remote if necessary.
type CachedGraphTarget struct {
	oras.GraphTarget

	Cache orascontent.Storage
}

// Fetch prefers to read from the local cache, falling back to the remote on misses.
func (c *CachedGraphTarget) Fetch(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error) {
	return fetch(ctx, c.Cache, target, c.GraphTarget.Fetch)
}

// Push caches blobs while pushing to the remote.
func (c *CachedGraphTarget) Push(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error {
	return push(ctx, c.Cache, expected, content, c.GraphTarget.Push)
}

// fetch prefers to read from a local cache (storage), falling back to the
// remoteFetcherFn on misses.
func fetch(ctx context.Context, storage orascontent.Storage, target ocispec.Descriptor, remoteFetcherFn func(ctx context.Context, target ocispec.Descriptor) (io.ReadCloser, error)) (io.ReadCloser, error) {
	log := logger.FromContext(ctx).With("digest", target.Digest)

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
		if err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
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
	log.InfoContext(ctx, "fetching blob from cache")
	return rc, nil
}

// push caches blobs while pushing with the provided remotePusherFn.
func push(ctx context.Context, storage orascontent.Storage, expected ocispec.Descriptor, content io.Reader, remotePusherFn func(ctx context.Context, expected ocispec.Descriptor, content io.Reader) error) error {
	log := logger.FromContext(ctx).With("digest", expected.Digest)

	cached, err := storage.Exists(ctx, expected)
	switch {
	case err != nil:
		log.DebugContext(ctx, "unable to check blob existence in cache", "error", err)
		fallthrough
	case cached:
		log.InfoContext(ctx, "blob already cached")
		err = remotePusherFn(ctx, expected, content)
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

		err = remotePusherFn(ctx, expected, tr)
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
			if cacheErr != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
				return fmt.Errorf("caching blob: %w", err)
			}
		}
	}

	return nil
}
