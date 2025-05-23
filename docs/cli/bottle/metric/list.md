---
title: ace-dt bottle metric list
description: list metric information from a bottle
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt bottle metric list

list metric information from a bottle

## Usage

```plaintext
ace-dt bottle metric list [flags]
```

## Aliases

```plaintext
ace-dt bottle metric ls
```

## Examples

```sh

List metrics from bottle in current working directory:
	ace-dt bottle metric list

List metric information from bottle at path <my/bottle/path>:
	ace-dt bottle metric list -d my/bottle/path

```

## Options

```plaintext
Options:
  -h, --help   help for list
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
