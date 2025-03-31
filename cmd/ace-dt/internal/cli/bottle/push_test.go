package bottle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
)

func Test_Functional_SinglePart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(1)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.SetTempBottleRef(rootReg)
	helper.RunCommand("push", helper.RegRef)

	assert.NoError(t, helper.CheckRegForBottle(helper.RegRef, ""))
}

func Test_Functional_TwoParts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(2)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.SetTempBottleRef(rootReg)
	helper.RunCommand("push", helper.RegRef)

	assert.NoError(t, helper.CheckRegForBottle(helper.RegRef, ""))
}

func Test_Functional_DirPart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryDirParts(1)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.SetTempBottleRef(rootReg)
	helper.RunCommand("push", helper.RegRef)

	assert.NoError(t, helper.CheckRegForBottle(helper.RegRef, ""))
}

func Test_Functional_ManyParts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(500)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.SetTempBottleRef(rootReg)
	helper.RunCommand("push", helper.RegRef)

	assert.NoError(t, helper.CheckRegForBottle(helper.RegRef, ""))
}

func Test_Functional_VeryCompressed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(10)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.SetTempBottleRef(rootReg)
	helper.RunCommand("push", helper.RegRef, "-z=max")

	assert.NoError(t, helper.CheckRegForBottle(helper.RegRef, ""))
}

func Test_Functional_LittleCompressed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(10)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.SetTempBottleRef(rootReg)
	helper.RunCommand("push", helper.RegRef, "-z=min")

	assert.NoError(t, helper.CheckRegForBottle(helper.RegRef, ""))
}

func Test_Functional_PushWriteBottleID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(4)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)
	helper.RunCommand("init")
	helper.SetTempBottleRef(rootReg)
	bottleIDFile := filepath.Join(helper.RootDir, ".dt", "bottleid")
	helper.RunCommand("push", helper.RegRef)

	err := helper.CheckRegForBottle(helper.RegRef, "")
	assert.NoError(t, err)
	helper.Load()
	helper.VerifyBottleIDFile(bottleIDFile)
}

func Test_Functional_Push_WithTelemetry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}
	t.Log("Using registry server", rootReg)

	telemetryURL := os.Getenv("TEST_TELEMETRY")
	if telemetryURL == "" {
		t.Skip("Skipping because TEST_TELEMETRY is not set")
	}
	t.Log("Using telemetry server", telemetryURL)

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(1)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)

	helper.RunCommand("init")
	helper.SetTempBottleRef(rootReg)

	// add the telem host to the config
	helper.RunCommand("push", helper.RegRef, "--telemetry", telemetryURL)
}

func Test_Functional_Push_DeprecatesWithTelemetry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}
	t.Log("Using registry server", rootReg)

	telemetryURL := os.Getenv("TEST_TELEMETRY")
	if telemetryURL == "" {
		t.Skip("Skipping because TEST_TELEMETRY is not set")
	}
	t.Log("Using telemetry server", telemetryURL)

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(1)

	// set the bottle dir for all other commands to use when running
	helper.SetBottleDir(helper.RootDir)

	helper.RunCommand("init")
	helper.SetTempBottleRef(rootReg)

	// add the telem host to the config
	helper.RunCommand("push", helper.RegRef, "--telemetry", telemetryURL)

	// save bottleID for comparison later
	helper.Load()
	bottleID := helper.Bottle.GetBottleID()

	// add a part to the bottle
	helper.AddArbitraryFileParts(1)

	// push the new bottle to deprecate the previous one
	helper.RunCommand("push", helper.RegRef, "--telemetry", telemetryURL)

	// assert that the previous bottle was deprecated
	helper.Load()
	assert.Len(t, helper.Bottle.Definition.Deprecates, 1)

	// assert that the bottleID is deprecated by the new bottle
	assert.Equal(t, bottleID, helper.Bottle.Definition.Deprecates[0])
}
