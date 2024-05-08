## ace-dt bottle push

Archives, compresses, and uploads bottle to an OCI registry

### Synopsis

Description:
  The files at the specified location are archived and compressed using
  Zstandard compression, and uploaded to the specified OCI registry.
	
  A bottle reference follows the form <registry>/<repository>/<name>:<tag>

  Pushing a bottle with altered data or metadata will automatically deprecate 
  the previous version (bottleID) of this bottle. This can be disabled 
  by passing the --no-deprecate flag.


```
ace-dt bottle push BOTTLE_REFERENCE [flags]
```

### Examples

```

To push the bottle TESTSET to the registry REGISTRY/REPO/NAME:TAG:
	ace-dt bottle push REGISTRY/REPO/NAME:TAG -d ./TESTSET

To add a telemetry server, and send metadata after the push, first use ace-dt config:
	ace-dt config --telemetry.url host.url.com
Then push like normal:
	ace-dt bottle push REGISTRY/REPO/NAME:TAG -d ./TESTSET

Share a bottle with other users by giving them the bottle reference
OR, share the bottle ID (--write-bottle-id flag) for Telemetry Server support.

```

### Options

```
  -z, --compression-level string   Overrides the compression level.
      --debug string               Puts UI into debug mode, dumping all UI events to the given path.
  -h, --help                       help for push
      --insecure                   Allow ace-dt to attempt to communicate over non-tls connections
      --no-deprecate               Disable deprecation of previous bottle version
  -n, --no-overwrite               Only push data if if doesn't already exist
      --no-term                    Disable terminal support for fancy printing
  -q, --quiet                      Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
      --telemetry string           Overrides the telemetry server configuration with the single telemetry server URL provided.  
                                   Modify the configuration file if multiple telemetry servers should be used or if auth is required.
      --write-bottle-id string     File path to write the bottle ID after a bottle operation
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

