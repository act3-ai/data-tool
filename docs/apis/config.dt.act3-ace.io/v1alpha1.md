# API Reference

## Packages
- [config.dt.act3-ace.io/v1alpha1](#configdtact3-aceiov1alpha1)


## config.dt.act3-ace.io/v1alpha1

Package v1alpha1 provides the ServerConfiguration used for configuring the Telemetry server.  Also include ClientConfiguration

Package v1alpha1 contains API schema definitions for managing ace-dt configuration.

### Resource Types
- [Configuration](#configuration)



#### Configuration



Configuration defines a set of configuration parameters.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `config.dt.act3-ace.io/v1alpha1`
| `kind` _string_ | `Configuration`
| `cachePruneMax` _Quantity_ | CachePruneMax is the maximum cache size after pruning in megabytes |
| `cachePath` _string_ | CachePath is the directory where the cache fields are stored |
| `compressionLevel` _string_ | CompressionLevel is the level used for compression.  Valid values are min, normal, max |
| `editor` _string_ | Editor is the text editor used for editing the files (e.g., ace-dt bottle edit) |
| `chunkSize` _Quantity_ | ChunkSize is the maximum size of a chunk used to upload blobs. 0 disables chunking |
| `concurrentHTTP` _integer_ | ConcurrentHTTP is the maximum number of HTTP requests that will be in flight at any given time. Must be positive. |
| `registryAuthFile` _string_ | RegistryAuthFile is the file to use for credentials to an OCI registry |
| `registryConfig` _[RegistryConfig](#registryconfig)_ | RegistryConfig stores custom registry mapping and http client options. |
| `hideProgress` _boolean_ | HideProgress will disable the process bar if true |
| `telemetry` _invalid type array_ | Telemetry is a list of telemetry server locations |
| `telemetryUserName` _string_ | TelemetryUserName is the name of this user for reporting to telemetry |
| `keys` _[SigningKey](#signingkey) array_ | SigningKeys is a list of signing key metadata. |


#### ConfigurationSpec



ConfigurationSpec is the actual configuration values.

_Appears in:_
- [Configuration](#configuration)

| Field | Description |
| --- | --- |
| `cachePruneMax` _Quantity_ | CachePruneMax is the maximum cache size after pruning in megabytes |
| `cachePath` _string_ | CachePath is the directory where the cache fields are stored |
| `compressionLevel` _string_ | CompressionLevel is the level used for compression.  Valid values are min, normal, max |
| `editor` _string_ | Editor is the text editor used for editing the files (e.g., ace-dt bottle edit) |
| `chunkSize` _Quantity_ | ChunkSize is the maximum size of a chunk used to upload blobs. 0 disables chunking |
| `concurrentHTTP` _integer_ | ConcurrentHTTP is the maximum number of HTTP requests that will be in flight at any given time. Must be positive. |
| `registryAuthFile` _string_ | RegistryAuthFile is the file to use for credentials to an OCI registry |
| `registryConfig` _[RegistryConfig](#registryconfig)_ | RegistryConfig stores custom registry mapping and http client options. |
| `hideProgress` _boolean_ | HideProgress will disable the process bar if true |
| `telemetry` _invalid type array_ | Telemetry is a list of telemetry server locations |
| `telemetryUserName` _string_ | TelemetryUserName is the name of this user for reporting to telemetry |
| `keys` _[SigningKey](#signingkey) array_ | SigningKeys is a list of signing key metadata. |


#### EndpointConfig



EndpointConfig contains the specified TLS configuration for an endpoint and the endpoint's referrers type. This value can be "tag" (for referrers not supported), "api", or "auto".

_Appears in:_
- [RegistryConfig](#registryconfig)

| Field | Description |
| --- | --- |
| `tls` _[TLS](#tls)_ |  |
| `supportsReferrers` _string_ |  |


#### Registry



Registry contains the custom configuration for a registry. This includes a slice of endpoints (including http:// or https:// scheme), and a map of any rewrites for pulling from mirrors. The endpoints will be attempted in the listed order and the first to resolve will be used.

_Appears in:_
- [RegistryConfig](#registryconfig)

| Field | Description |
| --- | --- |
| `endpoints` _string array_ | Endpoints is a list of string endpoint URLs (with scheme included) that are mirrors to the original registry. e.g., docker.io might have a mirror at https://example.mirror.com or http://localhost that the image can be pulled from instead of docker.io. |
| `rewritePull` _object (keys:string, values:string)_ | RewritePull to pull from a specified mirrored location that may be different from the original |


#### RegistryConfig



RegistryConfig holds the custom configuration data for registries and repositories.

_Appears in:_
- [Configuration](#configuration)
- [ConfigurationSpec](#configurationspec)

| Field | Description |
| --- | --- |
| `registries` _object (keys:string, values:[Registry](#registry))_ |  |
| `endpointConfig` _object (keys:string, values:[EndpointConfig](#endpointconfig))_ |  |


#### SigningKey



SigningKey is metadata which is added to signature annotations when signing.

_Appears in:_
- [Configuration](#configuration)
- [ConfigurationSpec](#configurationspec)

| Field | Description |
| --- | --- |
| `alias` _string_ | Alias is the unique user-defined name for the key. |
| `path` _string_ | Path to the signing key. |
| `api` _string_ | API used to access the key. |
| `userid` _string_ | Key owner's identity, typically a username for the KeyAPI. |
| `keyid` _string_ | Title of the key as indicated by the api. |


#### TLS



TLS defines the locations of the certificate files or can set insecureSkipVerify (which will define whether to verify the target registry's certificate).

_Appears in:_
- [EndpointConfig](#endpointconfig)

| Field | Description |
| --- | --- |
| `insecureSkipVerify` _boolean_ | InsecureSkipVerify skips the verification of the server's certificate (may allow MiTM attack so use sparingly). |


