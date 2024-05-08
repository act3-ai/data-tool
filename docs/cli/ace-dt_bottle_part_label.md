## ace-dt bottle part label

add key-value pair as a label to specified bottle part

### Synopsis

Description: 
  Add key-value pair as a label to specified bottle part

  A label key and value must begin with a letter or number, and may contain 
  letters, numbers, hyphens, dots, and underscores, up to  63 characters each.


```
ace-dt bottle part label <key>=<value>... PATH... [flags]
```

### Examples

```

Add label <foo=bar> to part <myPart.txt> bottle in current working directory:
	ace-dt bottle part label foo=bar myPart.txt

Add label <foo=bar> to many parts in current working directory:
	ace-dt bottle part label foo=bar myPart.txt myPicture.jpg myModel.model

Remove label "foo" from part "myPart.txt" in bottle at path "my/bottle/path":
	ace-dt bottle part label foo- my/bottle/path/myPart.txt -d my/bottle/path

```

### Options

```
  -h, --help   help for label
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

* [ace-dt bottle part](ace-dt_bottle_part.md)	 - A command group for bottle part operations

