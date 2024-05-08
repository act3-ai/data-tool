// Package registry provides options for remote OCI registry config and caching.
package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/adrg/xdg"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"

	"git.act3-ace.com/ace/data/tool/internal/httplogger"
	regcache "git.act3-ace.com/ace/data/tool/internal/registry/cache"
	"git.act3-ace.com/ace/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// CredentialOption applies optional credential configuration.
type CredentialOption func(credentials.Store)

// WithCredential adds credentials for a registry to an existing credential store.
func WithCredential(ctx context.Context, hostname string, cred auth.Credential) CredentialOption {
	return func(cs credentials.Store) {
		_ = cs.Put(ctx, credentials.ServerAddressFromHostname(hostname), cred)
	}
}

// CreateRepoWithCustomConfig creates a remote.Repository object and sets it up based off
// the custom parameters defined in registryConfig (inside ace-dt config file).
func CreateRepoWithCustomConfig(ctx context.Context, rc *v1alpha1.RegistryConfig, ref string,
	cache *regcache.RegistryCache, userAgent string, fallbackCredStores ...credentials.Store) (registry.Repository, error) {

	// handle auth
	storeOpts := credentials.StoreOptions{}
	mainCredStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential store: %w", err)
	}
	credStore := credentials.NewStoreWithFallbacks(mainCredStore, fallbackCredStores...)

	// parse the reference
	parsedRef, err := registry.ParseReference(ref)
	if err != nil {
		return nil, fmt.Errorf("invalid reference: %w", err)
	}

	// get the registry config and its existence
	// In the case of a non-existent entry in the config, we return a default repository.
	r := rc.Configs[parsedRef.Registry]

	// does the registry exist in our cache?
	reg, ok := cache.Exists(parsedRef.Registry)
	if ok {
		// we need to pass reg to some repo function that handles the rest of the bits
		return createRegistryRepository(ctx, reg, parsedRef)
	}

	fullRegistryPath := "https://" + parsedRef.Registry
	var endpoints []string
	if len(r.Endpoints) != 0 {
		endpoints = append(endpoints, r.Endpoints...)
	}
	endpoints = append(endpoints, fullRegistryPath)

	// handle endpoints
	// TODO support more than first endpoint when registry.Ping is better supported by registries (nvcr.io and quay.io)
	endpoint := endpoints[0]

	// get the http client config
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error parsing endpoint URL: %w", err)
	}

	var referrersType string
	var endpointTLS *v1alpha1.TLS

	// get the endpoint's config
	ecfg, ok := rc.EndpointConfig[endpoint]
	if ok {
		referrersType = ecfg.ReferrersType
		endpointTLS = ecfg.TLS
	}

	c, err := newHTTPClientWithOps(endpointTLS, endpointURL.Host, "")
	if err != nil {
		return nil, err
	}

	// create the endpoint registry object
	endpointReg := &remote.Registry{
		RepositoryOptions: remote.RepositoryOptions{
			Client: &auth.Client{
				Client: c,
				Header: http.Header{
					"User-Agent": {userAgent},
				},
				Cache: auth.DefaultCache,
				// Cache: auth.NewSingleContextCache(), // TODO could consider using this one
				Credential: credentials.Credential(credStore)},
			Reference: registry.Reference{
				Registry: endpointURL.Host,
				// we want to set the repository and reference after cacheing the registry
				// Repository: parsedRef.Repository,
				// Reference:  parsedRef.Reference,
			},
			PlainHTTP: endpointURL.Scheme == "http",
		},
	}

	// add the registry to the cache
	cachedReg := regcache.Registry{
		RemoteRegistry: endpointReg,
		ReferrersType:  referrersType,
		RewritePull:    r.RewritePull,
	}
	cache.AddRegistryToCache(parsedRef.Registry, cachedReg)
	return createRegistryRepository(ctx, cachedReg, parsedRef)
}

// if a nil TLS is passed, return a client with a logging transport (if ACE_DT_HTTP_LOG is set) wrapped in a retry transport.
// if a TLS config exists, search for TLS certs and append to client.
func newHTTPClientWithOps(cfg *v1alpha1.TLS, hostName, customCertPath string) (*http.Client, error) {
	var transport = http.DefaultTransport

	var certLocation string
	if customCertPath == "" {
		cLocation, err := resolveTLSCertLocation(getStandardCertLocations(hostName))
		if err != nil {
			return nil, err
		}
		certLocation = cLocation
	} else {
		certLocation = customCertPath
	}

	if certLocation != "" {
		ssl, err := fetchCertsFromLocation(certLocation)
		if err != nil {
			return nil, err
		}
		if cfg != nil {
			ssl.InsecureSkipVerify = cfg.InsecureSkipVerify
		}
		transport = &http.Transport{
			TLSClientConfig: ssl,
		}
	}

	// log requests to the logger (if verbosity is high enough)
	transport = &httplogger.LoggingTransport{
		Base: transport,
	}

	// we still want retry
	transport = retry.NewTransport(transport)

	return &http.Client{
		Transport: transport,
	}, nil
}

