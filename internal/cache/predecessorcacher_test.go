package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	tlog "git.act3-ace.com/ace/go-common/pkg/test"
)

// Tests ensure predecessors are established as expected.

func TestPredecessorCacher_Exists(t *testing.T) {
	ctx := logger.NewContext(context.Background(), tlog.Logger(t, 0))
	defer leaktest.Check(t)() //nolint

	t.Run("Basic", func(t *testing.T) {
		// push directly to the storage, to ensure the predecessor
		// is establish on Exists and not on the push itself
		storage := memory.New()
		pc := NewPredecessorCacher(storage)

		// manifestA is the predecessor of blobSubject
		blobSubject := []byte("subject")
		blobSubjectDesc := ocispec.Descriptor{
			MediaType: "application/octet-stream",
			Digest:    digest.FromBytes(blobSubject),
			Size:      int64(len(blobSubject)),
		}

		manifestA := ocispec.Manifest{
			MediaType: ocispec.MediaTypeImageManifest,
			Config:    ocispec.DescriptorEmptyJSON,
			Layers:    []ocispec.Descriptor{},
			Subject:   &blobSubjectDesc,
		}

		manifestABytes, err := json.Marshal(manifestA)
		assert.NoError(t, err)
		manifestADesc, err := oras.PushBytes(ctx, storage, ocispec.MediaTypeImageManifest, manifestABytes)
		assert.NoError(t, err)

		// establish predecessor
		exists, err := pc.Exists(ctx, manifestADesc)
		assert.NoError(t, err)
		if !exists {
			t.Errorf("PredecessorCacher.Exists() error = %v", err)
			return
		}

		// validation
		predecessors, err := pc.Predecessors(ctx, blobSubjectDesc)
		assert.NoError(t, err)
		for _, p := range predecessors {
			if p.Digest == manifestADesc.Digest {
				return
			}
		}
		t.Errorf("failed to find predecessors, expected = %v, got %v", blobSubjectDesc, predecessors)
	})
}

func TestPredecessorCacher_Fetch(t *testing.T) {
	ctx := logger.NewContext(context.Background(), tlog.Logger(t, 0))
	defer leaktest.Check(t)() //nolint

	t.Run("Basic", func(t *testing.T) {
		// push directly to the storage, to ensure the predecessor
		// is establish on Exists and not on the push itself
		storage := memory.New()
		pc := NewPredecessorCacher(storage)

		// manifestA is the predecessor of blobSubject
		blobSubject := []byte("subject")
		blobSubjectDesc := ocispec.Descriptor{
			MediaType: "application/octet-stream",
			Digest:    digest.FromBytes(blobSubject),
			Size:      int64(len(blobSubject)),
		}

		manifestA := ocispec.Manifest{
			MediaType: ocispec.MediaTypeImageManifest,
			Config:    ocispec.DescriptorEmptyJSON,
			Layers:    []ocispec.Descriptor{},
			Subject:   &blobSubjectDesc,
		}

		manifestABytes, err := json.Marshal(manifestA)
		assert.NoError(t, err)
		manifestADesc, err := oras.PushBytes(ctx, storage, ocispec.MediaTypeImageManifest, manifestABytes)
		assert.NoError(t, err)

		// establish predecessor
		rc, err := pc.Fetch(ctx, manifestADesc)
		assert.NoError(t, err)
		assert.NoError(t, rc.Close())

		// validation
		predecessors, err := pc.Predecessors(ctx, blobSubjectDesc)
		assert.NoError(t, err)
		for _, p := range predecessors {
			if p.Digest == manifestADesc.Digest {
				return
			}
		}
		t.Errorf("failed to find predecessors, expected = %v, got %v", blobSubjectDesc, predecessors)
	})
}

func TestPredecessorCacher_Push(t *testing.T) {
	ctx := logger.NewContext(context.Background(), tlog.Logger(t, 0))
	defer leaktest.Check(t)() //nolint

	t.Run("Basic", func(t *testing.T) {
		// here we use the PredecessorCacher push, which directly establishes the predecessor
		storage := memory.New()
		pc := NewPredecessorCacher(storage)

		// manifestA is the predecessor of blobSubject
		blobSubject := []byte("subject")
		blobSubjectDesc := ocispec.Descriptor{
			MediaType: "application/octet-stream",
			Digest:    digest.FromBytes(blobSubject),
			Size:      int64(len(blobSubject)),
		}

		manifestA := ocispec.Manifest{
			MediaType: ocispec.MediaTypeImageManifest,
			Config:    ocispec.DescriptorEmptyJSON,
			Layers:    []ocispec.Descriptor{},
			Subject:   &blobSubjectDesc,
		}

		manifestABytes, err := json.Marshal(manifestA)
		assert.NoError(t, err)

		manifestADesc := ocispec.Descriptor{
			MediaType: ocispec.MediaTypeImageManifest,
			Digest:    digest.FromBytes(manifestABytes),
			Size:      int64(len(manifestABytes)),
		}

		// establish predecessor
		err = pc.Push(ctx, manifestADesc, bytes.NewReader(manifestABytes))
		assert.NoError(t, err)

		// validation
		predecessors, err := pc.Predecessors(ctx, blobSubjectDesc)
		assert.NoError(t, err)
		for _, p := range predecessors {
			if p.Digest == manifestADesc.Digest {
				return
			}
		}
		t.Errorf("failed to find predecessors, expected = %v, got %v", blobSubjectDesc, predecessors)
	})
}
