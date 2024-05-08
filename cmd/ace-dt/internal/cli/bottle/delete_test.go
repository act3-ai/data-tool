package bottle

import (
	"os"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"oras.land/oras-go/v2/errdef"
)

func Test_Functional_DeleteBottle(t *testing.T) {
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
	remoteBottle.AddArbitraryFileParts(15)
	remoteBottle.SetTempBottleRef(rootReg)
	helper.SendBottleToReg(remoteBottle.RootDir, remoteBottle.RegRef)
	remoteBottle.Load()
	helper.PruneCache()

	helper.CommandHelper.RunCommand("delete", remoteBottle.RegRef)

	err := helper.CheckRegForBottle(remoteBottle.RegRef, "")
	assert.ErrorIs(t, err, errdef.ErrNotFound)

}
