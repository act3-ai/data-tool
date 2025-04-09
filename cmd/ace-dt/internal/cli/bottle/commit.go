/*
Command commit
*/

package bottle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "github.com/act3-ai/data-tool/internal/actions/bottle"
	"github.com/act3-ai/data-tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// commitCmd represents the commit command.
func newCommitCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Commit{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		GroupID: "basic",
		Use:     "commit",
		Short:   "Processes and commits local changes to a bottle",
		Long: `Processes any local additions, deletions, and modifications to known 
parts of a bottle to be processed and added to the bottle metadata.

Commits will automatically deprecate the previous version (bottleID) of this bottle.
This can be disabled by passing the --no-deprecate flag.
`,
		Example: `
Commit from current working directory:
	ace-dt bottle commit

Commit a bottle at path <TESTSET>:
	ace-dt bottle commit -d TESTSET

View information prior to commit: 
	ace-dt bottle status
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx)
			})
		},
	}

	CompressionLevelFlags(cmd.Flags(), &action.Compression)

	// Add flag no-deprecate to disable deprecation
	cmd.Flags().BoolVar(&action.NoDeprecate, "no-deprecate", false, "Disable deprecation of previous bottle version")

	// Add flag overrides function to override config with flags
	action.Config.AddConfigOverride(func(ctx context.Context, c *v1alpha1.Configuration) error {
		if action.Compression.Level != "" {
			c.CompressionLevel = action.Compression.Level
		}
		return nil
	})

	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	return cmd
}
