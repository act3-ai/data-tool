## ace-dt bottle init

Initialize metadata and tracking for a data bottle

### Synopsis

Description:
  With a specified root directory, initialize metadata relevant for the bottle.  
  The current working directory is used by default, and the data bottle directory
  is created if it currently does not exist.  An ".dt/entry.yaml" file is created 
  in the data bottle directory that can be edited to fill in relevant data


```
ace-dt bottle init [flags]
```

### Examples

```

To initialize the current working directory:
	ace-dt bottle init

Given a directory TESTSET:
	ace-dt bottle init TESTSET

Next steps:
  - add metadata, ace-dt bottle [annotate, artifact, author, label, metric, part, source]
  - edit metadata directly, ace-dt bottle edit
  - add to or edit bottle files, then commit changes, ace-dt bottle commit
  - Push bottle, ace-dt bottle push

```

### Options

```
  -f, --force   Perform initialization even if the data bottle appears to already be initialized
  -h, --help    help for init
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

