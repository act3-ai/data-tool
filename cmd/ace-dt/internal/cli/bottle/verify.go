package bottle

import (
	"context"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"
)

// verifyCmd represents the verify command.
func newVerifyCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Verify{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		GroupID: "basic",
		Use:     "verify",
		Short:   "Verifies all local signatures of a bottle's manifest digest.",
		Long: `Verifies all local signatures of the bottle's manifest digest. In order to ensure the local signatures are up-to-date use ace-dt bottle pull prior to signature verification.

It is optional to provide locally discovered public keys for verification. For each local public key provide the path to the key. The key's fingerprint will be used to correlate keys with signatures.
Ex: ace-dt bottle verify <path> <path> ...

Notice:
  If the signer provided insufficient metadata to discover the appropriate public key for verification,
  it will default to the no key management system verification method - which is notably insecure with ECDSA keys.


`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				return action.Run(ctx, args...)
			})
		},
	}

	// TODO: Do we need ui options?
	// ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	// cmd.Flags().StringVar(&action.DigestAlg, "digest-algorithm", "sha256", "Specify the algorithm to use when calculating digests.")
	// cmd.Flags().StringVar(&action.KeyPath, "public-key", "", "Verify with a local public keys identified by path and user identity. --public-key <path> <userID> <path> <userID> ...")

	cmd.Example = `
To verify a manifest digest:
	ace-dt bottle verify
`
	return cmd
}
