## ace-dt bottle part

A command group for bottle part operations

### Synopsis

Description: 
  This command group provides subcommands for interacting
  with metadata of bottle parts. You can enumerate the parts in this bottle 
  using the list command, and modify them using the label command. 

To add or remove parts, see the the 'bottle commit' command


### Options

```
  -h, --help   help for part
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
* [ace-dt bottle part label](ace-dt_bottle_part_label.md)	 - add key-value pair as a label to specified bottle part
* [ace-dt bottle part list](ace-dt_bottle_part_list.md)	 - show information about parts that are in this bottle

