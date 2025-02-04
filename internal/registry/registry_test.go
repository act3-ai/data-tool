package registry

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	mrand "math/rand"
	"net"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/credentials"

	regcache "gitlab.com/act3-ai/asce/data/tool/internal/registry/cache"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/test"
)

func TestRegistryConfig(t *testing.T) {
	log := test.Logger(t, 0)
	ctx := logger.NewContext(context.Background(), log)

	rne := require.New(t).NoError

	// Set up a fake registry
	s := httptest.NewServer(registry.New())
	tlss := httptest.NewUnstartedServer(registry.New())
	defer s.Close()
	defer tlss.Close()
	u, err := url.Parse(s.URL)
	rne(err)

	// create the temp directory and tls folder
	tmpdir := t.TempDir()
	tlsPath := filepath.Join(tmpdir, "testing")
	err = os.Mkdir(tlsPath, 0o777)
	rne(err)
	defer os.RemoveAll(tmpdir)
	endpoint := "http://" + u.Host

	serverCert, err := writeServerAndCACerts(tlsPath)
	rne(err)
	tlss.TLS = &tls.Config{
		Certificates: []tls.Certificate{*serverCert},
	}
	tlss.StartTLS()
	secureURL, err := url.Parse(tlss.URL)
	rne(err)

	t.Run("TestHTTPClient", func(t *testing.T) {
		client, err := newHTTPClientWithOps(nil, u.Host, "")
		assert.NoError(t, err)
		assert.NotNil(t, client.Transport)
	})

	t.Run("TestCreateRepository", func(t *testing.T) {
		regConfig := v1alpha1.RegistryConfig{
			Configs: map[string]v1alpha1.Registry{
				u.Host: {
					Endpoints: []string{endpoint},
				},
			},
		}
		rCache := regcache.NewRegistryCache()
		_, err := CreateRepoWithCustomConfig(ctx, &regConfig, u.Host+"/img1", rCache, "test-agent", credentials.NewMemoryStore())
		rne(err)
	})

	t.Run("TestCreateRepositoryWithCustomConfig", func(t *testing.T) {
		regConfig := v1alpha1.RegistryConfig{
			Configs: map[string]v1alpha1.Registry{
				u.Host: {
					Endpoints: []string{"http://" + u.Host},
					RewritePull: map[string]string{
						"img1": "ace/dt/img1",
					},
				},
			},
			EndpointConfig: map[string]v1alpha1.EndpointConfig{
				endpoint: {
					TLS: &v1alpha1.TLS{
						InsecureSkipVerify: true,
					},
					ReferrersType: "tag",
				},
			},
		}
		regCache := regcache.NewRegistryCache()
		r, err := CreateRepoWithCustomConfig(ctx, &regConfig, u.Host+"/img1", regCache, "test-agent", credentials.NewMemoryStore())
		rne(err)
		assert.NotNil(t, r)
	})

	t.Run("TestTLSCertLocation", func(t *testing.T) {
		certLocations := getStandardCertLocations(secureURL.Host)
		certLocations = append(certLocations, filepath.Join(tmpdir, "testing"))
		location, err := resolveTLSCertLocation(certLocations)
		rne(err)

		assert.Equal(t, filepath.Join(tmpdir, "testing"), location)
	})

	t.Run("TestCAFetch", func(t *testing.T) {
		tlsCfg, err := fetchCertsFromLocation(tlsPath)
		rne(err)
		assert.NotNil(t, tlsCfg.RootCAs)
	})

	t.Run("TestCertKeyFetchFailure", func(t *testing.T) {
		f, err := os.Create(filepath.Join(tlsPath, "cert.pem"))
		rne(err)
		k, err := os.Create(filepath.Join(tlsPath, "key.pem"))
		rne(err)
		_, err = fetchCertsFromLocation(tlsPath)
		if assert.Error(t, err) {
			assert.EqualError(t, err, "error reading the certificate and key files: tls: failed to find any PEM data in certificate input")
		}
		rne(f.Close())
		rne(k.Close())
		rne(os.Remove(filepath.Join(tlsPath, "cert.pem")))
		rne(os.Remove(filepath.Join(tlsPath, "key.pem")))
	})

	t.Run("TestCertKeyFetchSuccess", func(t *testing.T) {
		c, err := newHTTPClientWithOps(nil, secureURL.Host, tlsPath)
		rne(err)
		repo, err := remote.NewRepository(path.Join(secureURL.Host, "ebarkett"))
		rne(err)
		repo.Client = c

		rng := mrand.New(mrand.NewSource(1)) //nolint:gosec
		n := rng.Intn(100) + 1
		data := make([]byte, n)
		_, err = rng.Read(data)
		rne(err)
		_, err = oras.PushBytes(ctx, repo, "", data)
		rne(err)
	})
}

func writeServerAndCACerts(tlsPath string) (*tls.Certificate, error) {
	f, err := os.Create(filepath.Join(tlsPath, "ca.pem"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	caKey, err := os.Create(filepath.Join(tlsPath, "ca-key.pem"))
	if err != nil {
		return nil, err
	}
	defer caKey.Close()

	// create the CA cert
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2023),
		Subject: pkix.Name{
			Organization:  []string{"test-org"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"Dayton"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Minute * 10),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return nil, err
	}

	caPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	if err != nil {
		return nil, err
	}

	_, err = f.Write(caPEM.Bytes())
	if err != nil {
		return nil, err
	}

	_, err = caKey.Write(caPrivKeyPEM.Bytes())
	if err != nil {
		return nil, err
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2023),
		Subject: pkix.Name{
			Organization:  []string{"test-org"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"Dayton"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Minute * 10),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	if err != nil {
		return nil, err
	}

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		return nil, err
	}

	return &serverCert, nil
}
