package sign

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	mrand "math/rand"
	"os"
	"path/filepath"
	"syscall"
	"time"

	notationx509 "github.com/notaryproject/notation-core-go/x509"
	"github.com/opencontainers/go-digest"

	"golang.org/x/term"
)

const privateKeyType = "EC PRIVATE KEY"
const publicKeyType = "PUBLIC KEY"

// PrivateKeyProvider provide methods for accessing a private key and its corresponding public key.
type PrivateKeyProvider interface {
	// PrivateKey returns a pointer to an ecdsa private key for signing. NOTE: consider opts for future expansion.
	PrivateKey( /*opts ...PrivateKeyOption*/ ) (crypto.PrivateKey, error)
	// PublicKeyPEM returns a PEM public key based on a private key. This can load the private key if necessary.
	PublicKeyPEM() ([]byte, error)
	// PrivateKeyPEM returns a private key based on a private key. This can load the private key if necessary.
	PrivateKeyPEM() ([]byte, error)
	// Certificate returns a x509 certificate associated with a private key.
	Certificate() (*x509.Certificate, error)
}

// filePrivateKeyProvider provides a path to a private key and/or the ecdsa private key in its raw format.  If a .cert
// file with the same base name is found at the path, a x509 certificate is also loaded.
type filePrivateKeyProvider struct {
	privKeyPath string
	pKey        crypto.PrivateKey
	certPath    string
	cert        *x509.Certificate
}

// NewFilePrivateKeyProvider creates a private key provider with the given path.  Key loading is delayed until later
// optKey allows the private key to be generated elsewhere, in which case the path won't be used.
func NewFilePrivateKeyProvider(keyPath, certPath string) PrivateKeyProvider {
	return &filePrivateKeyProvider{
		privKeyPath: filepath.Clean(keyPath),
		pKey:        nil,
		certPath:    filepath.Clean(certPath),
		cert:        nil,
	}
}

// PrivateKeyPEM returns a PEM encoded private key.
func (pkf *filePrivateKeyProvider) PrivateKeyPEM() ([]byte, error) {
	pkey, err := pkf.PrivateKey()
	if err != nil {
		return nil, err
	}

	privPEM, err := rawPrivToPEM(pkey)
	if err != nil {
		return nil, err
	}

	return privPEM, nil
}

// PublicKeyPEM returns a PEM encoded public key derived from the private key.
func (pkf *filePrivateKeyProvider) PublicKeyPEM() ([]byte, error) {
	pkey, err := pkf.PrivateKey()
	if err != nil {
		return nil, err
	}

	ecdsaPriv, ok := pkey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key for retrieving public key PEM is not ecdsa")
	}
	pubPEM, err := rawPubToPEM(&ecdsaPriv.PublicKey)
	if err != nil {
		return nil, err
	}

	return pubPEM, nil
}

// PrivateKey returns the raw ecdsa private key, which may require loading from the
// keyPath if it is not already present in its raw format.
func (pkf *filePrivateKeyProvider) PrivateKey() (crypto.PrivateKey, error) {
	if pkf.pKey != nil {
		return pkf.pKey, nil
	}
	var err error
	pkf.pKey, err = notationx509.ReadPrivateKeyFile(pkf.privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("parsing private key from path %s: %w", pkf.privKeyPath, err)
	}
	return pkf.pKey, nil
}

// Certificate returns a x509 certificate associated with a private key, or nil if a cert file was not found with key.
func (pkf *filePrivateKeyProvider) Certificate() (*x509.Certificate, error) {
	if pkf.cert == nil {
		certs, err := notationx509.ReadCertificateFile(pkf.certPath)
		if err != nil {
			return nil, fmt.Errorf("parsing certificate from path %s: %w", pkf.certPath, err)
		}
		pkf.cert = certs[0] // cert chain should take the form: signer -> intermediates -> root
	}

	return pkf.cert, nil
}

// privateKeyProvider provides a private key.
type privateKeyProvider struct {
	pKey crypto.PrivateKey
}

// PrivateKeyPEM returns a PEM encoded private key.
func (pk *privateKeyProvider) PrivateKeyPEM() ([]byte, error) {
	pkey, err := pk.PrivateKey()
	if err != nil {
		return nil, err
	}

	privPEM, err := rawPrivToPEM(pkey)
	if err != nil {
		return nil, err
	}

	return privPEM, nil
}

// PublicKeyPEM returns a PEM encoded public key derived from the private key.
func (pk *privateKeyProvider) PublicKeyPEM() ([]byte, error) {
	pkey, err := pk.PrivateKey()
	if err != nil {
		return nil, err
	}
	ecdsaPriv, ok := pkey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key for retrieving public key PEM is not ecdsa")
	}
	pubPEM, err := rawPubToPEM(&ecdsaPriv.PublicKey)
	if err != nil {
		return nil, err
	}

	return pubPEM, nil
}

// PrivateKey returns the raw ecdsa private key.
func (pk *privateKeyProvider) PrivateKey() (crypto.PrivateKey, error) {
	if pk.pKey == nil {
		return nil, errors.New("privateKeyProvider does not contain a private key")
	}
	return pk.pKey, nil
}

// Certificate returns nil for a basic private key provider.
func (pk *privateKeyProvider) Certificate() (*x509.Certificate, error) {
	return nil, nil
}

// NewPrivateKeyProvider creates a simple private key provider with the raw private key.
func NewPrivateKeyProvider(key *ecdsa.PrivateKey) PrivateKeyProvider {
	return &privateKeyProvider{pKey: key}
}

