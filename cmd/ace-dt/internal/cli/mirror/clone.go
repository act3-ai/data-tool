package mirror

import (
	"context"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "git.act3-ace.com/ace/data/tool/internal/actions/mirror"
)

// newGatherCmd represents the mirror gather command.
func newCloneCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Clone{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		Use:   "clone SOURCES-FILE MAPPER",
		Short: "A command that copies images listed in SOURCES-FILE according to the mapper.",
		Long: `A command that copies images listed in SOURCES-FILE according to the mapper.
		
SOURCES-FILE is a text file with one OCI image reference per line.  Lines that begin with # are ignored.
Labels can be added to each source in the SOURCES-FILE by separating with a comma and following a key=value format. These will be added as annotations to that manifest:
reg.example.com/library/source1,component=core,module=test

The MAPPER types currently supported are nest, first-prefix (csv format), digests (csv format) and go-template.
The format of MAPPER is MAP-TYPE=MAP-ARG

If MAP-TYPE is "nest" then clone will nest all the images under MAP-ARG.
For example, is MAP-ARG is "reg.other.com" then a gathered image "foo.com/bar" will map to "reg.other.com/foo.com/bar".

Passing a first-prefix MAPPER requires a csv file that has formatted lines of: source,destination. 
The ace-dt mirror clone will send the source reference to the first prefix match that it makes.
This format also allows defining the source as a digest that is present in the source repository.

Passing a digests MAP-FILE requires a csv file that has formatted lines of: digest-string, destination.
Scatter will send each digest to the locations defined in the map file provided. 

Passing a go-template MAP-FILE allows a greater deal of flexibility in how references can be pushed
to destination repositories. Sprig functions are currently supported which allows for matching by 
prefix, digest, media-type, regex, etc. 

Example csv and go template files are located in the pkg/actions/mirror/test repository.
		`,
		Example: `To clone and scatter all the images contained in "sources.list" you can use
ace-dt mirror clone sources.list nest=ref.other.com/mirror

ace-dt mirror clone sources.list go-template=mapping.tmpl
ace-dt mirror clone sources.list first-prefix=mapping.csv
ace-dt mirror clone sources.list digests=mapping.csv
ace-dt mirror clone sources.list longest-prefix=mapping.csv
ace-dt mirror clone sources.list all-prefix=mapping.csv`,

		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, args[0], args[1])
			})
		},
	}

	cmd.PersistentFlags().StringSliceVarP(&action.Selectors, "selector", "l", []string{}, "Only scatter manifests tagged with annotation labels, e.g., component=core,module=test")
	cmd.PersistentFlags().BoolVar(&action.Check, "check", false, "Dry run- do not actually send to destination repositories")
	cmd.Flags().StringSliceVarP(&action.Platforms, "platforms", "p", []string{}, "Only gather images that match the specified platform(s). Warning: This will modify the manifest digest/reference..")
	ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	return cmd
}
