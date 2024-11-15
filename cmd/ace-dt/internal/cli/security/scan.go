// Package security contains the ace-dt security subcommands.
package security

import (
	"fmt"

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

		To scan a sourcefile (text or csv list of images):
		ace-dt security scan --source-file /path/to/sources.list

		To scan a gathered artifact:
		ace-dt security scan --gathered-image localhost:5000/gather:sync-1
		
		To scan a gathered artifact and push scan reports to the registry:
		ace-dt security scan --gathered-image localhost:5000/gather:sync-1 --push-reports

		To scan a gathered artifact and display CVE information:
		ace-dt security scan --gathered-image localhost:5000/gather:sync-1 --display-cve

		To set the lowest vulnerability level shown:
		ace-dt security scan --gathered-image localhost:5000/gather:sync-1 --vulnerability-level=low

		To get multiple formatted reports:
		ace-dt security scan --gathered-image localhost:5000/gather:sync-1 -o table=report.txt -o csv=report.csv -o markdown=report.md -o json=report.json
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if action.SourceFile == "" && action.GatherArtifactReference == "" {
				return fmt.Errorf("either --source-file or --gathered-image must be specified")
			}
			return action.Run(cmd.Context())
		},
	}
	cmd.Flags().StringVar(&action.SourceFile, "source-file", "", "Define a sources.list file to scan for vulnerabilities")
	cmd.Flags().StringVar(&action.GatherArtifactReference, "gathered-image", "", "Define an artifact reference created by Gather to scan for vulnerabilities")
	// cmd.Flags().StringVar(&action.SaveReport, "report-file", "", "Saves the vulnerability report to user-specified location")
	cmd.Flags().StringVar(&action.VulnerabilityLevel, "vulnerability-level", "medium", "The lowest level of vulnerability to display in reports and outputs. Options are 'critical', 'high', 'medium', 'low', 'negligable', or 'unknown'")
	cmd.Flags().StringSliceVarP(&action.Output, "output", "o", []string{"table"}, "Define how you would like the output displayed. Supported types are json (default), markdown, csv, and table. Multiple values are supported.")
	cmd.Flags().BoolVar(&action.DryRun, "check", false, "Outputs scanning information without generating SBOMS (only applicable to --gathered-image input)")
	cmd.Flags().BoolVar(&action.DisplayCVE, "display-cve", false, "Outputs the CVE information to file or stdout")
	cmd.Flags().BoolVar(&action.DisplayPlatforms, "display-platforms", false, "Outputs a table of platform information to file or stdout")
	cmd.Flags().BoolVar(&action.PushReport, "push-reports", false, "Pushes and attaches the vulnerability reports to each image.")

	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)
	return cmd
}
