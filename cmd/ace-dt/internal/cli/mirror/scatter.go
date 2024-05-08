package mirror

import (
	"context"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "git.act3-ace.com/ace/data/tool/internal/actions/mirror"
)

// newGatherCmd represents the mirror gather command.
func newScatterCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Scatter{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "scatter IMAGE MAPPER",
		Short: "A command that scatters images to destination registries defined in the MAPPER",
		Long: `A command that scatters images located in the source registry repo to multiple
remote repositories defined by the user with MAPPER.

The MAPPER types currently supported are nest, first-prefix (csv format), digests (csv format) and go-template.
The format of MAPPER is MAP-TYPE=MAP-ARG

If MAP-TYPE is "nest" then scatter will nest all the images under MAP-ARG.
For example, is MAP-ARG is "reg.other.com" then a gathered image "foo.com/bar" will map to "reg.other.com/foo.com/bar".

Passing a first-prefix MAPPER requires a csv file that has formatted lines of: source,destination. 
The ace-dt mirror scatter will send the source reference to the first prefix match that it makes.
This format also allows defining the source as a digest that is present in the source repository.

Passing a digests MAP-FILE requires a csv file that has formatted lines of: digest-string, destination.
Scatter will send each digest to the locations defined in the map file provided. 

Passing a go-template MAP-FILE allows a greater deal of flexibility in how references can be pushed
to destination repositories. Hermetic text Sprig functions are currently supported which allows for matching by 
prefix, digest, media-type, regex, etc.  The following additional functions are provides

Tag - Returns the tag of an OCI string
Repository - Returns the repository of an OCI string
Registry - Returns the registry of an OCI string
Package - Returns omits the registry from the OCI reference

Example csv and go template files are located in the pkg/actions/mirror/test repository.
		`,
		Example: `To put all the images nested under "reg.other.com/mirror" you can use
ace-dt mirror scatter reg.example.com/repo/data:sync-45 nest=ref.other.com/mirror

ace-dt mirror scatter reg.example.com/repo/data:sync-45 go-template=mapping.tmpl
ace-dt mirror scatter reg.example.com/repo/data:sync-45 first-prefix=mapping.csv
ace-dt mirror scatter reg.example.com/repo/data:sync-45 digests=mapping.csv
ace-dt mirror scatter reg.example.com/repo/data:sync-45 longest-prefix=mapping.csv
ace-dt mirror scatter reg.example.com/repo/data:sync-45 all-prefix=mapping.csv

To scatter by filtering on manifest labels, you can use
ace-dt mirror scatter reg.example.com/repo/data:sync-45 nest=ref.other.com/mirror --filter-labels=component=core,module=test
`,

		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, args[0], args[1])
			})
		},
	}
	cmd.PersistentFlags().BoolVar(&action.Check, "check", false, "Dry run- do not actually send to destination repositories")
	cmd.PersistentFlags().StringVar(&action.SourceFile, "subset", "", "Define a subset list of images to scatter with a sources.list file")
	cmd.PersistentFlags().StringSliceVarP(&action.Selectors, "selector", "l", []string{}, "Only scatter manifests tagged with annotation labels, e.g., component=core,module=test")
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	return cmd
}
