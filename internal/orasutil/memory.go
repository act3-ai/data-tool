package orasutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	"github.com/act3-ai/data-tool/internal/mirror/encoding"
)

// ErrMissingSuccessor indicates that a success is issing.
var ErrMissingSuccessor = errors.New("missing successor")

// CheckedStorage ensures that all successors are present before accepting a Push.
// This is to more closely mimic an OCI registry.
type CheckedStorage struct {
	Target oras.GraphTarget
}

// Push implements content.Fetcher and overrides that in the embedded oras.Target.
func (cs *CheckedStorage) Push(ctx context.Context, expected ocispec.Descriptor, r io.Reader) error {
	// only check successors on manifests
	if encoding.IsManifest(expected.MediaType) {
		data, err := content.ReadAll(r, expected)
		if err != nil {
			return fmt.Errorf("reading manifest: %w", err)
		}

		var fetcher content.FetcherFunc = func(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
			if desc.Digest != expected.Digest || desc.Size != expected.Size || desc.MediaType != expected.MediaType {
				return nil, errdef.ErrNotFound
			}
			return io.NopCloser(bytes.NewReader(data)), nil
		}

		// it is a manifest so check that the successors exist
		successors, err := content.Successors(ctx, fetcher, expected)
		if err != nil {
			return fmt.Errorf("finding successors to check: %w", err)
		}

		for _, successor := range successors {
			exists, err := cs.Target.Exists(ctx, successor) // TODO cache this
			if err != nil {
				return fmt.Errorf("checking existence of successor: %w", err)
			}

			if !exists {
				return fmt.Errorf("missing successor %+v: %w", successor, ErrMissingSuccessor)
			}
		}

		r = io.NopCloser(bytes.NewReader(data))
	}

	err := cs.Target.Push(ctx, expected, r)
	if err != nil {
		return fmt.Errorf("pushing content: %w", err)
	}

	return nil
}
