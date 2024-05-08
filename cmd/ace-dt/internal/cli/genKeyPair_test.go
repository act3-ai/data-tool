package cli

import (
	"testing"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/functesting"
)

func Test_GenKeyPair(t *testing.T) {
	rootCmd := rootTestCmd()

	helper := functesting.NewCommandHelper(t, rootCmd)

	testDir := t.TempDir()

	// test basic
	helper.RunCommand("util", "gen-key-pair", testDir)

	// test aliases
	helper.RunCommand("util", "keygen", testDir)
	helper.RunCommand("util", "genkeys", testDir)

	// test flags
	helper.RunCommand("util", "gen-key-pair", testDir, "--prefix", "testing")

	// test known error cases, errors wanted
	// helper.RunCommand("util", "gen-key-pair", "")
}
