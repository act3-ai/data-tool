// Package v1alpha1 provides the ServerConfiguration used for configuring the Telemetry server.  Also include ClientConfiguration
// +kubebuilder:object:generate=true
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	telemv1alpha2 "github.com/act3-ai/data-telemetry/v3/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

// +kubebuilder:object:root=true

// Configuration defines a set of configuration parameters.
type Configuration struct {
	metav1.TypeMeta `json:",inline"`

	ConfigurationSpec `json:",inline"`
}

// ConfigurationSpec is the actual configuration values.
type ConfigurationSpec struct {
	// CachePruneMax is the maximum cache size after pruning in megabytes
	CachePruneMax *resource.Quantity `json:"cachePruneMax,omitempty"`

	// CachePath is the directory where the cache fields are stored
	CachePath string `json:"cachePath,omitempty"`

	// CompressionLevel is the level used for compression.  Valid values are min, normal, max
	CompressionLevel string `json:"compressionLevel,omitempty"`

	// Editor is the text editor used for editing the files (e.g., ace-dt bottle edit)
	Editor string `json:"editor,omitempty"`

	// ChunkSize is the maximum size of a chunk used to upload blobs. 0 disables chunking
	ChunkSize *resource.Quantity `json:"chunkSize,omitempty"`

	// ConcurrentHTTP is the maximum number of HTTP requests that will be in flight at any given time. Must be positive.
	ConcurrentHTTP int `json:"concurrentHTTP,omitempty"`

	// RegistryAuthFile is the file to use for credentials to an OCI registry
	RegistryAuthFile string `json:"registryAuthFile,omitempty"`

	// RegistryConfig stores custom registry mapping and http client options.
	RegistryConfig RegistryConfig `json:"registryConfig,omitempty"`

	// HideProgress will disable the process bar if true
	HideProgress bool `json:"hideProgress"` // TODO this should be a flag

	// Telemetry is a list of telemetry server locations
	Telemetry []telemv1alpha2.Location `json:"telemetry,omitempty"`

	// TelemetryUserName is the name of this user for reporting to telemetry
	TelemetryUserName string `json:"telemetryUserName,omitempty"`

	// SigningKeys is a list of signing key metadata.
	SigningKeys []SigningKey `json:"keys,omitempty"`
}

// FIXME redact the telemetry config secrets

// SampleConfig is a sample Configuration snippet.
const SampleConfig = `# ACE Data Tool Configuration
apiVersion: config.dt.act3-ace.io/v1alpha1
kind: Configuration

# CachePruneMax is the maximum cache size after pruning
# cachePruneMax: 500Mi

# CachePath is the directory where the cache fields are stored
# cachePath:

# CompressionLevel is the level used for compression.  Valid values are min, normal, max
# compressionLevel: normal

# ChunkSize is the maximum size of a chunk used to upload blobs. 0 disables chunking
# chunkSize: 100Mi

# ConcurrentHTTP is the maximum number of HTTP requests that will be in flight at any given time. Must be positive.
# concurrentHTTP: 25

# RegistryAuthFile is the file to use for credentials to an OCI registry.  Defaults to ~/.docker/config.json
# registryAuthFile: ~/.docker/config.json

# Telemetry configuration
# telemetry:
# - name: lion
#   url: https://telemetry.lion.act3-ace.ai
# telemetryUserName: your-username

# Signing configuration
# keys:
# - alias: gitlabExampleKey
#   path: path/to/private.key
#   api: gitlab
#   userid: your-gitlab-username
#   keyid: key-title

# Registry configuration
# registryConfig:
#   registries:
#     index.docker.io:
#       endpoints:
#         - https://index.docker.io
#       rewritePull:
#         "^rancher/(.*)": "ace/dt/rancher-images/$1"
#         "^ubuntu/(.*)": "ace/dt/ubuntu-images/$1"
#     nvcr.io:
#       endpoints:
#         - https://nvcr.io
#     localhost:5000:
#       endpoints:
#         - http://localhost:5000
#   endpointConfig:
#     https://nvcr.io:
#       supportsReferrers: "tag"
#     http://localhost:5000
#       tls:
#         insecureSkipVerify: true
`

// SigningKey is metadata which is added to signature annotations when signing.
type SigningKey struct {
	// Alias is the unique user-defined name for the key.
	Alias string `json:"alias,omitempty"`

	// Path to the signing key.
	KeyPath string `json:"path,omitempty"`

	// API used to access the key.
	KeyAPI string `json:"api,omitempty"`

	// Key owner's identity, typically a username for the KeyAPI.
	UserIdentity string `json:"userid"`

	// Title of the key as indicated by the api.
	KeyID string `json:"keyid"`
}

// RegistryConfig holds the custom configuration data for registries and repositories.
type RegistryConfig struct {
	Configs        map[string]Registry       `json:"registries"`
	EndpointConfig map[string]EndpointConfig `json:"endpointConfig,omitempty"`
}

// Registry contains the custom configuration for a registry. This includes a slice of endpoints (including http:// or https:// scheme),
// and a map of any rewrites for pulling from mirrors. The endpoints will be attempted in the listed order and the first to resolve will be used.
type Registry struct {
	// Endpoints is a list of string endpoint URLs (with scheme included) that are mirrors to the original registry.
	// e.g., docker.io might have a mirror at https://example.mirror.com or http://localhost that the image can be pulled from instead of docker.io.
	Endpoints []string `json:"endpoints,omitempty"`
	// RewritePull to pull from a specified mirrored location that may be different from the original
	RewritePull map[string]string `json:"rewritePull,omitempty"`
	// NonCompliant indicates a registry is not OCI compliant
	NonCompliant bool `json:"noncompliant,omitempty"`
}

// EndpointConfig contains the specified TLS configuration for an endpoint and the endpoint's referrers type. This value can be
// "tag" (for referrers not supported), "api", or "auto".
type EndpointConfig struct {
	TLS           *TLS   `json:"tls,omitempty"`
	ReferrersType string `json:"supportsReferrers,omitempty"`
}

// TLS defines the locations of the certificate files or can set insecureSkipVerify (which will define whether to verify the target registry's certificate).
type TLS struct {
	// Expects client certificates to be named cert.pem and key.pem,.  Expects trusted server certificates to be in ca.pem.
	// The files must be in either default locations:
	// /etc/containerd/certs.d/{HOST_NAME}, /etc/docker/certs.d/{HOST_NAME} or ~/.docker/certs.d/{HOST_NAME}

	// InsecureSkipVerify skips the verification of the server's certificate (may allow MiTM attack so use sparingly).
	InsecureSkipVerify bool `json:"insecureSkipVerify"`
}
