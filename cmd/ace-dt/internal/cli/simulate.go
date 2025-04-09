package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/cmd/ace-dt/internal/cli/internal/ui"
	"github.com/act3-ai/data-tool/internal/actions"
)

// newSimulateCmd creates a new cobra.Command to visually test UI components.
func newSimulateCmd(tool *actions.DataTool) *cobra.Command {
	action := &actions.Simulate{DataTool: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "simulate",
		Short: "Simulate UI components",
		Long: `This is a debugging command to test the UI components.
		first argument is to change the number of "tasks" ran
		second argument is to change parallel task run count (max parallel)
	`,
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx)
			})
		},
	}

	// flag for NumTasks
	cmd.Flags().IntVarP(&action.NumTasks, "numTasks", "n", 1, "Number of tasks to run")

	// flag for NumMaxParallel
	cmd.Flags().IntVarP(&action.NumMaxParallel, "numMaxParallel", "m", 10, "Number of tasks to run in parallel")

	// flag for NumCountRecursive
	cmd.Flags().IntVarP(&action.NumCountRecursive, "numCountRecursive", "r", 2, "Number of recursive counting tasks to run")

	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	cmd.Hidden = true

	return cmd
}
