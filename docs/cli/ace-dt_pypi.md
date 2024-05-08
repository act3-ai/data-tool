## ace-dt pypi

A command group for performing python package syncing operations

### Examples

```
The first step is to fetch distribution files from remote sources.
$ ace-dt pypi to-oci reg.example.com/my/pypi numpy -l 'version.major=1,version.minor>5'

or with a requirements file
$ ace-dt pypi to-oci reg.example.com/my/pypi -r requirements.txt
		
After you have fetched you can serve up the PyPI compliant (PEP-691) package index with
$ ace-dt pypi serve reg.example.com/my/pypi

```

### Options

```
      --allow-yanked   Do not ignore yanked distribution files
  -h, --help           help for pypi
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
* [ace-dt pypi serve](ace-dt_pypi_serve.md)	 - Run the PyPI server
* [ace-dt pypi to-oci](ace-dt_pypi_to-oci.md)	 - Fetch python packages from Python package indexes and upload to the OCI-REPOSITORY as OCI images
* [ace-dt pypi to-pypi](ace-dt_pypi_to-pypi.md)	 - Pulls packages from OCI and uploads them to the python package index

