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
	helper.AddBottlePart(myPartName)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("author", "list")
	helper.RunCommand("author", "add", "testName", "testEmail@gmail.com", "--url=testURL.com")
	helper.RunCommand("author", "list")

	helper.Load()
	assert.Equal(t, 1, len(helper.Bottle.Definition.Authors))
	author := helper.Bottle.Definition.Authors[0]
	assert.Equal(t, "testName", author.Name)
	assert.Equal(t, "testURL.com", author.URL)
	assert.Equal(t, "testEmail@gmail.com", author.Email)

	helper.RunCommand("author", "remove", "testName")
	helper.Load()
	assert.Equal(t, 0, len(helper.Bottle.Definition.Authors))
}
