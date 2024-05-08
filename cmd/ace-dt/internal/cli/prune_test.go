package cli

import (
	"testing"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/functesting"
)

func Test_Functional_PruneParts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootCmd := rootTestCmd()

	helper := functesting.NewCommandHelper(t, rootCmd)

	helper.PopulateCache(300)
	helper.RunCommand("util", "prune")
}

func Test_Functional_PrunePartsToZero(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootCmd := rootTestCmd()

	helper := functesting.NewCommandHelper(t, rootCmd)

	helper.PopulateCache(300)
	helper.RunCommand("util", "prune", "-s=0")
}
