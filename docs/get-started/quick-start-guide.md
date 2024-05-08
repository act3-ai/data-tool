# Data Tool Quick Start Guide

## Intended Audience

This documentation is written for ACT3 researchers new to the Data Tool component of the ASCE Data application.

## Prerequisites

It is assumed that readers have:

- Completed the [ACT3 Onboarding process](https://www.git.act3-ace.com/onboarding/onboarding-prerequisites/)
- [Installed Data Tool](installation-guide.md)

## Installation

If you ran the [ACT3 Login script](https://gitlab.com/act3-ai/asce/up#act3-login) followed by installing [ASCE Tools](https://gitlab.com/act3-ai/asce/up#new-user-setup), you already have Data Tool installed.

To verify that you have Data Tool installed, run:

```sh
which ace-dt
```

> Team members who do not use Homebrew should consult the [installation guide](installation-guide.md) for alternative installation methods

## Configuration

Data Tool has a variety of optional configuration settings which are defined in a configuration file.

The default location for this file is `~/.config/ace/dt/config.yaml`.

If you ran the ACT3 Login script, you already have the `config.yaml` file in the default location with ACT3's Telemetry server pre-configured.

> If you ran the script but do not want your bottle activity registered and tracked by the ACT3 Telemetry server, edit your configuration file

To check your configuration settings, run:

```sh
ace-dt config
```

> Consult the [configuration guide](configuration-guide.md) for optional configuration settings

## Next Steps

[Learn More](../usage/user-guide.md){ .md-button }
