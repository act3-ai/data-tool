package bottle

import (
	"os"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"oras.land/oras-go/v2/errdef"
)

func Test_Functional_ShowBottle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint:revive
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	remoteBottle := NewBottleHelper(t)
	remoteBottle.AddArbitraryFileParts(15)
	remoteBottle.SetTempBottleRef(rootReg)
	helper.SendBottleToReg(remoteBottle.RootDir, remoteBottle.RegRef)
	helper.PruneCache()

	// TODO we are missing the case where the bottle is local

	helper.CommandHelper.RunCommand("show", remoteBottle.RegRef)
}

func Test_Functional_ShowSelector(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	remoteBottle := NewBottleHelper(t)
	remoteBottle.AddArbitraryFileParts(5)
	labeledPart := "labeledPart.txt"
	remoteBottle.AddBottlePart(labeledPart)
	assert.NoError(t, helper.SaveBottle(remoteBottle.RootDir))
	remoteBottle.Load()
	remoteBottle.AddPartLabel(labeledPart, "test", "true")
	remoteBottle.SetTempBottleRef(rootReg)
	helper.SendBottleToReg(remoteBottle.RootDir, remoteBottle.RegRef)
	helper.PruneCache()

	helper.CommandHelper.RunCommand("show", remoteBottle.RegRef, "-l", "test=true")
}

func Test_Functional_ShowFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.SetTempBottleRef(rootReg)

	err := helper.CommandHelper.RunCommandWithError("show", helper.BottleHelper.RegRef)
	assert.ErrorIs(t, err, errdef.ErrNotFound)
}
