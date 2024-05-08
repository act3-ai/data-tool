## ace-dt config

Show the current configuration

```
ace-dt config [flags]
```

### Examples

```
Configuration can be modified with the following environment variables:
ACE_DT_CACHE_PRUNE_MAX: Maximum cache prune size
ACE_DT_CACHE_PATH: directory to use for caching
ACE_DT_CHUNK_SIZE: Maximum chunk size for chunked uploads (set to "0" to disable)
ACE_DT_CONCURRENT_HTTP: Maximum concurrent network connections.
ACE_DT_REGISTRY_AUTH_FILE then REGISTRY_AUTH_FILE: Docker registry auth file
ACE_DT_EDITOR then VISUAL then EDITOR: Sets the editor to use for editing bottle schema.
ACE_DT_TELEMETRY_URL: If set will overwrite the telemetry configuration to only use this telemetry server URL.  Use the config file if you need multiple telemetry servers.
ACE_DT_TELEMETRY_USERNAME: Username to use for reporting events to telemetry.

To save the config to the default location run
$ ace-dt config -s > HOMEDIR/.config/ace/dt/config.yaml
then modify the configuration file as needed.

```

### Options

```
  -h, --help     help for config
  -s, --sample   Output a sample configuration that can be used in a configuration file.
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

