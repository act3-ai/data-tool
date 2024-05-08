## ace-dt git from-oci

Pull updates from an OCI artifact into a git repository.

### Synopsis

Initialize or update a remote git repository with all commits and references contained in an OCI artifact. All updated references are printed to stdout. View the commit artifact config to see the all references which may be updated.

		Rebuild prevents overwriting existing references by default to prevent undesired results. This behavior may be overidden with the --force flag, which may be interpreted as it is used in git push.
		

```
ace-dt git from-oci OCI_REFERENCE GIT_REMOTE [flags]
```

### Examples

```
Push updates in a sync manifest to a target repository:
		$ ace-dt git from-oci reg.example.com/my/libgit2:sync https://github.com/mypersonal/libgit2
		
		Force push updates in a sync manifest to a target repository:
		$ ace-dt git from-oci reg.example.com/my/libgit2:sync https://github.com/mypersonal/libgit2 --force
		
```

### Options

```
      --force   Force updates to repository references, analogous to git push --force.
  -h, --help    help for from-oci
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

