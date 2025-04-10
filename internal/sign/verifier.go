package sign

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"io"

	"github.com/notaryproject/notation-go/verifier/truststore"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	telemsig "github.com/act3-ai/data-telemetry/v3/pkg/signature"
)

// PublicKeyProvider returns a PublicKey associated with a digital signature.
type PublicKeyProvider interface {
	GetPublicKey(opts ...PublicKeyOption) (crypto.PublicKey, error)
}

// TrustStoreProvider is a notary trust store provider, that can help locate trusted x509 certificates.
type TrustStoreProvider interface {
	GetX509TrustStore() truststore.X509TrustStore
}

// Verifier verifies the digital signature using a specified public key.
// VerifyOption (included RPCOpts) was removed from VerifySignature.
type Verifier interface {
	PublicKeyProvider
	TrustStoreProvider
	VerifySignature(signedSubject ocispec.Descriptor, signature io.Reader) error
}

// simpleCertVerifier is a simple certificate base signature verifier where the certificates are provided directly
// versus existing in a trust store.
type simpleCertVerifier struct {
	certs     []*x509.Certificate
	publicKey *ecdsa.PublicKey
	hashFunc  crypto.Hash
}

// GetPublicKey returns the public key that is used to verify signatures by
// this verifier. As this value is held in memory, all options provided in arguments
// to this method are ignored.
func (scv simpleCertVerifier) GetPublicKey(_ ...PublicKeyOption) (crypto.PublicKey, error) {
	return scv.publicKey, nil
}

// GetX509TrustStore for mcv returns itself, the required interface to return certificates is implemented directly.
func (scv simpleCertVerifier) GetX509TrustStore() truststore.X509TrustStore {
	return scv
}

// GetCertificates returns the stored certificate collection from a memoryCertVerifier.
func (scv simpleCertVerifier) GetCertificates(ctx context.Context, storeType truststore.Type, namedStore string) ([]*x509.Certificate, error) {
	return scv.certs, nil
}

// VerifySignature for a simpleCertVerifier uses notary verification to verify a signature using the stored certificates
// and public key.
func (scv simpleCertVerifier) VerifySignature(signedSubject ocispec.Descriptor, signature io.Reader) error {
	// Read payload envelope
	sig, err := io.ReadAll(signature)
	if err != nil {
		return fmt.Errorf("reading jws envelope envelope: %w", err)
	}

	// verify the signature.  Failure is returned as an error, so we don't need to check outcome directly
	_, err = telemsig.ValidateSignatureNotary(context.Background(), signedSubject, sig, scv)
	// if failed, return a signature verification error
	if err != nil {
		return fmt.Errorf("verifying notation signature: %w", err)
	}

	return nil
}
