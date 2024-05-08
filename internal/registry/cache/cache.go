// Package cache implements tools for OCI registry caching.
package cache

import (
	"regexp"
	"strings"
	"sync"

	"oras.land/oras-go/v2/registry/remote"
)

// Registry represents the cached registry, HTTP client and repository-specific settings defined in the configuration.
type Registry struct {
	RemoteRegistry  *remote.Registry
	ReferrersType   string
	RewritePull     map[string]string
	signCertKeyPath string
}

// SignCertPath returns the signing certificate private key path for the associated registry.
func (r *Registry) SignCertPath() string {
	return r.signCertKeyPath
}

// SetSigningKeyPath sets the signing key path for a registry.
func (r *Registry) SetSigningKeyPath(keyFilePath string) {
	r.signCertKeyPath = keyFilePath
}

// IsValid returns true if the registry object exists (e.g. if remoteRegistry is not nil).
func (r *Registry) IsValid() bool {
	return r.RemoteRegistry != nil
}

// RegistryCache represents the in-memory storage of registries to limit duplication of remote.Registry and HTTP client objects.
type RegistryCache struct {
	mutex              sync.RWMutex
	existingRegistries map[string]Registry
}

// GetRegistryFromCache loads a registry from a RegistryCache based on a regex match with partial name, or an empty
// one if no match is found.
func GetRegistryFromCache(partialName string, cache *RegistryCache) Registry {
	reg, ok := cache.search(partialName)
	if ok {
		return reg
	}
	return Registry{}
}

// NewRegistryCache creates a new registry cache to store registry and client information.
func NewRegistryCache() *RegistryCache {
	return &RegistryCache{
		existingRegistries: make(map[string]Registry),
		mutex:              sync.RWMutex{},
	}
}

// AddRegistryToCache adds an entry to the cache.
func (cache *RegistryCache) AddRegistryToCache(hostname string, r Registry) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.existingRegistries[hostname] = r
}

// Exists checks the cache for an entry discovered by hostname.
func (cache *RegistryCache) Exists(hostname string) (Registry, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	r, ok := cache.existingRegistries[hostname]
	return r, ok
}

func (cache *RegistryCache) search(hostnameMatch string) (Registry, bool) {
	var foundReg Registry
	found := false
	var hostnameRegex *regexp.Regexp
	if strings.HasSuffix(hostnameMatch, "*") {
		hostnameRegex = regexp.MustCompile(`^([a-z0-9]+)`)
	} else {
		hostnameRegex = regexp.MustCompile(hostnameMatch)
	}
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	// regex to search for the beginning of a hostname given a partial match in the above
	for _, reg := range cache.existingRegistries {
		if ok := hostnameRegex.MatchString(reg.RemoteRegistry.Reference.Registry); ok {
			foundReg = reg
			found = true
			break
		}
	}
	return foundReg, found
}
