package bottle

import (
	"path/filepath"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
)

func Test_Functional_PartList(t *testing.T) {
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
	helper.RunCommand("part", "list")
}

func Test_Functional_PartLabel(t *testing.T) {
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
	helper.RunCommand("part", "list") // QUESTION what does this do?  There are no assertions.
	myPartPath := filepath.Join(helper.RootDir, myPartName)
	helper.RunCommand("part", "label", "test=value", myPartPath)
	helper.RunCommand("part", "list")

	helper.RequirePartNum(1)
	assert.Equal(t, labels.Set{"test": "value"}, helper.Bottle.GetPartByName(myPartName).GetLabels())

	helper.RunCommand("part", "label", "test-", myPartPath)
	helper.RequirePartNum(1)
	assert.Empty(t, helper.Bottle.GetPartByName(myPartName).GetLabels())

	t.Log("Now adding a new part and labelling it without calling init or commit")
	myPartName2 := "testPart2.txt"
	helper.AddBottlePart(myPartName2)
	myPartPath2 := filepath.Join(helper.RootDir, myPartName2)
	helper.RunCommand("part", "label", "test=value", myPartPath2)
	helper.RunCommand("part", "list")

	// do we expect this to be 1 or 2?  I would argue 2 because we have two parts now.
	helper.RequirePartNum(2)
}

func Test_Functional_PartMultiLabel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	myPartName1 := "testPart1.txt"
	helper.AddBottlePart(myPartName1)
	myPartName2 := "testPart2.txt"
	helper.AddBottlePart(myPartName2)
	myPartName3 := "testPart3.txt"
	helper.AddBottlePart(myPartName3)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("part", "list")
	myPartPath1 := filepath.Join(helper.RootDir, myPartName1)
	myPartPath2 := filepath.Join(helper.RootDir, myPartName2)
	myPartPath3 := filepath.Join(helper.RootDir, myPartName3)
	helper.RunCommand("part", "label", "test=value", myPartPath1, myPartPath2, myPartPath3)
	helper.RunCommand("part", "list")

	helper.RequirePartNum(3)
	assert.Equal(t, labels.Set{"test": "value"}, helper.Bottle.GetPartByName(myPartName1).GetLabels())
	assert.Equal(t, labels.Set{"test": "value"}, helper.Bottle.GetPartByName(myPartName2).GetLabels())
	assert.Equal(t, labels.Set{"test": "value"}, helper.Bottle.GetPartByName(myPartName3).GetLabels())

	helper.RunCommand("part", "label", "test-", myPartPath1, myPartPath2, myPartPath3)

	helper.RequirePartNum(3)
	assert.Empty(t, helper.Bottle.GetPartByName(myPartName1).GetLabels())
	assert.Empty(t, helper.Bottle.GetPartByName(myPartName2).GetLabels())
	assert.Empty(t, helper.Bottle.GetPartByName(myPartName3).GetLabels())
}
