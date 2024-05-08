package cli

import (
	_ "embed"
	"os"

	"github.com/spf13/cobra"
)

// newFilterCmd creates a new logging filter subcommand.
func newFilterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Print a filter to use when pretty printing logs",
	}

	cmd.AddCommand(
		newJqCmd(),
		newYqCmd(),
	)

	return cmd
}

//go:embed log.jq
var jqFilter string

func newJqCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jq",
		Short: "Print a jq filter to use when pretty printing logs",
		Long: `To use this filter you must have "jq" installed.
		ace-dt util filter jq > log.jq
	ace-dt ... | jq -j -f log.jq`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(jqFilter)
			return nil
		},
	}
	cmd.SetOut(os.Stdout)

	return cmd
}

//go:embed log.yq
var yqFilter string

func newYqCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yq",
		Short: "Print a yq filter to use when pretty printing logs",
		Long: `To use this filter you must have "yq" installed.
	ace-dt util filter yq > log.yq
	ace-dt ... | yq -p=json --from-file=log.yq`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(yqFilter)
			return nil
		},
	}
	cmd.SetOut(os.Stdout)

	return cmd
}
