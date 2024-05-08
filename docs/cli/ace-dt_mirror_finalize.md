## ace-dt mirror finalize

Finalize the destination existence file after performing a series of mirror fetch commands

### Synopsis

Description:
  Finalize the destination existence file after performing a series of mirror fetch commands

  Internally, this searches for the provided dest-file name with ".new" suffix, and removes the extension, replacing the
  dest-file as necessary.


```
ace-dt mirror finalize dest-file [flags]
```

### Options

```
      --debug string   Puts UI into debug mode, dumping all UI events to the given path.
  -h, --help           help for finalize
      --no-term        Disable terminal support for fancy printing
  -q, --quiet          Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
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

