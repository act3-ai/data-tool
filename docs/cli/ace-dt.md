## ace-dt

data management tool for bottles and artifacts

### Synopsis

Description:
  Provides data transfer facilities for obtaining data bottles based 
  on a data registry, as well as capabilities for adding new data bottles to a
  data registry.
	


### Options

```
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -h, --help                       help for ace-dt
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [error])
```

### SEE ALSO

* [ace-dt bottle](ace-dt_bottle.md)	 - A command group for common data bottle operations
* [ace-dt completion](ace-dt_completion.md)	 - Generate completion script
* [ace-dt config](ace-dt_config.md)	 - Show the current configuration
* [ace-dt gendocs](ace-dt_gendocs.md)	 - Generate documentation for the tool in various formats
* [ace-dt git](ace-dt_git.md)	 - A command group for performing Git to OCI or OCI to Git operations
* [ace-dt info](ace-dt_info.md)	 - View detailed documentation for the tool
* [ace-dt login](ace-dt_login.md)	 - Provide authentication credentials for OCI push and pull operations
* [ace-dt logout](ace-dt_logout.md)	 - Logout from a remote registry
* [ace-dt mirror](ace-dt_mirror.md)	 - A command group for performing mirror operations such as fetch and push
* [ace-dt oci](ace-dt_oci.md)	 - A command group for performing raw OCI operations
* [ace-dt pypi](ace-dt_pypi.md)	 - A command group for performing python package syncing operations
* [ace-dt util](ace-dt_util.md)	 - A command group for common bottle utility functions
* [ace-dt version](ace-dt_version.md)	 - Print the version

