---
title: ace-dt oci tree
description: Show the tree view of the OCI data graph for a remote image or a local OCI directory.
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt oci tree

Show the tree view of the OCI data graph for a remote image or a local OCI directory.

## Synopsis

IMAGE is an OCI image reference.
If --oci-layout is set then the positionaly argument, OCILAYOUT, is used to specify an OCI-Layout directory.  It may be specified as a path and tag (path/to/dir:tag) or a path and digest (path/to/dir@sha256:deedbeef...).

## Usage

```plaintext
ace-dt oci tree [--oci-layout] [IMAGE|OCILAYOUT] [flags]
```

## Examples

```sh
 To display the tree view of a remote image "reg.example.com/image" and referrers:
		ace-dt oci tree reg.example.com/image

	To display the tree view of a remote image "reg.example.com/image" with all predecessors (not just referrers):
		ace-dt oci tree reg.example.com/image --only-referrers=true

	To display the tree view of a local OCI directory in ~/imageDir at at my-tag:
		ace-dt oci tree --oci-layout ~/imageDir:my-tag
	
```

## Options

```plaintext
Options:
      --artifact-type string   Limit predecessors to this artifact type
      --depth int              Maximum depth of the tree to display (default 10)
  -h, --help                   help for tree
      --oci-layout             Argument is a path and tag/digest in OCI image layout format
      --only-referrers         When true this will only show referrers (those who's subject field matches the node).  When false this will display all known immediate predecessors irregardless of if reference is using the subject field. (default true)
  -s, --short-digests          For brevity, display only 12 hexadecimal digits of the digest (omits the algorithm as well) (default true)
      --show-annotations       Show annotations of manifests and descriptors
      --show-blobs             Display the blob and config descriptors.  Still shows subjects. (default true)
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
