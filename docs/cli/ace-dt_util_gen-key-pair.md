## ace-dt util gen-key-pair

generates a key pair used for signing/verifying data bottles, writing them to the destination path.

### Synopsis

Description:
  Generates an ECDSA public-private key pair, which is used for signing and verifying
  signatures of manifest digests. The public key is written to DESTINATION_PATH/bottle.pub
  while the private key is written to DESTINATION_PATH/bottle.key. The prefix "bottle" may
  be optionally changed with the --prefix flag. Any existing key names will be overwritten with
  the new key pair.


```
ace-dt util gen-key-pair DESTINATION_PATH [flags]
```

### Examples

```

	To generate keys with default naming (bottle.key, bottle.pub):
gen-key-pair DESTINATION_PATH
	To generate keys with custom naming (<prefix>.key, <prefix>.pub):
gen-key-pair DESTINATION_PATH --prefix PREFIX

```

### Options

```
  -h, --help            help for gen-key-pair
  -p, --prefix string   Set the prefix of the key names. Default is 'bottle'. (default "bottle")
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

