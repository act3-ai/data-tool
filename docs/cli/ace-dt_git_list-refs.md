## ace-dt git list-refs

List all references in an OCI sync artifact.

### Synopsis

Lists all head and tag references in an OCI sync artifact along with the commits they reference.

```
ace-dt git list-refs OCI_REFERENCE [flags]
```

### Examples

```
List all references at reg.example.com/my/libgit2:sync:
		$ ace-dt git list-refs reg.example.com/my/libgit2:sync
```

### Options

```
  -h, --help   help for list-refs
```

### Options inherited from parent commands

```
      --alt-git string             Provide a path to an alternative git executable.
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
      --lfs                        Include git LFS tracked files.
      --lfs-server string          Directly specify LFS server URL or override a repositories config lfs.url if it already exists.
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [error])
```

### SEE ALSO

* [ace-dt git](ace-dt_git.md)	 - A command group for performing Git to OCI or OCI to Git operations

