---
title: ace-dt util simulate
description: Simulate UI components
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# ace-dt util simulate

Simulate UI components

## Synopsis

This is a debugging command to test the UI components.
		first argument is to change the number of "tasks" ran
		second argument is to change parallel task run count (max parallel)
	

## Usage

```plaintext
ace-dt util simulate [flags]
```

## Options

```plaintext
Options:
      --debug string            Puts UI into debug mode, dumping all UI events to the given path.
  -h, --help                    help for simulate
      --no-term                 Disable terminal support for fancy printing
  -r, --numCountRecursive int   Number of recursive counting tasks to run (default 2)
  -m, --numMaxParallel int      Number of tasks to run in parallel (default 10)
  -n, --numTasks int            Number of tasks to run (default 1)
  -q, --quiet                   Quiet mode.  Do not output any status to standard output.  Errors are still output to standard error.
```

## Options inherited from parent commands

```plaintext
Global options:
      --config stringArray         configuration file location (setable with env "ACE_DT_CONFIG").
                                   The first configuration file present is used.  Others are ignored.
                                    (default [ace-dt-config.yaml,/root/.config/ace/dt/config.yaml,/etc/ace/dt/config.yaml])
  -v, --verbosity strings[=warn]   Logging verbosity level (also setable with environment variable ACE_DT_VERBOSITY)
                                   Aliases: error=0, warn=4, info=8, debug=12 (default [warn])
```
