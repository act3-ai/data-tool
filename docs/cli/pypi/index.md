---
title: ace-dt pypi
description: Python package syncing operations
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt pypi

Python package syncing operations

## Synopsis

Python package index and OCI credentials are both retrieved from the docker credential store.
Use "ace-dt login --no-auth-check" to add your python package index credentials to the store.
Only the hostname (and port) is used for the credential lookup.  The full URL path to the index is not used.

## Examples

```sh
The first step is to fetch distribution files from remote sources.
$ ace-dt pypi to-oci reg.example.com/my/pypi numpy -l 'version.major=1,version.minor>5'

or with a requirements file
$ ace-dt pypi to-oci reg.example.com/my/pypi -r requirements.txt
		
After you have fetched you can serve up the PyPI compliant (PEP-691) package index with
$ ace-dt pypi serve reg.example.com/my/pypi

```

## Options

```plaintext
Options:
      --allow-yanked   Do not ignore yanked distribution files
  -h, --help           help for pypi
```

## Options inherited from parent commands

```plaintext
Global options:
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,/root/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [warn])
```

## Subcommands

- [`ace-dt pypi labels`](labels.md) - Displays all the computed labels for a given python distribution filename DISTNAME
- [`ace-dt pypi serve`](serve.md) - Run the PyPI server
- [`ace-dt pypi to-oci`](to-oci.md) - Fetch python packages from Python package indexes and upload to the OCI-REPOSITORY as OCI images
- [`ace-dt pypi to-pypi`](to-pypi.md) - Pulls packages from OCI and uploads them to the python package index
