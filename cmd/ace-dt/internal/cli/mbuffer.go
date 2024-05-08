package cli

import (
	"os"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/internal/flag"
	"gitlab.com/act3-ai/asce/data/tool/internal/mirror/blockbuf"
)

// newMbufferCmd creates a new multiplexer subcommand.
func newMbufferCmd() *cobra.Command {
	mbufOpts := flag.MemoryBufferOptions{
		BlockSize: 1024 * 1024, // 1MiB is the default block size
	}

	cmd := &cobra.Command{
		Use:     "mbuffer",
		Short:   "",
		Example: `ace-dt util mbuffer -m 6Gi > /dev/nst0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			n, bs, hwm := mbufOpts.Options()
			return blockbuf.Copy(os.Stdout, os.Stdin, n, bs, hwm)
		},
	}

	flag.AddMemoryBufferFlags(cmd.Flags(), &mbufOpts)
	cobra.CheckErr(cmd.MarkFlagRequired("buffer-size"))

	return cmd
}

/*
Testing
head -c 10G </dev/urandom >../bigfile
cat ../bigfile | ace-dt util mbuffer -m 1Gi | pv -L 100m >../dest
top | grep ace-dt
*/
