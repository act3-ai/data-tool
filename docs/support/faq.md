# Frequently Asked Questions

## What is a data bottle?

Bottles are simply data with associated metadata (e.g., authors, description, sources) that are stored in any OCI compliant registry.

## What are some benefits of a data bottle?

ACE Data Bottles have the following benefits:

- Bottles are made of parts (which can be a file or directory)
- Directory parts are archived  with tar
- Parts that benefit from compression are zstd compressed
- Metadata is stored along-side but separate from the data
- Thousands of parts are possible.  Docker images only allow 127 layers due to a hard limit on the number of arguments to a syscal on Linux.
- Partial bottles can be downloaded by using selectors (selecting parts with specific labels).
- Inter-bottle relationships are a first class citizen
- Bottle metadata can be indexed by ACE Telemetry to aid in data discovery
- Multi-threaded
- Supports efficient OCI to OCI mirroring with one-way communication useful for air-gapped networks.

## What is the anatomy of a bottle?

Bottles are made up of parts and parts contain the actual bulk data of the bottle as files. See the [Bottle Anatomy Tutorial](../usage/tutorials/bottle-anatomy.md) for a discussion of how a bottle is created.

## What is a bottle part?

Each subset of a bottle that has been labeled as a bottle part can be viewed as a "sub bottle" that contributes to the entire bottle's understanding.

## How do I decide which bottle files constitute a bottle part?

Files and directories should be logically grouped and labeled as a bottle part so consumers can pull a subset of a bottle's contents independently without having to pull the whole bottle.

The organization and searchability of those files, known as bottle parts, is dependent on a bottle's author creating meaningful labels using the command `ace-dt bottle part label`.

Conceptual Example: Consider a bottle whose data relate to an object carrying an experimental payload. Each directory and subsequent file grouping that can stand independent of the bottle should be documented.

Bottle part labels:

- vehicle architecture
  - internal
  - external
- experiments run
  - agricultural
    - experiment q
    - experiment r
  - material
    - experiment x
    - experiment y
- flight variables
  - speed
    - 7.7 km/s
    - 7.8 km/s
    - 7.9 km/s
  - inclination angle
    - 15 degrees
    - 20 degrees
    - 25 degrees
    - 30 degrees

In this example a consumer could choose to pull the following parts based on the labels assigned by the author:

External vehicle architecture
flight variable inclination angles 20 degrees - 30 degrees.

## What is an example of Data Tool usage?

To download and extract the latest test data bottle to the `mybottle` directory run

```bash
ace-dt bottle pull us-central1-docker.pkg.dev/aw-df16163b-7044-4662-93fa-ec0/public-down-auth-up/mnist:v1.6 -d mybottle
```

See the [user guide](../usage/user-guide.md) for more example usage. For CLI documentation see the [CLI reference](../cli/ace-dt.md).

## How can a bottle be referenced?

Bottles may be referenced in many ways.

