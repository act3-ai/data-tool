// Package bottle provides command and subcommands for manipulating bottles and bottle metadata.
package bottle

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	actions "gitlab.com/act3-ai/asce/data/tool/internal/actions"
	bottleactions "gitlab.com/act3-ai/asce/data/tool/internal/actions/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
)

// NewBottleCmd is the top level command that aggregates subcommands for interacting with bottles.
func NewBottleCmd(tool *actions.DataTool) *cobra.Command {
	action := &bottleactions.Action{DataTool: tool}

	cmd := &cobra.Command{
		GroupID: "core",
		Use:     "bottle",
		Aliases: []string{"b"},
		Short:   "Bottle operations",
		Long: `A data bottle is a directory of data with optional rich metadata about the bottle.
This command group provides subcommands for common bottle operations, including push, pull, commit and init.
Use label command to add bottle level labels.
Use part command to add part level labels.
`,
	}

	cwd, err := os.Getwd()
	cobra.CheckErr(err)
	cmd.PersistentFlags().StringVarP(&action.Dir, "bottle-dir", "d", cwd, "Specify bottle directory")

	cmd.AddGroup(
		&cobra.Group{
			ID:    "basic",
			Title: "Basic bottle commands",
		},
		&cobra.Group{
			ID:    "remote",
			Title: "Commands that access remote OCI registries",
		},
		&cobra.Group{
			ID:    "metadata",
			Title: "Commands to edit bottle metadata",
		},
	)

	cmd.AddCommand(
		newInitCmd(action),
		newCommitCmd(action),
		newPushCmd(action),
		newShowCmd(action),
		newPullCmd(action),
		newEditCmd(action),
		newDeleteCmd(action),
		newStatusCmd(action),
		newGuiCmd(action),
		newLabelCmd(action),
		newBtlAnnotateCmd(action),
		newBtlPartCmd(action),
		newBtlAuthorCmd(action),
		newBtlMetricCmd(action),
		newBtlSourceCmd(action),
		newBtlArtifactCmd(action),
		newBtlDescribeCmd(action),
		newVerifyCmd(action),
		newSignCmd(action),
	)
	return cmd
}

// TODO the below functions should take a FlagSet instead of a command

// WriteBottleIDFlags adds flag for writing bottle ID to a cobra command.
func WriteBottleIDFlags(flags *pflag.FlagSet, action *bottleactions.WriteBottleOptions) {
	flags.StringVar(&action.BottleIDFile, "write-bottle-id", "", "File path to write the bottle ID after a bottle operation")
}

// PartSelectorFlags adds flags for part selectors.
func PartSelectorFlags(flags *pflag.FlagSet, action *bottle.PartSelectorOptions) {
	flags.BoolVar(&action.Empty, "empty", false, "retrieve empty bottle, only containing metadata")
	flags.StringArrayVarP(&action.Selectors, "selector", "l", []string{}, "Provide selectors for which parts to retrieve. Format \"name=value\"")
	flags.StringArrayVarP(&action.Parts, "part", "p", []string{}, "Parts to retrieve")
	flags.StringArrayVarP(&action.Artifacts, "artifact", "u", []string{}, "Retrieve only parts containing the provided public artifact type")
}

// CompressionLevelFlags adds a flag for changing the default compression level for part compression.
func CompressionLevelFlags(flags *pflag.FlagSet, action *bottleactions.CompressionLevelOptions) {
	flags.StringVarP(&action.Level, "compression-level", "z", "",
		`Overrides the compression level.`)
}
