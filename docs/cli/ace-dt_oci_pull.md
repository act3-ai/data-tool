## ace-dt oci pull

Pull remote images by reference and store their contents locally and store in the standard OCI image layout

```
ace-dt oci pull IMAGE... DIRECTORY [flags]
```

### Options

```
  -c, --cache_path string   Path to cache image layers
  -h, --help                help for pull
```

### Options inherited from parent commands

```
      --config stringArray   configuration file location (setable with env "ACE_DT_CONFIG").
                             The first configuration file present is used.  Others are ignored.
                              (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
      --insecure             Allow ace-dt to attempt to communicate over non-tls connections as a fallback if a registry is insecure
  -v, --verbosity int8[=1]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
```

### SEE ALSO

* [ace-dt oci](ace-dt_oci.md)	 - A command group for performing raw OCI operations

