## ace-dt bottle show

Display information about a remote or local data bottle

### Synopsis

Description:
This command connects to a registry and queries information
  about a specified bottle and tag when specified. it can 
  also show bottle information about a local bottle if the 
  path is specified.
	
The information provided is a list of files contained within the
  bottle, their digests (sha256), sizes, and a list of labels associated
  with each one.
	
This list can be filtered with selectors, similar to the pull command.
  Only files matching the provided selector will be returned, revealing 
  the expected result of pulling a data bottle with the selector set.
 
A bottle reference uses one of the forms
  by tag                <registry>/<repository>/<name>:<tag>
  by name (latest tag)  <registry>/<repository>/<name>
  by digest             <registry>/<repository>/<name>@sha256:<sha>


```
ace-dt bottle show [BOTTLE_REFERENCE] [flags]
```

### Options

```
  -u, --artifact stringArray   Retrieve only parts containing the provided public artifact type
      --empty                  retrieve empty bottle, only containing metadata
  -h, --help                   help for show
      --insecure               Allow ace-dt to attempt to communicate over non-tls connections
  -p, --part stringArray       Parts to retrieve
  -l, --selector stringArray   Provide selectors for which parts to retrieve. Format "name=value"
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

