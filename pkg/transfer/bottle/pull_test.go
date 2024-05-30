package bottle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"

	"gitlab.com/act3-ai/asce/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/data/tool/pkg/conf"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

type DestHelper struct {
	dirs []string
}

func (h *DestHelper) SetupPullDir() (string, error) {
	pullPath, err := os.MkdirTemp("", "bottle-dest-*")
	if err != nil {
		return "", fmt.Errorf("failed to setup bottle destination: %w", err)
	}
	return pullPath, err
}

func (h *DestHelper) Cleanup() error {
	var errs []error
	for _, dir := range h.dirs {
		if err := os.RemoveAll(dir); err != nil {
			errs = append(errs, fmt.Errorf("removing bottle at path '%s': %w", dir, err))
		}
	}
	return errors.Join(errs...)
}

var (
	ociReference   string
	registryConfig v1alpha1.RegistryConfig
	pullDirHelper  DestHelper
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

// testMain allows us to use defer, returning the appropriate exit code.
func testMain(m *testing.M) int {
	log := slog.New(slog.Default().Handler())
	ctx := logger.NewContext(context.Background(), log)

	// setup registry
	log.InfoContext(ctx, "setting up source registry")
	s := httptest.NewServer(registry.New())
	defer func() {
		log.InfoContext(ctx, "closing source registry")
		s.Close()
	}()
	u, err := url.Parse(s.URL)
	if err != nil {
		panic(fmt.Sprintf("parsing registry url: %v", err))
	}

	// setup pull source
	ociReference = u.Host + "/repo/example:v1" // ref used for pull
	log.InfoContext(ctx, "setting up source bottle", "reference", ociReference)
	if err := setupExampleBottle(ctx, ociReference); err != nil {
		panic(fmt.Sprintf("Failed to setup source bottle: %v\n", err))
	}

	registryConfig = v1alpha1.RegistryConfig{
		Configs: map[string]v1alpha1.Registry{
			u.Host: {
				Endpoints: []string{"http://" + u.Host}, // enable plain-http
			},
		},
	}

	// setup pullDirHelper
	pullDirHelper = DestHelper{dirs: make([]string, 2)} // alloc test count
	defer func() {
		log.InfoContext(ctx, "cleaning up pulled bottles")
		if err := pullDirHelper.Cleanup(); err != nil {
			panic(fmt.Sprintf("cleaning up bottle pull directories: %v", err))
		}
	}()

	// run tests
	return m.Run()
}

func ExamplePull() {
	log := slog.New(slog.Default().Handler())
	ctx := logger.NewContext(context.Background(), log)

	// setup pull destination
	bottleDir, err := pullDirHelper.SetupPullDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to set up bottle pull directory: %v", err))
	}

	// define default configuration
	// overrides are only necessary if the desired configuration is not
	// available by default or loaded from a file with config.AddConfigFiles().
	config := conf.New()
	config.AddConfigOverride(conf.WithRegistryConfig(registryConfig)) // configure testing registry for plain-http

	src, desc, err := Resolve(ctx, ociReference, config, TransferOptions{})
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve OCI reference: %v", err))
	}

	// also fails on send telemetry event failure
	err = Pull(ctx, src, desc, bottleDir, PullOptions{})
	if err != nil {
		panic(fmt.Sprintf("Bottle pull failed: %v\n", err))
	}
	fmt.Fprintf(os.Stdout, "Success") //nolint

	// Output: Success
}

func ExamplePull_partselection() {
	log := slog.New(slog.Default().Handler())
	ctx := logger.NewContext(context.Background(), log)

	// setup pull destination
	pullDir, err := pullDirHelper.SetupPullDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to set up bottle pull directory: %v", err))
	}

	// define default configuration
	// overrides are only necessary if the desired configuration is not
	// available by default or loaded from a file with config.AddConfigFiles().
	config := conf.New()
	config.AddConfigOverride(conf.WithRegistryConfig(registryConfig)) // configure testing registry for plain-http

	src, desc, err := Resolve(ctx, ociReference, config, TransferOptions{})
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve OCI reference: %v", err))
	}

	// define part selection
	labelSelectors := []string{"foo=bar"}

	// define a cache, keeping track of non-selected (virtual) parts
	// while also allowing the bottle to be pushed later should we choose.
	cachePath, err := os.MkdirTemp("", "cache-*")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize cache path: %v", err))
	}

	pullOpts := PullOptions{
		TransferOptions: TransferOptions{
			CachePath: cachePath,
		},
		PartSelectorOptions: PartSelectorOptions{
			Labels: labelSelectors,
		},
	}

	// also fails on send telemetry event failure
	err = Pull(ctx, src, desc, pullDir, pullOpts)
	if err != nil {
		panic(fmt.Sprintf("Bottle pull failed: %v\n", err))
	}
	fmt.Fprintf(os.Stdout, "Success") //nolint

	// Output: Success
}

