package bottle

import (
	"context"
	"path/filepath"
	"testing"

	actions "git.act3-ace.com/ace/data/tool/internal/actions"
)

func Test_Functional_SignBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

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

	helper.CommandHelper.RunCommand("sign", "--key-path", privKeyPath, "--key-api", "no-kms", "--user-id", "testingUser", "--key-id", "BottleKey")
}
