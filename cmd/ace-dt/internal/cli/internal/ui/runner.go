// Package ui provides UI helpers for the CLI.
package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
	"golang.org/x/term"

	"github.com/act3-ai/data-tool/internal/ui"
	"github.com/act3-ai/data-tool/internal/util"
)

// RunUI runs the command with a UI.
func RunUI(ctx context.Context, options Options, run func(ctx context.Context) error) error {
	g, ctx := errgroup.WithContext(ctx)

	// construct the UI
	var rootUI ui.UI

	switch {
	case options.quiet:
		rootUI = ui.NewSilentUI()
	case options.debugPath != "":
		// If the path is relative, make it absolute with the current working directory
		var err error
		options.debugPath, err = filepath.Abs(options.debugPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for debug output: %w", err)
		}
		logfile := filepath.Join(options.debugPath, "logs.txt")

		// create folder for debug output
		if err := util.CreatePathForFile(logfile); err != nil {
			return fmt.Errorf("failed to create debug folder given path %s, err: %w", options.debugPath, err)
		}

		out, err := os.Create(logfile)
		// create sub log file within output folder
		if err != nil {
			return fmt.Errorf("failed to create debug file: %w", err)
		}
		defer out.Close()
		rootUI = ui.NewDebugUI(out)
	default:
		out := os.Stdout

		// check if file descriptor associated with writer is a terminal
		if !options.disableTerminal && term.IsTerminal(int(out.Fd())) {
			rootUI = ui.NewComplexUI(out)
		} else {
			rootUI = ui.NewSimpleUI(out)
		}
	}

	rootTask := rootUI.Root(ctx)
	ctx = ui.NewContext(ctx, rootTask)

	// Run the UI
	g.Go(func() error {
		return rootUI.Run(ctx)
	})

	// Do the actual action (work)
	g.Go(func() error {
		defer rootUI.Shutdown()
		defer rootTask.Complete()
		return run(ctx)
	})

	return g.Wait()
}
