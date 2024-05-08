## ace-dt bottle annotate list

Lists annotation associated with a bottle

```
ace-dt bottle annotate list [flags]
```

### Examples

```
 
List annotation on bottle at path <my/bottle/path>:
	ace-dt bottle annotate list foo=bar -d my/bottle/path

List annotation on bottle in current working directory:
	ace-dt bottle annotate list

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

* [ace-dt bottle annotate](ace-dt_bottle_annotate.md)	 - (advanced) Adds or removes an annotation as key-value pair to specified bottle

