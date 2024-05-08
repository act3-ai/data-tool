## ace-dt util mbuffer



```
ace-dt util mbuffer [flags]
```

### Examples

```
ace-dt util mbuffer -m 6Gi > /dev/nst0
```

### Options

```
  -b, --block-size bytes    Block size used for writes.  Si suffixes are supported. (default 1.0 MB (1048576 B))
  -m, --buffer-size bytes   Size of the memory buffer. Si suffixes are supported. (default 0 B (0 B))
  -h, --help                help for mbuffer
      --hwm int             Percentage of buffer to fill before writing (default 90)
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