// FingerprintECDSA returns the fingerprint of an ECDSA public key in PEM format.  This is a sha256 digest of the data,
// and can be used to compare keys.
func FingerprintECDSA(rawKey crypto.PublicKey) (digest.Digest, error) {
	pemBytes, err := rawPubToPEM(rawKey)
	if err != nil {
		return "", err
	}
	return FingerprintPEM(pemBytes), nil
}

// FingerprintPEM returns a fingerprint of a key in PEM format. Similar to FingerprintECDSA, but does not convert a key
// to pem format first.
func FingerprintPEM(pemBytes []byte) digest.Digest {
	return digest.FromBytes(pemBytes)
}

// RawPubToPEM converts a raw ecdsa public key into PEM encoded byte format, returned as a string.
func rawPubToPEM(rawPub crypto.PublicKey) ([]byte, error) {

	// public key to DER form
	pubDER, err := x509.MarshalPKIXPublicKey(rawPub)
	if err != nil {
		return nil, fmt.Errorf("marshaling public key to DER: %w", err)
	}

	// public DER to PEM encode
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  publicKeyType,
		Bytes: pubDER,
	})

	return pubPEM, nil
}

// rawPrivToPEM converts a raw ecdsa private key into PEM encoded byte format, returned as a string.
func rawPrivToPEM(rawPriv crypto.PrivateKey) ([]byte, error) {
	// private key to DER form
	ecdsaPriv, ok := rawPriv.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key for conversion to PEM is not ecdsa")
	}
	privDER, err := x509.MarshalECPrivateKey(ecdsaPriv)
	if err != nil {
		return nil, fmt.Errorf("marshaling private key: %w", err)
	}

	// private DER to PEM encode
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  privateKeyType,
		Bytes: privDER,
	})

	return privPEM, nil
}

// GenerateKeyPair generates a public private ECDSA key pair.
func GenerateKeyPair() (PrivateKeyProvider, error) {

	// generate the private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating private key: %w", err)
	}

	// Make a key provider with an empty path.
	return NewPrivateKeyProvider(privateKey), nil

}

// EcdsaCertPair is an association between a certificate and an ecdsa private key.
type EcdsaCertPair struct {
	Cert       *x509.Certificate
	PrivateKey *ecdsa.PrivateKey
}

// MakeEcdsaCertPair creates a new certificate and associates it with a private key.  If issuer is not nil, then the
// provided issuer certificate is a parent certificate, otherwise a root certificate is created.
func MakeEcdsaCertPair(cn string, privateKey PrivateKeyProvider, issuer *EcdsaCertPair) EcdsaCertPair {
	if privateKey == nil {
		privateKey, _ = GenerateKeyPair()
	}
	privKey, _ := privateKey.PrivateKey()
	ecdsaprivKey, ok := privKey.(*ecdsa.PrivateKey)
	if !ok {
		panic("private key is not ecdsa in MakeEcdsaCertPair")
	}
	template := makeCertTemplate(cn, issuer == nil)

	var certBytes []byte
	var err error
	if issuer != nil {
		certBytes, err = x509.CreateCertificate(rand.Reader, template, issuer.Cert, &issuer.PrivateKey.PublicKey, issuer.PrivateKey)
	} else {
		certBytes, err = x509.CreateCertificate(rand.Reader, template, template, &ecdsaprivKey.PublicKey, privKey)
	}
	if err != nil {
		panic(err)
	}

	cert, _ := x509.ParseCertificate(certBytes)
	return EcdsaCertPair{
		Cert:       cert,
		PrivateKey: ecdsaprivKey,
	}
}

func makeCertTemplate(cn string, isRoot bool) *x509.Certificate {
	template := &x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{"act3-ace"},
			Country:      []string{"US"},
			Province:     []string{"OH"},
			Locality:     []string{"Dayton"},
			CommonName:   cn,
		},
		NotBefore:   time.Now(),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
	}

	if isRoot {
		template.SerialNumber = big.NewInt(int64(mrand.Intn(200000000)))
		template.NotAfter = time.Now().AddDate(2, 0, 0)
		template.KeyUsage = x509.KeyUsageDigitalSignature
		template.BasicConstraintsValid = true
		template.IsCA = false
	} else {
		template.SerialNumber = big.NewInt(int64(mrand.Intn(200000000)))
		template.NotAfter = time.Now().AddDate(2, 0, 0)
	}

	return template
}

// GetPassFromTerm prompts the user to provide a password from the terminal.
func GetPassFromTerm(confirm bool) ([]byte, error) {
	_, err := fmt.Fprint(os.Stderr, "Enter password for private key: ")
	if err != nil {
		return nil, fmt.Errorf("prompting password: %w", err)
	}
	// Unnecessary convert of syscall.Stdin on *nix, but Windows is a uintptr
	pw1, err := term.ReadPassword(int(syscall.Stdin)) //nolint:unconvert
	if err != nil {
		return nil, fmt.Errorf("reading password from terminal: %w", err)
	}
	_, err = fmt.Fprintln(os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("printing line to stderr: %w", err)
	}
	if !confirm {
		return pw1, nil
	}
	_, err = fmt.Fprint(os.Stderr, "Enter password for private key again: ")
	if err != nil {
		return nil, fmt.Errorf("prompting password: %w", err)
	}
	// Unnecessary convert of syscall.Stdin on *nix, but Windows is a uintptr
	confirmpw, err := term.ReadPassword(int(syscall.Stdin)) //nolint:unconvert
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr)
		return nil, fmt.Errorf("failed reading password from terminal: %w", err)
	}

	if string(pw1) != string(confirmpw) {
		return nil, errors.New("passwords do not match")
	}
	return pw1, nil
}
