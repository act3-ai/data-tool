package bottle

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/ui"
	actions "git.act3-ace.com/ace/data/tool/internal/actions/bottle"
	"git.act3-ace.com/ace/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// signCmd represents the sign command.
func newSignCmd(tool *actions.Action) *cobra.Command {
	action := &actions.Sign{Action: tool}
	uiOptions := ui.Options{}

	cmd := &cobra.Command{
		GroupID: "basic",
		Use:     "sign PRIVATE_KEY_ALIAS",
		Short:   "Signs a bottle manifest digest with a private key.",
		Long: `Signs the bottle's manifest digest using a unique user-defined alias name for a private key whose metadata is outlined in the ace-dt config file. This metadata is signed alongside the manifest digest and is used during verification to locate the public key associated with the signing private key.

The signature layer, manifest, and config are written to the .signature directory within the bottle directory. The layer is named by its digest. The manifest is named by the hash algorithm used, the manifest digest itself, with a .sig suffix.

It is possible to directly sign a bottle manifest digest without adding private key metadata to ace-dt config. This use case requires using all --key-path, --key-api, --key-id, and --user-id flags. Supported key management system apis include gitlab.

Current supported api's include no-kms and gitlab.

Signing a bottle with altered data or metadata will automatically deprecate the previous version (bottleID) of this bottle. This can be disabled by passing the --no-deprecate flag.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunUI(cmd.Context(), uiOptions, func(ctx context.Context) error {
				if len(args) < 1 {
					return action.Run(ctx, "overridekey")
				}
				return action.Run(ctx, args[0])
			})
		},
	}

	// TODO: Do we need ui options?
	// ui.AddOptionsFlags(cmd.Flags(), &uiOptions)

	cmd.Flags().BoolVar(&action.NoDeprecate, "no-deprecate", false, "Disable deprecation of previous bottle version")
	// cmd.Flags().StringVar(&action.DigestAlg, "digest-algorithm", "sha256", "Algorithm to use when calculating digests.")
	cmd.Flags().StringVar(&action.KeyPath, "key-path", "", "Path to a PEM formatted private key.")
	cmd.Flags().StringVar(&action.KeyAPI, "key-api", "", "The API added to the signed annotations, which is used during verification to locate the corresponding public key. Current api's supported are no-kms and gitlab.")
	cmd.Flags().StringVar(&action.KeyID, "key-id", "", "The title/ID of the signing key.")
	cmd.Flags().StringVar(&action.UserIdentity, "user-id", "", "The key owner's identity, typically a username for the KeyAPI.")

	// Add flag overrides function to override config with flags.
	action.Config.AddConfigOverride(func(ctx context.Context, c *v1alpha1.Configuration) error {

		// override config if all key metadata is specified, otherwise fail if some but not all is specified.
		if action.KeyPath != "" && action.KeyAPI != "" && action.KeyID != "" && action.UserIdentity != "" {
			c.SigningKeys = []v1alpha1.SigningKey{
				{
					Alias:        "overridekey",
					KeyPath:      action.KeyPath,
					KeyAPI:       action.KeyAPI,
					UserIdentity: action.UserIdentity,
					KeyID:        action.KeyID},
			}
		} else if action.KeyPath != "" || action.KeyAPI != "" || action.KeyID != "" || action.UserIdentity != "" {
			return fmt.Errorf("insufficient signing key metadata, please ensure to specify all metadata with flags when directly providing a key: KeyPath = %s, KeyAPI = %s, UserIdentity = %s, KeyID = %s", action.KeyPath, action.KeyAPI, action.UserIdentity, action.KeyID)
		}

		return nil
	})

	cmd.Example = `
To sign a manifest digest with alias 'MyPrivateKey':
	ace-dt bottle sign MyPrivateKey

To sign a manifest digest by directly providing private key metadata:
	ace-dt bottle sign --key-path PATH/TO/PRIVATE.KEY --key-api KMS_API --key-id KEY_TITLE --user-id KEY_OWNER_ID
`
	return cmd
}
