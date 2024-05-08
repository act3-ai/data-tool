package cli

import (
	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
)

// verifyCmd represents the verify command.
func newGenKeyCmd(tool *actions.DataTool) *cobra.Command {
	action := &actions.GenKeyPair{DataTool: tool}

	cmd := &cobra.Command{
		// GroupID: "keys",
		Use:     "gen-key-pair DESTINATION_PATH",
		Aliases: []string{"keygen", "genkeys"},
		Short:   "generates a key pair used for signing/verifying data bottles, writing them to the destination path.",
		Long:    `Generates an ECDSA public-private key pair, which is used for signing and verifying signatures of manifest digests. The public key is written to DESTINATION_PATH/bottle.pub while the private key is written to DESTINATION_PATH/bottle.key. The prefix "bottle" may be optionally changed with the --prefix flag. Any existing key names will be overwritten with the new key pair.`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args[0], action.Prefix)
		},
	}

	cmd.Flags().StringVarP(&action.Prefix, "prefix", "p", "bottle", "Set the prefix of the key names. Default is 'bottle'.")
	cmd.Flags().BoolVarP(&action.MakeCert, "cert", "c", false, "Set to true to generate a self signed certificate tied to the keys.")

	cmd.Example = `
	To generate keys with default naming (bottle.key, bottle.pub):
gen-key-pair DESTINATION_PATH
	To generate keys with custom naming (<prefix>.key, <prefix>.pub):
gen-key-pair DESTINATION_PATH --prefix PREFIX
	To generate a self signed certificate keys with custom naming (<prefix>.key, <prefix>.pub, <prefix>.crt):
gen-key-pair DESTINATION_PATH --cert --prefix PREFIX
`
	return cmd
}
