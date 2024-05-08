## ace-dt mirror fetch

Retrieve selected objects according to the list provided in the sources-file

### Synopsis

Description:
  Retrieve selected objects according to the list provided in the sources-file


```
ace-dt mirror fetch sources-file dest-exists-file destination-location [flags]
```

### Options

```
      --arch string       Use ARCH instead of the running architecture for choosing images
  -c, --check             Get manifests for each entry in the source list, showing any errors
      --debug string      Puts UI into debug mode, dumping all UI events to the given path.
  -h, --help              help for fetch
      --insecure          Allow ace-dt to attempt to communicate over non-tls connections
      --no-term           Disable terminal support for fancy printing
      --os string         Use OS instead of the running OS for choosing images
      --platform string   Specify a platform for selecting images e.g. linux/amd64. Conflicts with os and arch arguments
  -q, --quiet             Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
      --variant string    See VARIANT instead of the running architecture variant for choosing images
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

