# Data Tool Configuration Guide

## Intended Audience

This documentation is written for Data Tool users who want to specify optional configuration settings for:

- A telemetry server so that when bottles are transferred they become discoverable to others and their lineage is automatically tracked
- Locations of certificate keys and custom metadata for digitally signing bottles
- One or more OCI repositories (mirrors) that can be used to efficiently manage movement of images or artifacts

## Prerequisites

It is assumed that readers have:

- Completed the [ACT3 Onboarding process](https://www.git.act3-ace.com/onboarding/onboarding-prerequisites/)
- [Installed Data Tool](installation-guide.md)
- Completed the [Data Tool Quick Start Guide](quick-start-guide.md)

## Workflow Overview

To make use of optional configuration settings, update the relevant sections of your `config.yaml` file.

Data Tool's configuration can also be modified with environment variables.

> Consult the [environment variables](#using-environment-variables) section for syntax and options

## Generating a Config File

If you don't have a `config.yaml` file in the default location, you can use Data Tool's CLI to generate one and place it here:

```sh
ace-dt config -s > ~/.config/ace/dt/config.yaml
```

To see a sample `config.yaml` file, run:

```sh
ace-dt config --sample
```

## Updating the Config File

To make use of optional configuration settings, update the relevant sections of your `config.yaml` file.

To include a telemetry host, update the section corresponding to **telemetry**:

```yaml
telemetry:
  - name: "host"
    url: "host.url.com"
```

To specify certificate key locations and signature metadata for use when signing bottles, update the section corresponding to **Signing configuration**:

```sh
Signing configuration
keys:
- alias: gitlabExampleKey
  path: path/to/private.key
  api: gitlab
  userid: your-gitlab-username
  keyid: key-title
```

To set default values for OCI registries and repositories, update the section corresponding to **Registry configuration**:

```yaml
registryConfig:
     index.docker.io:
       endpoints:
         - https://index.docker.io
         - http://localhost:5000
       rewritePull:
         "^rancher/(.*)": "ace/dt/rancher-images/$1"
         "^ubuntu/(.*)": "ace/dt/ubuntu-images/$1"
     nvcr.io:
       endpoints:
         - https://nvcr.io
   endpointConfig:
     https://nvcr.io:
       supportsReferrers: "tag"
     https://index.docker.io:
       tls:
         insecureSkipVerify: true
```

Registry configurations allow you to specify settings on a per-registry basis to avoid rate limits, registry errors, and time waste.

Registry configuration is supported by the following `ace-dt` commands:

- `ace-dt mirror`
- `ace-dt oci`
- `ace-dt pypi`

Configuration options include the following fields:

- `registries`: a map of registry host names to their respective registry config settings
  - `endpoints`: a sub-field of `registries`; accepts an array of fully-qualified registry endpoints
  - `rewritePull`: a sub-field of `registries`; this field enforces a renaming scheme such that repositories with names matching the regular expression are replaced with the second field; e.g. ubuntu/fuzzylion-20.22.4 is replaced with ace/dt/ubuntu-images/fuzzylion-20.22.4
- `endpointConfig`: a key-value field that contains endpoint-specific settings for TLS and referrers support
  - `supportsReferrers`: a sub-field of `endpointConfig`; can be queried via a GET request to a specific referrers path; not supported by all registries
  - `tls`: a sub-field of `endpointConfig`: TLS certificates are automatically checked when interacting with a registry/endpoint
    - `insecureSkipVerify`: a sub-field of `tls`; allows a user to interact with the registry without SSL validation

See below for registry configuration options:

### registries

The `registries` field is a map of registry host names to their respective registry config settings. In the example above, the registries that we have defined in the config are `index.docker.io`, `reg.example.com`, and `localhost:5000`. The values in the `registries` field should not include scheme as that is defined in the endpoint (or assumed `https` if no endpoint is assigned or registry does not exist in the config).

### endpoints

As a sub-field of `registries`, the `endpoints` field accepts an array of fully-qualified registry endpoints. Currently only the first endpoint is supported but in the future this will expand to allow multiple endpoints. `ace-dt` will pull all images from the first endpoint defined. Endpoints are useful if a mirror already exists of a specific registry and can help avoid rate-limiting from registries like `index.docker.io`.

In the example above, `index.docker.io` has an endpoint defined: `http://localhost:5000`. If a user runs `ace-dt mirror gather` with this command and with `docker.io` images in their `sources.list` file, `ace-dt` will resolve the image location from `http://localhost:5000`.

### rewritePull

The `rewritePull` sub-field of `registries` allows a user to define regex to pull an image from a different *repository location* than what may be defined in a `sources.list` file. This is useful when used in conjunction with `endpoints`, where mirrored image locations may have a nested repository structure compared to the source registry.

In the example above, we can see that there are two image rewrite defined:

```yaml
rewritePull:
        "^rancher/(.*)": "ace/dt/rancher-images/$1"
        "^library/ubuntu": "ace/dt/ubuntu-images/ubuntu$1"
```

Assuming a user were to have this in their registry config file with rancher and ubuntu images in their `ace-dt mirror gather` `sources.list` file, AND that the `http://localhost:5000` endpoint is valid, `ace-dt` would look for rancher at the location `localhost:5000/ace/dt/rancher-images/rancher:{TAG}`. Likewise, it would also look for the `ubuntu` image at `localhost:5000/ace/dt/ubuntu-images/ubuntu:{TAG}`.

### endpointConfig

The `endpointConfig` field contains some endpoint-specific settings for TLS and referrers support.
Its key value should be the fully-qualified endpoint URL. In the example above, we have one configuration defined for endpoint `https://reg.example.com`.

```yaml
endpointConfig:
    https://reg.example.com:
      supportsReferrers: "tag" # values "tag", "api", "auto"
      tls:
        insecureSkipVerify: true
```

### supportsReferrers

Some artifacts may reference manifests in their `Subject` field, defining the manifest as a dependency. With the [referrers API](https://github.com/opencontainers/distribution-spec/blob/main/spec.md#listing-referrers), these can be queried via a GET request to a specific referrers path:

```bash
GET /v2/{repository}/_oras/artifacts/referrers?digest={digest}
```

Some registries do not support this query, however. In that case, you can specify `supportReferrers` in your endpoint config. The default behavior assumes that referrers are supported, so if a specific registry returns an error upon querying the `/referrers` endpoint, you can set the `supportsReferrers` value to `tag` to turn off the API-querying behavior like so:

```yaml
supportsReferrers: "tag"
```

### tls

Certificate paths *do not* need to be specified in the `tls` section of the endpoint config ([see the section on adding TLS certificates](#adding-tls-certificates)).

`insecureSkipVerify` can be set in the `tls` section of the `endpointConfig` as outlined below.

### insecureSkipVerify

Setting `insecureSkipVerify` to `true` in the `endpointConfig` allows you to interact with the registry without SSL validation.

### Adding TLS Certificates

TLS certificates are automatically checked when interacting with a registry/endpoint. The `ace-dt` script will expect the appropriate `.pem` files to be located in one of 3 default locations (as expected from `containerd` and `docker`):

- `/etc/containerd/certs.d/{HOST_NAME}`
- `~/.docker/certs.d/{HOST_NAME}`
- `/etc/docker/certs.d/{HOST_NAME}`

A default naming convention of `cert.pem`, `key.pem`, and/or `ca.pem` is expected.

## Using Environment Variables

Data Tool's configuration can also be modified with environment variables.

The syntax for setting environment variables is:

```sh
export ACE_DT_TELEMETRY_URL=https://mytelemetry.example.com
```

Options include:

- ACE_DT_CACHE_PRUNE_MAX: Maximum cache prune size
- ACE_DT_CACHE_PATH: directory to use for caching
- ACE_DT_CHUNK_SIZE: Maximum chunk size for chunked uploads (set to "0" to disable)
- ACE_DT_CONCURRENT_HTTP: Maximum concurrent network connections
- ACE_DT_REGISTRY_AUTH_FILE then REGISTRY_AUTH_FILE: Docker registry auth file
- ACE_DT_EDITOR then VISUAL then EDITOR: Sets the editor to use for editing bottle schema
- ACE_DT_TELEMETRY_URL: If set will overwrite the telemetry configuration to only use this telemetry server URL.  Use the config file if you need multiple telemetry servers
- ACE_DT_TELEMETRY_USERNAME: Username to use for reporting events to telemetry
