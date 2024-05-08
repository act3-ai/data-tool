## ace-dt mirror push

push objects to the locations specified in dest-file from the source-location, recording transmitted blobs to blob-ledger

### Synopsis

Description:
  push objects to the locations specified in dest-file from the source-location, recording transmitted blobs to blob-ledger


```
ace-dt mirror push source-location blob-ledger dest-file [flags]
```

### Options

```
  -c, --check              Display repository manifest destinations, but do not push
      --debug string       Puts UI into debug mode, dumping all UI events to the given path.
  -h, --help               help for push
      --insecure           Allow ace-dt to attempt to communicate over non-tls connections
      --no-term            Disable terminal support for fancy printing
  -q, --quiet              Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
      --telemetry string   Overrides the telemetry server configuration with the single telemetry server URL provided.  
                           Modify the configuration file if multiple telemetry servers should be used or if auth is required.
```

### Options inherited from parent commands

```
      --config stringArray   configuration file location (setable with env "ACE_DT_CONFIG").
                             The first configuration file present is used.  Others are ignored.
                              (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -v, --verbosity int8[=1]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
```

### SEE ALSO

* [ace-dt mirror](ace-dt_mirror.md)	 - A command group for performing mirror operations such as fetch and push