// resolveTLSCertLocation first searches for the registry certs in containerd's default TLS config path.
// If it is not located there it falls back to docker's default TLS config path.
// If there is no cert repository it will return an empty string.
// More info on containerd: https://github.com/containerd/containerd/blob/main/docs/hosts.md
// More info on docker: https://docs.docker.com/engine/reference/commandline/dockerd/#insecure-registries
func resolveTLSCertLocation(paths []string) (string, error) {
	// locations to search for certs
	// containerdPath := filepath.Join("/etc/containerd/certs.d", hostName)
	// dockerPath := filepath.Join(xdg.Home, ".docker/certs.d", hostName)
	// etcDockerPath := filepath.Join("/etc/docker/certs.d", hostName)

	for _, certPath := range paths {
		_, err := os.Stat(certPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", fmt.Errorf("error accessing the TLS certificates in %s: %w", certPath, err)
		}
		return certPath, nil
	}
	return "", nil
}

func fetchCertsFromLocation(certDir string) (*tls.Config, error) {
	certFilePath := filepath.Join(certDir, "cert.pem")
	keyFilePath := filepath.Join(certDir, "key.pem")
	caFilePath := filepath.Join(certDir, "ca.pem")

	tlscfg := &tls.Config{}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("error reading the certificate and key files: %w", err)
		}
	}
	tlscfg.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := os.ReadFile(caFilePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return tlscfg, nil
		}
		return nil, fmt.Errorf("error reading the caFile: %w", err)
	}

	// Only trust this CA for this host
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlscfg.RootCAs = caCertPool
	return tlscfg, nil
}

func createRegistryRepository(ctx context.Context, r regcache.Registry, parsedRef registry.Reference) (registry.Repository, error) {
	// handle repo rewrite
	rewrittenRepo, err := handleRepoRewrite(parsedRef.Repository, r.RewritePull)
	if err != nil {
		return nil, err
	}

	rr, err := r.RemoteRegistry.Repository(ctx, rewrittenRepo)
	if err != nil {
		return nil, fmt.Errorf("error creating registry repository: %w", err)
	}
	repo := rr.(*remote.Repository)
	repo.Reference.Reference = parsedRef.Reference

	// handle referrers type (api/tag/auto)
	switch r.ReferrersType {
	case "tag":
		err := repo.SetReferrersCapability(false)
		if err != nil {
			return nil, fmt.Errorf("error during repository setting of referrers capability: %w", err)
		}
	case "api":
		err := repo.SetReferrersCapability(true)
		if err != nil {
			return nil, fmt.Errorf("error during repository setting of referrers capability: %w", err)
		}
	case "auto":
		// do nothing
	case "":
		// do nothing
	default:
		return nil, fmt.Errorf("invalid referrers capability set")
	}
	return repo, nil
}

func handleRepoRewrite(repoRef string, rewritePull map[string]string) (string, error) {
	// handle repository rewrite
	// for each image in rewrite map, check and see if regex matches target repository
	for k, v := range rewritePull {
		rx, err := regexp.Compile(k)
		if err != nil {
			return "", fmt.Errorf("error parsing the regex provided in the registry config: %w", err)
		}
		if rx.MatchString(repoRef) {
			return rx.ReplaceAllString(repoRef, v), nil
		}
	}
	return repoRef, nil
}

// Currently, there are three standard locations checked for TLS certificates in ace-dt (modeled after containerd's implementation).
// First we check the standard containerd location for certs in /etc/containerd/certs.d/{HOSTNAME}.
// If it is not located there, we follow containerd's fallback location checks in docker's 2 certificate locations, located in /etc/docker/certs.d/{HOSTNAME} and ~/.docker/certs.d/{HOSTNAME} respectively.
func getStandardCertLocations(hostName string) []string {
	return []string{
		filepath.Join("/etc/containerd/certs.d", hostName), filepath.Join("/etc/docker/certs.d", hostName), filepath.Join(xdg.Home, ".docker/certs.d", hostName),
	}
}
