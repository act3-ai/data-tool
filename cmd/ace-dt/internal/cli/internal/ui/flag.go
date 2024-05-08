package ui

import (
	"github.com/spf13/pflag"
)

// Options for the CLI UI.
type Options struct {
	quiet           bool   // suppress output
	disableTerminal bool   // disables the UI features that use the terminal (redrawing lines)
	debugPath       string // file created for UI debug output
}

// AddOptionsFlags adds options flags to the flagset.
func AddOptionsFlags(flags *pflag.FlagSet, options *Options) {
	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.")
	flags.BoolVar(&options.disableTerminal, "no-term", false, "Disable terminal support for fancy printing")
	flags.StringVar(&options.debugPath, "debug", "", "Puts UI into debug mode, dumping all UI events to the given path.")
}
