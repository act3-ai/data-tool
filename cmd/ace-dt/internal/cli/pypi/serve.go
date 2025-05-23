package pypi

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/actions/pypi"
	"github.com/act3-ai/go-common/pkg/config"
)

// NnwServeCmd creates a new "serve" subcommand.
func newServeCmd(base *pypi.Action) *cobra.Command {
	action := &pypi.Serve{Action: base}

	cmd := &cobra.Command{
		Use:   "serve [-l PORT] REPOSITORY",
		Short: "Run the PyPI server",
		Long:  `This runs a PyPI (simple) compatible server pulling packages from the (OCI) REPOSITORY as needed to serve the content.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0])
		},
	}

	cmd.Flags().StringVarP(&action.Listen, "listen", "l", config.EnvOr("ACE_DT_PYPI_LISTEN", "localhost:8101"),
		`Interface and port to listen on.
Use :8101 to listen all on interfaces on the standard port.`)

	return cmd
}
