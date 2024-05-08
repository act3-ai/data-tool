## ace-dt bottle verify

Verifies all local signatures of a bottle's manifest digest.

### Synopsis

Description:
  Verifies all local signatures of the bottle's manifest digest. In order to ensure the local
  signatures are up-to-date use ace-dt bottle pull prior to signature verification.

  It is optional to provide locally discovered public keys for verification. For each local public key
  provide the path to the key. The key's fingerprint will be used to correlate keys with signatures.
  Ex: ace-dt bottle verify <path> <path> ...

  Notice:
  If the signer provided insufficient metadata to discover the appropriate public key for verification,
  it will default to the no key management system verification method - which is notably insecure with ECDSA keys.




```
ace-dt bottle verify [flags]
```

### Examples

```

To verify a manifest digest:
	ace-dt bottle verify

```

### Options

```
  -h, --help   help for verify
```

### Options inherited from parent commands

```
  -d, --bottle-dir string          Specify bottle directory (default "/builds/ace/data/tool")
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [error])
```

### SEE ALSO

* [ace-dt bottle](ace-dt_bottle.md)	 - A command group for common data bottle operations

