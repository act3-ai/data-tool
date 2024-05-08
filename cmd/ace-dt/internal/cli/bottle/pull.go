/* Command pull
 */

package bottle

import (
	"context"

	telemv1alpha1 "git.act3-ace.com/ace/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/flag"
	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/oci"
	"git.act3-ace.com/ace/go-common/pkg/redact"

	"github.com/spf13/cobra"

	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"
	"git.act3-ace.com/ace/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// pullCmd represents the pull command.
func newPullCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Pull{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		GroupID: "remote",
		Use:     "pull BOTTLE_REFERENCE",
		Short:   "Retrieves a bottle from remote OCI storage",
		Long: `Retrieves a bottle from remote OCI storage, based on the bottle name and tag; stores the resulting files in the current directory, or at the directory supplied with the -d option.

A bottle reference follows the form <registry>/<repository>/<name>:<tag>
  
Pull a bottle using any of the following bottle reference forms
	by name (latest tag)  <registry>/<repository>/<name>
	by tag                <registry>/<repository>/<name>:<tag>
	by digest             <registry>/<repository>/<name>@<digest>
	by bottle ID          bottle:<digest>
where <digest> is often of the form sha256:<sha256 digest, lower case hex encoded>.`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: oci.RefCompletion(action.DataTool),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, args[0])
			})
		},
	}

	cmd.Flags().StringVar(&action.CheckBottleID, "check-bottle-id", "", "Supply a bottle id (eg. sha256:abcdef...) to have ace-dt verify the bottle configuration contents after pull")

	WriteBottleIDFlags(cmd.Flags(), &action.Write)
	PartSelectorFlags(cmd.Flags(), &action.PartSelector)
	flag.TelemetryURLFlags(cmd.Flags(), &action.Telemetry)
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	// Add flag overrides function to override config with flags
	action.Config.AddConfigOverride(func(ctx context.Context, c *v1alpha1.Configuration) error {
		if action.Telemetry.URL != "" {
			c.Telemetry = []telemv1alpha1.Location{
				{URL: redact.SecretURL(action.Telemetry.URL)},
			}
		}
		return nil
	})

	cmd.Example =
		`Pull the tagged bottle TESTSET:TAG from registry REG/REPO/TESTSET:TAG to path PATH:
  ace-dt bottle pull REG/REPO/TESTSET:TAG --bottle-dir PATH

Use bottle ID to pull a bottle from the best available location:
  ace-dt bottle pull bottle:sha256:abc123...

Pull a bottle that is publicly available:
  ace-dt bottle pull us-central1-docker.pkg.dev/aw-df16163b-7044-4662-93fa-ec0/public-down-auth-up/mnist:v2.1 -d mnist

`
	return cmd
}
