package bottle

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gitlab.com/act3-ai/asce/data/tool/pkg/conf"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"

	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/orasutil"
	tbottle "gitlab.com/act3-ai/asce/data/tool/internal/transfer/bottle"
	reg "gitlab.com/act3-ai/asce/data/tool/pkg/registry"
	tbtl "gitlab.com/act3-ai/asce/data/tool/pkg/transfer/bottle"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	tlog "gitlab.com/act3-ai/asce/go-common/pkg/test"
)

// StoreInfo defines an oras.GraphTarget, a "reference" to it, and a function to access it.
type StoreInfo struct {
	Ref         string // our fake registry ref, ultimately used by the VirtualPartTracker
	Store       *orasutil.CheckedStorage
	NewTargetFn reg.NewGraphTargetFn
}

func Test_VirtualParts(t *testing.T) {
	ctx := context.Background()
	log := tlog.Logger(t, 0)
	ctx = logger.NewContext(ctx, log)

	config := conf.NewConfiguration("testagent")
	cfg := config.Get(ctx)

	// init bottle
	t.Log("Creating test bottle")
	btlDir := t.TempDir()
	partNames := []string{"alpha.txt", "beta.txt", "gamma.txt"}
	createBottle(t, btlDir, partNames)

	btl, err := LoadAndUpgradeBottle(ctx, cfg, btlDir)
	if err != nil {
		t.Fatalf("loading test bottle: error = %v", err)
	}

	// commit bottle
	if err := commit(ctx, cfg, btl, "", false); err != nil {
		t.Fatalf("committing bottle: error = %v", err)
	}

	// collect original descriptors for future validation
	origDescs := make([]ocispec.Descriptor, 0, len(partNames)+2)
	origDescs = append(origDescs, btl.Manifest.GetManifestDescriptor())
	origDescs = append(origDescs, btl.Manifest.GetConfigDescriptor())
	origDescs = append(origDescs, btl.Manifest.GetLayerDescriptors()...)
	t.Logf("Constructed origDescs: %v", origDescs)

	// init src and dest oras.GraphTargets
	srcInfo, destInfo := initSrcDest()

	t.Log("Pushing bottle to source registry")
	push(t, ctx, config, btlDir, srcInfo) // src starts with bottle

	// pull part of the bottle from src, i.e. setup virtual parts
	t.Log("Pulling bottle from source registry with part selection")
	pullDir := t.TempDir()
	partSelection := selectParts(partNames, 1) // select one part
	pull(t, ctx, config, pullDir, srcInfo, partSelection)

	// push bottle to destination
	t.Log("Pushing bottle to desintation registry")
	push(t, ctx, config, pullDir, destInfo)

	// verify that the entire bottle exists in the destination
	t.Log("Validating bottle at destination")

	// validate manifest
	for _, desc := range origDescs {
		exists, err := destInfo.Store.Target.Exists(ctx, desc)
		switch {
		case err != nil:
			t.Errorf("checking descriptor existence in destination repository: mediatype = '%s', digest = '%s', error = %v", desc.MediaType, desc.Digest, err)
		case !exists:
			t.Errorf("descriptor not found in desintation repository: mediatype = '%s', digest = '%s'", desc.MediaType, desc.Digest)
		default:
			// if the manifest exists, everythign else should as well but we continue to be safe
			t.Logf("Successfully verified bottle manifest existence in destination, mediatype = '%s', digest = '%s", desc.MediaType, desc.Digest)
		}
	}
	t.Log("Validation successful")
}

