package bottle

import (
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
)

func Test_Functional_Metric(t *testing.T) {
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
	helper.CommandHelper.RunCommand("metric", "list")
	helper.CommandHelper.RunCommand("metric", "add", "testMetric", "0.42", "--desc=testDesc")
	helper.CommandHelper.RunCommand("metric", "list")

	helper.BottleHelper.Load()
	assert.Equal(t, "testDesc", helper.BottleHelper.Bottle.Definition.Metrics[0].Description)
	assert.Equal(t, "testMetric", helper.BottleHelper.Bottle.Definition.Metrics[0].Name)
	assert.Equal(t, "0.42", helper.BottleHelper.Bottle.Definition.Metrics[0].Value)

	helper.CommandHelper.RunCommand("metric", "remove", "testMetric")
	helper.BottleHelper.Load()
	assert.Equal(t, 0, len(helper.BottleHelper.Bottle.Definition.Metrics))
}
