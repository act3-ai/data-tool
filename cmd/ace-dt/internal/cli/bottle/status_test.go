package bottle

import (
	"testing"

	"github.com/fortytw2/leaktest"
)

func Test_Functional_StatusBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.AddArbitraryFileParts(10)

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.CommandHelper.RunCommand("commit")
	helper.CommandHelper.RunCommand("status")
}
