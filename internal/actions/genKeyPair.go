package actions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	sigcustom "github.com/act3-ai/data-tool/internal/sign"
	"github.com/act3-ai/data-tool/internal/ui"
	"github.com/act3-ai/go-common/pkg/logger"
)

// GenKeyPair represents the generate key pair action.
type GenKeyPair struct {
	*DataTool

	Prefix   string // file name prefix to add to key names
	MakeCert bool   // true to generate a self-signed x509 certificate as well
}

// Run runs the gen-key-pair action.
func (action *GenKeyPair) Run(ctx context.Context, destPath, prefix string) error {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	log.InfoContext(ctx, "Util Generate Key Pair command activated")

	err := GenAndWriteKeyPair(ctx, destPath, prefix, action.MakeCert)
	if err != nil {
		return fmt.Errorf("ace-dt gen-key-pair: %w", err)
	}

	rootUI.Infof("Private Key written to %s, ", filepath.Join(destPath, prefix+".key"))
	rootUI.Infof("Public Key written to %s, ", filepath.Join(destPath, prefix+".pub"))

	return nil
}

// GenAndWriteKeyPair creates a public private ecdsa key pair, writing them to the specified destination. The
// public and private key files are named <prefix>.pub and <prefix>.key respectively.
func GenAndWriteKeyPair(ctx context.Context, destPath, prefix string, wantCert bool) error {
	log := logger.V(logger.FromContext(ctx), 1)

	// check destination path provided on command line.
	if destPath == "" {
		return errors.New("no destination path provided")
	}

	// set key paths.
	privPath := filepath.Join(destPath, prefix+".key")
	pubPath := filepath.Join(destPath, prefix+".pub")

	log.InfoContext(ctx, "Generating key pair")
	keypair, err := sigcustom.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("error generating key pair: %w", err)
	}

	// Get the private key in PEM format.
	privPEM, err := keypair.PrivateKeyPEM()
	if err != nil {
		return fmt.Errorf("error PEM formatting private key: %w", err)
	}

	// Get the public key in PEM format.
	pubPEM, err := keypair.PublicKeyPEM()
	if err != nil {
		return fmt.Errorf("error PEM formatting public key: %w", err)
	}

	// write the private key to the destination path.
	log.InfoContext(ctx, "Writing key pair to destination path", "destPath", destPath)
	err = os.WriteFile(privPath, privPEM, 0o600)
	if err != nil {
		return fmt.Errorf("writing private key to PEM file: %w", err)
	}

	// write the public key to the destination path
	err = os.WriteFile(pubPath, pubPEM, 0o600)
	if err != nil {
		return fmt.Errorf("writing public key to PEM file: %w", err)
	}

	// if the user is also generating a certificate
	if wantCert {
		log.InfoContext(ctx, "generating x509 certificate for private key")
		certPair := sigcustom.MakeEcdsaCertPair("ace-dt self signed signing cert", keypair, nil)
		cert := certPair.Cert.Raw

		// write the certificate to the destination path
		certPath := filepath.Join(destPath, prefix+".crt")
		err = os.WriteFile(certPath, cert, 0o600)
		if err != nil {
			return fmt.Errorf("writing certificate to ASN.1 DER file: %w", err)
		}
	}

	return nil
}
