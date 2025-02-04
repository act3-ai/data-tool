# Mirror Guide

## Intended Audience

This documentation is written for Data Tool users who want to efficiently move OCI images from one repository to another.

> Consult the [Data Tool User Guide](../user-guide.md) to review Data Tools's key concepts and common usage patterns

## Prerequisites

This guide does not have prerequisites for the intended audience

## Overview

This guide outlines the mirror workflow, indicating which parts are completed on the low-side and high-side repositories when moving data to an air gapped environment, and provides a conceptual overview of each part of the workflow.

> For step-by-step guidance, [consult the mirror tutorial](../tutorials/mirror.md)

![Mirroring Process](../../resources/diagrams/mirror-process.drawio.svg)

## General Workflow Overview

The workflow for using Data Tool's mirror functionality is:

- Update the Data Tool configuration file (if needed)
- Prepare a `sources.list` file to gather sources into a target repository (can be remote or local)
- Execute the `gather` command to populate the target repository
- Execute the `serialize` command to write a tar file to storage media (e.g a tape drive)
- Execute the `deserialize` command to move the contents of the tar file to a target destination repository
- Prepare a `scatter.tmpl` file using Go templating to map locations between the initial repository and the destination repository
- Execute the `scatter` command to distribute the contents of the tar file to the locations designated in the `scatter.tmpl` file

## Air Gap Workflow Overview

The workflow outlined above can be used any time it is necessary to efficiently move OCI images from one repository in an OCI registry to another.  When working in an air gapped environment, the workflow above can be adapted with the first group of activities happening on the low side and the second group of activities happening on the high side.

Lo-side repository:

- Update the Data Tool configuration file (if needed)
- Prepare a `sources.list` file to gather sources into a low side registry
- Execute the `gather` command to populate the low side registry
- Execute the `serialize` command to write a tar file to storage media (e.g a tape drive)

High-side repository:

- Execute the `deserialize` command to move the contents of the tar file to the high side registry
- Prepare a `scatter.tmpl` file using Go templating to map locations between the low side and high side
- Execute the `scatter` command to distribute the contents of the tar file to the locations designated in the `scatter.tmpl` file

## Usage

### Update Configuration File

Data Tool's configuration file specifies values that are used to transfer OCI images from one registry to another.

View a sample config:

```sh
ace-dt config --sample
```

The relevant section of the sample config for this guide is:

```sh
# registryConfig:
#      index.docker.io:
#        endpoints:
#          - https://index.docker.io
#          - http://localhost:5000
#        rewritePull:
#          "^rancher/(.*)": "ace/dt/rancher-images/$1"
#          "^ubuntu/(.*)": "ace/dt/ubuntu-images/$1"
#      nvcr.io:
#        endpoints:
#          - https://nvcr.io
#    endpointConfig:
#      https://nvcr.io:
#        supportsReferrers: "tag"
#      https://index.docker.io:
#        tls:
#          insecureSkipVerify: true
```

### Sources

This step in the workflow assumes that you have already identified a list of OCI images that need to be transferred.

The `sources.list` file specifies those images. Data Tool will uses the `sources.list` when the `gather` command is run.

The syntax is to list one fully-qualified OCI image reference per line.

Example usage:

```sh
quay.io/ceph/ceph:v17.2
docker.io/curlimages/curl:7.73.0
docker.io/konstin2/maturin@sha256:a203e1071d73c6452715eb819701cb49ca18e0dcd82fe13928de2724c4f2861f
```

### Gather

The `gather` command uses the `sources.list` file to populate a repository with OCI images.

The syntax is:

```sh
ace-dt mirror gather SOURCES-FILE IMAGE [flags]
```

<!--
```sh
ace-dt mirror gather source-file remote-repository-image [flags]
``` -->

The `gather` command can be executed against a local repository or a remote repository.

When executed against a remote repository, all specified OCI images and indexes are copied to a singular remote registry without using any local disk.

After adding all user-specified images and indexes to the target registry, the `gather` command creates an index for that repository.

Example usage:

If a user wanted to collect all remote images referenced in a local `sources.list` file, send them to a target repository `/gather` within registry `reg.example.com` and indicate that it is the 45th time that `gather` has been run, they would run the following command:

