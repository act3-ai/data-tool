package bottle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/functesting"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/pkg/conf"
	"gitlab.com/act3-ai/asce/go-common/pkg/fsutil"
)

// TestHelper is a struct that contains helpers to make it easier to design functional tests.
type TestHelper struct {
	*functesting.CommandHelper
	*BottleHelper

	t testing.TB
}

// NewTestHelper returns a new functional tests with,
// the given root cobra command used to run all commands with CommandHelper
// the root bottle directory used to run all BottleHelper functions
// The registry used to make bottle references when pushing / pulling to a registry.
func NewTestHelper(tb testing.TB, cmd *cobra.Command) *TestHelper {
	tb.Helper()
	return &TestHelper{
		CommandHelper: functesting.NewCommandHelper(tb, cmd),
		BottleHelper:  NewBottleHelper(tb),
		t:             tb,
	}
}

func (h *TestHelper) getContext() context.Context {
	return h.Context()
}

// CheckRegForBottle returns an error if no bottle is found on the registry,
// Returns nil if a bottle is found.
func (h *TestHelper) CheckRegForBottle(regRef string, expectedDigest digest.Digest) error {
	ctx := h.getContext()
	rcfg := h.GetConfig().RegistryConfig

	c := conf.New()
	c.AddConfigOverride(conf.WithRegistryConfig(rcfg))

	gt, err := c.Repository(ctx, regRef)
	if err != nil {
		return fmt.Errorf("creating repository %q: %w", regRef, err)
	}

	desc, err := gt.Resolve(ctx, regRef)
	if err != nil {
		return fmt.Errorf("resolving bottle reference: %w", err)
	}

	if expectedDigest != "" && desc.Digest != expectedDigest {
		return fmt.Errorf("resolved digest does not match expected, want '%s', got '%s'", expectedDigest, desc.Digest)
	}
	return nil
}

// RemoveBottleFromReg deletes the given bottle reference from the registry.
func (h *TestHelper) RemoveBottleFromReg(regRef string) error {
	return h.RunCommandWithError("delete", regRef)
}

// SaveBottle will either init or commit a bottle. Saves any updates to local filesystem.
func (h *TestHelper) SaveBottle(btlDir string) error {
	err := h.RunCommandWithError("--bottle-dir", btlDir, "init")
	if errors.Is(err, bottle.ErrBottleAlreadyInit) {
		// ErrFilesystem is returned if bottle is already init, so we need to try to commit
		if err := h.RunCommandWithError("--bottle-dir", btlDir, "init"); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// SendBottleToReg is a shortcut to sending a bottle at the given bottle directory to the given registry reference
// NOTE: If you do not TestHelper.PruneCache() after this func, the parts will be in local cache.
func (h *TestHelper) SendBottleToReg(btlDir string, registry string) {
	h.t.Helper()
	assert.NoError(h.t, h.sendBottleToReg(btlDir, registry))
}

func (h *TestHelper) sendBottleToReg(btlDir string, registry string) error {
	if err := bottle.CheckIfCanInitialize(btlDir, false); err == nil {
		// bottle needs init
		err := h.RunCommandWithError("--bottle-dir", btlDir, "init")
		if err != nil {
			return err
		}
	} else if !errors.Is(err, bottle.ErrBottleAlreadyInit) {
		return err
	}
	// bottle push
	return h.RunCommandWithError("push", registry, "--bottle-dir", btlDir)
}

// VerifyBottleIDFile compares bottle id in f.Bottle to given bottleID file
// Expects bottle to be loaded with BottleHelper.load().
func (h *TestHelper) VerifyBottleIDFile(path string) {
	h.t.Helper()
	assert.NoError(h.t, h.verifyBottleIDFile(path))
}

func (h *TestHelper) verifyBottleIDFile(path string) error {
	btlDigest := h.Bottle.GetBottleID()
	d, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading bottle id file: %w", err)
	}
	dgst, err := digest.Parse(string(d))
	if err != nil {
		return err
	}
	// TODO we should use this instead
	// return h.Bottle.VerifyBottleID(dgst)
	if btlDigest != dgst {
		return fmt.Errorf("bottle digest does not match bottle id file digest. %v != %v", btlDigest, dgst)
	}
	return nil
}

// GetNumLocalParts looks at the given bottleDir and returns how many parts there are in that bottle dir.
func (h *TestHelper) GetNumLocalParts(bottleDir string) (int, error) {
	numParts := 0
	dirEntries, err := os.ReadDir(bottleDir)
	if err != nil {
		return numParts, fmt.Errorf("error reading bottle dir: %w", err)
	}
	for _, dirEntry := range dirEntries {
		// if not hidden dir, then bottle part
		if !strings.HasPrefix(dirEntry.Name(), ".") {
			numParts++
		}
	}
	return numParts, nil
}

// RequirePartNum is a shortcut to compare expected partNum the amount of parts in a bottle, post bottle load.
func (h *TestHelper) RequirePartNum(partNum int) {
	h.t.Helper()
	h.Load()
	require.Equal(h.t, partNum, h.Bottle.NumParts())
}

// EqualBottles compares bottle directories for equality
// checks bottle ID and the file and directory names.
func (h *TestHelper) EqualBottles(bottleDir1 string, bottleDir2 string) {
	h.t.Helper()
	a := assert.New(h.t)

	btl1, err := bottle.LoadBottle(bottleDir1, bottle.WithLocalLabels())
	a.NoError(err)
	btl2, err := bottle.LoadBottle(bottleDir2, bottle.WithLocalLabels())
	a.NoError(err)

	a.Equal(btl1.GetBottleID(), btl2.GetBottleID())
	a.NoError(fsutil.EqualFilesystem(os.DirFS(bottleDir1), os.DirFS(bottleDir2), fsutil.DefaultComparisonOpts))
}
