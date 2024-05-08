## ace-dt oci

A command group for performing raw OCI operations

### Options

```
  -h, --help       help for oci
      --insecure   Allow ace-dt to attempt to communicate over non-tls connections as a fallback if a registry is insecure
```

### Options inherited from parent commands

```
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [error])
```

### SEE ALSO

* [ace-dt](ace-dt.md)	 - data management tool for bottles and artifacts
* [ace-dt oci pushdir](ace-dt_oci_pushdir.md)	 - Push local directory as an OCI image to a remote registry
* [ace-dt oci tree](ace-dt_oci_tree.md)	 - Show the tree view of the OCI data graph for a remote image or a local OCI directory.

