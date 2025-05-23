---
title: ace-dt bottle label
description: add key-value pair as a label to specified bottle
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt bottle label

add key-value pair as a label to specified bottle

## Synopsis

Add key-value pair label to the bottle

A label key and value must begin with a letter or number, and may contain 
  letters, numbers, hyphens, dots, and underscores, up to  63 characters each.

Do not confuse bottle labels with part labels.  See "ace-dt bottle part label -h" for more information about how to add labels to parts.


## Usage

```plaintext
ace-dt bottle label <key>=<value> [flags]
```

## Examples

```sh

Add label <foo=bar> to bottle at path <my/bottle/path>:
	ace-dt bottle label foo=bar -d my/bottle/path

Add multiple labels to a bottle in current working directory:
	ace-dt bottle label foo1=bar1 foo2=bar2 foo3=bar3

Remove label <foo> from bottle <bar> at path <my/bottle/path>:
	ace-dt bottle label foo- -d my/bottle/path

```

## Options

```plaintext
Options:
  -h, --help   help for label
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

## Subcommands

- [`ace-dt bottle label list`](list.md) - list labels applied on specified bottle
