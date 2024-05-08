## ace-dt git to-oci

Store a git repository as an OCI sync artifact.

### Synopsis


		Update or create a new OCI artifact with all commits leading up to and including the commits referenced by the rev-list. In a typical scenario rev-list is a list of git tag and head references. See the git rev-list manual for more information on the rev-list argument.
		
		A to-oci action strictly updates the refs provided during its execution. The resulting OCI artifact may contain additional references only if a prior to-oci action has occurred. All references included in the OCI artifact will be accessible in the git repository updated with ace-dt git from-oci, see ace-dt git from-oci for more details. 
		

```
ace-dt git to-oci GIT_REMOTE OCI_REFERENCE REV-LIST... [flags]
```

### Examples

```
Create a new commit manifest with the v1.6.1 tag reference:
		$ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/my/libgit2:sync v1.6.1

		Overwrite an existing commit manifest with a new base layer with the 
		main head reference (does not include the v1.6.1 tag ref):
		$ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/my/libgit2:sync main --clean

		Include git LFS tracked files:
		$ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/my/libgit2:sync main --lfs
		
```

### Options

```
      --clean   Start a clean commit manifest regardless whether or not a tag exists in the target repository. This will overwrite an existing manifest.
  -h, --help    help for to-oci
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

