---
title: ace-dt mirror scatter
description: A command that scatters images to destination registries defined in the MAPPER
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt mirror scatter

A command that scatters images to destination registries defined in the MAPPER

## Synopsis

A command that scatters images located in the source registry repo to multiple
remote repositories defined by the user with MAPPER.

The MAPPER types currently supported are nest, first-prefix (csv format), digests (csv format) and go-template.
The format of MAPPER is MAP-TYPE=MAP-ARG

If MAP-TYPE is "nest" then scatter will nest all the images under MAP-ARG.
For example, is MAP-ARG is "reg.other.com" then a gathered image "foo.com/bar" will map to "reg.other.com/foo.com/bar".

Passing a first-prefix MAPPER requires a csv file that has formatted lines of: source,destination. 
The ace-dt mirror scatter will send the source reference to the first prefix match that it makes.
This format also allows defining the source as a digest that is present in the source repository.

Passing a digests MAP-FILE requires a csv file that has formatted lines of: digest-string, destination.
Scatter will send each digest to the locations defined in the map file provided. 

Passing a go-template MAP-FILE allows greater flexibility in how references can be pushed
to destination repositories. Hermetic text Sprig functions are currently supported which allows for matching by 
prefix, digest, media-type, regex, etc.  The following additional functions are provided:

Tag - Returns the tag of an OCI string
Repository - Returns the repository of an OCI string
Registry - Returns the registry of an OCI string
Package - Returns omits the registry from the OCI reference

Example csv and go template files are located in the pkg/actions/mirror/test repository.
		

## Usage

```plaintext
ace-dt mirror scatter IMAGE MAPPER [flags]
```

## Examples

```sh
To put all the images nested under "reg.other.com/mirror" you can use
ace-dt mirror scatter reg.example.com/repo/data:sync-45 nest=ref.other.com/mirror

ace-dt mirror scatter reg.example.com/repo/data:sync-45 go-template=mapping.tmpl
ace-dt mirror scatter reg.example.com/repo/data:sync-45 first-prefix=mapping.csv
ace-dt mirror scatter reg.example.com/repo/data:sync-45 digests=mapping.csv
ace-dt mirror scatter reg.example.com/repo/data:sync-45 longest-prefix=mapping.csv
ace-dt mirror scatter reg.example.com/repo/data:sync-45 all-prefix=mapping.csv

To scatter by filtering on manifest labels, you can use
ace-dt mirror scatter reg.example.com/repo/data:sync-45 nest=ref.other.com/mirror --filter-labels=component=core,module=test

```

## Options

```plaintext
Options:
      --check              Dry run- do not actually send to destination repositories
      --debug string       Puts UI into debug mode, dumping all UI events to the given path.
  -h, --help               help for scatter
      --no-term            Disable terminal support for fancy printing
  -q, --quiet              Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
  -l, --selector strings   Only scatter manifests tagged with annotation labels, e.g., component=core,module=test
      --subset string      Define a subset list of images to scatter with a sources.list file
```

## Options inherited from parent commands

```plaintext
Global options:
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,/root/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -r, --recursive                  recursively copy the referrers
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [warn])
```
