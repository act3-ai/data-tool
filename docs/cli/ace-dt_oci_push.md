## ace-dt oci push

Push local image contents to a remote registry

### Synopsis

PATH will be read as an OCI image layout
		
		All match options must apply for an image to be included.

```
ace-dt oci push PATH IMAGE [flags]
```

### Options

```
  -h, --help                           help for push
      --index                          push a collection of images as a single index, currently required if PATH contains multiple images
      --legacy                         use legacy (a.k.a., docker) media types instead of the newer OCI media types.  This is useful for backwards compatibility with old registries (e.g., gitlab registry).
      --match-annotation stringArray   selectors to use to filter the image list based on annotations.  Only applicable to OCI format.
                                       	To filter by original image name use "original=busybox".
      --match-digest string            digest of image to select.  Only applicable to OCI format.
      --match-name string              name of image to select (the one with "org.opencontainers.image.ref.name" annotation that matches).  Only applicable to OCI format.
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