// push pushes a bottle to an oras.GraphTarget identified by destInfo.
func push(t *testing.T, ctx context.Context, config *conf.Configuration, btlDir string, destInfo StoreInfo) { //nolint
	t.Helper()
	cfg := config.Get(ctx)
	btl, err := LoadAndUpgradeBottle(ctx, cfg, btlDir)
	if err != nil {
		t.Fatalf("loading test bottle: error = %v", err)
	}

	// commit bottle
	if err := commit(ctx, cfg, btl, "", false); err != nil {
		t.Fatalf("committing bottle: error = %v", err)
	}

	// build transfer options
	opts := []tbtl.TransferOption{
		tbtl.WithNewGraphTargetFn(destInfo.NewTargetFn),
	}

	transferCfg := tbtl.NewTransferConfig(ctx, destInfo.Ref, btlDir, config, opts...)
	if err := tbottle.PushBottle(ctx, btl, *transferCfg, tbottle.WithSignatures()); err != nil {
		t.Fatalf("pushing bottle to source repository: error = %v", err)
	}
}

// pull pulls a bottle, with parts identified by partSelection, from an oras.GraphTarget identified by srcInfo.
func pull(t *testing.T, ctx context.Context, config *conf.Configuration, pullDir string, srcInfo StoreInfo, partSelection map[string]bool) { //nolint
	t.Helper()

	selection := make([]string, 0)
	for part, selected := range partSelection {
		if selected {
			selection = append(selection, part)
		}
	}

	// build transfer options
	opts := []tbtl.TransferOption{
		tbtl.WithNewGraphTargetFn(srcInfo.NewTargetFn),
		tbtl.WithPartSelection(nil, selection, nil),
	}

	transferCfg := tbtl.NewTransferConfig(ctx, srcInfo.Ref, pullDir, config, opts...)

	_, err := tbtl.Pull(ctx, *transferCfg)
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

// initSrcDest initializes an oras.GraphTarget for both a source and destination, backed by a memory.Store.
// Returns the store itself and associated metadata within a StoreInfo struct.
//
// NOTE: We may want to support env vars here such that we can test with
// specific zot, harbor, etc. registries.
func initSrcDest() (srcInfo StoreInfo, destInfo StoreInfo) {
	// define source information and handler func

	srcInfo.Ref = "sourceregistry/repository:tag"
	srcInfo.Store = &orasutil.CheckedStorage{
		Target: memory.New(),
	}
	srcInfo.NewTargetFn = func(ctx context.Context, ref string) (oras.GraphTarget, error) {
		return srcInfo.Store.Target, nil
	}

	// define dest info and handler func
	destInfo.Ref = "destinationregistry/repository:tag"
	destInfo.Store = &orasutil.CheckedStorage{
		Target: memory.New(),
	}
	destInfo.NewTargetFn = func(ctx context.Context, ref string) (oras.GraphTarget, error) {
		// typically we would connect to the registry here
		switch {
		case ref == srcInfo.Ref:
			// this case hits when we discover our virtual parts in another target
			return srcInfo.Store.Target, nil
		case ref == destInfo.Ref:
			// this case hits when we're simply connecting to the destination target
			return destInfo.Store.Target, nil
		default:
			return destInfo.Store.Target, nil
		}
	}
	return
}

// createBottle builds and saves a bottle, returning its path and part names.
func createBottle(t *testing.T, btlDir string, partNames []string) {
	t.Helper()

	createParts(t, btlDir, partNames)

	opts := []bottle.BOption{
		bottle.WithLocalPath(btlDir),
		bottle.DisableCache(true),
	}
	btl, err := bottle.NewBottle(opts...)
	if err != nil {
		t.Fatalf("creating test bottle: error = %v", err)
	}
	if err := btl.Save(); err != nil {
		t.Fatalf("saving test bottle: error = %v", err)
	}
}

// createParts builds parts for a bottle.
func createParts(t *testing.T, btlDir string, partNames []string) {
	t.Helper()
	for _, pn := range partNames {
		createPart(t, btlDir, pn)
	}

	if err := bottle.CreateBottle(btlDir, true); err != nil {
		t.Fatalf("creating bottle: error = %v", err)
	}
}

// createPart adds a file to the bottle directory.
func createPart(t *testing.T, btlDir, name string) {
	t.Helper()

	partPath := filepath.Join(btlDir, name)
	if err := os.WriteFile(partPath, []byte("testing part "+name), 0o666); err != nil {
		t.Fatalf("creating part %s: error = %v", name, err)
	}
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
