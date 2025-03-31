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
	helper.AddBottlePart(myPartName)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("artifact", "list")
	myPartPath := filepath.Join(helper.RootDir, myPartName)
	helper.RunCommand("artifact", "set", "test", myPartPath)
	helper.RunCommand("artifact", "list")

	helper.Load()
	require.Len(t, helper.Bottle.Definition.PublicArtifacts, 1)
	assert.Equal(t, "test", helper.Bottle.Definition.PublicArtifacts[0].Name)

	helper.RunCommand("artifact", "remove", myPartPath)
	helper.Load()
	assert.Len(t, helper.Bottle.Definition.PublicArtifacts, 0)
}

func Test_Functional_ArtifactMimeType(t *testing.T) {
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
	helper.RunCommand("artifact", "list")
	myPartPath := filepath.Join(helper.RootDir, myPartName)
	helper.RunCommand("artifact", "set", "test", myPartPath, "--media-type", "text/plain")
	helper.RunCommand("artifact", "list")

	helper.Load()
	require.Len(t, helper.Bottle.Definition.PublicArtifacts, 1)
	assert.Equal(t, "test", helper.Bottle.Definition.PublicArtifacts[0].Name)

	helper.RunCommand("artifact", "remove", myPartPath)
	helper.Load()
	assert.Len(t, helper.Bottle.Definition.PublicArtifacts, 0)
}
