package sign

import (
	"crypto"
	"crypto/x509"
	"fmt"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
)

// Keys for OCI layer annotations.
const (
	AnnotationKeyID     = "keyID"      // AnnotationKeyID is the key identifier. A layer annotation key.
	AnnotationUserID    = "identity"   // AnnotationUserID is the key owner's identity. A layer annotation key.
	AnnotationVerifyAPI = "verify-api" // AnnotationVerifyAPI is the api to use when verifying. A layer annotation key.
	// AnnotationKeyHost   = "key-host"   // AnnotationKeyHost is host details for retrieving public keys. A layer annotation key.
)

// KeyRetriever provide a means of fetching public keys from a remote location and defaults to using the attached
// public key or certificate.
// WARNING: An attached key / certificate does not automatically link a signer identity with signature, without which
// a signature does not provide a guarantee of authenticity or security.
type KeyRetriever interface {
	RetrieveVerifier(crypto.Hash) (Verifier, error)
}

// certBasicKeyRetriever is used for certificate based signing or verification.
type certBasicKeyRetriever struct {
	cert *x509.Certificate
}

// RetrieveVerifier returns a certificate based verifier containing the public key and cert used for signature verification
// from the signed and unsigned payload annotations.
//
// WARNING: This verifier does not validate an associated identity with the certificate, or pin the certificate to a
// known certificate authority, and thus does not provide cryptographic assurance or security.
func (cbkr *certBasicKeyRetriever) RetrieveVerifier(hashFunc crypto.Hash) (Verifier, error) {
	return &simpleCertVerifier{
		certs:    []*x509.Certificate{cbkr.cert},
		hashFunc: hashFunc,
	}, nil
}

// NewCertRetriever constructs a certificate based key retriever for a jws or cose envelope.
func NewCertRetriever(mediatype string, payload []byte) (KeyRetriever, error) {
	var env signature.Envelope
	var err error
	switch mediatype {
	case jws.MediaTypeEnvelope:
		env, err = jws.ParseEnvelope(payload)
		if err != nil {
			return nil, fmt.Errorf("parsing jws signature envelope: %w", err)
		}
	case cose.MediaTypeEnvelope:
		env, err = cose.ParseEnvelope(payload)
		if err != nil {
			return nil, fmt.Errorf("parsing cose signature envelope: %w", err)
		}
	default:
		return nil, &signature.UnsupportedSignatureFormatError{MediaType: mediatype}
	}
	if err != nil {
		return nil, fmt.Errorf("parsing signature envelope: %w", err)
	}

	content, err := env.Content()
	if err != nil {
		return nil, fmt.Errorf("extracting '%s' envelope content: %w", mediatype, err)
	}

	rootCA := content.SignerInfo.CertificateChain[len(content.SignerInfo.CertificateChain)-1]

	return &certBasicKeyRetriever{
		cert: rootCA,
	}, nil
}
