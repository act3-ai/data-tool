package bottle

import (
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
)

func Test_Functional_Author(t *testing.T) {
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
	helper.CommandHelper.RunCommand("author", "list")
	helper.CommandHelper.RunCommand("author", "add", "testName", "testEmail@gmail.com", "--url=testURL.com")
	helper.CommandHelper.RunCommand("author", "list")

	helper.BottleHelper.Load()
	assert.Equal(t, 1, len(helper.BottleHelper.Bottle.Definition.Authors))
	author := helper.BottleHelper.Bottle.Definition.Authors[0]
	assert.Equal(t, "testName", author.Name)
	assert.Equal(t, "testURL.com", author.URL)
	assert.Equal(t, "testEmail@gmail.com", author.Email)

	helper.CommandHelper.RunCommand("author", "remove", "testName")
	helper.BottleHelper.Load()
	assert.Equal(t, 0, len(helper.BottleHelper.Bottle.Definition.Authors))
}
