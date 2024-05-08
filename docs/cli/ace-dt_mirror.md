## ace-dt mirror

A command group for performing mirror operations such as fetch and push

### Synopsis

Description
  A command group for performing mirror operations such as fetch and push.


### Options

```
  -h, --help        help for mirror
  -r, --recursive   recursively copy the referrers
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
* [ace-dt mirror clone](ace-dt_mirror_clone.md)	 - A command that copies images listed in SOURCES-FILE according to the mapper.
* [ace-dt mirror deserialize](ace-dt_mirror_deserialize.md)	 - Deserializes OCI images from SOURCE-FILE and writes them to IMAGE.
* [ace-dt mirror gather](ace-dt_mirror_gather.md)	 - Efficiently copies images listed in SOURCES-FILE to the IMAGE
* [ace-dt mirror scatter](ace-dt_mirror_scatter.md)	 - A command that scatters images to destination registries defined in the MAPPER
* [ace-dt mirror serialize](ace-dt_mirror_serialize.md)	 - Serialize image data from IMAGE to DEST assuming that all blobs in the EXISTING-IMAGE(s) do not need to be sent.

