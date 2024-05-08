/* Test this with
podman run -d -p 127.0.0.1:5000:5000 docker.io/library/registry:2
go run ./cmd/ace-dt pypi to-oci localhost:5000/pypi pyzmq
*/

package pypi

import (
	"context"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	"git.act3-ace.com/ace/data/tool/internal/actions/pypi"
)

// newToOCICmd creates a new cobra.Command for the pypi to-oci subcommand.
func newToPyPICmd(base *pypi.Action) *cobra.Command {
	action := &pypi.ToPyPI{Action: base}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "to-pypi OCIREF PYPI-REPOSITORY",
		Short: "Pulls packages from OCI and uploads them to the python package index",
		Long: `Pulls packages from OCI at OCIREF and uploads them to the python package index at PYPI-REPOSITORY.

OCIREF is a repository reference (no tag).
PYPI_REPOSITORY is a python package index.  This is the same URL used by "twine"'s "TWINE_REPOSITORY_URL" setting.  It does not include the trailing "/simple" that is used by "pip".
`,
		Example: `To upload all the packages in OCI at reg.example.com to Gitlab PyPI at https://git.example.com/api/v4/projects/1234/packages/pypi run the command
ace-dt pypi to-pypi reg.example.com/mypypi https://git.example.com/api/v4/projects/1234/packages/pypi`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ociRepo := args[0]
			pypiRepo := args[1]
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, ociRepo, pypiRepo)
			})
		},
	}

	cmd.Flags().BoolVar(&action.DryRun, "dry-run", false, "Dry run by only determining what work needs to be done.  Does not upload distribution files to the python package index.")

	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	return cmd
}
