package cli

import (
	"context"
	"path/filepath"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror/testing"
)

func newRunRecipe() *cobra.Command {
	var validate bool
	var ociDir string
	uiOptions := ui.Options{}
	cmd := &cobra.Command{
		Use:   "run-recipe RECIPE",
		Short: "Execute the RECIPE and storing the test data in OCI image layout",
		Long: `The RECIPE file is a JSONL file.  Each line is a JSON file with the following fields:
file - relative file name
mediaType - media type to use for this content (default is application/octet-stream)
algorithm - digest algorithm to use (default is sha256)

All the Hermetic Text Sprig functions are available along with the following functions:

FileDescriptor(filename, mediaType, algorithm) Descriptor
FileDigest(filename, algorithm) string
Tar(filename, ...) []byte
ToData(str string) []byte
Gzip(data []byte) []byte
Zstd(data []byte) []byte
`,
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			recipe := args[0]
			if ociDir == "" {
				ociDir = filepath.Join(filepath.Dir(recipe), "oci")
			}

			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return testing.ProcessRecipe(ctx, recipe, ociDir, validate)
			})
		},
	}

	cmd.Flags().BoolVar(&validate, "validate", true, "Validate the generated OCI artifact")
	cmd.Flags().StringVar(&ociDir, "oci-dir", "", "The OCI image layout directory to store the data in.  If empty data stored in the directory of the recipe file under \"oci\"")
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	return cmd
}
