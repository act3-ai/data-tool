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
	helper.AddBottlePart(myPartName)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.RunCommand("metric", "list")
	helper.RunCommand("metric", "add", "testMetric", "0.42", "--desc=testDesc")
	helper.RunCommand("metric", "list")

	helper.Load()
	assert.Equal(t, "testDesc", helper.Bottle.Definition.Metrics[0].Description)
	assert.Equal(t, "testMetric", helper.Bottle.Definition.Metrics[0].Name)
	assert.Equal(t, "0.42", helper.Bottle.Definition.Metrics[0].Value)

	helper.RunCommand("metric", "remove", "testMetric")
	helper.Load()
	assert.Equal(t, 0, len(helper.Bottle.Definition.Metrics))
}
