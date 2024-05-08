## ace-dt mirror clone

A command that copies images listed in SOURCES-FILE according to the mapper.

### Synopsis

A command that copies images listed in SOURCES-FILE according to the mapper.
		
SOURCES-FILE is a text file with one OCI image reference per line.  Lines that begin with # are ignored.

The MAPPER types currently supported are nest, first-prefix (csv format), digests (csv format) and go-template.
The format of MAPPER is MAP-TYPE=MAP-ARG

If MAP-TYPE is "nest" then clone will nest all the images under MAP-ARG.
For example, is MAP-ARG is "reg.other.com" then a gathered image "foo.com/bar" will map to "reg.other.com/foo.com/bar".

Passing a first-prefix MAPPER requires a csv file that has formatted lines of: source,destination. 
The ace-dt mirror clone will send the source reference to the first prefix match that it makes.
This format also allows defining the source as a digest that is present in the source repository.

Passing a digests MAP-FILE requires a csv file that has formatted lines of: digest-string, destination.
Scatter will send each digest to the locations defined in the map file provided. 

Passing a go-template MAP-FILE allows a greater deal of flexibility in how references can be pushed
to destination repositories. Sprig functions are currently supported which allows for matching by 
prefix, digest, media-type, regex, etc. 

Example csv and go template files are located in the pkg/actions/mirror/test repository.
		

```
ace-dt mirror clone SOURCES-FILE MAPPER [flags]
```

### Examples

```
To clone and scatter all the images contained in "sources.list" you can use
ace-dt mirror clone sources.list nest=ref.other.com/mirror

ace-dt mirror clone sources.list go-template=mapping.tmpl
ace-dt mirror clone sources.list first-prefix=mapping.csv
ace-dt mirror clone sources.list digests=mapping.csv
ace-dt mirror clone sources.list longest-prefix=mapping.csv
ace-dt mirror clone sources.list all-prefix=mapping.csv
```

### Options

```
      --check          Dry run- do not actually send to destination repositories
      --debug string   Puts UI into debug mode, dumping all UI events to the given path.
  -h, --help           help for clone
      --no-term        Disable terminal support for fancy printing
  -q, --quiet          Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
```

### Options inherited from parent commands

```
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,HOMEDIR/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -r, --recursive                  recursively copy the referrers
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [error])
```

### SEE ALSO

* [ace-dt mirror](ace-dt_mirror.md)	 - A command group for performing mirror operations such as fetch and push

