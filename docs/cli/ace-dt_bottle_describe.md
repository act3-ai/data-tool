## ace-dt bottle describe

Adds a description to specified bottle

### Synopsis

Description:
  Add description and useful details to a bottle. Short description can be added
  directly from the command line, and long description can be added from a specified text 
  file.
 

```
ace-dt bottle describe [DESCRIPTION TEXT] [flags]
```

### Examples

```

Add a short description to bottle at path <my/bottle/path>:
  ace-dt bottle describe "The context of this bottle is foobar." -d my/bottle/path

Add description text from a file to a bottle at path <my/bottle/path>:
  ace-dt bottle describe --from-file ./my-description.txt -d my/bottle/path

```

### Options

```
  -f, --from-file string   add description text from a file to a bottle
  -h, --help               help for describe
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

