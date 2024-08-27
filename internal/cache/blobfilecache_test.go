package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	tlog "git.act3-ace.com/ace/go-common/pkg/test"
)

func TestFileCache_Push(t *testing.T) {
	ctx := logger.NewContext(context.Background(), tlog.Logger(t, 0))
	defer leaktest.Check(t)() //nolint

	blobA := []byte("alpha")
	blobADesc := ocispec.Descriptor{
		MediaType: "application/octet-stream",
		Digest:    digest.FromBytes(blobA),
		Size:      int64(len(blobA)),
	}

	type args struct {
		expected ocispec.Descriptor
		content  io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Basic", args{blobADesc, bytes.NewReader(blobA)}, false},
		// Duplicate is idempotentence test, with no idication that optimizations were successful
		{"Duplicate", args{blobADesc, bytes.NewReader(blobA)}, false},
		{"InvalidDigest",
			args{
				ocispec.Descriptor{
					MediaType: "application/octet-stream",
					Digest:    digest.Digest("sha256:BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
					Size:      int64(len(blobA)),
				},
				bytes.NewReader(blobA),
			},
			true},
	}

	defer leaktest.Check(t)() //nolint

	cacheRoot := t.TempDir()
	fc, err := NewFileCache(cacheRoot)
	if err != nil {
		t.Fatalf("initializing file cache, error = %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fc.Push(ctx, tt.args.expected, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("FileCache.Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileCache_Fetch(t *testing.T) {
	ctx := logger.NewContext(context.Background(), tlog.Logger(t, 0))
	defer leaktest.Check(t)() //nolint

	blobA := []byte("alpha")
	blobADesc := ocispec.Descriptor{
		MediaType: "application/octet-stream",
		Digest:    digest.FromBytes(blobA),
		Size:      int64(len(blobA)),
	}

	type args struct {
		desc ocispec.Descriptor
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{"Basic", args{blobADesc}, nil},
		{"NonExistant",
			args{
				ocispec.Descriptor{
					MediaType: "application/octet-stream",
					Digest:    digest.Digest("sha256:BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
					Size:      int64(len(blobA))},
			},
			errdef.ErrNotFound},
	}

	cacheRoot := t.TempDir()
	fc, err := NewFileCache(cacheRoot)
	if err != nil {
		t.Fatalf("initializing file cache, error = %v", err)
	}
	if err := fc.Push(ctx, blobADesc, bytes.NewReader(blobA)); err != nil {
		t.Fatalf("initializing cache with blobA, error = %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc, err := fc.Fetch(ctx, tt.args.desc)
			switch {
			case !errors.Is(err, tt.wantErr):
				// unexpected behavior
				t.Errorf("FileCache.Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			case err != nil && errors.Is(err, tt.wantErr):
				// expected error
				// noop
			default:
				res, err := io.ReadAll(rc)
				if err != nil {
					t.Errorf("reading fetch result, error = %v", err)
				}
				rc.Close()

				if string(res) != string(blobA) {
					t.Errorf("fetch result does not match original, got '%s', want '%s'", res, blobA)
				}
			}
		})
	}
}

// TODO: test Exists, Basic, AfterPushDuplicate, expected failure (NotExists)

func TestFileCache_Exists(t *testing.T) {
	ctx := logger.NewContext(context.Background(), tlog.Logger(t, 0))
	defer leaktest.Check(t)() //nolint

	blobA := []byte("alpha")
	blobADesc := ocispec.Descriptor{
		MediaType: "application/octet-stream",
		Digest:    digest.FromBytes(blobA),
		Size:      int64(len(blobA)),
	}
	malformedDesc := ocispec.Descriptor{
		MediaType: "application/octet-stream",
		Digest:    digest.Digest("sha256:BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
		Size:      int64(len(blobA)),
	}

	type args struct {
		target ocispec.Descriptor
	}
	type testCase struct {
		name    string
		args    args
		want    bool
		wantErr error
	}
	tests := []testCase{
		{"Basic", args{blobADesc}, true, nil},
		{"NonExistant", args{malformedDesc}, false, nil},
	}

	cacheRoot := t.TempDir()
	fc, err := NewFileCache(cacheRoot)
	if err != nil {
		t.Fatalf("initializing file cache, error = %v", err)
	}
	if err := fc.Push(ctx, blobADesc, bytes.NewReader(blobA)); err != nil {
		t.Fatalf("initializing cache with blobA, error = %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := fc.Exists(ctx, tt.args.target)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FileCache.Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if exists != tt.want {
				t.Errorf("FileCache.Exists() = %v, want %v", exists, tt.want)
			}
		})
	}

	t.Run("DuplicatePush", func(t *testing.T) {
		// duplicate push
		if err := fc.Push(ctx, blobADesc, bytes.NewReader(blobA)); err != nil {
			t.Fatalf("initializing cache with blobA, error = %v", err)
		}

		exists, err := fc.Exists(ctx, blobADesc)
		if err != nil {
			t.Errorf("FileCache.Exists() error = %v, wantErr %v", err, nil)
			return
		}
		if !exists {
			t.Errorf("FileCache.Exists() = %v, want %v", exists, true)
		}
	})

	t.Run("NotExistAfterPushFailure", func(t *testing.T) {
		// failed push
		if err := fc.Push(ctx, malformedDesc, bytes.NewReader(blobA)); err == nil {
			t.Fatalf("expected push failure succeeded")
		}

		exists, err := fc.Exists(ctx, malformedDesc)
		if err != nil {
			t.Errorf("FileCache.Exists() error = %v, wantErr %v", err, nil)
			return
		}
		if exists {
			t.Errorf("FileCache.Exists() = %v, want %v", exists, false)
		}
	})
}

func TestFileCache_Delete(t *testing.T) {
	ctx := logger.NewContext(context.Background(), tlog.Logger(t, 0))
	defer leaktest.Check(t)() //nolint

	blobA := []byte("alpha")
	blobADesc := ocispec.Descriptor{
		MediaType: "application/octet-stream",
		Digest:    digest.FromBytes(blobA),
		Size:      int64(len(blobA)),
	}

	type args struct {
		desc ocispec.Descriptor
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{"Basic", args{blobADesc}, nil},
		// NonExistant is an idempotentcy test
		{"NonExistant", args{blobADesc}, nil},
	}

	cacheRoot := t.TempDir()
	fc, err := NewFileCache(cacheRoot)
	if err != nil {
		t.Fatalf("initializing file cache, error = %v", err)
	}
	if err := fc.Push(ctx, blobADesc, bytes.NewReader(blobA)); err != nil {
		t.Fatalf("initializing cache with blobA, error = %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fc.Delete(ctx, tt.args.desc); !errors.Is(err, tt.wantErr) {
				t.Errorf("FileCache.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			_, err := os.Stat(fc.blobPath(tt.args.desc.Digest))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				t.Errorf("stat-ing deleted blob, error = %v", err)
			} else if err == nil {
				t.Errorf("blob still exists, digest = '%s'", tt.args.desc.Digest)
			}
		})
	}
}

func TestFileCache_Predecessors(t *testing.T) {
	ctx := logger.NewContext(context.Background(), tlog.Logger(t, 0))
	defer leaktest.Check(t)() //nolint

	blobA := []byte("alpha")
	blobADesc := ocispec.Descriptor{
		MediaType: "application/octet-stream",
		Digest:    digest.FromBytes(blobA),
		Size:      int64(len(blobA)),
	}

	manifestA := ocispec.Manifest{
		MediaType: ocispec.MediaTypeImageManifest,
		Config:    ocispec.DescriptorEmptyJSON,
		Subject:   &blobADesc,
	}

	manABytes, err := json.Marshal(manifestA)
	if err != nil {
		t.Fatalf("encoding manifest, error = %v", err)
	}

	cacheRoot := t.TempDir()
	fc, err := NewFileCache(cacheRoot, WithPredecessors())
	if err != nil {
		t.Fatalf("initializing file cache, error = %v", err)
	}

	manADesc, err := oras.PushBytes(ctx, fc, ocispec.MediaTypeImageManifest, manABytes)
	if err != nil {
		t.Fatalf("pushing manifest to file cache, error = %v", err)
	}

	type args struct {
		node ocispec.Descriptor
	}
	tests := []struct {
		name    string
		args    args
		want    []ocispec.Descriptor
		wantErr bool
	}{
		{"Basic", args{blobADesc}, []ocispec.Descriptor{manADesc}, false},
		{"None", args{manADesc}, []ocispec.Descriptor{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := fc.Predecessors(ctx, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileCache.Predecessors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(tt.want) == 0 && len(got) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileCache.Predecessors() = %v, want %v", got, tt.want)
			}
		})
	}
}
