---
title: ace-dt pypi to-pypi
description: Pulls packages from OCI and uploads them to the python package index
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt pypi to-pypi

Pulls packages from OCI and uploads them to the python package index

## Synopsis

Pulls packages from OCI at OCIREF and uploads them to the python package index at PYPI-REPOSITORY.

OCIREF is a repository reference (no tag).
PYPI_REPOSITORY is a python package index.  This is the same URL used by "twine"'s "TWINE_REPOSITORY_URL" setting.  It does not include the trailing "/simple" that is used by "pip".


## Usage

```plaintext
ace-dt pypi to-pypi OCIREF PYPI-REPOSITORY [flags]
```

## Examples

```sh
To upload all the packages in OCI at reg.example.com to Gitlab PyPI at https://git.example.com/api/v4/projects/1234/packages/pypi run the command
ace-dt pypi to-pypi reg.example.com/mypypi https://git.example.com/api/v4/projects/1234/packages/pypi
```

## Options

```plaintext
Options:
      --debug string   Puts UI into debug mode, dumping all UI events to the given path.
      --dry-run        Dry run by only determining what work needs to be done.  Does not upload distribution files to the python package index.
  -h, --help           help for to-pypi
      --no-term        Disable terminal support for fancy printing
  -q, --quiet          Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
```

## Options inherited from parent commands

```plaintext
Global options:
      --allow-yanked               Do not ignore yanked distribution files
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,/root/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [warn])
```
