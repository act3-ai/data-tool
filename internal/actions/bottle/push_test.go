package bottle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"git.act3-ace.com/ace/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/data/tool/pkg/conf"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	orasreg "oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	tlog "git.act3-ace.com/ace/go-common/pkg/test"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/orasutil"
	tbtl "gitlab.com/act3-ai/asce/data/tool/internal/transfer/bottle"
	tbottle "gitlab.com/act3-ai/asce/data/tool/pkg/transfer/bottle"
)

var (
	srcInfo          *SrcStoreInfo        // source registry information
	destInfo         *DestStoreInfo       // destination registry information
	config           *conf.Configuration  // handle plain-http registries
	blobInfoCacheDir string               // virtual part tracking
	origDescs        []ocispec.Descriptor // ensure entire bottle makes it to the destination, MUST have order: [manifest, config, parts...]
)

// SrcStoreInfo defines an oras.GraphTarget, a "reference" to it, and a function to access it.
type SrcStoreInfo struct {
	Ref   string // our fake registry ref, ultimately used by the VirtualPartTracker
	Store *orasutil.CheckedStorage
}

// GraphTarget implements registry.GraphTargeter
func (s *SrcStoreInfo) GraphTarget(ctx context.Context, ref string) (oras.GraphTarget, error) {
	return s.Store.Target, nil
}

// ParseEndpointReference implements registry.EndpointParser.
func (s *SrcStoreInfo) ParseEndpointReference(reference string) (orasreg.Reference, error) {
	return orasreg.ParseReference(reference)
}

// ReadOnlyGraphTarget implements registry.GraphTargeter
func (s *SrcStoreInfo) ReadOnlyGraphTarget(ctx context.Context, ref string) (oras.ReadOnlyGraphTarget, error) {
	return s.Store.Target, nil
}

// DestStoreInfo defines an oras.GraphTarget, a "reference" to it, and a function to access it.
type DestStoreInfo struct {
	Ref           string // our fake registry ref, ultimately used by the VirtualPartTracker
	Store         *orasutil.CheckedStorage
	VirtualTarget oras.GraphTarget // for vitual part handling
}

// GraphTarget implements registry.GraphTargeter
func (d *DestStoreInfo) GraphTarget(ctx context.Context, ref string) (oras.GraphTarget, error) {
	switch {
	case ref == srcInfo.Ref:
		// this case hits when we discover our virtual parts in another target
		return d.VirtualTarget, nil
	case ref == destInfo.Ref:
		// this case hits when we're simply connecting to the destination target
		return d.Store.Target, nil
	default:
		return d.Store.Target, nil
	}
}

// ParseEndpointReference implements registry.EndpointParser.
func (d *DestStoreInfo) ParseEndpointReference(reference string) (orasreg.Reference, error) {
	return orasreg.ParseReference(reference)
}

// ReadOnlyGraphTarget implements registry.GraphTargeter
func (d *DestStoreInfo) ReadOnlyGraphTarget(ctx context.Context, ref string) (oras.ReadOnlyGraphTarget, error) {
	switch {
	case ref == srcInfo.Ref:
		// this case hits when we discover our virtual parts in another target
		return d.VirtualTarget, nil
	case ref == destInfo.Ref:
		// this case hits when we're simply connecting to the destination target
		return d.Store.Target, nil
	default:
		return d.Store.Target, nil
	}
}

// NOTE: We may want to support env vars here such that we can test with
// specific zot, harbor, etc. registries.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	log := slog.New(slog.Default().Handler())
	ctx := logger.NewContext(context.Background(), log)

	config = conf.New()
	config.AddConfigOverride(
		conf.WithConcurrency(1),
		conf.WithCachePath(blobInfoCacheDir),
	)

	var err error
	srcReg := httptest.NewServer(registry.New())
	defer srcReg.Close()
	destReg := httptest.NewServer(registry.New())
	defer destReg.Close()

	log.InfoContext(ctx, "initializing source and destination registries", "source", srcReg.URL, "dest", destReg.URL)
	srcInfo, destInfo, err = initRegistries(ctx, srcReg.URL, destReg.URL)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize source registry: %v", err))
	}

	// init blobInfoCacheDir
	blobInfoCacheDir, err = os.MkdirTemp("", "blobinfocache-*")
	if err != nil {
		panic(fmt.Sprintf("initializing blobinfocache directory: %v", err))
	}
	defer func() {
		err := os.RemoveAll(blobInfoCacheDir)
		if err != nil {
			panic(fmt.Sprintf("cleaning up blob info cache: %v", err))
		}
	}()
	log.InfoContext(ctx, "initialized cache directory", "cacheDir", blobInfoCacheDir)

	log.InfoContext(ctx, "setting up source bottle", "reference", srcInfo.Ref)
	origDescs, err = setupExampleBottle(ctx, srcInfo.Ref)
	if err != nil {
		panic(fmt.Sprintf("Failed to setup source bottle: %v\n", err))
	}

	// run tests
	return m.Run()
}

