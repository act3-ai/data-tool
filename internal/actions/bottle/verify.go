package bottle

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/notaryproject/notation-core-go/signature"

	"git.act3-ace.com/ace/data/tool/internal/actions"
	sigcustom "git.act3-ace.com/ace/data/tool/internal/sign"
	"git.act3-ace.com/ace/data/tool/internal/ui"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// Verify is the action structure.
type Verify struct {
	*Action

	// DigestAlg   string
	Write     WriteBottleOptions
	Telemetry actions.TelemetryOptions
}

// Run runs the bottle verify action.
func (action *Verify) Run(ctx context.Context, args ...string) error {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	log.InfoContext(ctx, "Bottle verify command activated")

	// prepare bottle and it's manifest for verification
	log.InfoContext(ctx, "Prepping Bottle for signing", "bottlePath", action.Dir)
	_, bottle, err := action.prepare(ctx)
	if err != nil {
		return err
	}
	err = bottle.ConstructManifest()
	if err != nil {
		return fmt.Errorf("constructing bottle manifest handler: %w", err)
	}

	botManifestDesc := bottle.Manifest.GetManifestDescriptor()
	sigPath := filepath.Join(action.Dir, ".signature")

	// Load a Notary style signature handler, which expects each signature to have its own manifest and signature
	// blob
	sigsHandler, err := sigcustom.LoadLocalSignatures(ctx, botManifestDesc, sigPath)
	if err != nil {
		return err
	}

	// verify the bottle.
	log.InfoContext(ctx, "Beginning verification process", "signedManifestDigest", botManifestDesc.Digest)
	pass, err := sigsHandler.Verify(ctx)
	switch {
	case errors.Is(err, signature.SignatureNotFoundError{}):
		// exit gracefully if no signatures are found
		rootUI.Infof("No signatures found.")
		return nil
	case pass:
		rootUI.Infof("Bottle passed verification.")
		rootUI.Infof("Warning: Current verification methods only support integrity verification.")
	case !pass:
		rootUI.Infof("Bottle failed verification.")
	}

	for _, fp := range sigsHandler.VerifiedSignatures() {
		rootUI.Infof("    Verified Signature Digest: %s", fp)
	}
	for _, fp := range sigsHandler.FailedSignatures() {
		rootUI.Infof("    Failed Signature Digest: %s", fp)
	}
	if err != nil {
		return fmt.Errorf("ace-dt bottle verify: %w", err) // verification failed.
	}
	return nil
}
