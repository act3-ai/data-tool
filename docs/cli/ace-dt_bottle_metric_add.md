## ace-dt bottle metric add

add metric information to a bottle

```
ace-dt bottle metric add [METRIC] [VALUE] [flags]
```

### Examples

```

Add metric precision of value 0.87 to bottle in current directory:
	ace-dt bottle metric add "precision" 0.87

Add metric recall of value 0.68 to bottle to bottle at path <my/bottle/path>:
	ace-dt bottle metric add "recall" 0.68 --desc="recall of car classifier" -d my/bottle/path

Add metric loss with a negative value of -3.14 (use '--' when value is negative):
	ace-dt bottle metric --desc="loss value" -- loss "-3.14"

```

### Options

```
      --desc string   add description of metric
  -h, --help          help for add
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

