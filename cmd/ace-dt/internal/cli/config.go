package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/actions"
	"github.com/act3-ai/go-common/pkg/config"
)

const configExample = `Configuration can be modified with the following environment variables:
ACE_DT_CACHE_PRUNE_MAX: Maximum cache prune size
ACE_DT_CACHE_PATH: directory to use for caching
ACE_DT_CHUNK_SIZE: Maximum chunk size for chunked uploads (set to "0" to disable)
ACE_DT_CONCURRENT_HTTP: Maximum concurrent network connections.
ACE_DT_REGISTRY_AUTH_FILE then REGISTRY_AUTH_FILE: Docker registry auth file
ACE_DT_EDITOR then VISUAL then EDITOR: Sets the editor to use for editing bottle schema.
ACE_DT_TELEMETRY_URL: If set will overwrite the telemetry configuration to only use this telemetry server URL.  Use the config file if you need multiple telemetry servers.
ACE_DT_TELEMETRY_USERNAME: Username to use for reporting events to telemetry.

To save the config to the default location run
$ ace-dt config -s > %s
then modify the configuration file as needed.
`

// newConfigCmd creates a new "config" subcommand.
func newConfigCmd(tool *actions.DataTool) *cobra.Command {
	action := &actions.Config{
		DataTool: tool,
	}

	cmd := &cobra.Command{
		GroupID: "setup",
		Use:     "config",
		Short:   "Show the current configuration",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
		Example: fmt.Sprintf(configExample, config.DefaultConfigPath("ace", "dt", "config.yaml")),
	}

	cmd.Flags().BoolVarP(&action.Sample, "sample", "s", false,
		"Output a sample configuration that can be used in a configuration file.")

	return cmd
}
