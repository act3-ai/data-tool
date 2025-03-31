package bottle

import (
	"os"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
)

func Test_Functional_Source(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	myPartName := "testPart1.txt"
	helper.AddBottlePart(myPartName)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("source", "list")
	helper.RunCommand("source", "add", "testSource", "test.com")
	helper.RunCommand("source", "list")

	helper.Load()
	assert.Equal(t, "testSource", helper.Bottle.Definition.Sources[0].Name)
	assert.Equal(t, "test.com", helper.Bottle.Definition.Sources[0].URI)

	helper.RunCommand("source", "remove", "testSource")
	helper.Load()
	assert.Equal(t, 0, len(helper.Bottle.Definition.Sources))
}

func Test_Functional_Source_Reference(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	rootCmd := rootTestCmd()
	helper := NewTestHelper(t, rootCmd)

	remoteBottle := NewBottleHelper(t)
	remoteBottle.AddArbitraryFileParts(10)
	remoteBottle.SetTempBottleRef(rootReg)
	helper.SendBottleToReg(remoteBottle.RootDir, remoteBottle.RegRef)
	helper.PruneCache()
	remoteBottle.Load()
	parentBottleID := "bottle:" + remoteBottle.Bottle.GetBottleID().String()

	myPartName := "testPart1.txt"
	helper.AddBottlePart(myPartName)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("source", "add", "testSource", remoteBottle.RegRef, "--ref")

	assert.NoError(t, helper.load())
	assert.Equal(t, "testSource", helper.Bottle.Definition.Sources[0].Name)
	assert.Equal(t, parentBottleID, helper.Bottle.Definition.Sources[0].URI)
	assert.Len(t, helper.Bottle.Definition.Sources, 1)

}

func Test_Functional_SourceBottleID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	parentBottle := NewBottleHelper(t)
	parentBottle.AddArbitraryFileParts(10)
	parentBottle.SetTempBottleRef(rootReg)
	helper.SendBottleToReg(parentBottle.RootDir, parentBottle.RegRef)
	helper.PruneCache()
	parentBottle.Load()
	parentBottleID := "bottle:" + parentBottle.Bottle.GetBottleID().String()

	myPartName := "testPart1.txt"
	helper.AddBottlePart(myPartName)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("source", "list")
	helper.RunCommand("source", "add", "--path", "bottleSource", parentBottle.RootDir)
	helper.RunCommand("source", "list")

	helper.Load()
	assert.Equal(t, "bottleSource", helper.Bottle.Definition.Sources[0].Name)
	assert.Equal(t, parentBottleID, helper.Bottle.Definition.Sources[0].URI)

	helper.RunCommand("source", "remove", "bottleSource")
	helper.Load()
	assert.Equal(t, len(helper.Bottle.Definition.Sources), 0)
}
