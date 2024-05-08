package bottle

import (
	"path/filepath"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
)

func Test_Functional_PostInitFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.AddArbitraryFileParts(4)

	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	// Assert number of parts post command is equal to the number of parts created
	helper.BottleHelper.Load()
	helper.RequirePartNum(4)

	helper.BottleHelper.AddArbitraryFileParts(4)
	helper.CommandHelper.RunCommand("commit")

	helper.BottleHelper.Load()
	// Assert number of parts post command is equal to the number of parts created
	helper.RequirePartNum(8)
}

func Test_Functional_PostInitDirs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	helper.BottleHelper.AddArbitraryDirParts(5)
	helper.CommandHelper.RunCommand("commit")
	// Assert number of parts post command is equal to the number of parts created
	helper.RequirePartNum(5)
}

func Test_Functional_PostInitDeleteFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.AddBottlePart("test1.txt")
	helper.BottleHelper.AddBottlePart("test2.txt")

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)

	helper.CommandHelper.RunCommand("init")
	// Assert number of parts post command is equal to the number of parts created
	helper.RequirePartNum(2)

	helper.BottleHelper.RemoveBottlePart("test2.txt")

	helper.CommandHelper.RunCommand("commit")
	// Assert number of parts post command is equal to the number of parts created
	helper.RequirePartNum(1)
}

func Test_Functional_PostCommitDeleteFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.AddBottlePart("test1.txt")
	helper.BottleHelper.AddBottlePart("test2.txt")

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	helper.CommandHelper.RunCommand("commit")
	helper.RequirePartNum(2)

	helper.BottleHelper.RemoveBottlePart("test2.txt")

	helper.CommandHelper.RunCommand("commit")
	helper.RequirePartNum(1)
}

func Test_Functional_ManyCommitNewFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	for i := 0; i < 10; i++ {
		helper.BottleHelper.AddArbitraryFileParts(1)
		helper.CommandHelper.RunCommand("commit")
	}
	helper.RequirePartNum(14)
}

func Test_Functional_ManyCommitNewDirs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	for i := 0; i < 10; i++ {
		helper.BottleHelper.AddArbitraryDirParts(1)
		helper.CommandHelper.RunCommand("commit")
	}
}

func Test_Functional_CommitWriteBottleID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	bottleIDFile := filepath.Join(t.TempDir(), "bottleID")
	helper.CommandHelper.RunCommand("commit", "--write-bottle-id", bottleIDFile)
	helper.RequirePartNum(4)

	helper.BottleHelper.Load()
	helper.VerifyBottleIDFile(bottleIDFile)
}

func Test_Functional_UpdatedBottleID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.CommandHelper.RunCommand("commit")

	helper.BottleHelper.Load()
	firstBottleID := helper.BottleHelper.Bottle.GetBottleID()

	helper.BottleHelper.AddArbitraryFileParts(4)
	helper.CommandHelper.RunCommand("commit")
	helper.RequirePartNum(8)

	helper.BottleHelper.Load()
	secondBottleID := helper.BottleHelper.Bottle.GetBottleID()

	assert.NotEqual(t, firstBottleID, secondBottleID)
}

func Test_Functional_DeprecateBottleID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.CommandHelper.RunCommand("commit")

	helper.BottleHelper.Load()
	firstBottleID := helper.BottleHelper.Bottle.GetBottleID()

	// assert that first commit doesn't have any deprecates
	assert.Len(t, helper.BottleHelper.Bottle.Definition.Deprecates, 0)

	helper.BottleHelper.AddArbitraryFileParts(4)
	helper.CommandHelper.RunCommand("commit")
	helper.RequirePartNum(8)

	helper.BottleHelper.Load()
	secondBottleID := helper.BottleHelper.Bottle.GetBottleID()

	assert.NotEqual(t, firstBottleID, secondBottleID)

	assert.Len(t, helper.BottleHelper.Bottle.Definition.Deprecates, 1)
	assert.Equal(t, firstBottleID, helper.BottleHelper.Bottle.Definition.Deprecates[0])

	// test that double commit (commit with no changes) doesn't add duplicate deprecates
	helper.CommandHelper.RunCommand("commit")
	helper.BottleHelper.Load()
	assert.Len(t, helper.BottleHelper.Bottle.Definition.Deprecates, 1)
}

func Test_Functional_NoDeprecate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.CommandHelper.RunCommand("commit")

	helper.BottleHelper.Load()
	firstBottleID := helper.BottleHelper.Bottle.GetBottleID()

	// assert that first commit doesn't have any deprecates
	assert.Len(t, helper.BottleHelper.Bottle.Definition.Deprecates, 0)

	helper.BottleHelper.AddArbitraryFileParts(4)
	helper.CommandHelper.RunCommand("commit", "--no-deprecate")
	helper.RequirePartNum(8)

	helper.BottleHelper.Load()
	secondBottleID := helper.BottleHelper.Bottle.GetBottleID()

	assert.NotEqual(t, firstBottleID, secondBottleID)

	assert.Len(t, helper.BottleHelper.Bottle.Definition.Deprecates, 0)
}
