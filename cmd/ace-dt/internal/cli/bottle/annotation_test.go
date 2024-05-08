package bottle

import (
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
)

func Test_Functional_Annotate(t *testing.T) {
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
	helper.CommandHelper.RunCommand("annotate", "list")
	helper.CommandHelper.RunCommand("annotate", "test=true")
	helper.CommandHelper.RunCommand("annotate", "list")

	helper.BottleHelper.Load()
	assert.Equal(t, "true", helper.BottleHelper.Bottle.Definition.Annotations["test"])

	helper.CommandHelper.RunCommand("annotate", "test-")
	helper.BottleHelper.Load()
	assert.Equal(t, "", helper.BottleHelper.Bottle.Definition.Annotations["test"])
}
