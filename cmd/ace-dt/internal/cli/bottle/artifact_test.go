package bottle

import (
	"path/filepath"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Functional_Artifact(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	myPartName := "testPart1.txt"
	helper.BottleHelper.AddBottlePart(myPartName)

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.CommandHelper.RunCommand("artifact", "list")
	myPartPath := filepath.Join(helper.BottleHelper.RootDir, myPartName)
	helper.CommandHelper.RunCommand("artifact", "set", "test", myPartPath)
	helper.CommandHelper.RunCommand("artifact", "list")

	helper.BottleHelper.Load()
	require.Len(t, helper.BottleHelper.Bottle.Definition.PublicArtifacts, 1)
	assert.Equal(t, "test", helper.BottleHelper.Bottle.Definition.PublicArtifacts[0].Name)

	helper.CommandHelper.RunCommand("artifact", "remove", myPartPath)
	helper.BottleHelper.Load()
	assert.Len(t, helper.BottleHelper.Bottle.Definition.PublicArtifacts, 0)
}

func Test_Functional_ArtifactMimeType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	myPartName := "testPart1.txt"
	helper.BottleHelper.AddBottlePart(myPartName)

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.CommandHelper.RunCommand("artifact", "list")
	myPartPath := filepath.Join(helper.BottleHelper.RootDir, myPartName)
	helper.CommandHelper.RunCommand("artifact", "set", "test", myPartPath, "--media-type", "text/plain")
	helper.CommandHelper.RunCommand("artifact", "list")

	helper.BottleHelper.Load()
	require.Len(t, helper.BottleHelper.Bottle.Definition.PublicArtifacts, 1)
	assert.Equal(t, "test", helper.BottleHelper.Bottle.Definition.PublicArtifacts[0].Name)

	helper.CommandHelper.RunCommand("artifact", "remove", myPartPath)
	helper.BottleHelper.Load()
	assert.Len(t, helper.BottleHelper.Bottle.Definition.PublicArtifacts, 0)
}
