## ace-dt bottle

A command group for common data bottle operations

### Synopsis

Description:
  This command group provides subcommands for common
  bottle operations, including push, pull, commit and init.
  Use label command to add bottle level labels.
  Use part command to add part level labels.


### Options

```
  -d, --bottle-dir string   Specify bottle directory (default "/builds/ace/data/tool")
  -h, --help                help for bottle
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
* [ace-dt bottle annotate](ace-dt_bottle_annotate.md)	 - (advanced) Adds or removes an annotation as key-value pair to specified bottle
* [ace-dt bottle artifact](ace-dt_bottle_artifact.md)	 - A command group for bottle artifacts operations
* [ace-dt bottle author](ace-dt_bottle_author.md)	 - A command group for bottle author operations
* [ace-dt bottle commit](ace-dt_bottle_commit.md)	 - Processes and commits local changes to a bottle
* [ace-dt bottle delete](ace-dt_bottle_delete.md)	 - Remove a bottle from remote oci storage
* [ace-dt bottle describe](ace-dt_bottle_describe.md)	 - Adds a description to specified bottle
* [ace-dt bottle edit](ace-dt_bottle_edit.md)	 - Open a data bottle configuration in the system editor
* [ace-dt bottle gui](ace-dt_bottle_gui.md)	 - Open browser to a local web GUI for editing a bottle
* [ace-dt bottle init](ace-dt_bottle_init.md)	 - Initialize metadata and tracking for a data bottle
* [ace-dt bottle label](ace-dt_bottle_label.md)	 - add key-value pair as a label to specified bottle
* [ace-dt bottle metric](ace-dt_bottle_metric.md)	 - A command group for bottle metric operations
* [ace-dt bottle part](ace-dt_bottle_part.md)	 - A command group for bottle part operations
* [ace-dt bottle pull](ace-dt_bottle_pull.md)	 - Retrieves a bottle from remote OCI storage
* [ace-dt bottle push](ace-dt_bottle_push.md)	 - Archives, compresses, and uploads bottle to an OCI registry
* [ace-dt bottle show](ace-dt_bottle_show.md)	 - Display information about a remote or local data bottle
* [ace-dt bottle sign](ace-dt_bottle_sign.md)	 - Signs a bottle manifest digest with a private key, writing the signature to the .signature directory within the bottle directory.
* [ace-dt bottle source](ace-dt_bottle_source.md)	 - A command group for bottle source operations
* [ace-dt bottle status](ace-dt_bottle_status.md)	 - Show status of items in the data bottle
* [ace-dt bottle verify](ace-dt_bottle_verify.md)	 - Verifies all local signatures of a bottle's manifest digest.

