## ace-dt login

Provide authentication credentials for OCI push and pull operations

### Synopsis

Description:
  Provide authentication credentials for OCI push and pull operations
	
  This will prompt for a user name and password, and will authenticate to the
  provided registry. If successful, the credentials will be used for future 
  interactions with that registry by adding an entry to your ~/.docker/config.json.


```
ace-dt login REGISTRY [flags]
```

### Options

```
  -h, --help              help for login
  -p, --password string   password credential for login
  -u, --username string   username credential for login
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

* [ace-dt](ace-dt.md)	 - data management tool for bottles and artifacts