```sh
ace-dt mirror gather source-images.txt reg.example.com/gather:sync-45
```

#### Index Fallback

Nested indexes (index-of-index) are specified in [The OCI 1.1 Image Index Specification](https://github.com/opencontainers/image-spec/blob/main/image-index.md#image-index-property-descriptions). However, not all registries support index-of-index behavior (nesting an index within another index).

In these cases, the `gather` command has an `--index-fallback` flag. When this flag is used, `gather` will still push indexes specified in the `sources.list` file to the destination repository, but it will not add them to the main index's manifest list. Instead, it pushes their references to the main index's annotations where they are automatically handled by the subsequent mirror commands.

Example usage:

If a user wanted to push the above `sources.list` file to a registry that does not support nested indexes (e.g., Jfrog as of Sep. 2023), they would need to run the gather operation with the `index-fallback` flag:

```sh
ace-dt mirror gather source-images.txt reg.example.com/gather:sync-45 --index-fallback
```

The `--index-fallback` flag tells `gather` to reference the indexes in the annotation field `vnd.act3-ace.extra-manifests` instead of adding them to the main index's manifest list (which would trigger a registry error). The subsequent mirror steps (`serialize`, `deserialize`, and `scatter`) will automatically handle the parsing of the nested indexes in the annotation.

The `gather` command reports the total index size and the deduplicated size in the top-level image index under the annotation fields `vnd.act3-ace.layer.size.total` and `vnd.act3-ace.layer.size.deduplicated` respectively.

#### Annotations

The `gather` command also accepts additional annotations as key-value pairs that the user can append to the index of the gathered image. These annotations can be defined with the `--annotations` or `-a` flags. Annotations are useful for adding any metadata that may not be appropriate as a label.

Data Tool automatically adds the following to the annotations of the gathered image:

- Version of `ace-dt`
- Total size of the gather image
- Deduplicated size of the gather image (the total size minus the size of any duplicated blobs across objects)

To set additional annotations, run the `ace-dt mirror gather` command with the `--annotations` flag.

Example usage:

```sh
ace-dt mirror gather source-images.txt reg.example.com/project/repo:sync-45 --annotations=key1=value1,key2=value2
```

### Serialize

The `serialize` command is used to create a tar file. The tar file can be saved to a local machine or can be directed to write to a tape drive with custom buffer and block size flags when transferring images to an air gapped environment.

The syntax is:

```sh
ace-dt mirror serialize IMAGE DEST [EXISTING-IMAGE...] [flags]
```

The `serialize` command includes several optional flags to customize memory buffer options and to save or resume from a generated checkpoint file in the case of a network interruption during the serialization process.

Example usage:

If a user wanted to serialize all images in the remote repository `reg.example.com/gather:sync-45` to a tape drive destination `/dev/nst0`, they would run the following command:

```sh
ace-dt mirror serialize reg.example.com/gather:sync-45 /dev/nst0
```

#### Optional Serialize Commands

The `serialize` command optionally offers checkpoint functionality. A checkpoint is used to avoid having to repeat the entire serialize process in the case of any failure when the `serialize` command is run.

Using the `--checkpoint` flag after the `serialize` command saves a checkpoint file locally. The checkpoint file includes the blobs written and their byte position which can be used to estimate the offset for the tape.

In the case of failure(s) during the `serialize` command, a user can reference the checkpoint files created in the previous runs to avoid copying blobs that already exist on the destination.

> A user could also estimate offset by inspecting the tape position at failure

The `--resume-from-checkpoint` flag allows a user to resume the serialize process from a specific image in the index (and avoid duplicated serialization to the tape drive/file).

If the user needs to `--resume-from-checkpoint` they will also need to use the `--resume-from-offset` flag to start the serialization process at the correct point on the tape drive.

Example usage:

If a user wanted to continue the serialization process of `reg.example.com/gather:sync-45` after 3 network failures, they would specify a new tar file (or the tape drive after manually setting it to start at the total offset), a new checkpoint file and each checkpoint file previously created and its respective offset as shown below:

```sh
ace-dt mirror serialize reg.example.com/gather:sync-45 sync45-4.tar --checkpoint cp4.txt --existing-from-checkpoint cp1.txt:34752000 --existing-from-checkpoint cp2.txt:5724672 --existing-from-checkpoint cp3.txt:53126656
```

The serialize command also optionally accepts a list of images that are already in the isolated environment and can be skipped.

Syntax:

```sh
ace-dt mirror serialize image destination [EXISTING-IMAGES...] [flags]
```

Example usage:

If a user wanted to serialize the images in the remote repository `reg.example.com/gather:sync-45` EXCEPT for `quay.io/ceph/ceph:v17.2` and `docker.io/curlimages/curl:7.73.0` to a local tar file `sync45.tar`, they would run the following command:

```sh
ace-dt mirror serialize reg.example.com/gather:sync-45 sync45.tar quay.io/ceph/ceph:v17.2 docker.io/curlimages/curl:7.73.0
```

### Deserialize

The `deserialize` command is used to reconstruct the contents of the tar file from its serialized form. This command also pushes the images from the deserialized tar file to the target remote repository and stages them so they can then be distributed (or scattered) in the next step of the mirror workflow.

The syntax is:

```sh
ace-dt mirror deserialize SOURCE-FILE IMAGE [flags]
```

<!--

```sh
ace-dt mirror deserialize tape.tar remote-repository [flags]
```
-->

Example usage:

If a user wanted to deserialize everything on `/dev/nst0` to a target repository `/scatter` within the isolated registry `reg.high.example.com` and indicate that this is the 45th mirror sync process, they would run the command:

```sh
ace-dt mirror deserialize /dev/nst0 reg.high.example.com/scatter:sync-45
```

### Scatter

The `scatter` command uses the `scatter.tmpl` file to distribute or *scatter* the contents of the tar file to one or more designated location(s).

Create a `scatter.tmpl` file:

```sh
touch scatter.tmpl
```

Use Go templating in an editor of your choice to define the locations where the OCI images should be distributed.

Sample syntax:

```go
{{- $annotation := index .Annotations "org.opencontainers.image.ref.name" -}}
localhost:5000/{{ $annotation }}
```

Example usage from the sample template above:

- `index.docker.io/library/busybox:1.36.1` would be mapped to `localhost:5000/index.docker.io/library/busybox:1.36.1`.
- `reg.git.act3-ace.com/ace/data/tool/bottle/mnist:v1.6` would be mapped to `localhost:5000/reg.git.act3-ace.com/ace/data/tool/bottle/mnist:v1.6`.

After the `scatter.tmpl` file has been created, the scatter command can be run.

This command accepts a remote repository (the same that was the target of the deserialize command) and a `destfile` (formatted set of rules).

The syntax is:

```sh
ace-dt mirror scatter IMAGE MAPPER [flags]
```

<!-- 

```sh
ace-dt mirror scatter remote-repository [ruleSet]=path/to/destfile
```
-->

Example usage:

If a user wanted to distribute the images in `reg.high.example.com/scatter:sync-45` using an `all-prefix` ruleset defined in a local CSV file called `destfile.csv`, they would run the following:

```sh
ace-dt mirror scatter reg.high.example.com/scatter:sync-45 all-prefix=destfile.csv
```

#### Optional ruleset Types

There are 5 rulesets available for the `destfile.csv`:

- `all-prefix`
- `first-prefix`
- `longest-prefix`
- `digests`
- `go-template`

##### all-prefix

The `all-prefix` allows a user to send one image to many remote registries.

Use the `all-prefix` ruleset to send images matching any prefix to their respective target repository, as defined in the `destfile.csv`.

Syntax:

```sh
ace-dt mirror scatter reg.high.example.com/scatter:sync-45 all-prefix=destfile.csv
```

Example usage:

If the `reg.high.example.com/scatter:sync-45` repository contained image `docker.io/konstin2/maturin` and the previous scatter command were run with the following `destfile.csv`:

```sh
docker.io/,secret.reg.example.com/docker.io/
docker.io/,secret.froggy.example.com/docker.io/
docker.io/library/,secret.reg.example.com/docker.io/library
```

The `docker.io/konstin2/maturin` image in the example above would be distributed to the locations of `secret.reg.example.com/docker.io/konstin2/maturin` and `secret.froggy.example.com/docker.io/konstin2/maturin`. It would not be sent to the third repository because `docker.io` and `docker.io/library` are not considered a match.

##### first-prefix

Use the `first-prefix` ruleset to send each image matching the first prefix to the target repository defined in the `destfile.csv`.

Syntax:

```sh
ace-dt mirror scatter reg.high.example.com/scatter:sync-45 first-prefix=destfile.csv
```

Example usage:

If the `reg.high.example.com/scatter:sync-45` repository contained image `docker.io/konstin2/maturin` and the scatter command were run with the following `destfile.csv`:

```csv
docker.io/,secret.reg.example.com/docker.io/
docker.io/,secret.froggy.example.com/docker.io/
docker.io/library/,secret.reg.example.com/docker.io/library
```

The `docker.io/konstin2/maturin` image would only be distributed to the first location `secret.reg.example.com/docker.io/konstin2/maturin`.

##### longest-prefix

Use the `longest-prefix` ruleset to send each image matching the longest prefix to the target repository defined in the `destfile.csv`.

Syntax:

```sh
ace-dt mirror scatter reg.high.example.com/scatter:sync-45 longest-prefix=destfile.csv
```

Example usage:

If the `reg.high.example.com/scatter:sync-45` repository contained image `docker.io/konstin2/maturin` and the scatter command were run with the following `destfile.csv`:

```text
docker.io/,secret.reg.example.com/docker.io/
docker.io/konstin2/,secret.froggy.example.com/docker.io/
docker.io/library/,secret.reg.example.com/docker.io/library
```

The `docker.io/konstin2/maturin` image would be distributed to the second location `secret.froggy.example.com/docker.io/konstin2/maturin` because it has the longest prefix match.

##### digests

Use the `digest` ruleset to send each image with a matching digest to the appropriate target repository defined in the `destfile.csv`. This rule set also allows sending one image to multiple remote registries.

Syntax:

```sh
ace-dt mirror scatter reg.high.example.com/scatter:sync-45 digests=destfile.csv
```

Example usage:

If the `reg.high.example.com/scatter:sync-45` repository contained image `docker.io/konstin2/maturin@sha256:a203e1071d73c6452715eb819701cb49ca18e0dcd82fe13928de2724c4f2861f` and the scatter command were run with the following `destfile.csv`:

```text
sha256:a203e1071d73c6452715eb819701cb49ca18e0dcd82fe13928de2724c4f2861f,secret.reg.example.com/docker.io/
sha256:4a59234a43f552820a807abafd092ccfd2786440de873e12588d9926e3216f02,secret.froggy.example.com/docker.io/
sha256:ed5262412dc05cfd143d62c7791a04bf5694ad68c0663a4b42cf0c09a2716733,secret.reg.example.com/docker.io/library
```

The `docker.io/konstin2/maturin` image would be distributed to the first location `secret.reg.example.com/docker.io/konstin2/maturin` because the digests match.

##### go-template

Using the `go-template` ruleset allows you to use Go language [templates](https://pkg.go.dev/text/template) to map to remote registries. This option also supports all the [Hermetic text sprig functions](http://masterminds.github.io/sprig), which offer useful features such as prefix matching and regex matching. There are also some other functions defined:

- `tag`: Returns the tag of an OCI string
- `repository`: Returns the repository of an OCI string
- `registry`: Returns the registry of an OCI string
- `package`: Omits the registry from the OCI reference

Syntax:

```sh
ace-dt mirror scatter reg.high.example.com/scatter:sync-45 go-template=destfile.tmpl
```

Example usage:

Given the following `destfile.tmpl` (with added numbered lines for reference):

```go
1 {{- $name := index .Annotations "vnd.act3-ace.manifest.source -}}
2 secret.reg.example.com/{{ trimPrefix "reg.example.com/" $name -}}
3 {{ if hasPrefix "docker.io" $name }}
4 secret.reg.example.com/high/scatter/{{ trimPrefix "reg.example.com" $name -}}
5 {{- end -}}
```

The first rule is not conditional (see lines 1 and 2 of the example above). It defines that all images are to be sent to `secret.reg.example.com/[image-name]`. Their originating reference is stored in the image annotations. The Go template uses the `org.opencontainers.image.ref.name` annotation to craft the new destination.

Given the image `docker.io/konstin2/maturin`, it would be sent to `secret.reg.example.com/docker.io/konstin2/maturin`.

The second (conditional) rule uses the sprig function `hasPrefix` (see lines 3-5 of the example above). This rule checks to see if an image originates from a `docker.io` registry. If so, it completes the rule and creates the repository to send the image.

Given the image `docker.io/konstin2/maturin`, it would *also* be sent to `secret.reg.example.com/high/scatter/docker.io/konstin2/maturin`.

### Clone

The `ace-dt mirror clone` command takes the input file of `gather` and the mapping file from `scatter` and clones the images in a `sources.list` file by scattering them according to the mapping ruleset passed. Outside of air gapped environments, the `ace-dt mirror clone` command is also useful when a user simply wants to scatter a list of images to new locations.

The `ruleset` file can be any of the [rulesets](#optional-ruleset-types) defined in `mirror scatter`.

Syntax:

```sh
ace-dt mirror clone source-file [ruleSet]=path/to/destfile
```

Example usage:

Given a `sources.list` file and a go-template `ruleset` as shown below:

```text
quay.io/ceph/ceph:v17.2
docker.io/curlimages/curl:7.73.0
docker.io/konstin2/maturin@sha256:a203e1071d73c6452715eb819701cb49ca18e0dcd82fe13928de2724c4f2861f
```

```go
1 {{- $name := index .Annotations "vnd.act3-ace.manifest.source -}}
2 secret.reg.example.com/{{ trimPrefix "reg.example.com/" $name -}}
```

If a user wanted to clone the three images in the `sources.list` file, they would be cloned to these locations respectively:

- `secret.reg.example.com/quay.io/ceph/ceph:v17.2`
- `secret.reg.example.com/docker.io/curlimages/curl:7.73.0`
- `secret.reg.example.com/docker.io/konstin2/maturin@sha256:a203e1071d73c6452715eb819701cb49ca18e0dcd82fe13928de2724c4f2861f`

### Archive

The `ace-dt mirror archive` command takes the input file of `gather`, a local `tar` destination path, and a tag and creates a `tar` file of the gathered artifact. It is a combination of `ace-dt mirror gather` and `ace-dt mirror serialize` that is useful when the user does not require an intermediate remote repository on the low side for auditing purposes.

The `source-file` should contain a list of the fully-qualified image references (with the source registry) to gather into the target repository.

Syntax:

```sh
ace-dt mirror archive sources.list path/to/destfile.tar sync-1
```

Example usage:

For example, given a `sources.list` file (shown below), a local `archive.tar` path, and a `sync-1` tag:

```text
quay.io/ceph/ceph:v17.2
docker.io/curlimages/curl:7.73.0
docker.io/konstin2/maturin@sha256:a203e1071d73c6452715eb819701cb49ca18e0dcd82fe13928de2724c4f2861f
```

The 3 images in the above `sources.list` file would be gathered to local cache, tagged as `sync-1`, and then serialized to the local `archive.tar` file.

### Unarchive

The `ace-dt mirror unarchive` command takes the output `tar` file of `archive` or `serialize`, reconstructs the contents to local cache, and scatters them according to the [ruleset](#optional-ruleset-types) defined by the user. It is a combination of `ace-dt mirror deserialize` and `ace-dt mirror scatter` that is useful when the user does not require an intermediary registry for auditing purposes.

Syntax:

```sh
ace-dt mirror unarchive path/to/destfile.tar [ruleSet]=path/to/ruleset sync-1
```

Example usage:

For example, given an existing `archive.tar` file that was tagged as `sync-1` and a given ruleset of `nest=localhost:5000`, the 3 images in the tar file would be scattered as follows:

- `localhost:5000/quay.io/ceph/ceph:v17.2`
- `localhost:5000/docker.io/curlimages/curl:7.73.0`
- `localhost:5000/docker.io/konstin2/maturin@sha256:a203e1071d73c6452715eb819701cb49ca18e0dcd82fe13928de2724c4f2861f`

## The Mirror Batch Commands

The mirror batch commands (`ace-dt mirror batch-serialize` and `ace-dt mirror batch-deserialize`) were created to address the need to transfer as little data as possible over an air gap by eliminating duplicative blob copies. These commands sequentially exist after the `mirror gather` command and before the `mirror scatter` command.

The purpose of these commands is to serialize (remote artifact to tar on low side) and deserialize (tar to remote repository on high side) multiple gather artifacts, keeping track of those which have already been serialized and/or deserialized. Keeping track of the artifacts that have already been processed allows the command to skip the copying of manifests and blobs that have already been moved across the air gap.

To efficiently track the status of artifacts, files are created in the SYNC-DIRECTORY during `ace-dt mirror batch-serialize` and `ace-dt mirror batch-deserialize`. The user should not need to edit these files unless they need to repeat a serialization or deserialization of a specific image, in which case they can delete that entry from the applicable file. The two files that are created are:

- `SYNC-DIRECTORY/record-keeping.csv`: The `record-keeping.csv` file is generated during batch-serialize. It keeps track of the images that were serialized and the filename where they exist.
- `SYNC-DIRECTORY/successful_syncs.csv`: The `successful_syncs.csv` file is generated during batch-deserialize. It keeps track of the files that were deserialized along with the date they were pushed to the remote repository.

### Batch-Serialize

> For step-by-step guidance, [consult the mirror batch-serialize tutorial](../tutorials/mirror/mirror-batch-serialize.md)

The `ace-dt mirror batch-serialize` command serializes multiple artifacts to separate files within the SYNC-DIRECTORY. In a typical user workflow, multiple artifacts may need to get transferred over an air gap on a semi-regular basis. The gather artifact that has been compiled may contain many of the same images as an artifact that was serialized at an earlier time. To account for this, `ace-dt` consults the `record-keeping.csv` file and adds the previously serialized images to the [EXISTING-IMAGES...] slice when it eventually runs the `serialize` action.

The `batch-serialize` command requires 2 arguments: the `BATCH-LIST` and the `SYNC-DIRECTORY`. The `BATCH-LIST` is a user-generated `.csv` file that includes the name of the artifact and the remote location. For example, if a user wanted to serialize 3 artifacts named `tools`, `ops`, and `cicd`, they might have a `BATCH-LIST` named `batch.csv` that looks like this:

```csv
name, artifact
tools,reg.example.com/dev-tools:v3.0.6
ops,reg.example.com/ops:v0.2.5
cicd,reg.example.com/cicd:v8.1
```

To serialize these artifacts into a directory `sync/data`, the appropriate syntax would be:

Syntax:

```sh
ace-dt mirror batch-serialize batch.csv sync/data 
```

The command will fetch each of the artifacts in `batch.csv` and serialize them each into a numbered and named tar file within the `sync/data` directory. After execution, the `sync/data` directory will look like:

```txt
sync/
|_ data/
   |_ record-keeping.csv
   |_ 000001-tools.tar
   |_ 000002-ops.tar
   |_ 000003-cicd.tar
```

Opening the `sync/data/record-keeping.csv` file will reveal the following information:

```csv
sync_name, artifact, digest
000001-tools.tar, reg.example.com/dev-tools:v3.0.6, sha256:0e3a39687...
000002-ops.tar, reg.example.com/ops:v0.2.5, sha256:9a339f60e...
000003-cicd.tar, reg.example.com/cicd:v8.1, sha256:62c5eb3a4...
```

After a period of time, the ops artifact needs to be updated because one of the images has an urgent update. The end user can update the image and gather into a new artifact `reg.example.com/ops:v0.2.6` and can add (or update the old entry) that to their `batch.csv` file:

```csv
name, artifact
tools, reg.example.com/dev-tools:v3.0.6
ops, reg.example.com/ops:v0.2.6
cicd, reg.example.com/cicd:v8.1
```

Running the same batch command again will create a new file for the new ops version while optimizing the amount of data being serialized:

```sh
ace-dt mirror batch-serialize batch.csv sync/data 
```

```txt
sync/
|_ data/
   |_ record-keeping.csv
   |_ 000001-tools.tar
   |_ 000002-ops.tar
   |_ 000003-cicd.tar
   |_ 000004-ops.tar
```

During the command execution, ace-dt will pull the `record-keeping.csv` file if it exists and collect all of the previously serialized `ops` artifacts and pass them into the `serialize` action as the [EXISTING-IMAGES...] argument. Doing this creates a much smaller serialized file for the new version of `ops` that only contains the diffs of the current artifact and the previously-serialized artifacts. It is essentially running the equivalent command:

```sh
ace-dt mirror serialize reg.example.com/ops:v0.2.6 sync/data/4-ops.tar reg.example.com/ops:v0.2.5
```

For example, given an artifact `reg.example.com/ops:v0.2.5` containing manifests:

```sh
sha256:318fb637823f...
sha256:94ffabc893ee...
```

and the updated artifact `reg.example.com/ops:v0.2.6` containing manifests:

```sh
sha256:0dc2e6c0f9de...
sha256:94ffabc893ee...
```

Only the manifest `sha256:0dc2e6c0f9de...` and its unique blobs would be serialized to the file `000004-ops.tar`.

After the `ace-dt mirror batch-serialize` command is complete and the files are generated, the end-user moves the images across the air gap (for example, via `rsync` to a head node). From there, the user can run `ace-dt mirror batch-deserialize` and then `ace-dt mirror scatter` to distribute the images to their final locations.

### Batch-Deserialize

The `ace-dt mirror batch-deserialize` command will deserialize all new files in the `SYNC-DIRECTORY` to the `DESTINATION` (a remote repository). It generates or consults (if existing) the `SYNC-DIRECTORY/successful_syncs.csv` file to ensure that the same file is not deserialized multiple times. After successful execution, it appends the deserialized filename and timestamp to the `SYNC-DIRECTORY/successful_syncs.csv` file.

For example, given a `sync/data` directory:

```txt
sync/
|_ data/
   |_ record-keeping.csv
   |_ 000001-tools.tar
   |_ 000002-ops.tar
   |_ 000003-cicd.tar
```

If the user wanted to deserialize the files contained in `sync/data` to the remote image `registry.example.com/sync`, the syntax would be:

Syntax:

```sh
ace-dt mirror batch-deserialize sync/data registry.example.com/sync
```

Upon execution, the command will deserialize each of those files to `registry.example.com/sync` and tag them as their respective image name:

- The file `000001-tools.tar` would be deserialized to location `registry.example.com/sync:tools`.
- The file `000002-ops.tar` would be deserialized to location `registry.example.com/sync:ops`.
- The file `000003-cicd.tar` would be deserialized to location `registry.example.com/sync:cicd`.

A `SYNC-DIRECTORY/successful_syncs.csv` file will be created (if it does not exist) or appended to (if it does exist):

```csv
filename,artifact,timestamp
000001-tools.tar,2024-08-15 16:03:34.875740174 -0400 EDT m=+0.096890310
000002-ops.tar,2024-08-15 16:03:34.958255535 -0400 EDT m=+0.179405761
000003-cicd.tar,2024-08-15 16:03:35.589223735 -0400 EDT m=+0.088605142
```

Like in the `batch-serialize` example, if the file `000004-ops.tar` were later created and existed during a subsequent execution of the `batch-deserialize` command, it would be pushed to `registry.example.com/sync:ops` (assuming the user passes the same destination image repository). The command would consult the `successful_syncs.csv` file and see that the first 3 tar files have already been pushed, so it would only deserialize the `000004-ops.tar` file. The `SYNC-DIRECTORY/successful_syncs.csv` file would then look like:

```csv
filename,artifact,timestamp
000001-tools.tar,2024-08-15 16:03:34.875740174 -0400 EDT m=+0.096890310
000002-ops.tar,2024-08-15 16:03:34.958255535 -0400 EDT m=+0.179405761
000003-cicd.tar,2024-08-15 16:03:35.589223735 -0400 EDT m=+0.088605142
000004-ops.tar,2024-08-22 10:19:55.779132898 -0400 EDT m=+0.127524968
```

After the `batch-deserialize` command has completed, the user can then scatter the artifacts to their final destinations:

- `ace-dt mirror scatter registry.example.com/sync:tools nest=reg.high.example.com`
- `ace-dt mirror scatter registry.example.com/sync:ops nest=reg.high.example.com`
- `ace-dt mirror scatter registry.example.com/sync:cicd nest=reg.high.example.com`

See [Mirror Scatter](#scatter) documentation for more information.

## See Also

- [Mirror Tutorial](../tutorials/mirror.md)
