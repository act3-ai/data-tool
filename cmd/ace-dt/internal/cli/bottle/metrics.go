package bottle

import (
	"github.com/spf13/cobra"

	actions "github.com/act3-ai/data-tool/internal/actions/bottle"
)

// newBtlPartCmd is the top level command that aggregates subcommands for interacting with bottles parts fields.
func newBtlMetricCmd(tool *actions.Action) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "metadata",
		Use:     "metric",
		Aliases: []string{"metrics"},
		Short:   "Bottle metric operations",
		Long:    `This command group provides subcommands for interacting with metrics associated with a bottle. You can list metrics on this bottle.`,
	}

	cmd.AddCommand(
		newMetricAddCmd(tool),
		newMetricRemoveCmd(tool),
		newMetricListCmd(tool),
	)
	return cmd
}

func newMetricAddCmd(tool *actions.Action) *cobra.Command {
	action := &actions.MetricAdd{Action: tool}

	addMetricCmd := &cobra.Command{
		Use:   "add [METRIC] [VALUE]",
		Short: "add metric information to a bottle",
		Example: `
Add metric precision of value 0.87 to bottle in current directory:
	ace-dt bottle metric add "precision" 0.87

Add metric recall of value 0.68 to bottle to bottle at path <my/bottle/path>:
	ace-dt bottle metric add "recall" 0.68 --desc="recall of car classifier" -d my/bottle/path

Add metric loss with a negative value of -3.14 (use '--' when value is negative):
	ace-dt bottle metric --desc="loss value" -- loss "-3.14"
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], args[1], cmd.OutOrStdout())
		},
	}

	addMetricCmd.Flags().StringVar(&action.Description, "desc", "", "add description of metric")

	return addMetricCmd
}

func newMetricListCmd(tool *actions.Action) *cobra.Command {
	action := &actions.MetricList{Action: tool}

	listMetricCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list metric information from a bottle",
		Example: `
List metrics from bottle in current working directory:
	ace-dt bottle metric list

List metric information from bottle at path <my/bottle/path>:
	ace-dt bottle metric list -d my/bottle/path
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	return listMetricCmd
}

func newMetricRemoveCmd(tool *actions.Action) *cobra.Command {
	action := &actions.MetricRemove{Action: tool}

	removeCmd := &cobra.Command{
		Use:     "remove [NAME]",
		Aliases: []string{"rm"},
		Short:   "Remove metric entry from a bottle",
		Long:    `Remove a metric entry from a bottle. The metric to be removed is identified by using the name of the metric as a key.`,
		Example: `
Remove metric <mse> from bottle in current working directory:
	ace-dt bottle metric remove "mse" 
  
Remove source <f1-score> from bottle at path <my/bottle/path>:
	ace-dt bottle metric rm "f1-score" -d my/bottle/path
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], cmd.OutOrStdout())
		},
	}

	return removeCmd
}
