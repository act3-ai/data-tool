package bottle

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/act3-ai/go-common/pkg/logger"

	sigcustom "github.com/act3-ai/data-tool/internal/sign"
	"github.com/act3-ai/data-tool/internal/ui"
	"github.com/act3-ai/data-tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// Sign represents a manifest digest sign action.
type Sign struct {
	*Action

	// DigestAlg    string
	KeyPath      string
	KeyAPI       string
	UserIdentity string
	KeyID        string

	NoDeprecate bool // Don't deprecate existing bottle when committing.
}

// Run runs the bottle verify action.
func (action *Sign) Run(ctx context.Context, keyAlias string) error {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	log.InfoContext(ctx, "Bottle sign command activated")

	// ensure bottle is ready to be signed.
	log.InfoContext(ctx, "Prepping Bottle for signing", "bottlePath", action.Dir)
	cfg, bottle, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	// commit creates a bottle manifest handler
	if err := commit(ctx, cfg, bottle, action.NoDeprecate); err != nil {
		return err
	}

	bottleManifestDescriptor := bottle.Manifest.GetManifestDescriptor()
	sigPath := filepath.Join(action.Dir, ".signature")
	unsignedAnnos := make(map[string]string)
	var sigHandler sigcustom.SigsHandler
	var foundKey v1alpha1.SigningKey

	// search for signing key in config.
	for _, key := range cfg.SigningKeys {
		if key.Alias == keyAlias {
			foundKey = key
			log.InfoContext(ctx, "key found", "alias", keyAlias)
			break
		}
	}
	// fail if key is not found or if incomplete metadata for key.
	if foundKey.Alias == "" {
		return fmt.Errorf("private key not found in configuration")
	} else if foundKey.KeyPath == "" || foundKey.KeyAPI == "" || foundKey.UserIdentity == "" || foundKey.KeyID == "" {
		return fmt.Errorf("private key metadata incomplete, please check ace-dt config: Alias: %s, KeyPath = %s, KeyAPI = %s, UserIdentity = %s, KeyID = %s", foundKey.Alias, foundKey.KeyPath, foundKey.KeyAPI, foundKey.UserIdentity, foundKey.KeyID)
	}

	// construct the map for the annotations to be signed.
	unsignedAnnos[sigcustom.AnnotationUserID] = foundKey.UserIdentity
	unsignedAnnos[sigcustom.AnnotationVerifyAPI] = foundKey.KeyAPI
	unsignedAnnos[sigcustom.AnnotationKeyID] = foundKey.KeyID

	// Load a Notary style signature handler, which expects each signature to have its own manifest and signature
	// blob
	sigHandler, err = sigcustom.LoadLocalSignatures(ctx, bottleManifestDescriptor, sigPath)
	if err != nil {
		return err
	}

	// HACK: We should come up with a more robust method for specifying cert files.
	name := filepath.Base(foundKey.KeyPath)
	i := strings.LastIndex(name, ".")
	certPath := filepath.Join(filepath.Dir(foundKey.KeyPath), name[:i]+".crt")

	// Create a file based private key provider. Note the key is not loaded until the key is used during signing.
	log.InfoContext(ctx, "Constructing private key provider", "privateKeyPath", foundKey.KeyPath)
	pkProvider := sigcustom.NewFilePrivateKeyProvider(foundKey.KeyPath, certPath)

	// sign the bottle with the annotations.
	log.InfoContext(ctx, "Beginning signing process")
	err = sigHandler.Sign(ctx, pkProvider, unsignedAnnos, nil)
	if err != nil {
		rootUI.Infof("Unable to sign Bottle %s", action.Dir)
		return fmt.Errorf("ace-dt bottle sign: %w", err)
	}

	// signature created.
	rootUI.Infof("Bottle %s successfully signed.", action.Dir)
	return nil
}
