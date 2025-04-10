package sbom

import (
	"fmt"

	"github.com/spf13/cobra"

	sbomActions "github.com/act3-ai/data-tool/internal/actions/sbom"
)

// NewSBOMListCommand creates a new sbom list sub-command.
func NewSBOMListCommand(tool *sbomActions.Action) *cobra.Command {
	action := &sbomActions.List{Action: tool}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the SBOM(s) for a given image or gather artifact",
		Example: `
		To list the SBOMs available for an image
		ace-dt sbom list --image reg.example.com/image:v1.0
		
		To list all SBOMs and their digests within a gathered artifact:
		ace-dt sbom list --gathered-image localhost:5000/gather:sync-1
		
		To list the SBOM of a specific platform(s) within a multi-architecture image
		ace-dt sbom list --image reg.example.com/image:v1.0 --platforms=linux/amd64 

		To list the SBOM of a specific image within a gather artifact
		ace-dt sbom list --gathered-image localhost:5000/gather:sync-1 --image reg.example.com/image:v1.0
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if action.SourceImage == "" && action.GatherArtifactReference == "" {
				return fmt.Errorf("either --image or --gathered-image must be specified")
			}
			return action.Run(cmd.Context())
		},
	}
	// TODO update output flag doc
	cmd.Flags().StringVar(&action.SourceImage, "image", "", "Define an artifact reference created by Gather to scan for vulnerabilities")
	cmd.Flags().StringVar(&action.GatherArtifactReference, "gathered-image", "", "Define an artifact reference created by Gather to scan for vulnerabilities")
	cmd.Flags().StringSliceVarP(&action.Platforms, "platforms", "p", []string{}, "Only gather images that match the specified platform(s). Warning: This will modify the manifest digest/reference.")
	return cmd
}
