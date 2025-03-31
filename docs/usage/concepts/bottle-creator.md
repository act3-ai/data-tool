# Bottle Creator Guide

## Intended Audience

This concept guide is written for Data Tool users who want to create and modify bottles, also known as *bottle authors* or *bottle creators*.

> Consult the [Data Tool User Guide](../user-guide.md) to review Data Tools's key concepts and common usage patterns

## Prerequisites

This guide does not have prerequisites for the intended audience.

## Overview

This guide is focused on specific considerations and known best practices for creating high quality bottles.

## Ethics

Responsible AI (rAI) requires practitioners to understand and document the data used to design, develop, or apply AI tools. Bottle creators play an integral role in developing responsible AI (rAI) for the research and development of Air Force tools and applications.

Bottle creators should follow the best practices defined in this guide to ethically represent and document their data sets.

### Best Practices

Rich and accurate metadata are essential to rAI and will enhance the overall utility of a bottle through inclusion of authors, labels, metrics and sources.

Metadata fields implemented in Data Tool enable bottle creators to accomplish this task.

### Types of Metadata

Metadata fields for **sources and authors** supplement a bottle's automatically tracked data lineage. **Authors** direct consumers to the individual human(s) responsible for creating the bottle. **Sources** direct consumers to the origins of the bottle’s contents.

**Metrics** are associated with a bottle's contents and are used to rank a bottle’s contents in comparison to other similar bottles. Bottle creators should add accurate metrics applicable to their data sets. Accurate metrics help researchers establish validity and reliability during the experimental process.

### Using Labels

Labels classify and identify bottles. Bottle creators should create key-value pairs for labels that are clear and accurate for their bottles and bottle parts. Inaccurate labels hinder users who are locating bottles and can pose an ethical concern if they misrepresent the bottle's contents.

Metadata associated with a bottle can provide vital information to AI/ML practitioners who are required to assume responsibility for understanding and correctly documenting the underlying data used to design, develop, or apply AI tools.

> ACE Data Tool cannot prevent any user from misusing data--that is the shared responsibility of bottle creators and consumers

## Bottle Creation

*Bottle creation* is the generalized phrase used to describe the multi-step process of packaging and documenting a data set.

Bottles can be saved locally or published to an OCI registry. When a bottle is pushed to a registry, it is also recorded in ACE Telemetry, which can then be used by other researchers searching for bottles.

### Workflow Overview

To package and publish a bottle to a telemetry server:

- Initialize a directory as a data bottle
- Add Metadata to the bottle
- Commit
- Push

### Initialize

A directory is transformed into a data bottle when it is initialized.

The syntax is:

```sh
ace-dt bottle init
```

> Throughout this guide, it is assumed that the current working directory is the the top-level directory containing the data that will become a bottle.

To specify a location, use the following flag to define the path to a directory:

```sh
ace-dt bottle init --bottle-dir
```

or

```sh
ace-dt bottle init -d <directory path>
```

Bottles are typically initialized:

- On a local computer
- On a computer accessed remotely
- In ACE Hub
- In a computing environment equivalent to ACE Hub
- Programmatically

After a bottle is initialized, it exists in the local environment and not yet on a registry because it has not been pushed.

