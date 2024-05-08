package actions

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"path/filepath"
	"testing"

	sigcustom "git.act3-ace.com/ace/data/tool/internal/sign"
)

func Test_GenAndWriteKeyPair(t *testing.T) {
	ctx := context.Background()
	testingDir := t.TempDir()

	// generate a key pair
	err := GenAndWriteKeyPair(ctx, testingDir, "testing", false)
	if err != nil {
		t.Errorf("GenerateKeyPair() error = %v", err)
	}

	// ensure the key is parsable
	pkProvider := sigcustom.NewFilePrivateKeyProvider(filepath.Join(testingDir, "testing.key"), filepath.Join(testingDir, "testing.crt"))

	privKeyPEM, err := pkProvider.PrivateKeyPEM()
	if err != nil {
		t.Fatalf("loading private key err = %v", err)
	}

	privKeyDER, rest := pem.Decode(privKeyPEM)
	if privKeyDER == nil || privKeyDER.Type != "EC PRIVATE KEY" {
		t.Fatalf("decoding PEM key to DER format; privKeyStr = %s, type = %s, rest = %s", privKeyPEM, privKeyDER.Type, rest)
	}

	parsedPrivKey, err := x509.ParseECPrivateKey(privKeyDER.Bytes)
	if err != nil {
		t.Fatalf("parsing DER bytes err = %v", err)
	}

	generatedKey, err := pkProvider.PrivateKey()
	if err != nil {
		t.Fatalf("extracting raw priv key from pk provider error = %v", err)
	}

	ecdsakey, ok := generatedKey.(*ecdsa.PrivateKey)
	if !ok {
		t.Fatalf("generated key is not ecdsa type")
	}

	if !ecdsakey.Equal(parsedPrivKey) {
		t.Fatalf("generated and parsed keys do not match generated = %s parsed = %s", generatedKey, parsedPrivKey)
	}
}
