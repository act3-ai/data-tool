package bottle

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"

	actions "gitlab.com/act3-ai/asce/data/tool/internal/actions"
)

// Was for testing cryptographic agility, which has been temporarily taken out
func Test_Functional_Verify(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	defer leaktest.Check(t)() //nolint:revive

	rootCmd := rootTestCmd()
	helper := NewTestHelper(t, rootCmd)
	testingDir := t.TempDir()
	ctx := context.Background()

	// prep a sample bottle
	helper.BottleHelper.AddArbitraryFileParts(2)
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	// prep a private key
	privKeyPath := filepath.Join(testingDir, "testing.key")
	err := actions.GenAndWriteKeyPair(ctx, testingDir, "testing", true)
	if err != nil {
		t.Fatalf("generating key pair: %v", err)
	}
	t.Log("private key created")

	// sign
	helper.CommandHelper.RunCommand("sign", "--key-path", privKeyPath, "--key-api", "cert-basic", "--user-id", "testingUser", "--key-id", "BottleKey")

	// verify
	helper.CommandHelper.RunCommand("verify")
}

func Test_Functional_SignPushPullVerify(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint:revive

	rootCmd := rootTestCmd()
	helper := NewTestHelper(t, rootCmd)
	helper.BottleHelper.AddArbitraryFileParts(2)
	pullDir := t.TempDir()
	keyDir := t.TempDir()

	ctx := context.Background()

	// prep a private key
	privKeyPath := filepath.Join(keyDir, "testing.key")
	err := actions.GenAndWriteKeyPair(ctx, keyDir, "testing", true)
	if err != nil {
		t.Fatalf("generating key pair: %v", err)
	}

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.BottleHelper.SetTempBottleRef(rootReg)

	// sign
	helper.CommandHelper.RunCommand("sign", "--key-path", privKeyPath, "--key-api", "cert-basic", "--user-id", "testingUser", "--key-id", "BottleKey")

	// push
	helper.CommandHelper.RunCommand("push", helper.BottleHelper.RegRef) // sigs pushed as well

	assert.NoError(t, helper.CheckRegForBottle(helper.BottleHelper.RegRef, ""))

	// pull
	helper.CommandHelper.SetBottleDir(pullDir)
	helper.CommandHelper.RunCommand("pull", helper.BottleHelper.RegRef, "--bottle-dir", pullDir)

	helper.EqualBottles(pullDir, helper.BottleHelper.RootDir) // filesystem check should validate sig dir as well

	// verify
	helper.CommandHelper.RunCommand("verify")
}