func Test_VirtualParts(t *testing.T) {
	ctx := logger.NewContext(context.Background(), tlog.Logger(t, -6))

	partNames := []string{"part1.txt", "part2.txt"}

	// pull part of the bottle from src, i.e. setup virtual parts
	t.Log("Pulling bottle from source registry with part selection")
	pullDir := t.TempDir()
	partSelection := selectParts(partNames, 1) // select one part
	pull(t, ctx, pullDir, srcInfo, partSelection)

	// push bottle to destination
	t.Log("Pushing bottle to desintation registry")
	push(t, ctx, pullDir, destInfo)

	// verify that the entire bottle exists in the destination
	t.Log("Validating bottle at destination")

	// validate manifest
	validateDescs := origDescs
	for _, desc := range validateDescs {
		exists, err := destInfo.Store.Target.Exists(ctx, desc)
		switch {
		case err != nil:
			t.Errorf("checking descriptor existence in destination repository: mediatype = '%s', digest = '%s', error = %v", desc.MediaType, desc.Digest, err)
		case !exists:
			t.Errorf("descriptor not found in destination repository: mediatype = '%s', digest = '%s'", desc.MediaType, desc.Digest)
		default:
			// if the manifest exists, everythign else should as well but we continue to be safe
			t.Logf("Successfully verified descriptor in destination, mediatype = '%s', digest = '%s", desc.MediaType, desc.Digest)
		}
	}
	t.Log("Validation successful")
}

// push pushes a bottle to an oras.GraphTarget identified by destInfo.
func push(t *testing.T, ctx context.Context, btlDir string, destInfo *DestStoreInfo) { //nolint
	t.Helper()
	cfg := config.Get(ctx)
	btl, err := bottle.LoadBottle(btlDir,
		bottle.WithCachePath(blobInfoCacheDir),
		bottle.WithBlobInfoCache(blobInfoCacheDir),
	)
	if err != nil {
		t.Fatalf("loading test bottle: error = %v", err)
	}
	if err := btl.LoadLocalLabels(); err != nil {
		t.Fatalf("loading test bottle labes: error = %v", err)
	}

	// commit bottle
	if err := commit(ctx, cfg, btl, false); err != nil {
		t.Fatalf("committing bottle: error = %v", err)
	}

	pushOpts := tbtl.PushOptions{
		TransferOptions: tbottle.TransferOptions{
			CachePath: blobInfoCacheDir,
		},
	}

	if err := tbtl.PushBottle(ctx, btl, destInfo, destInfo.Ref, pushOpts); err != nil {
		t.Fatalf("pushing bottle to source repository: error = %v", err)
	}
}

// pull pulls a bottle, with parts identified by partSelection, from an oras.GraphTarget identified by srcInfo.
func pull(t *testing.T, ctx context.Context, pullDir string, srcInfo *SrcStoreInfo, //nolint
	partSelection map[string]bool,
) {
	t.Helper()

	selection := make([]string, 0)
	for part, selected := range partSelection {
		if selected {
			selection = append(selection, part)
		}
	}

	transferOpts := tbottle.TransferOptions{
		Concurrency: 1,
		CachePath:   blobInfoCacheDir,
	}

	src, desc, err := tbottle.Resolve(ctx, srcInfo.Ref, config, transferOpts)
	if err != nil {
		t.Fatalf("resolving source reference: %v", err)
	}

	pullOpts := tbottle.PullOptions{
		TransferOptions: transferOpts,
		PartSelectorOptions: tbottle.PartSelectorOptions{
			Names: selection,
		},
	}

	err = tbottle.Pull(ctx, src, desc, pullDir, pullOpts)
	if err != nil {
		t.Fatalf("pulling bottle from source: error = %v", err)
	}

	// ensure appropriate parts were selected
	for part := range partSelection {
		t.Logf("PartName: %s, Selected: %t", part, partSelection[part])
		partPath := filepath.Join(pullDir, part)

		_, err := os.Stat(partPath)
		switch {
		case err == nil && !partSelection[part]:
			t.Errorf("unselected part found in pulled bottle")
		case errors.Is(err, fs.ErrNotExist) && partSelection[part]:
			t.Errorf("selected part not found in pulled bottle")
		case err != nil && !errors.Is(err, fs.ErrNotExist):
			t.Errorf("unwanted error when validating pulled bottle: error = %v", err)
		default:
			// expected outcome
		}
	}
}

