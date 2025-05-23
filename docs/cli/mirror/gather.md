---
title: ace-dt mirror gather
description: Efficiently copies images listed in SOURCES-FILE to the IMAGE
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt mirror gather

Efficiently copies images listed in SOURCES-FILE to the IMAGE

## Synopsis

Efficiently copies images listed in SOURCES-FILE to the IMAGE.
		
SOURCES-FILE is a text file with one OCI image reference per line.  Lines that begin with # are ignored. 
Labels can be added to each source in the SOURCES-FILE by separating with a comma and following a key=value format. These will be added as annotations to that manifest:
reg.example.com/library/source1,component=core,module=test

IMAGE is an OCI image reference that will be used to push all the missing blobs and manifests.
The manifest at the tag will be a OCI Image Index.

Many gather commands can be run to gather images from different registries.  Ensure that they push to different tags in the destination repository.

Often the next command run is the "ace-dt mirror serialize" command.

## Usage

```plaintext
ace-dt mirror gather SOURCES-FILE IMAGE [flags]
```

## Examples

```sh

ace-dt mirror gather repos.list reg.example.com/project/repo:sync-45

To gather with custom annotations:
ace-dt mirror gather repos.list reg.example.com/project/repo:sync-45 --annotations=key1=value1,key2=value2

To gather to a repository that does not support nested indexes:
ace-dt mirror gather repos.list reg.example.com/project/repo:sync-45 --index-fallback

To gather to a repository and only include manifests for specific platforms:
ace-dt mirror gather repos.list reg.example.com/project/repo:sync-45 -p linux/arm/v8 -p linux/amd64
```

## Options

```plaintext
Options:
  -a, --annotations stringToString   Define any additional annotations to add to the index of the gather repository.
      --debug string                 Puts UI into debug mode, dumping all UI events to the given path.
  -h, --help                         help for gather
      --index-fallback               Tells ace-dt to add indexes in annotations for registries that do not support nested indexes (i.e., not OCI 1.1 compliant).  This makes the references to the sub-indexes not real references therefore a garbage collection process might incorrectly delete the sub-indexes.  Therefore, this should only be used when necessary (e.g., when targeting Artifactory).
      --no-term                      Disable terminal support for fancy printing
  -p, --platforms strings            Only gather images that match the specified platform(s). Warning: This will modify the manifest digest/reference.
  -q, --quiet                        Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
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
