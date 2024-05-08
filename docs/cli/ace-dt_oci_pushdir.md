## ace-dt oci pushdir

Push local directory as an OCI image to a remote registry

### Synopsis

PATH is a directory that will be used as the contents of the image.

```
ace-dt oci pushdir PATH IMAGE [flags]
```

### Options

```
  -h, --help                help for pushdir
      --legacy              use legacy (a.k.a., docker) media types instead of the newer OCI media types.  This is useful for backwards compatibility with old registries (e.g., gitlab registry).
      --platform platform   platform to use for uploading image.  The format is os/arch (e.g., linux/amd64) (default all)
      --reproducible        Makes the uploaded artifact have a consistent (reproducible) digest.  This removes timestamps.  The benefit is that this produces layers that are possibly better suited for mirroring.
```

### Options inherited from parent commands

```
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
      --insecure                   Allow ace-dt to attempt to communicate over non-tls connections as a fallback if a registry is insecure
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [error])
```

### SEE ALSO

* [ace-dt oci](ace-dt_oci.md)	 - A command group for performing raw OCI operations

