package v1alpha1

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ConfigurationDefault defaults the configuration values.
func ConfigurationDefault(obj *Configuration) {
	// This is called after we decode the values (from file) so we need to be careful not to overwrite values that are already set.
	// We can use pointers if we need to know that a value has been set or not.

	// These might not be set in some cases
	obj.APIVersion = GroupVersion.String()
	obj.Kind = "Configuration"

	if obj.CachePruneMax == nil {
		cachePruneMax := resource.MustParse("500Mi")
		obj.CachePruneMax = &cachePruneMax
	}

	if obj.CachePath == "" {
		obj.CachePath = filepath.Join(xdg.CacheHome, "ace", "dt")
	}

	if obj.CompressionLevel == "" {
		obj.CompressionLevel = "normal"
	}

	if obj.Editor == "" {
		// TODO make this a reasonable default based on operating system
		obj.Editor = "nano" // vi
	}

	if obj.ChunkSize == nil {
		chunkSize := resource.MustParse("100Mi")
		obj.ChunkSize = &chunkSize
	}

	if obj.ConcurrentHTTP < 1 {
		obj.ConcurrentHTTP = 5
	}

	if obj.RegistryAuthFile == "" {
		// skopeo/podman supports the env REGISTRY_AUTH_FILE
		// There default location is ${XDG_RUNTIME_DIR}/containers/auth.json
		// filepath.Join(xdg.RuntimeDir, "containers", "auth.json")
		// see https://docs.podman.io/en/latest/markdown/podman-login.1.html
		// man containers-auth.json
		// We need to change how login, logout, and registry auth for pull/push operate before we change the default to auth.json so that it falls back to the .docker/config.json and .dockercfg
		obj.RegistryAuthFile = filepath.Join(xdg.Home, ".docker", "config.json")
	}
}
