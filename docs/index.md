---
hide:
  - navigation
  # - toc
---

# Welcome to ASCE Data Tool Docs

[![version](https://gitlab.com/act3-ai/asce/data/tool/-/badges/release.svg)](https://gitlab.com/act3-ai/asce/data/tool/-/releases)

Data Tool is a command line interface (CLI) designed for researchers and AI/ML practitioners who manage large amounts of data, such as datasets for experimentation or machine learning models.

Data Tool supports the practice of [responsible AI (rAI)](./usage/user-guide.md#ethics) through the concept of the [ASCE Data Bottle](./usage/user-guide.md#data-bottles), which uses the OCI standard to encapsulate data sets as immutable images while providing a way to associate a variety of metadata fields with each bottle.

Features include:

- An Open Container Initiative (OCI) data model approach that is used to package data sets as data bottles
- Ability to assign rich metadata to data bottles
- Optional configuration for pushing data bottles as OCI artifacts to one or more registries
- Optional configuration with one or more telemetry server(s) to make data bottles discoverable, show bottle lineage, and allow users to compare bottles
- A Custom Container Storage Interface (CSI) Driver that converts ASCE Data data bottles into volumes for usability in testing environments
- Ability to mirror and distribute OCI images efficiently

Data Tool is supported and tested on Linux, macOS, and WSL2 running Ubuntu 22.04 (Windows is only supported through WSL2).

Data Tool is portable because it is a self-contained statically-linked executable.

[Get Started](./get-started/quick-start-guide.md){ .md-button }

<!-- ## Integrations

Data Tool is included by default in the Hub cluster templates for VSCode Server and Custom Jupyter with the `ace-dt` command immediately available. -->

<!-- [Get started](quick-start-guide.md){ .md-button .md-button--primary }
[Learn more](user-guide.md){ .md-button } -->
