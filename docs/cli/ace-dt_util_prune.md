## ace-dt util prune

Prune bottle cache, removing least recently used files

### Synopsis

Description:
  Prune bottle cache, removing least recently used files.  Use
  maxsize option to choose a maximum size of data to keep, in MiB.


```
ace-dt util prune [flags]
```

### Options

```
  -h, --help          help for prune
  -s, --maxsize int   A maximum size to keep in the cache, in MiB (default -1)
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

* [ace-dt util](ace-dt_util.md)	 - A command group for common bottle utility functions

