package bottle

import (
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
)

func Test_Functional_Label(t *testing.T) {
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
	helper.RunCommand("label", "list")
	helper.RunCommand("label", "test=value")
	helper.RunCommand("label", "list")

	helper.Load()
	assert.Equal(t, "value", helper.Bottle.Definition.Labels["test"])

	helper.RunCommand("label", "test-")
	helper.Load()
	assert.Equal(t, "", helper.Bottle.Definition.Labels["test"])
}

func Test_Functional_MultiLabel(t *testing.T) {
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
	helper.RunCommand("label", "list")
	helper.RunCommand("label", "test1=value1", "test2=value2", "test3=value3")
	helper.RunCommand("label", "list")

	helper.Load()
	assert.Equal(t, "value1", helper.Bottle.Definition.Labels["test1"])
	assert.Equal(t, "value2", helper.Bottle.Definition.Labels["test2"])
	assert.Equal(t, "value3", helper.Bottle.Definition.Labels["test3"])

	helper.RunCommand("label", "test1-")
	helper.RunCommand("label", "test2-")
	helper.RunCommand("label", "test3-")
	helper.Load()
	assert.Equal(t, "", helper.Bottle.Definition.Labels["test1"])
	assert.Equal(t, "", helper.Bottle.Definition.Labels["test2"])
	assert.Equal(t, "", helper.Bottle.Definition.Labels["test3"])
}