func setupExampleBottle(ctx context.Context, fullRef string) ([]ocispec.Descriptor, error) {
	log := logger.FromContext(ctx)
	i := strings.LastIndex(fullRef, ":")

	ref := fullRef[:i]
	tag := fullRef[i+1:]

	r, err := remote.NewRepository(ref)
	if err != nil {
		return nil, fmt.Errorf("connecting to source repository: %w", err)
	}
	r.PlainHTTP = true

	part1Desc, err := pushPart1(ctx, r)
	if err != nil {
		return nil, err
	}
	log.InfoContext(ctx, "pushed part 1", "digest", part1Desc.Digest)

	part2Desc, err := pushPart2(ctx, r)
	if err != nil {
		return nil, err
	}
	log.InfoContext(ctx, "pushed part 2", "digest", part2Desc.Digest)

	configDesc, err := pushConfig(ctx, r)
	if err != nil {
		return nil, err
	}
	log.InfoContext(ctx, "pushed config", "digest", configDesc.Digest)

	manDesc, err := pushManifest(ctx, r, tag)
	if err != nil {
		return nil, err
	}
	log.InfoContext(ctx, "pushed manifest", "digest", manDesc.Digest)

	descs := make([]ocispec.Descriptor, 0, 4) // alloc: 2 parts, 1 config, 1 manifest
	descs = append(descs, manDesc, configDesc, part1Desc, part2Desc)
	return descs, nil
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

// selectParts selects n parts from partNames, deterministically.
func selectParts(partNames []string, n int) map[string]bool {
	slices.Sort(partNames)
	selectedParts := make(map[string]bool, n)
	for i, part := range partNames {
		switch {
		case i <= n-1:
			selectedParts[part] = true
		default:
			selectedParts[part] = false
		}
	}
	return selectedParts
}

// initRegistries initializes a source and destinationregistry, adding plain-http endpoints to the provided configuration.
func initRegistries(ctx context.Context, srcURL, destURL string) (*SrcStoreInfo, *DestStoreInfo, error) {
	// build the oci reference
	srcurl, err := url.Parse(srcURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing source registry url: %w", err)
	}
	srcRef := srcurl.Host + "/sourcerepo/name:v1"

	// build the oci reference
	desturl, err := url.Parse(destURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing destination registry url: %w", err)
	}
	destRef := desturl.Host + "/destinationrepo/name:v1"

	// add servers to config for plain-http
	rcfg := v1alpha1.RegistryConfig{
		Configs: map[string]v1alpha1.Registry{
			srcurl.Host: {
				Endpoints: []string{"http://" + srcurl.Host},
			},
			desturl.Host: {
				Endpoints: []string{"http://" + desturl.Host},
			},
		},
	}
	config.AddConfigOverride(conf.WithRegistryConfig(rcfg))

	// define source information
	s := &SrcStoreInfo{}
	s.Ref = srcRef
	srcRepo, err := config.GraphTarget(ctx, s.Ref)
	if err != nil {
		return nil, nil, fmt.Errorf("initializing source target: %w", err)
	}
	s.Store = &orasutil.CheckedStorage{
		Target: srcRepo,
	}

	// define dest info
	d := &DestStoreInfo{}
	d.Ref = destRef
	destRepo, err := config.GraphTarget(ctx, d.Ref)
	if err != nil {
		return nil, nil, fmt.Errorf("initializing destination target: %w", err)
	}
	d.Store = &orasutil.CheckedStorage{
		Target: destRepo,
	}
	d.VirtualTarget = s.Store.Target

	return s, d, nil
}
