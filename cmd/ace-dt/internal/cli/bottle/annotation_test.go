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
	helper.AddBottlePart(myPartName)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("annotate", "list")
	helper.RunCommand("annotate", "test=true")
	helper.RunCommand("annotate", "list")

	helper.Load()
	assert.Equal(t, "true", helper.Bottle.Definition.Annotations["test"])

	helper.RunCommand("annotate", "test-")
	helper.Load()
	assert.Equal(t, "", helper.Bottle.Definition.Annotations["test"])
}
