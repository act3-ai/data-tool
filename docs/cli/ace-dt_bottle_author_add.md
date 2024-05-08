## ace-dt bottle author add

Adds author information to specified bottle

```
ace-dt bottle author add [NAME] [EMAIL] [flags]
```

### Examples

```

Add author <John Doe> to bottle in current working directory:
	ace-dt bottle author add "John Doe" "jdoe@example.com"

Add author <Alice Wonders> to bottle at path <my/bottle/path>:
	ace-dt bottle author add "Alice Wonders" "alicew@example.com" --url="university.example.com/~awonders" -d my/bottle/path

```

### Options

```
  -h, --help         help for add
      --url string   specify author's webpage, or preferred public url
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

* [ace-dt bottle author](ace-dt_bottle_author.md)	 - A command group for bottle author operations

