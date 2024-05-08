## ace-dt gendocs md

Generate documentation in Markdown format

```
ace-dt gendocs md [dir] [flags]
```

### Options

```
  -f, --flat            generate docs in a flat directory structure
  -h, --help            help for md
  -i, --index           generate a README.md index file (default true)
      --only-commands   only generate command documentation
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

* [ace-dt gendocs](ace-dt_gendocs.md)	 - Generate documentation for the tool in various formats