func setupExampleBottle(ctx context.Context, fullRef string) error {
	log := logger.FromContext(ctx)
	i := strings.LastIndex(fullRef, ":")

	ref := fullRef[:i]
	tag := fullRef[i+1:]

	r, err := remote.NewRepository(ref)
	if err != nil {
		return fmt.Errorf("connecting to source repository: %w", err)
	}
	r.PlainHTTP = true

	part1Desc, err := pushPart1(ctx, r)
	if err != nil {
		return err
	}
	log.InfoContext(ctx, "pushed part 1", "digest", part1Desc.Digest)

	part2Desc, err := pushPart2(ctx, r)
	if err != nil {
		return err
	}
	log.InfoContext(ctx, "pushed part 2", "digest", part2Desc.Digest)

	configDesc, err := pushConfig(ctx, r)
	if err != nil {
		return err
	}
	log.InfoContext(ctx, "pushed config", "digest", configDesc.Digest)

	manDesc, err := pushManifest(ctx, r, tag)
	if err != nil {
		return err
	}
	log.InfoContext(ctx, "pushed manifest", "digest", manDesc.Digest)

	return nil
}

func pushPart1(ctx context.Context, target oras.Target) (ocispec.Descriptor, error) {
	content := []byte("test part one\n")
	desc := ocispec.Descriptor{
		MediaType: mediatype.MediaTypeLayer,
		Digest:    digest.FromBytes(content),
		Size:      14,
	}

	if err := target.Push(ctx, desc, bytes.NewReader(content)); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("pushing bottle contents: %w", err)
	}
	return desc, nil
}

func pushPart2(ctx context.Context, target oras.Target) (ocispec.Descriptor, error) {
	content := []byte("test part two\n")
	desc := ocispec.Descriptor{
		MediaType: mediatype.MediaTypeLayer,
		Digest:    digest.FromBytes(content),
		Size:      14,
	}

	if err := target.Push(ctx, desc, bytes.NewReader(content)); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("pushing bottle contents: %w", err)
	}
	return desc, nil
}

func pushConfig(ctx context.Context, target oras.Target) (ocispec.Descriptor, error) {
	// If you change the config, update the matchBottleID in ExamplePull_validation
	config := []byte(`{"kind":"Bottle","apiVersion":"data.act3-ace.io/v1","parts":[{"name":"part1.txt","size":14,"digest":"sha256:0a587a815606ceadb036832f1989f5e868296b6fa98ef39564b447e951cad78c","labels":{"foo":"bar"}},{"name":"part2.txt","size":14,"digest":"sha256:5f2802faa177eff7526372ada8f37e52251f321b003979811aa8e9fff10427b8"}]}`)
	desc := ocispec.Descriptor{
		MediaType: mediatype.MediaTypeBottleConfig,
		Digest:    digest.FromBytes(config),
		Size:      313,
	}

	if err := target.Push(ctx, desc, bytes.NewReader(config)); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("pushing bottle manifest: %w", err)
	}
	return desc, nil
}

func pushManifest(ctx context.Context, target oras.Target, tag string) (ocispec.Descriptor, error) {
	manifest := []byte(`{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json","artifactType":"application/vnd.act3-ace.bottle","config":{"mediaType":"application/vnd.act3-ace.bottle.config.v1+json","digest":"sha256:f47fbb9257c6d0dd1bdce6517c969f2cbef2e0f2d053c02ea96506e7fd3fafda","size":313},"layers":[{"mediaType":"application/vnd.act3-ace.bottle.layer.v1","digest":"sha256:0a587a815606ceadb036832f1989f5e868296b6fa98ef39564b447e951cad78c","size":14},{"mediaType":"application/vnd.act3-ace.bottle.layer.v1","digest":"sha256:5f2802faa177eff7526372ada8f37e52251f321b003979811aa8e9fff10427b8","size":14}]}`)
	desc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifest),
		Size:      int64(len(manifest)),
	}

	if err := target.Push(ctx, desc, bytes.NewReader(manifest)); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("pushing bottle manifest: %w", err)
	}

	if err := target.Tag(ctx, desc, tag); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("tagging bottle manifest: %w", err)
	}
	return desc, nil
}
