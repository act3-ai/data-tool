## ace-dt mirror batch-deserialize

Deserialize multiple tar files within SYNC-REPOSITORY to DESTINATION, while avoiding pushing blobs that already exist.

### Synopsis

SYNC-REPOSITORY is the repository holding the tar files to be deserialized. All tar files will be deserialized to DESTINATION.
DESTINATION is a remote repository WITHOUT a tag. Tags will be automatically generated based off of the image name within the tar file name.
For example, given a file "SYNC-REPOSITORY/0-image1.tar", the blobs will be deserilaized to DESTINATION and tagged as "image1".

```
ace-dt mirror batch-deserialize SYNC-REPOSITORY DESTINATION [flags]
```

### Examples

```
ace-dt mirror batch-deserialize sync/data/ registry.example.com/image
```

### Options

```
      --debug string           Puts UI into debug mode, dumping all UI events to the given path.
  -h, --help                   help for batch-deserialize
      --no-term                Disable terminal support for fancy printing
  -q, --quiet                  Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
      --sync-filename string   used to override the default sync-file name. (default "successful-syncs.csv")
```

### Options inherited from parent commands

```
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -r, --recursive                  recursively copy the referrers
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [error])
```

### SEE ALSO

* [ace-dt mirror](ace-dt_mirror.md)	 - A command group for performing mirror operations such as fetch and push

