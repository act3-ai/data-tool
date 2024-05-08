## ace-dt bottle gui

Open browser to a local web GUI for editing a bottle

### Synopsis

Description:
		Open your default web browser to a page to edit the given bottle.
		This command will run a local webserver to show the GUI in your browser.


```
ace-dt bottle gui [flags]
```

### Options

```
  -h, --help            help for gui
      --listen string   Address and port for the server to listen for new connections (default "localhost:0")
      --no-browser      Automatically open browser to the GUI
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

