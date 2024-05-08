## ace-dt util demux

Demultiplex stdin into the files specified

### Synopsis

The data will be appended to the files given and created if they do not exist.

```
ace-dt util demux [out1 out2 ...]  [flags]
```

### Examples

```

Interleave file1, file2, and file3 onto the same stream and separate them back out.  The file1 and file1-dup will be identical.  Likewise for the other files.
	ace-dt util mux file1 file2 file3 | ace-dt util demux file1-dup file2-dup file3-dup

This can be combined with named pipes or unnamed pipes:
	ace-dt util mux <(command to stream from A) <(command to stream from B) > f
	ace-dt util demux >(command to stream to A) <(command to stream to B) < f
		
Can be combined with the mbuffer subcommand:
	ace-dt util mux <(command to stream from A) <(command to stream from B) | ace-dt util mbuffer -n 6Gi > /dev/nst0

```

### Options

```
  -h, --help   help for demux
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

* [ace-dt util](ace-dt_util.md)	 - A command group for common bottle utility functions

