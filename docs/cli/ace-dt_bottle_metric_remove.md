## ace-dt bottle metric remove

remove metric entry from a bottle

### Synopsis

Description: 
  remove a metric entry from from a bottle. The metric to be 
  removed is identified by using the name of the metric as a key


```
ace-dt bottle metric remove [NAME] [flags]
```

### Examples

```

Remove metric <mse> from bottle in current working directory:
	ace-dt bottle metric remove "mse" 
  
Remove source <f1-score> from bottle at path <my/bottle/path>:
	ace-dt bottle metric rm "f1-score" -d my/bottle/path

```

### Options

```
  -h, --help   help for remove
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

* [ace-dt bottle metric](ace-dt_bottle_metric.md)	 - A command group for bottle metric operations

