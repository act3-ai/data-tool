// Package security contains the ace-dt security subcommands.
package security

import (
	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	securityActions "gitlab.com/act3-ai/asce/data/tool/internal/actions/security"
)

func newScanCommand(tool *securityActions.Action) *cobra.Command {
	action := &securityActions.Scan{Action: tool}
	uiOptions := ui.Options{}
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "",
		Example: `
		ace-dt security scan --source-file /path/to/sources.list
		ace-dt security scan --gathered-image localhost:5000/gather:sync-1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&action.SourceFile, "source-file", "", "Define a sources.list file to scan for vulnerabilities")
	cmd.Flags().StringVar(&action.GatherArtifactReference, "gathered-image", "", "Define an artifact reference created by Gather to scan for vulnerabilities")
	cmd.Flags().StringVarP(&action.Output, "output", "o", "table", "Define how you would like the output displayed. Supported types are json (default), markdown, csv, and table.")
	cmd.Flags().BoolVar(&action.DryRun, "check", false, "Outputs scanning information without generating SBOMS (only applicable to --gathered-image input)")
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)
	return cmd
}
