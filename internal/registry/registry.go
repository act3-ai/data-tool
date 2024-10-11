// Package registry provides options for remote OCI registry config and caching.
package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/adrg/xdg"
	"golang.org/x/net/proxy"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/httplogger"
	regcache "gitlab.com/act3-ai/asce/data/tool/internal/registry/cache"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// CreateRepoWithCustomConfig creates a remote.Repository object and sets it up based off
// the custom parameters defined in registryConfig (inside ace-dt config file).
func CreateRepoWithCustomConfig(ctx context.Context, rc *v1alpha1.RegistryConfig, ref string,
	cache *regcache.RegistryCache, userAgent string, credStore credentials.Store) (registry.Repository, error) {
	log := logger.FromContext(ctx)

	// parse the original reference, without endpoint resolution
	parsedRef, err := registry.ParseReference(ref)
	if err != nil {
		return nil, fmt.Errorf("invalid reference %s: %w", ref, err)
	}

	// get the registry config and its existence
	// In the case of a non-existent entry in the config, we return a default repository.
	r := rc.Configs[parsedRef.Registry]

	// does the registry exist in our cache?
	cachedReg, ok := cache.Exists(parsedRef.Registry)
	if ok {
		// we need to pass reg to some repo function that handles the rest of the bits
		if cachedReg.RemoteRegistry.Reference.Registry != parsedRef.Registry {
			log.InfoContext(ctx, "using alternate endpoint defined by registry configuration", "original", parsedRef.Registry, "alternate", cachedReg.RemoteRegistry.Reference.Registry)
		}
		return createRegistryRepository(ctx, cachedReg, parsedRef)
	}

	endpointURL, err := ResolveEndpoint(rc, parsedRef)
	if err != nil {
		return nil, fmt.Errorf("resolving endpoint: %w", err) // only potential failure is url parsing
	}
	if endpointURL.Host != parsedRef.Registry {
		log.InfoContext(ctx, "using alternate endpoint defined by registry configuration", "original", parsedRef.Registry, "alternate", endpointURL.Host)
	}

	var referrersType string
	var endpointTLS *v1alpha1.TLS

	// get the endpoint's config
	ecfg, ok := rc.EndpointConfig[endpointURL.String()]
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
				Credential: credentials.Credential(credStore),
			},
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
	newReg := regcache.Registry{
		RemoteRegistry: endpointReg,
		ReferrersType:  referrersType,
		RewritePull:    r.RewritePull,
	}
	cache.AddRegistryToCache(parsedRef.Registry, newReg)
	return createRegistryRepository(ctx, newReg, parsedRef)
}

// if a nil TLS is passed, return a client with a logging transport wrapped in a retry transport.
// if a TLS config exists, search for TLS certs and append to client.
func newHTTPClientWithOps(cfg *v1alpha1.TLS, hostName, customCertPath string) (*http.Client, error) {
	nd := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	// defaultTransport is a new instance of the default transport
	var defaultTransport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           nd.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

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

	ssl, err := fetchCertsFromLocation(certLocation)
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		ssl.InsecureSkipVerify = cfg.InsecureSkipVerify
	}

	defaultTransport.TLSClientConfig = ssl

	// get the proxy from the environment
	dialer := proxy.FromEnvironment()

	if dialer != nil {
		defaultTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	}

	// log requests to the logger (if verbosity is high enough)
	lt := &httplogger.LoggingTransport{
		Base: defaultTransport,
	}

	// we still want retry
	rt := retry.NewTransport(lt)

	return &http.Client{
		Transport: rt,
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

	// add system level certs
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("fetching system certs: %w", err)
	}

	if certDir != "" {

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

		// caCertPool := x509.NewCertPool()
		// Only trust this CA for this host
		caCertPool.AppendCertsFromPEM(caCert)

	}

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
