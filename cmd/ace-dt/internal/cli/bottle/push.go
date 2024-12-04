/*
Command push
*/

package bottle

import (
	"context"

	"github.com/spf13/cobra"

	telemv1alpha2 "git.act3-ace.com/ace/data/telemetry/v2/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
	"git.act3-ace.com/ace/go-common/pkg/redact"
	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/flag"
	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "gitlab.com/act3-ai/asce/data/tool/internal/actions/bottle"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// pushCmd represents the push command.
func newPushCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Push{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		GroupID: "remote",
		Use:     "push BOTTLE_REFERENCE",
		Short:   "Archives, compresses, and uploads bottle to an OCI registry",
		Long: `The files at the specified location are archived and compressed using Zstandard compression, and uploaded to the specified OCI registry.
	
A bottle reference follows the form <registry>/<repository>/<name>:<tag>

Pushing a bottle with altered data or metadata will automatically deprecate 
the previous version (bottleID) of this bottle. This can be disabled 
by passing the --no-deprecate flag.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				action.Ref = args[0]
				return action.Run(ctx)
			})
		},
	}

	cmd.Flags().BoolVarP(&action.NoOverwrite, "no-overwrite", "n", false, "Only push data if if doesn't already exist")
	cmd.Flags().BoolVar(&action.NoDeprecate, "no-deprecate", false, "Disable deprecation of previous bottle version")

	CompressionLevelFlags(cmd.Flags(), &action.Compression)
	flag.TelemetryURLFlags(cmd.Flags(), &action.Telemetry)
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	// Add flag overrides function to override config with flags
	action.Config.AddConfigOverride(func(ctx context.Context, c *v1alpha1.Configuration) error {
		if action.Telemetry.URL != "" {
			c.Telemetry = []telemv1alpha2.Location{
				{URL: redact.SecretURL(action.Telemetry.URL)},
			}
		}
		if action.Compression.Level != "" {
			c.CompressionLevel = action.Compression.Level
		}
		return nil
	})

	cmd.Example = `
To push the bottle TESTSET to the registry REGISTRY/REPO/NAME:TAG:
	ace-dt bottle push REGISTRY/REPO/NAME:TAG -d ./TESTSET

To add a telemetry server, and send metadata after the push, set the telemetry URL and username (see ace-dt config --help):
	export ACE_DT_TELEMETRY_URL=http://127.0.0.1:8100
	export ACE_DT_TELEMETRY_USERNAME=exampleuser
Then push like normal:
	ace-dt bottle push REGISTRY/REPO/NAME:TAG -d ./TESTSET

Share a bottle with other users by giving them the bottle reference
OR, share the bottle ID for Telemetry Server support.
`

	return cmd
}
