## ace-dt bottle artifact list

Lists public artifacts in a bottle

```
ace-dt bottle artifact list [flags]
```

### Examples

```

List public artifacts of bottle in current working directory:
	ace-dt bottle artifact list

List public artifacts of bottle at the path my/bottle/path:
	ace-dt bottle artifact list -d "my/bottle/path/"

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

* [ace-dt bottle artifact](ace-dt_bottle_artifact.md)	 - A command group for bottle artifacts operations

