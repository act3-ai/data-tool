---
title: ace-dt mirror archive
description: Efficiently copies images listed in SOURCES-FILE to the DEST-FILE in TAR format
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt mirror archive

Efficiently copies images listed in SOURCES-FILE to the DEST-FILE in TAR format

## Synopsis

Efficiently copies images listed in SOURCES-FILE to the DEST-FILE in TAR format. If it is a tape archive better performance can be had by setting --buffer-size=1Gi or larger.  The tar file can also be written to the tape after serialization is completed (see "ace-dt util mbuffer").
		Because this is a combination of mirror gather and mirror serialize, it inherits all of the flags and options defined in those commands.
		EXISTING-IMAGE(s) are images that we use to extract blob references from to determine if we need to serialize the blob.
		
SOURCES-FILE is a text file with one OCI image reference per line.  Lines that begin with # are ignored. 
Labels can be added to each source in the SOURCES-FILE by separating with a comma and following a key=value format. These will be added as annotations to that manifest:
reg.example.com/library/source1,component=core,module=test

DEST-FILE is the name of the TAR file to be created on the local system.

The optional reference flag is a sync tag to assign to the archive when it is stored in CAS. E.g., "sync-1". 


## Usage

```plaintext
ace-dt mirror archive SOURCES-FILE DEST-FILE [flags]
```

## Examples

```sh

To archive all the images in a file named sources.list to a local file.tar, tagging it as sync-3, you can use
ace-dt mirror archive sources.list file.tar

To specify a directory ./test to cache the oci images locally before archiving them, you can use
ace-dt mirror archive sources.list file.tar --cache ./test

```

## Options

```plaintext
Options:
  -a, --annotations stringToString         Define any additional annotations to add to the index of the gather repository.
  -b, --block-size bytes                   Block size used for writes.  Si suffixes are supported. (default 1.0 MB (1048576 B))
  -m, --buffer-size bytes                  Size of the memory buffer. Si suffixes are supported. (default 0 B (0 B))
      --checkpoint string                  Save checkpoint file to file.  Can be provided to --resume-from and --resume-from-checkpoint to continue an incomplete serialize operation from where it left off.
      --compression string                 Supports zstd and gzip compression methods. (Default behavior is no compression.)
      --debug string                       Puts UI into debug mode, dumping all UI events to the given path.
      --existing-from-checkpoint strings   List of checkpoint files and their offsets. e.g, checkpoint.txt:12345, checkpoint2.txt:23456
  -h, --help                               help for archive
      --hwm int                            Percentage of buffer to fill before writing (default 90)
      --index-fallback                     Tells ace-dt to add indexes in annotations for registries that do not support nested indexes (i.e., not OCI 1.1 compliant).  This makes the references to the sub-indexes not real references therefore a garbage collection process might incorrectly delete the sub-indexes.  Therefore, this should only be used when necessary (e.g., when targeting Artifactory).
      --manifest-json                      Save a manifest.json file similar to the output of 'ctr images export' (fully compatible) or 'docker image save' (not fully compatible). Recommended to be used on images gathered with one platform specified.
      --no-term                            Disable terminal support for fancy printing
  -p, --platforms strings                  Only gather images that match the specified platform(s). Warning: This will modify the manifest digest/reference.
  -q, --quiet                              Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
      --reference string                   Tag the gathered image on disk with this reference, if not set, latest will be used. (default "latest")
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
