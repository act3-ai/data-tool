package oci

import (
	"context"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	"git.act3-ace.com/ace/data/tool/internal/actions/oci"
)

// newTreeCmd creates a new cobra.Command for the pull subcommand.
func newTreeCmd(base *oci.Action) *cobra.Command {
	action := &oci.Tree{Action: base}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "tree [--oci-layout] [IMAGE|OCILAYOUT]",
		Short: "Show the tree view of the OCI data graph for a remote image or a local OCI directory.",
		Long: `IMAGE is an OCI image reference.
If --oci-layout is set then the positionaly argument, OCILAYOUT, is used to specify an OCI-Layout directory.  It may be specified as a path and tag (path/to/dir:tag) or a path and digest (path/to/dir@sha256:deedbeef...).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, cmd.OutOrStdout(), args[0])
			})
		},
		Example: ` To display the tree view of a remote image "reg.example.com/image" and referrers:
		ace-dt oci tree reg.example.com/image

	To display the tree view of a remote image "reg.example.com/image" with all predecessors (not just referrers):
		ace-dt oci tree reg.example.com/image --only-referrers=true

	To display the tree view of a local OCI directory in ~/imageDir at at my-tag:
		ace-dt oci tree --oci-layout ~/imageDir:my-tag
	`,
	}

	// input options
	cmd.Flags().BoolVar(&action.OCILayout, "oci-layout", false, "Argument is a path and tag/digest in OCI image layout format")

	// tree traversal options
	cmd.Flags().IntVar(&action.Depth, "depth", 10, "Maximum depth of the tree to display")
	// cmd.Flags().BoolVarP(&action.Predecessors, "predecessors", "p", false, "Display the predecessor tree instead of successors")
	cmd.Flags().BoolVar(&action.OnlyReferrers, "only-referrers", true, "When true this will only show referrers (those who's subject field matches the node).  When false this will display all known immediate predecessors irregardless of if reference is using the subject field.")
	cmd.Flags().StringVar(&action.ArtifactType, "artifact-type", "", "Limit predecessors to this artifact type")

	// formatting options
	cmd.Flags().BoolVar(&action.ShowBlobs, "show-blobs", true, "Display the blob and config descriptors.  Still shows subjects.")
	cmd.Flags().BoolVarP(&action.ShortDigests, "short-digests", "s", true, "For brevity, display only 12 hexadecimal digits of the digest (omits the algorithm as well)")
	cmd.Flags().BoolVar(&action.ShowAnnotations, "show-annotations", false, "Show annotations of manifests and descriptors")

	return cmd
}
