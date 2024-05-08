## ace-dt util filter jq

Print a jq filter to use when pretty printing logs

### Synopsis

To use this filter you must have "jq" installed.
		ace-dt util filter jq > log.jq
	ace-dt ... | jq -j -f log.jq

```
ace-dt util filter jq [flags]
```

### Options

```
  -h, --help   help for jq
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

* [ace-dt util filter](ace-dt_util_filter.md)	 - Print a filter to use when pretty printing logs

