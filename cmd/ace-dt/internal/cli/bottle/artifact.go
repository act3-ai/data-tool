package bottle

import (
	actions "github.com/act3-ai/data-tool/internal/actions/bottle"

	"github.com/spf13/cobra"
)

// newBtlPartCmd is the top level command that aggregates subcommands for interacting with bottles parts fields.
func newBtlArtifactCmd(tool *actions.Action) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "metadata",
		Use:     "artifact",
		Short:   "Bottle artifacts operations",
	}

	cmd.AddCommand(
		setArtifactCmd(tool),
		removeArtifactCmd(tool),
		listArtifactCmd(tool),
	)
	return cmd
}

func setArtifactCmd(tool *actions.Action) *cobra.Command {
	action := &actions.ArtifactSet{Action: tool}

	setArtifactCmd := &cobra.Command{
		Use:   "set [NAME] [PATH]",
		Short: "Sets a file as public artifact of bottle",
		Long: `Sets a file as a public artifact of specified bottle.

A public artifact is comprised of four fields: 'media type', 'digest', 'name', 'path'. 
'ace-dt' computes media-type and digest when a public artifact is added.
Name and path are required and must be set by the user. Relative paths are allowed.
  
This command only supports adding artifacts that are already
committed, and listed as a part of the bottle. 
- To see parts listed in the bottle, run 'ace-dt bottle part list' command
- To add a file to a bottle, see the 'bottle commit' command
`,
		Example: `
Set public artifact at the path <./detect-food.csv> in current working directory:
	ace-dt bottle artifact set "food detection" ./detect-food.csv

Set public artifact <inference example> at the path <my/bottle/path/infer.py>:
	ace-dt bottle artifact set --bottle-dir my/bottle "inference usage" my/bottle/path/infer.py

Set public artifact <food model> at the path <./food.model> with its media type <application/octet-stream>:
	ace-dt bottle artifact set "food model" ./food.model --media-type "application/octet-stream"
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], args[1], cmd.OutOrStdout())
		},
	}

	setArtifactCmd.Flags().StringVarP(&action.MediaType, "media-type", "m", "", "specify artifact's media type")

	return setArtifactCmd
}

func listArtifactCmd(tool *actions.Action) *cobra.Command {
	action := &actions.ArtifactList{Action: tool}

	listArtCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Lists public artifacts in a bottle",
		Example: `
List public artifacts of bottle in current working directory:
	ace-dt bottle artifact list

List public artifacts of bottle at the path my/bottle/path:
	ace-dt bottle artifact list -d "my/bottle/path/"
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	return listArtCmd
}

func removeArtifactCmd(tool *actions.Action) *cobra.Command {
	action := &actions.ArtifactRemove{Action: tool}

	removeArtCmd := &cobra.Command{
		Use:     "remove [PATH]",
		Aliases: []string{"rm"},
		Short:   "Removes item from public artifact list of a bottle",
		Long:    `Removes an item from public artifact list of a bottle using the file's path.`,
		Example: `
Remove artifact mnist_public.zip  nested in /dataset directory of current bottle:
	ace-dt bottle artifact remove dataset/mnist_public.zip
	
Remove artifact kaggle_data.csv from bottle at path my/bottle/path:
	ace-dt bottle artifact rm kaggle_data.csv -d my/bottle/path
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], cmd.OutOrStdout())
		},
	}

	return removeArtCmd
}
