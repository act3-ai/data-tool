---
title: ace-dt bottle pull
description: Retrieves a bottle from remote OCI storage
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt bottle pull

Retrieves a bottle from remote OCI storage

## Synopsis

Retrieves a bottle from remote OCI storage, based on the bottle name and tag; stores the resulting files in the current directory, or at the directory supplied with the -d option.

A bottle reference follows the form <registry>/<repository>/<name>:<tag>
  
Pull a bottle using any of the following bottle reference forms
	by name (latest tag)  <registry>/<repository>/<name>
	by tag                <registry>/<repository>/<name>:<tag>
	by digest             <registry>/<repository>/<name>@<digest>
	by bottle ID          bottle:<digest>
where <digest> is often of the form sha256:<sha256 digest, lower case hex encoded>.

## Usage

```plaintext
ace-dt bottle pull BOTTLE_REFERENCE [flags]
```

## Examples

```sh
Pull the tagged bottle TESTSET:TAG from registry REG/REPO/TESTSET:TAG to path PATH:
  ace-dt bottle pull REG/REPO/TESTSET:TAG --bottle-dir PATH

Use bottle ID to pull a bottle from the best available location:
  ace-dt bottle pull bottle:sha256:abc123...

Pull a bottle that is publicly available:
  ace-dt bottle pull us-central1-docker.pkg.dev/aw-df16163b-7044-4662-93fa-ec0/public-down-auth-up/mnist:v2.1 -d mnist


```

## Options

```plaintext
Options:
  -u, --artifact stringArray   Retrieve only parts containing the provided public artifact type
      --debug string           Puts UI into debug mode, dumping all UI events to the given path.
      --empty                  retrieve empty bottle, only containing metadata
  -h, --help                   help for pull
      --no-term                Disable terminal support for fancy printing
  -p, --part stringArray       Parts to retrieve
  -q, --quiet                  Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
  -l, --selector stringArray   Provide selectors for which parts to retrieve. Format "name=value"
      --telemetry string       Overrides the telemetry server configuration with the single telemetry server URL provided.  
                               Modify the configuration file if multiple telemetry servers should be used or if auth is required.
```

## Options inherited from parent commands

```plaintext
Global options:
  -d, --bottle-dir string          Specify bottle directory (default "/work/src")
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,/root/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [warn])
```
