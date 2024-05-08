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
	helper.BottleHelper.AddBottlePart(myPartName)
	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.CommandHelper.RunCommand("label", "list")
	helper.CommandHelper.RunCommand("label", "test=value")
	helper.CommandHelper.RunCommand("label", "list")

	helper.BottleHelper.Load()
	assert.Equal(t, "value", helper.BottleHelper.Bottle.Definition.Labels["test"])

	helper.CommandHelper.RunCommand("label", "test-")
	helper.BottleHelper.Load()
	assert.Equal(t, "", helper.BottleHelper.Bottle.Definition.Labels["test"])
}

func Test_Functional_MultiLabel(t *testing.T) {
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
	helper.CommandHelper.RunCommand("label", "list")
	helper.CommandHelper.RunCommand("label", "test1=value1", "test2=value2", "test3=value3")
	helper.CommandHelper.RunCommand("label", "list")

	helper.BottleHelper.Load()
	assert.Equal(t, "value1", helper.BottleHelper.Bottle.Definition.Labels["test1"])
	assert.Equal(t, "value2", helper.BottleHelper.Bottle.Definition.Labels["test2"])
	assert.Equal(t, "value3", helper.BottleHelper.Bottle.Definition.Labels["test3"])

	helper.CommandHelper.RunCommand("label", "test1-")
	helper.CommandHelper.RunCommand("label", "test2-")
	helper.CommandHelper.RunCommand("label", "test3-")
	helper.BottleHelper.Load()
	assert.Equal(t, "", helper.BottleHelper.Bottle.Definition.Labels["test1"])
	assert.Equal(t, "", helper.BottleHelper.Bottle.Definition.Labels["test2"])
	assert.Equal(t, "", helper.BottleHelper.Bottle.Definition.Labels["test3"])
}
