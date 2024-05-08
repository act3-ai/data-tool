## ace-dt bottle annotate

(advanced) Adds or removes an annotation as key-value pair to specified bottle

### Synopsis

Description:
  Annotations are typically used to encode arbitrary metadata into the bottle..

  An annotation key may be up to 63 characters and must begin with a letter or number. The key may contain
  letters, numbers, punctuation characters. The value can be any string of arbitrary length.
 

```
ace-dt bottle annotate <key>=<value> [flags]
```

### Examples

```

Add annotation <foo=bar> to bottle in current working directory:
	ace-dt bottle annotate foo=bar

Remove annotation <foo> from bottle <bar> at path <my/bottle/path>:
	ace-dt bottle annotate foo- -d my/bottle/path

List all annotations for bottle in current working directory:
	ace-dt bottle annotate list

```

### Options

```
  -h, --help   help for annotate
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
* [ace-dt bottle annotate list](ace-dt_bottle_annotate_list.md)	 - Lists annotation associated with a bottle

