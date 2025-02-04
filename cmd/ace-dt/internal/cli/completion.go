package cli

import (
	"bytes"
	"fmt"
	"strings"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"github.com/spf13/cobra"
)

const completionLong = `
To load completions:

Bash:

$ source <(ace-dt completion bash)

# To load completions for each session, execute once:
Linux:
$ ace-dt completion bash > /etc/bash_completion.d/ace-dt
MacOS:
$ ace-dt completion bash > /usr/local/etc/bash_completion.d/ace-dt

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ ace-dt completion zsh > "${fpath[1]}/_ace-dt"

# You will need to start a new shell for this setup to take effect.

Fish:

$ ace-dt completion fish | source

# To load completions for each session, execute once:
$ ace-dt completion fish > ~/.config/fish/completions/ace-dt.fish
`

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		GroupID:               "setup",
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate completion script",
		Long:                  completionLong,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := logger.FromContext(ctx)

			logger.V(log, 1).InfoContext(ctx, "completion command activated")
			var err error
			switch args[0] {
			case "bash":
				// HACK to remove ":" from the BASH stopwords
				var buf bytes.Buffer
				err = cmd.Root().GenBashCompletion(&buf)
				if err != nil {
					return fmt.Errorf("error generating bash completion: %w", err)
				}
				str := buf.String()
				str = strings.Replace(str, "_init_completion -s || return", `_init_completion -n ":" -s || return`, 1)
				str = strings.Replace(str, `__ace-dt_init_completion -n "=" || return`, `__ace-dt_init_completion -n "=:" || return`, 1)
				_, err = fmt.Fprint(cmd.OutOrStdout(), str)
			case "zsh":
				err = cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				err = cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				err = cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
			}
			logger.V(log, 1).InfoContext(ctx, "completion command completed")
			if err != nil {
				return fmt.Errorf("error generating completion: %w", err)
			}
			return nil
		},
	}
	return cmd
}