| Syntax                                    | Description                        | Mutable | Example                                                                                                |
|-------------------------------------------|------------------------------------|---------|--------------------------------------------------------------------------------------------------------|
| `<registry>/<repository>/<name>:<tag>`    | name and tag                       | yes     | `reg.example.com/repo/dataset:1.2`                                                                     |
| `<registry>/<repository>/<name>:latest`   | name and tag "latest"              | yes     | `reg.example.com/repo/dataset:latest`                                                                         |
| `<registry>/<repository>/<name>@<digest>` | Manifest ID (i.e., manifest digest | no      | `reg.example.com/repo/dataset@sha256:1234123412341234123412341234123412341234123412341234123412341234` |
| `bottle:<digest>`                         | Bottle ID (i.e., bottle digest)    | no      | `bottle:sha256:deedbeef12341234123412341234123412341234123412341234123412341234`                       |

The first three examples are identical to how Docker references images. The last one is unique to Data Tool.

## How is a bottle reference structured?

A bottle reference follows the form:

`<registry>/<repository>/<name>:<tag>`

## What are labels and selectors?

Bottle authors classify and identify sets of bottles using labels.

Bottle consumers use label selectors (often simply called selectors) to restrict the set of bottles returned when executing a query in ACE Telemetry.

See the [Labels and Selectors Guide](../usage/concepts/labels-selectors.md) for a discussion of how labels and selectors are related.

## Does ACE Data Tool automatically version a data bottle?

Bottles are immutable. When the contents of a bottle or its metadata are edited and pushed, the resulting bottle is a new bottle with a unique Bottle ID.

The source command should be used to define a parent bottle. ACE Telemetry will then infer descendant bottles.

An individual researcher can use label metadata to keep track of their own versions of a bottle, e.g. `version=v1.2.3`, if needed.

## Why are some bottle data compressed and not others?

Bottles are designed for effective access to data. ACE Data Tool automates the compression process for bottle parts.

Compression is only used when it provided a measurable benefit in file size. When data are compressed, they must also be decompressed before a consumer can access them. If decompression will take more vital time away from a researcher running an experiment than the benefits gained by a compressed file size, ACE Data Tool will leave the part uncompressed.

If ACE Data Tool's logic determines that compression is beneficial, it will handle the process when a bottle is pushed.

When compression and encryption are not used, the Layer ID and the Content ID can be the same.

## How do I keep my bottle private so no one else uses my data?

ACE Data Tool registers bottles in ACE Telemetry when a bottle is pushed to an OCI registry. The metadata associated with a bottle are recorded and become searchable on ACE Telemetry. The data in the bottle are stored separately in the designated OCI registry. Each registry handles access controls to the actual data, allowing researchers to restrict access to the data within the bottle to only authorized users. Metadata should be considered public. No sensitive data should disclosed via a bottle's metadata fields.

## What OCI registries can I use with ACE Data Tool?

The following registries are known to work:

- [Google Artifact Registry](https://cloud.google.com/artifact-registry) (GAR)
- [Docker distribution](https://github.com/distribution/distribution) (i.e., the docker image registry:2)
- [GitLab Container Registry](https://docs.gitlab.com/ee/user/packages/container_registry/)
- [JFrog Container Registry](https://jfrog.com/container-registry/)

Other registries:

- Nexus
  - Only partially OCI compliant
  - May have limitations when used with ACE Data Tool
- Docker Hub
  - Is thus not OCI compliant
  - Only supports Docker images
  - Does not support arbitrary artifacts
  - Does not work with ace-dt

## Which registry should be used when pushing a bottle from ACE Hub?

You are able to push to any OCI registry.  If working in GitLab repository, you can also push to that repo's registry.

Example:

Push a data bottle to GitLab repo `/mygroup/myproject/bottle`

```shell
ace-dt bottle push reg.git.act3-ace.commygroup/myproject/bottle/mybottle:v1
```

Keep in mind, bottles pushed to `zot.lion.act3-ace.ai` are for testing only and have no permission controls.

## Which metadata field is the most important?

Metadata are used to describe a bottle's contents from multiple consumer perspectives. Therefore all metadata can be a validated priority.

Bottle creators should think about how a consumer will use metadata to locate and pull a bottle so that the information associated with a bottle is as rich, meaningful, and accurate as possible. The future of a bottle's application could otherwise be limited by skewed or partial representation in a bottle's metadata fields.

The description is useful for humans to find bottles previously unknown to them.  Bottle labels are useful for selecting bottles for leaderboards and other tooling.  A bottle annotation is just a catch-all field that can be used to store other information not appropriate elsewhere in the metadata.  A metric is a real number that that has an implied ordering use for ranking similar bottles.

## Is the Bottle ID the name of my bottle?

Bottles do not have names or titles. Neither a name nor a title could be guaranteed to be unique (i.e. two researchers could come up with the same name and use it for two entirely different bottles). The Bottle ID is the unique identifier for a bottle. It is generated automatically by ACE Data Tool.

Each bottle has one, and only one, Bottle ID.

Any time a bottle is modified, including edits or additions to the bottle's metadata, a new Bottle ID is generated.

## Why does my Bottle ID change every time I edit and commit my bottle?

Bottles are immutable and cannot be changed. When a bottle is edited or modified, a new bottle is created and a new corresponding Bottle ID is generated. This functionality prevents bottles or their parts from being overwritten and creates a data lineage documenting the evolution of a dataset.

## What are selectors and how are they used?

Selectors are search criteria used to filter bottle queries in ACE Telemetry.

Selectors filter queries for a consumer by matching keywords in bottle metadata. The ability to locate a bottle is dependent on the author's accurate representation of their dataset to a consuming audience using metadata such as annotations, authors, descriptions, labels, or metrics.

## Can I directly mount a bottle into a kubernetes cluster?

A Bottle can be directly mounted into Kubernetes pods on Lion as shown below.  This approach uses a special CSI driver for Kubernetes.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.13-alpine
    ports:
    - containerPort: 80
    volumeMounts:
    - name: data
      mountPath: /usr/share/nginx/html
  volumes:
  - name: data
    csi:
      driver: bottle.csi.act3-ace.io
      # Add a secret to pull the bottle (if auth is required by the registry)
      # nodePublishSecretRef:
      #   name: test-secret
      volumeAttributes:
        # Specify the bottle to pull
        bottle: us-central1-docker.pkg.dev/aw-df16163b-7044-4662-93fa-ec0/public-down-auth-up/mnist:v1.6
        # or by digest (this pulls the bottle above but uses ACE Telemetry to find a suitable repository to pull from)
        # bottle: bottle:sha256:8d90d933cffe2c82c383e1a2ecd6da700fc714a9634144dd7a822a1d77432566
        
        # Optionally select what subset of data you want to pull down
        # selector: "subset=train,component=image|type=usage"
```

Bottle pulls by CSI are registered with the [telemetry server on lion](https://telemetry.lion.act3-ace.ai).  When the registry requires authentication, a `nodePublishSecretRef` must be provided.  

The `nodePublishSecretRef` is a Kubernetes secret, which can easily be created with `kubectl create secret docker-registry`.

## Can I use Data Tool to mirror container images from one registry to another?

See the Mirror User Guide for a discussion of the `ace-dt mirror` command group, which is used to transfer container images from one registry to another.

The objective of `ace-dt mirror` is to efficiently perform data movement while minimizing the user's tracking effort.  The tool keeps track of the data that has already been sent to the target environment to ensure that it is not sent again.  This minimizes the amount of data needed to be scanned and transferred along from the source to the target.

Potential uses cases of `ace-dt mirror` are include:

- Mirroring images from a network with internet access to a secured air-gapped environment
- Mirror images from one registry to any number of other registries, performing renaming of images along the way
