## ace-dt bottle part list

show information about parts that are in this bottle

### Synopsis

Description:
This commands shows information about parts that are in this bottle.
By default, the parts information shown are name, size, and labels.
User has the options of showing digest of a part, in lieu of labels,
by passing in the flag --digest, -D


```
ace-dt bottle part list [flags]
```

### Examples

```

List parts that are in bottle at current working directory:
	ace-dt bottle part list

List parts that are in the bottle at path <my/bottle/path>:
	ace-dt bottle part ls -d my/bottle/path
 
List parts that are in the bottle with digest information:
	ace-dt bottle part list -D 

```

### Options

```
  -D, --digest   show parts information with digest
  -h, --help     help for list
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

