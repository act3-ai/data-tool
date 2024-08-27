package orasutil

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/memory"

	"git.act3-ace.com/ace/data/tool/internal/cache"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	tlog "git.act3-ace.com/ace/go-common/pkg/test"
)

func Test_push(t *testing.T) {
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
		expected ocispec.Descriptor
		content  io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Basic", args{blobADesc, bytes.NewReader(blobA)}, false},
		{"InvalidDigest", args{malformedDesc, bytes.NewReader(blobA)}, true},
		{"Duplicate", args{blobADesc, bytes.NewReader(blobA)}, false},
	}

	cacheRoot := t.TempDir()
	fc, err := cache.NewFileCache(cacheRoot)
	if err != nil {
		t.Fatalf("initializing file cache, error = %v", err)
	}

	target := &CheckedStorage{Target: memory.New()}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := push(ctx, fc, tt.args.expected, tt.args.content, target.Push)
			switch {
			case (err != nil) != tt.wantErr:
				// unexpected behavior
				t.Errorf("push() error = %v, wantErr %v", err, tt.wantErr)
				return
			case err != nil && tt.wantErr:
				// expected error
				// noop
			default:
				exists, err := fc.Exists(ctx, tt.args.expected)
				if err != nil {
					t.Errorf("failed to check blob existence in cache, error = %v", err)
					return
				}
				if !exists {
					t.Errorf("blob does not exist in cache, error = %v", err)
				}
			}
		})
	}
}

func Test_fetch(t *testing.T) {
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
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"Basic", args{blobADesc}, false},
		{"InvalidDigest", args{malformedDesc}, true},
	}

	cacheRoot := t.TempDir()
	fc, err := cache.NewFileCache(cacheRoot)
	if err != nil {
		t.Fatalf("initializing file cache, error = %v", err)
	}
	if err := fc.Push(ctx, blobADesc, bytes.NewReader(blobA)); err != nil {
		t.Fatalf("initializing cache with blobA, error = %v", err)
	}

	source := memory.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc, err := fetch(ctx, fc, tt.args.target, source.Fetch)
			switch {
			case (err != nil) != tt.wantErr:
				// unexpected behavior
				t.Errorf("fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			case err != nil && tt.wantErr:
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
