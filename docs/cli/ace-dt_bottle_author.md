## ace-dt bottle author

A command group for bottle author operations

### Synopsis

Description:
  This command group provides subcommands for interacting
  with author information of bottle parts.


### Options

```
  -h, --help   help for author
```

### Options inherited from parent commands

```
  -d, --bottle-dir string          Specify bottle directory (default "/builds/ace/data/tool")
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [error])
```

### SEE ALSO

* [ace-dt bottle](ace-dt_bottle.md)	 - A command group for common data bottle operations
* [ace-dt bottle author add](ace-dt_bottle_author_add.md)	 - Adds author information to specified bottle
* [ace-dt bottle author list](ace-dt_bottle_author_list.md)	 - Lists author information of a bottle
* [ace-dt bottle author remove](ace-dt_bottle_author_remove.md)	 - Removes author's information from a bottle

