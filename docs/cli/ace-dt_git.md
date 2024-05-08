## ace-dt git

A command group for performing Git to OCI or OCI to Git operations

### Options

```
      --alt-git string      Provide a path to an alternative git executable.
  -h, --help                help for git
      --lfs                 Include git LFS tracked files.
      --lfs-server string   Directly specify LFS server URL or override a repositories config lfs.url if it already exists.
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
* [ace-dt git from-oci](ace-dt_git_from-oci.md)	 - Pull updates from an OCI artifact into a git repository.
* [ace-dt git list-refs](ace-dt_git_list-refs.md)	 - List all references in an OCI sync artifact.
* [ace-dt git to-oci](ace-dt_git_to-oci.md)	 - Store a git repository as an OCI sync artifact.

