// Package main is the entrypoint for ace-dt, which provides a set of tools that facilitate data set transfer to and
// from OCI registry storage, focused on OCI objects called bottles, but additionally working with general OCI objects.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/cli"
	"git.act3-ace.com/ace/data/tool/docs"
	"git.act3-ace.com/ace/go-common/pkg/cmd"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"git.act3-ace.com/ace/go-common/pkg/runner"
	vv "git.act3-ace.com/ace/go-common/pkg/version"
)

// getVersionInfo retrieves the proper version information for this executable.
func getVersionInfo() vv.Info {
	info := vv.Get()
	if version != "" {
		info.Version = version
	}
	return info
}

// main launches the root CLI handler, implemented with cobra.
func main() {
	info := getVersionInfo()
	root := cli.NewToolCmd(info.Version)
	root.SilenceUsage = true // silence usage when directly called

	handler := runner.SetupLoggingHandler(root, "ACE_DT_VERBOSITY") // create log handler
	l := slog.New(handler)
	ctx := logger.NewContext(context.Background(), l)
	root.SetContext(ctx) // inject context for data commands

	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		log := logger.FromContext(cmd.Context()) //nolint:contextcheck
		log.InfoContext(ctx, "Software", "version", info.Version)
		log.InfoContext(ctx, "Software details", "info", info)
	}

	// Add embedded documentation commands
	embeddedDocs := docs.Embedded(root)
	root.AddCommand(
		cmd.NewVersionCmd(info),
		cmd.NewInfoCmd(embeddedDocs),
		cmd.NewGendocsCmd(embeddedDocs),
		cmd.NewGenschemaCmd(docs.Schemas(), docs.SchemaAssociations),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
