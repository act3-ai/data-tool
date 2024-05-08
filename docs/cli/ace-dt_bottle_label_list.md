## ace-dt bottle label list

list labels applied on specified bottle

```
ace-dt bottle label list [flags]
```

### Examples

```

List label on bottle at path <my/bottle/path>:
	ace-dt bottle label list -d my/bottle/path

List the label of the bottle of the current directory:
	ace-dt bottle label list	

```

### Options

```
  -h, --help   help for list
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

* [ace-dt bottle label](ace-dt_bottle_label.md)	 - add key-value pair as a label to specified bottle

