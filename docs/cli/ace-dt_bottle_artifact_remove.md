## ace-dt bottle artifact remove

Removes item from public artifact list of a bottle

### Synopsis

Description:
  Removes an item from public artifact list of a bottle using the file's path.


```
ace-dt bottle artifact remove [PATH] [flags]
```

### Examples

```

Remove artifact mnist_public.zip  nested in /dataset directory of current bottle:
	ace-dt bottle artifact remove dataset/mnist_public.zip
	
Remove artifact kaggle_data.csv from bottle at path my/bottle/path:
	ace-dt bottle artifact rm kaggle_data.csv -d my/bottle/path

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

* [ace-dt bottle artifact](ace-dt_bottle_artifact.md)	 - A command group for bottle artifacts operations

