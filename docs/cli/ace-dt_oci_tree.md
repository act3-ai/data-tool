## ace-dt oci tree

Show the tree view of the OCI data graph for a remote image or a local OCI directory.

```
ace-dt oci tree [-i IMAGE | -d DIR] [flags]
```

### Examples

```
 To display the tree view of a remote image "reg.example.com/image" and its predecessor (if applicable), you can run the command:
		ace-dt oci tree --image reg.example.com/image

	To display the tree view of a remote image "reg.example.com/image" without displaying any predecessors, you can run the commands:
		ace-dt oci tree --image reg.example.com/image --predecessors=off

	To display the tree view of a remote image "reg.example.com/image" and display all predecessors (including predecessors of predecessors), you can run the commands:
		ace-dt oci tree --image reg.example.com/image --predecessors=recursive

	To display the tree view of a local OCI directory in ~/imageDir, you would run the command:
		ace-dt oci tree --dir ~/imageDir
	
```

### Options

```
  -d, --dir string            Directory path to OCI directory
  -h, --help                  help for tree
  -i, --image string          OCI reference to a remote image or image index
  -p, --predecessors string   Define how show predecessors by setting levels: shallow (default- shows one level of predecessors), recursive, or off (default "shallow")
  -s, --short-digests         For brevity, display only 12 hexadecimal digits of the digest (omits the algorithm as well) (default true)
      --show-annotations      Show annotations of manifests and descriptors
```

### Options inherited from parent commands

```
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
      --insecure                   Allow ace-dt to attempt to communicate over non-tls connections as a fallback if a registry is insecure
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [error])
```

### SEE ALSO

* [ace-dt oci](ace-dt_oci.md)	 - A command group for performing raw OCI operations

