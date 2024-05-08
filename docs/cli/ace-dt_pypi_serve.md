## ace-dt pypi serve

Run the PyPI server

### Synopsis

This runs a PyPI (simple) compatible server pulling packages from the (OCI) REPOSITORY as needed to serve the content.

```
ace-dt pypi serve [-l PORT] REPOSITORY [flags]
```

### Options

```
  -h, --help            help for serve
  -l, --listen string   Interface and port to listen on.
                        Use :8101 to listen all on interfaces on the standard port. (default "localhost:8101")
```

### Options inherited from parent commands

```
      --allow-yanked               Do not ignore yanked distribution files
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [error])
```

### SEE ALSO

* [ace-dt pypi](ace-dt_pypi.md)	 - A command group for performing python package syncing operations

