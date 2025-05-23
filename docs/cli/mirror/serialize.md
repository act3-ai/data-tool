---
title: ace-dt mirror serialize
description: Serialize image data from IMAGE to DEST assuming that all blobs in the EXISTING-IMAGE(s) do not need to be sent.
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt mirror serialize

Serialize image data from IMAGE to DEST assuming that all blobs in the EXISTING-IMAGE(s) do not need to be sent.

## Synopsis

IMAGE is a reference to an OCI image index to use as the source.  All the images in the image index will be sent to DEST.
DEST is a tar file or a tape archive.  If it is a tape archive better performance can be had by setting --buffer-size=1Gi or larger.  The tar file can also be written to the tape after serialization is completed (see "ace-dt util mbuffer").
EXISTING-IMAGE(s) are images that we use to extract blob references from to determine if we need to serialize the blob.

Checkpointing can be accomplished by added the --checkpoint flag.
If serialize fails for any reason, provide the --resume-from-checkpoint flag with the checkpoint file from the previous run.  Also inspect the media (file size or tape archive position, to determine a conservative (lower value is more conservative) for the number of bytes that were properly written to the media and provide that to --resume-from-offset.

## Usage

```plaintext
ace-dt mirror serialize IMAGE DEST [EXISTING-IMAGE...] [flags]
```

## Examples

```sh
ace-dt mirror serialize reg.example.com/project/repo:sync-45 /dev/nst0 reg.example.com/project/repo:complete
```

## Options

```plaintext
Options:
  -b, --block-size bytes                   Block size used for writes.  Si suffixes are supported. (default 1.0 MB (1048576 B))
  -m, --buffer-size bytes                  Size of the memory buffer. Si suffixes are supported. (default 0 B (0 B))
      --checkpoint string                  Save checkpoint file to file.  Can be provided to --resume-from and --resume-from-checkpoint to continue an incomplete serialize operation from where it left off.
      --compression string                 Supports zstd and gzip compression methods. (Default behavior is no compression.)
      --debug string                       Puts UI into debug mode, dumping all UI events to the given path.
      --existing-from-checkpoint strings   List of checkpoint files and their offsets. e.g, checkpoint.txt:12345, checkpoint2.txt:23456
  -h, --help                               help for serialize
      --hwm int                            Percentage of buffer to fill before writing (default 90)
      --manifest-json                      Save a manifest.json file similar to the output of 'ctr images export' (fully compatible) or 'docker image save' (not fully compatible). Recommended to be used on images gathered with one platform specified.
      --no-term                            Disable terminal support for fancy printing
  -q, --quiet                              Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
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