Creating bottles programmatically can automate the process of adding a complete set of metadata to a bottle. See the [source code example](https://git.act3-ace.com/ace/examples/ace-hub-demo/-/blob/master/tb/tb.py#L236) for an example of how to create a bottle using Python.

Metadata should be added before the bottle is pushed to an OCI registry.

### Add Metadata

Consumers use metadata to query and pull bottles from ACE Telemetry. Rich and accurate metadata are essential for:

- Creating a high-quality bottle
- Adhering to the best practices defined as rAI

Add and edit metadata using the following command group:

```sh
ace-dt bottle <subcommand>
```

Metadata commands can be categorized based on how they document or represent a bottle for consumers. Commands listed below in **bold** can be queried in a telemetry server.

Documentation commands are:

- **`author`**: used to document bottle creator(s)
- `source`: used to identify parent data
- `artifact`: used to make small files associated with a bottle visible in a telemetry server; may include public examples to benefit bottle consumers

Representation commands are:

- **`describe`**: used to add an abstract (paragraph description) to a bottle as a supplement other metadata
- **`label`**: used to add keywords (short content descriptors) structured as key-value pairs
- **`part`**: used to add keywords corresponding to specific parts that define logical subset(s) of bottle files; can have labels applied for searchability in a telemetry server
- **`metric`**: used to add scalar benchmarks that measure a data set's performance
- `annotate`: used to add supplemental author descriptions and appendices that are relevant to a bottle but not searchable in a telemetry server

#### See Also

- [Additional Standards and Conventions](#additional-standards-and-conventions) section for a discussion of the metadata conventions defined by ACT3 for bottles that contain ML models;  - - ["Model Cards for Model Reporting"](https://arxiv.org/pdf/1810.03993.pdf) for the basis on which ACT3's metadata conventions were developed
- [Bottle Metadata Guide](../concepts/bottle-metadata.md) for syntax and usage patterns

### Commit

A bottle can be committed many times while working locally.

The syntax is:

```sh
ace-dt bottle commit
```

When a bottle is committed, the bottle and its parts are automatically compressed (if needed) and they tracked in the `.dt/entry.yaml` file.

### Push

When all relevant metadata have been added to a bottle and the last commit is made, it is ready to be pushed. Note that, unlike git, pushing a bottle will automatically commit all changes.

The syntax is:

```sh
ace-dt bottle push OCI_REF
```

Note that an OCI reference argument is required when a bottle is pushed.

Example of an OCI_REF:

```sh
<registry>/<repository>/<name>{:<tag>|@<digest>}
```

When a bottled is pushed, the OCI image of the data is sent to the registry (or registries) designated by the bottle creator.

#### Push with Telemetry

If one or more telemetry servers are configured, event data (when and who pushed the bottle) as well as bottle metadata are automatically tracked. The bottle will subsequently be discoverable by others who have access to the telemetry server(s), often using metrics or labels (see [Labels and Selectors](https://telemetry.lion.act3-ace.ai/www/about.html#labels-and-selectors)). The telemetry server only stores metadata about the bottle. The bottle itself is stored in the OCI registry, which also handles access control to the data. For more information on telemetry visit the [Telemetry Write-up](https://gitlab.com/act3-ai/asce/data/telemetry/-/blob/main/docs/writeup.md?ref_type=heads).

See the [Configuration Guide](../../get-started/configuration-guide.md) to configure one or more telemetry servers.

Example:

```sh
ACE_DT_TELEMETRY_URL=https://telemetry.lion.act3-ace.ai ace-dt bottle push OCI_REF
```

The ACT3 [telemetry server](https://telemetry.lion.act3-ace.ai) offers a graphical interface where you can discover bottles created by others.

## Other Concerns

### Immutability

Data bottles are immutable. Any changes to the data or metadata will produce a different bottle (because the digests change). If a telemetry server is configured when a new bottle is pushed, the previous bottle is automatically deprecated. The lineage of the bottle is automatically tracked in the configured telemetry server.

### Data Security

By design, the metadata of bottles (including artifacts) are relatively unrestricted.  Metadata help users searching in ACE Telemetry find bottles of interest. Because ACE Telemetry does not store the actual data, a researcher could find relevant bottles even if they do not have access to the data that are stored in a restricted OCI registry.

The registry's access controls permit the contents of a bottle (its parts) to contain sensitive information while ACE Telemetry makes the existence of those bottles discoverable but not accessible.

Care should be taken to not expose any sensitive information in the metadata of the bottle which include the artifacts (called `publicArtifacts` in the `.dt/entry.yaml` for this very reason).

### Additional Standards and Conventions

Bottles can contain ML models. Mitchell et al.'s [Model Cards](https://arxiv.org/pdf/1810.03993.pdf) paper describes a collection of concerns that should be documented in the metadata of bottled containing ML models.

ACT3 has logically mapped the conventions from Model Card to ACE Data Bottles as follows:

- "Model Details", "Intended Use", "Ethical Considerations", and "Caveats and Recommendations" belong in the bottle description unless they can be structured; if structured, then labels or annotations should be used instead
- "Factors" belong in the bottle description
- "Training Data" and "Evaluation Data" are bottle sources
- "Quantitative Analyses" are bottle artifacts
- "Metrics" are bottle metrics

### Experiments

Model training can be viewed as a data science experiment. A properly constructed bottle can capture and describe a computational experiment.

In these cases:

- Labels should be used to represent an experiment's independent variables (to keep track of inputs for the experiment and related metadata)
- Metrics should be used to represent the dependent variables (to keep track of output values from the experiment)
- Annotations should be used to represent structured information that does not fit into labels or metrics
- Sources should be used to represent the (bulk) input data for the experiment

<!-- Typically one would not include the (bulk) input data in the bottle directly. However, if the same bottle part is in both the source bottle and the experiment bottle and stored in the same registry then there would be no duplication of data for storage or data transfer. -->
