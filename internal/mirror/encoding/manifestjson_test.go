package encoding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	tlog "gitlab.com/act3-ai/asce/go-common/pkg/test"

	"git.act3-ace.com/ace/data/tool/internal/ref"
)

func TestBuildManifestJSON(t *testing.T) {
	ctx := context.Background()
	ctx = logger.NewContext(ctx, tlog.Logger(t, -2))
	cas := memory.New()

	// blobs
	blobAlpha := []byte("blob alpha")
	blobAlphaDesc := ocispec.Descriptor{
		MediaType: "application/octet-stream",
		Digest:    digest.FromBytes(blobAlpha),
		Size:      int64(len(blobAlpha)),
	}
	if err := cas.Push(ctx, blobAlphaDesc, bytes.NewReader(blobAlpha)); err != nil {
		panic(fmt.Errorf("pushing blob alpha to cas: %w", err))
	}
	blobBeta := []byte("blob beta")
	blobBetaDesc := ocispec.Descriptor{
		MediaType: "application/octet-stream",
		Digest:    digest.FromBytes(blobBeta),
		Size:      int64(len(blobBeta)),
	}
	if err := cas.Push(ctx, blobBetaDesc, bytes.NewReader(blobBeta)); err != nil {
		panic(fmt.Errorf("pushing blob beta to cas: %w", err))
	}

	// platforms
	linuxamd64 := &ocispec.Platform{
		Architecture: "amd64",
		OS:           "linux",
	}
	windowsamd64 := &ocispec.Platform{
		Architecture: "amd64",
		OS:           "windows",
	}

	// "Image" Alpha, linux/amd64
	cfgAlpha := ocispec.ImageConfig{
		User: "userAlpha", // for unique digest
	}
	cfgAlphaRaw, err := json.Marshal(cfgAlpha)
	if err != nil {
		panic(fmt.Errorf("encoding manifest config alpha: %w", err))
	}
	cfgAlphaDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    digest.FromBytes(cfgAlphaRaw),
		Size:      int64(len(cfgAlphaRaw)),
		Platform:  linuxamd64,
	}
	manAlphaRef := "oras.memory.io/repo/alpha:v1.0.0-build20240903"
	manAlphaOpts := oras.PackManifestOptions{
		Layers:              []ocispec.Descriptor{blobAlphaDesc}, // update test cases if more layers are added
		ManifestAnnotations: map[string]string{ref.AnnotationSrcRef: manAlphaRef},
		ConfigDescriptor:    &cfgAlphaDesc,
	}
	manAlphaDesc, err := oras.PackManifest(ctx, cas, oras.PackManifestVersion1_1, "", manAlphaOpts)
	if err != nil {
		panic(fmt.Errorf("pushing manifest alpha: %w", err))
	}
	manAlphaDesc.Platform = linuxamd64 // for index

	// "Image" Beta, windows/amd64
	cfgBeta := ocispec.ImageConfig{
		User: "userBeta", // for unique digest
	}
	cfgBetaRaw, err := json.Marshal(cfgBeta)
	if err != nil {
		panic(fmt.Errorf("encoding manifest config beta: %w", err))
	}
	cfgBetaDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    digest.FromBytes(cfgBetaRaw),
		Size:      int64(len(cfgBetaRaw)),
		Platform:  windowsamd64,
	}
	manBetaRef := "oras.memory.io/repo/beta:v1.0.0-build20240903"
	manBetaOpts := oras.PackManifestOptions{
		Layers:              []ocispec.Descriptor{blobBetaDesc}, // update test cases if more layers are added
		ManifestAnnotations: map[string]string{ref.AnnotationSrcRef: manBetaRef},
		ConfigDescriptor:    &cfgBetaDesc,
	}
	manBetaDesc, err := oras.PackManifest(ctx, cas, oras.PackManifestVersion1_1, "", manBetaOpts)
	if err != nil {
		panic(fmt.Errorf("pushing manifest beta: %w", err))
	}
	manBetaDesc.Platform = windowsamd64 // for index

	// Index Alpha
	idxAlpha := ocispec.Index{
		MediaType:   ocispec.MediaTypeImageIndex,
		Manifests:   []ocispec.Descriptor{manAlphaDesc},
		Annotations: map[string]string{ref.AnnotationSrcRef: manAlphaRef},
	}
	idxAlphaRaw, err := json.Marshal(idxAlpha)
	if err != nil {
		panic(fmt.Errorf("encoding index manifest alpha: %w", err))
	}
	idxAlphaDesc := ocispec.Descriptor{
		MediaType:   ocispec.MediaTypeImageIndex,
		Digest:      digest.FromBytes(idxAlphaRaw),
		Size:        int64(len(idxAlphaRaw)),
		Annotations: idxAlpha.Annotations,
	}
	if err := cas.Push(ctx, idxAlphaDesc, bytes.NewReader(idxAlphaRaw)); err != nil {
		panic(fmt.Errorf("pushing index manifest alpha: %w", err))
	}

	// Index Alpha-Beta, contains both alpha and beta image manifests
	idxAlphaBetaRef := "oras.memory.io/repo/alphabeta:v1.0.0-build20240903"
	idxAlphaBeta := ocispec.Index{
		MediaType:   ocispec.MediaTypeImageIndex,
		Manifests:   []ocispec.Descriptor{manAlphaDesc, manBetaDesc},
		Annotations: map[string]string{ref.AnnotationSrcRef: idxAlphaBetaRef},
	}
	idxAlphaBetaRaw, err := json.Marshal(idxAlphaBeta)
	if err != nil {
		panic(fmt.Errorf("encoding index manifest alpha-beta: %w", err))
	}
	idxAlphaBetaDesc := ocispec.Descriptor{
		MediaType:   ocispec.MediaTypeImageIndex,
		Digest:      digest.FromBytes(idxAlphaBetaRaw),
		Size:        int64(len(idxAlphaBetaRaw)),
		Annotations: idxAlphaBeta.Annotations,
	}
	if err := cas.Push(ctx, idxAlphaBetaDesc, bytes.NewReader(idxAlphaBetaRaw)); err != nil {
		panic(fmt.Errorf("pushing index manifest alpha-beta: %w", err))
	}

	// Index Gamma, the same as Index Alpha but has a different reference annotation (same digests)
	idxGammaRef := "oras.memory.io/repo/gamma:v1.0.0-build20240903"
	idxGamma := ocispec.Index{
		MediaType:   ocispec.MediaTypeImageIndex,
		Manifests:   []ocispec.Descriptor{manAlphaDesc},
		Annotations: map[string]string{ref.AnnotationSrcRef: idxGammaRef},
	}
	idxGammaRaw, err := json.Marshal(idxGamma)
	if err != nil {
		panic(fmt.Errorf("encoding index manifest gamma: %w", err))
	}
	idxGammaDesc := ocispec.Descriptor{
		MediaType:   ocispec.MediaTypeImageIndex,
		Digest:      digest.FromBytes(idxGammaRaw),
		Size:        int64(len(idxGammaRaw)),
		Annotations: idxGamma.Annotations,
	}
	if err := cas.Push(ctx, idxGammaDesc, bytes.NewReader(idxGammaRaw)); err != nil {
		panic(fmt.Errorf("pushing index manifest gamma: %w", err))
	}

	type args struct {
		manifests []ocispec.Descriptor
	}
	tests := []struct {
		name    string
		args    args
		want    ManifestJSON
		wantErr bool
	}{
		// manifest
		{"ImageManifest",
			args{manifests: []ocispec.Descriptor{manAlphaDesc}},
			ManifestJSON{
				Manifests: []ManifestInfo{
					{dgst: manAlphaDesc.Digest,
						Config:   addPrefix(cfgAlphaDesc.Digest),
						RepoTags: []string{manAlphaRef},
						Layers:   []string{addPrefix(blobAlphaDesc.Digest)},
					},
				},
			},
			false,
		},
		{"IndexManifestOneImage",
			args{manifests: []ocispec.Descriptor{idxAlphaDesc}},
			ManifestJSON{
				Manifests: []ManifestInfo{
					{dgst: manAlphaDesc.Digest,
						Config:   addPrefix(cfgAlphaDesc.Digest),
						RepoTags: []string{manAlphaRef},
						Layers:   []string{addPrefix(blobAlphaDesc.Digest)},
					},
				},
			},
			false,
		},
		{"IndexManifestTwoImages",
			args{manifests: []ocispec.Descriptor{idxAlphaBetaDesc}},
			ManifestJSON{
				Manifests: []ManifestInfo{
					{dgst: manAlphaDesc.Digest,
						Config:   addPrefix(cfgAlphaDesc.Digest),
						RepoTags: []string{idxAlphaBetaRef}, // ref inherited from idx annotations
						Layers:   []string{addPrefix(blobAlphaDesc.Digest)},
					},
				},
			},
			false,
		},
		{"IndexManifestsDuplicateRefs",
			args{manifests: []ocispec.Descriptor{idxGammaDesc}},
			ManifestJSON{
				Manifests: []ManifestInfo{
					{dgst: manAlphaDesc.Digest,
						Config:   addPrefix(cfgAlphaDesc.Digest),
						RepoTags: []string{idxGammaRef},
						Layers:   []string{addPrefix(blobAlphaDesc.Digest)},
					},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildManifestJSON(ctx, cas, tt.args.manifests)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildManifestJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.Manifests) != len(tt.want.Manifests) {
				t.Errorf("count of ManifestJSON entries does not match, got = %d, want = %d", len(got.Manifests), len(tt.want.Manifests))
			}
			// only check what is marshalled
			for i := range tt.want.Manifests {
				if got.Manifests[i].Config != tt.want.Manifests[i].Config {
					t.Errorf("ManifestJSON Config = %v, want %v", got.Manifests[i].Config, tt.want.Manifests[i].Config)
				}

				if !reflect.DeepEqual(got.Manifests[i].RepoTags, tt.want.Manifests[i].RepoTags) {
					t.Errorf("ManifestJSON RepoTags = %v, want %v", got.Manifests[i].RepoTags, tt.want.Manifests[i].RepoTags)
				}

				if !reflect.DeepEqual(got.Manifests[i].Layers, tt.want.Manifests[i].Layers) {
					t.Errorf("ManifestJSON Layers = %v, want %v", got.Manifests[i].Layers, tt.want.Manifests[i].Layers)
				}
			}
		})
	}
}
