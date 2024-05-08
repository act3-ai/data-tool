## ace-dt bottle edit

Open a data bottle configuration in the system editor

### Synopsis

Description:
  Open a data bottle configuration in the system editor
	
The editor opened is either the $EDITOR environment variable, or 
  vim if no editor is specified there.
	
By default, the current directory is searched for a data bottle configuration,
  but a path to a data bottle may be specified for an alternate location..


```
ace-dt bottle edit [flags]
```

### Options

```
  -h, --help   help for edit
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

