package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/mirror/multiplex"
)

const muxExample = `
Interleave file1, file2, and file3 onto the same stream and separate them back out.  The file1 and file1-dup will be identical.  Likewise for the other files.
	ace-dt util mux file1 file2 file3 | ace-dt util demux file1-dup file2-dup file3-dup

This can be combined with named pipes or unnamed pipes (requires bash to use process substitution):
	ace-dt util mux <(command to stream from A) <(command to stream from B) > f
	ace-dt util demux >(command to stream to A) >(command to stream to B) < f
		
Can be combined with the mbuffer subcommand:
	ace-dt util mux <(command to stream from A) <(command to stream from B) | ace-dt util mbuffer -n 6Gi > /dev/nst0
`

// newMuxCmd creates a new multiplexer subcommand.
func newMuxCmd() *cobra.Command {
	var bsStr string
	cmd := &cobra.Command{
		Use:     "mux [in1 in2 ...] ",
		Args:    cobra.MinimumNArgs(1),
		Short:   "Multiplex via interleaving, the inputs into stdout in such a way that they can be recovered with demux",
		Example: muxExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			// parse block size
			bs, err := humanize.ParseBytes(bsStr)
			if err != nil {
				return fmt.Errorf("parsing block size: %w", err)
			}

			ins := make([]io.Reader, len(args))
			for i, inFile := range args {
				ins[i], err = os.Open(inFile)
				if err != nil {
					return fmt.Errorf("mux input file: %w", err)
				}
			}

			return multiplex.Mux(int(bs), os.Stdout, ins...)
		},
	}

	cmd.Flags().StringVarP(&bsStr, "block-size", "b", "32kiB", "Block size (may provide Si suffixes) is the maximum size of each read")

	return cmd
}

// newDemuxCmd creates a new demultiplexer subcommand.
func newDemuxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "demux [out1 out2 ...] ",
		Args:    cobra.MinimumNArgs(1),
		Short:   "Demultiplex stdin into the files specified",
		Long:    `The data will be appended to the files given and created if they do not exist.`,
		Example: muxExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			outs := make([]io.Writer, len(args))
			for i, outFile := range args {
				outs[i], err = os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
				if err != nil {
					return fmt.Errorf("demux output file: %w", err)
				}
			}

			return multiplex.Demux(os.Stdin, outs...)
		},
	}

	return cmd
}
