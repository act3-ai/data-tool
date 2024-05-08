# ACE Data Tool

ACE Data Tool is the primary way for researchers, data scientists, and AI/ML practitioners to work with large data sets in the ACT3 ecosystem.

ACE Data Tool:

- Uses the concept of a data bottle to encapsulate data sets
- Allows for the identification and documentation of metadata associated with a data bottle
- Facilitates reuse of a data set used while conducting research or an experiment
- Documents the lineage (starting point, any modifications, and current state) of a data set, when connected to the Telemetry Server
- Makes heavy use of the OCI data model and content addressable storage to avoid duplicate work

Data Tool provides command groups that can be used via the CLI to:

- Create (initialize) data bottles
- Add metadata to bottles
- Commit and push bottles to OCI registries and to the ACE Telemetry Server
- Pull and use bottles for experiments, research, and AL/ML projects
- Mirror container images from one registry to another, especially secure air-gapped environments

Data Tool is designed to help researchers and AI/ML practitioners:

- **Practice responsible AI** by providing a variety of metadata options for data bottles
- **Automate tracking of data lineage** by using the ACE Telemetry Server

## Documentation

The documentation for ACE Data Tool is organized as follows:

- **[Quick Start Guide](docs/quick-start-guide.md)**: provides documentation for downloading, installing, and configuring Data Tool
- **[User Guide](docs/user-guide.md)**: provides an overview of Data Tool, introduces new users to key concepts, and explains basic usage
- **[Bottle Creator Guide](docs/bottle-creator-guide.md)**: documents the process and known best practices for creating high quality bottles
- **[Mirror User Guide](docs/mirror-user-guide.md)**: documents specific considerations and known best practices for using Data Tool's mirror features
- **[Labels and Selectors Guide](docs/labels.md)**: provides an overview of the relationship between labels created in ACE Data Tool and selectors used in ACE Telemetry Server to build queries and locate data bottles

## How to Contribute

See the **[Developer Guide](docs/developer-guide.md)** for contributing code to the ACE Data Tool repository.

## Support

- **[Mattermost channel](https://chat.git.act3-ace.com/act3/channels/act3-pt)**: create a post in the ACT3 Data Tool channel for assistance
- **[FAQ](docs/faq.md)**: see the frequently asked (and answered) questions for ACE Data Tool
<!-- TODO reactivate when functional - **[Create a GitLab issue by email](mailto:incoming+ace-data-tool-238-cpdx5kax2g659873veqpf97dt-issue@mail.act3-ace.com)** -->
