## ace-dt bottle status

Show status of items in the data bottle

### Synopsis

Description:
  Show status of items in the data bottle, including
  whether items are cached, changed, new, or deleted. The current working
  directory is used as the source of the bottle, but a path can 
  be provided to inspect an alternate location


```
ace-dt bottle status [flags]
```

### Options

```
  -D, --details   Show file paths within sub directories for changed files
  -h, --help      help for status
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

