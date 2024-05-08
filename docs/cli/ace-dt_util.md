## ace-dt util

A command group for common bottle utility functions

### Synopsis

Description:
  This command group provides shortcut commands for useful
  bottle related tasks, such as clearing bottle cache or generating
  auto completion scripts.


### Options

```
  -h, --help   help for util
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
* [ace-dt util demux](ace-dt_util_demux.md)	 - Demultiplex stdin into the files specified
* [ace-dt util filter](ace-dt_util_filter.md)	 - Print a filter to use when pretty printing logs
* [ace-dt util gen-key-pair](ace-dt_util_gen-key-pair.md)	 - generates a key pair used for signing/verifying data bottles, writing them to the destination path.
* [ace-dt util mbuffer](ace-dt_util_mbuffer.md)	 - 
* [ace-dt util mux](ace-dt_util_mux.md)	 - Multiplex via interleaving, the inputs into stdout in such a way that they can be recovered with demux
* [ace-dt util prune](ace-dt_util_prune.md)	 - Prune bottle cache, removing least recently used files

