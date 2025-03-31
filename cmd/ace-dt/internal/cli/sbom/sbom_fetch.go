package sbom

import (
	"fmt"

	"github.com/spf13/cobra"

	sbomActions "gitlab.com/act3-ai/asce/data/tool/internal/actions/sbom"
)

// NewSBOMFetchCommand creates a new sbom fetch sub-command.
func NewSBOMFetchCommand(tool *sbomActions.Action) *cobra.Command {
	action := &sbomActions.FetchSBOM{Action: tool}
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch the SBOM(s) for a given image or gather artifact and save them to file or print to standard out",
		Example: `
		To save the SBOMs for an image to a destination directory:
		ace-dt sbom fetch -- image reg.example.com/image1:tag1 -o dest/dir

		To print the SBOMs for a gathered image to standard out:
		ace-dt sbom fetch --gathered-image localhost:5000/gather:sync-1 -o -

		To save the linux/amd64 SBOM for a specific image within a gather artifact:
		ace-dt sbom fetch --gathered-image localhost:5000/gather:sync-1 --image docker.io/library/image:v1.0 --platform=linux/amd64 -o dest/dir`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if action.SourceImage == "" && action.GatherArtifactReference == "" {
				return fmt.Errorf("either --image or --gathered-image must be specified")
			}
			return action.Run(cmd.Context())
		},
	}
	cmd.Flags().StringVar(&action.SourceImage, "image", "", "Define a sources.list file to scan for vulnerabilities")
	cmd.Flags().StringVar(&action.GatherArtifactReference, "gathered-image", "", "Define an artifact reference created by Gather to scan for vulnerabilities")
	cmd.Flags().StringVarP(&action.Output, "output", "o", "-", "- for stdout or pass a directory path to save to file")
	cmd.Flags().StringSliceVarP(&action.Platforms, "platforms", "p", []string{}, "Only fetch SBOMs that match the specified platform(s).")
	return cmd
}
