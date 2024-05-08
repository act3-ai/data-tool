package mirror

import (
	"context"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "gitlab.com/act3-ai/asce/data/tool/internal/actions/mirror"
)

// newGatherCmd represents the mirror gather command.
func newGatherCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Gather{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "gather SOURCES-FILE IMAGE",
		Short: "Efficiently copies images listed in SOURCES-FILE to the IMAGE",
		Long: `Efficiently copies images listed in SOURCES-FILE to the IMAGE.
		
SOURCES-FILE is a text file with one OCI image reference per line.  Lines that begin with # are ignored. 
Labels can be added to each source in the SOURCES-FILE by separating with a comma and following a key=value format. These will be added as annotations to that manifest:
reg.example.com/library/source1,component=core,module=test

IMAGE is an OCI image reference that will be used to push all the missing blobs and manifests.
The manifest at the tag will be a OCI Image Index.

Many gather commands can be run to gather images from different registries.  Ensure that they push to different tags in the destination repository.

Often the next command run is the "ace-dt mirror serialize" command.`,
		Example: `
ace-dt mirror gather repos.list reg.example.com/project/repo:sync-45

To gather with custom annotations:
ace-dt mirror gather repos.list reg.example.com/project/repo:sync-45 --annotations=key1=value1,key2=value2

To gather to a repository that does not support nested indexes:
ace-dt mirror gather repos.list reg.example.com/project/repo:sync-45 --index-fallback

To gather to a repository and only include manifests for specific platforms:
ace-dt mirror gather repos.list reg.example.com/project/repo:sync-45 -p linux/arm/v8 -p linux/amd64`,

		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, args[0], args[1])
			})
		},
	}
	cmd.Flags().BoolVar(&action.IndexFallback, "index-fallback", false, "Tells ace-dt to add indexes in annotations for registries that do not support nested indexes (i.e., not OCI 1.1 compliant).  This makes the references to the sub-indexes not real references therefore a garbage collection process might incorrectly delete the sub-indexes.  Therefore, this should only be used when necessary (e.g., when targeting Artifactory).")
	cmd.Flags().StringToStringVarP(&action.ExtraAnnotations, "annotations", "a", map[string]string{}, "Define any additional annotations to add to the index of the gather repository.")
	cmd.Flags().StringSliceVarP(&action.Platforms, "platforms", "p", []string{}, "Only gather images that match the specified platform(s). Warning: This will modify the manifest digest/reference.")
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	return cmd
}
