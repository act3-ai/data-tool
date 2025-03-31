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

	helper.AddArbitraryFileParts(4)

	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	// Assert number of parts post command is equal to the number of parts created
	helper.Load()
	helper.RequirePartNum(4)

	helper.AddArbitraryFileParts(4)
	helper.RunCommand("commit")

	helper.Load()
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
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")

	helper.AddArbitraryDirParts(5)
	helper.RunCommand("commit")
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

	helper.AddBottlePart("test1.txt")
	helper.AddBottlePart("test2.txt")

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)

	helper.RunCommand("init")
	// Assert number of parts post command is equal to the number of parts created
	helper.RequirePartNum(2)

	helper.RemoveBottlePart("test2.txt")

	helper.RunCommand("commit")
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

	helper.AddBottlePart("test1.txt")
	helper.AddBottlePart("test2.txt")

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")

	helper.RunCommand("commit")
	helper.RequirePartNum(2)

	helper.RemoveBottlePart("test2.txt")

	helper.RunCommand("commit")
	helper.RequirePartNum(1)
}

func Test_Functional_ManyCommitNewFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")

	for i := 0; i < 10; i++ {
		helper.AddArbitraryFileParts(1)
		helper.RunCommand("commit")
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
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")

	for i := 0; i < 10; i++ {
		helper.AddArbitraryDirParts(1)
		helper.RunCommand("commit")
	}
}

func Test_Functional_CommitWriteBottleID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	bottleIDFile := filepath.Join(helper.RootDir, ".dt", "bottleid")
	helper.RunCommand("commit")
	helper.RequirePartNum(4)

	helper.Load()
	helper.VerifyBottleIDFile(bottleIDFile)
}

func Test_Functional_UpdatedBottleID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("commit")

	helper.Load()
	firstBottleID := helper.Bottle.GetBottleID()

	helper.AddArbitraryFileParts(4)
	helper.RunCommand("commit")
	helper.RequirePartNum(8)

	helper.Load()
	secondBottleID := helper.Bottle.GetBottleID()

	assert.NotEqual(t, firstBottleID, secondBottleID)
}

func Test_Functional_DeprecateBottleID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("commit")

	helper.Load()
	firstBottleID := helper.Bottle.GetBottleID()

	// assert that first commit doesn't have any deprecates
	assert.Len(t, helper.Bottle.Definition.Deprecates, 0)

	helper.AddArbitraryFileParts(4)
	helper.RunCommand("commit")
	helper.RequirePartNum(8)

	helper.Load()
	secondBottleID := helper.Bottle.GetBottleID()

	assert.NotEqual(t, firstBottleID, secondBottleID)

	assert.Len(t, helper.Bottle.Definition.Deprecates, 1)
	assert.Equal(t, firstBottleID, helper.Bottle.Definition.Deprecates[0])

	// test that double commit (commit with no changes) doesn't add duplicate deprecates
	helper.RunCommand("commit")
	helper.Load()
	assert.Len(t, helper.Bottle.Definition.Deprecates, 1)
}

func Test_Functional_NoDeprecate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("commit")

	helper.Load()
	firstBottleID := helper.Bottle.GetBottleID()

	// assert that first commit doesn't have any deprecates
	assert.Len(t, helper.Bottle.Definition.Deprecates, 0)

	helper.AddArbitraryFileParts(4)
	helper.RunCommand("commit", "--no-deprecate")
	helper.RequirePartNum(8)

	helper.Load()
	secondBottleID := helper.Bottle.GetBottleID()

	assert.NotEqual(t, firstBottleID, secondBottleID)

	assert.Len(t, helper.Bottle.Definition.Deprecates, 0)
}
