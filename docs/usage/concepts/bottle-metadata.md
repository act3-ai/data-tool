# Bottle Metadata Guide

## Intended Audience

This concept guide is written for Data Tool users who want to add metadata when creating and modifying bottles.

## Prerequisites

This guide does not have prerequisites for the intended audience.

## Annotate

Annotations are used to add or remove unregulated author descriptions as key-value pairs to a bottle. The typical usage is to encode arbitrary metadata into the bottle.

Annotations are not required to have any semantic meaning to the system. A current example of using annotations that have a semantic meaning is the [Artifact Viewer](../tutorials/artifact-viewer.md) feature.

Bottle consumers do not use annotations to search for or identify a bottle, but rather to access arbitrary structured data, which the bottle's creator has encoded into the bottle's metadata.

Conceptual examples:

- Use annotations for proposed new fields  
  - If they mature, the field can be included into the next version of the formal bottle schema  
  - This aids in forward and backward compatibility of bottles
- Use annotations to encode metrics that are non-scalar
  - This provides an alternative to the bottle metric command, which requires scalar values
  - Note that this use of an annotation does not produce a metric suitable for leaderboard view on a telemetry server

The syntax is:

```sh
ace-dt bottle annotate <key>=<value> [flags]
```

Example usage:

```sh
# Add annotation <foo=bar> to bottle in current working directory:

ace-dt bottle annotate foo=bar
```

```sh
# Remove annotation <foo> from bottle <bar> at path <~/mypath>:

ace-dt bottle annotate foo- -d ~/mypath
```

```sh
# Edit annotation <foo> to become <foo1> for bottle <bar> at path <~/mypath>:

ace-dt bottle edit foo foo1 -d ~/mypath
```

```sh

# List all bottle annotations:

ace-dt bottle annotate list
```

Syntactically, the requirements for the key are the same as [K8s labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set), but an annotation value can be any string of arbitrary length and may contain punctuation characters (e.g., 100KB base64 encoded image or a large JSON document).

## Artifact

An artifact is simply a small file within a bottle with an associated name and media type. Unlike bottle parts, artifacts are visible in a telemetry server if one is configured when a bottle is pushed. As such, artifacts should be considered public to whomever can access the configured telemetry server.

The syntax to associate an artifact is:

```sh
ace-dt bottle artifact [flags]
```

Examples of command usage are below:

```sh
# Add artifact foobar.png to bottle in current working directory:

ace-dt bottle artifact add foobar.png
```

```sh
# Remove artifact foobar.png from bottle:

ace-dt bottle artifact remove foobar.png
```

The maximum size of an artifact is limited by the specific deployment of the telemetry server. Typically, artifacts are around 1MiB. Artifact [media types](https://en.wikipedia.org/wiki/Media_type) that are recognized and visualized by the telemetry server include the following:

- HTML: `text/html`
- Images: `image/jpeg`, `image/png`, `image/svg+xml`, `image/webp`, `image/gif`
- Jupyter Notebook: `application/x.jupyter.notebook+json`
- Markdown: `text/markdown`
- Tabular: `text/csv`, `text/tab-separated-values`
- Text: `text/plain`

Media types that are not recognized by the telemetry server may still be added to a bottle. In such cases, a bottle consumer can download the artifact as a raw file. In addition, it is possible to create an Artifact Viewer and embed that within the bottle to make viewing complex artifacts easier.

> Consult the [Artifact Viewer Tutorial](../tutorials/artifact-viewer.md) for a walkthrough of creating one

## Author

The `author` field is used to document bottle creator(s). Bottles may have any number of authors. When an author is added, the entry can include a URL associated with the author.

Bottle consumers can discover bottles associated with an author in a telemetry server.

The syntax to add or remove an author is:

`ace-dt bottle author [flags]`

Examples of command usage are below:

```sh
# Add author <John Doe> to bottle in current working directory:

ace-dt bottle author add "John Doe" "jdoe@example.com"
```

```sh
# Remove author <John Doe> to bottle in current working directory:

ace-dt bottle author remove "John Doe" "jdoe@example.com"
```

```sh
# Add author <Alice Wonders> to bottle at path <my/bottle/path>:
ace-dt bottle author add "Alice Wonders" "alicew@example.com" --url="university.example.com/~awonders" -d my/bottle/path
```

## Describe

The `describe` field is used to add an abstract (paragraph description) to a bottle as a supplement other metadata. A description is a textual paragraph written for bottle consumers to read. A well-written description can help researchers locate bottles and can help the researcher evaluate the bottle's suitability for an experiment.

Bottle consumers can query bottles in a telemetry server by matching words in a bottle’s description.

The syntax is:

`ace-dt bottle describe [flags]`

Examples of command usage are below:

```sh
# Add a short description to bottle in current working directory

ace-dt bottle describe "The context of this bottle is foobar."
```

```sh
# Add description text from <./my-description.txt> file to a bottle at path <my/bottle/path>

ace-dt bottle describe --from-file ./my-description.txt -d my/bottle/path
```

## Label

Labels are key-value pairs that provide a structured way to represent the contents of a bottle. Labels should fully describe the data contained in the bottle.

As a reminder about [rAI](../user-guide.md#ethics), labels are used to ethically represent and document data sets.

There are no limits to the number of labels that can be added to a bottle. In practice, manual data entry may be time consuming. Creating bottles programmatically can automate the process of adding labels to a bottle.

Bottle consumers can use labels to locate bottles in a telemetry server.

The syntax is:

`ace-dt bottle label <key>=<value> [flags]`

Examples of command usage are below:

```sh
# Add label <foo=bar> to bottle at path <my/bottle/path>:

ace-dt bottle label foo=bar -d my/bottle/path
```

```sh
# Add multiple labels to a bottle in current working directory:

ace-dt bottle label key1=val1 key2=val2 key3=val3
```

```sh
# Add label to bottle part:

ace-dt bottle part label foo=bar
```

```sh
# Remove label <foo> from bottle <bar> at path <my/bottle/path>:

ace-dt bottle label foo- -d my/bottle/path
```

Some well known labels include:

- `data.act3-ace.io/projectName` denotes the project name
- `data.act3-ace.io/contractNumber` denotes the contract number
<!-- - `data.act3-ace.io/compression` when this label is set to `none` then compress is not attempted on the part.  `ace-dt` will avoid compressing incompressable data but the only way to do that is to try to compress it and see if it worked.  This special label sort circuits that and disables compression on the part so that committing the part is faster.  In the future other values might be supported here like the type of compression and/or compression parameters. -->

Other common labels include:

- `experiment`
- `run`
- (hyper)parameters used in training (e.g., `layers=13`, `model=resnet`)
- version

## Metric

Metrics are key-value pairs in which the values are named scalars that are orderable and may be accompanied by a description. They are used to measure a bottle’s performance and thus help researchers compare bottles returned from a given query in the ACE Telemetry.

A metric differs from a bottle label because it must be a floating-point number. Known types of metrics that represent a bottle’s real-world application include:

- Decision thresholds
- Variation approaches
- Training accuracy
- Evaluation accuracy
- Runtime
- Memory usage

When bottle creators add metrics and labels in tandem, consumers can locate a group of bottles by subject label, then evaluate the bottles’ performance metrics for their needs.

The syntax is:

`ace-dt bottle metric add [METRIC] [VALUE]`

Examples of command usage are below.

In the following example, the metric of `blue completion` with a value of `0` is added

```sh
ace-dt bottle metric add "blue completion" 0 --desc "Blue agent's results"
```

To view the metric:

```sh
ace-dt bottle metric list
```

The expected output is:

```sh
METRIC             DESCRIPTION         
blue completion=0  Blue agent's results
```

## Part

Part is a command group that provides subcommands for interacting with metadata of bottle parts. All metadata commands that can be applied to bottles can also be applied to bottle parts.

When labels are applied to a bottle part, different logical parts of the data are grouped and made searchable.

Bottle consumers can then access bottle parts via a telemetry server.

The syntax to add or remove bottle part labels is:

`ace-dt bottle part label key1=val1`

Example Usage:

```sh
# Add label <foo=bar> to part <myPart.txt> bottle in current working directory:

ace-dt bottle part label foo=bar myPart.txt
```

```sh
# Add label <foo=bar> to many parts in current working directory:

ace-dt bottle part label foo=bar myPart.txt myPicture.jpg myModel.model
```

```sh
# Add label <foo=bar> to parts matching a wildcard:

ace-dt bottle part label foo=bar *.txt
```

```sh

# Remove label <foo> from part <myPart.txt> at path <my/bottle/path>:

ace-dt bottle part label foo- myPart.txt -d my/bottle/path
```

Bottle part labels can also be added or edited by directly modifying `**/.labels.yaml`.

To use best practices, bottle creators should:

- Consider how another researcher might use the contents of a bottle when deciding how to apply part labels
- Add meaningful label bottle parts so consumers can download only the parts they need to accomplish their work  

Conceptual example: imagine a dataset where the images are grouped by inclination angle. This should likely be one bottle with a part that is labeled for each inclination angle group. Then, a bottle consumer can pull all images taken from angles that are relevant to their work.

## Source

Sources are citations for bottle parts. Sources contribute to the bottle’s data lineage connecting current datasets to previous iterations. Data Lineage allows researchers to follow a dataset back to its origins. Using the source command to accurately document a new bottle is part of practicing [rAI](../user-guide.md#ethics).

The syntax to add or remove a source is:

`ace-dt bottle source [flags]`

Examples of command usage are below:

```sh
# Add a source <FooBar> to bottle in current working directory

ace-dt bottle source add FooBar
```

```sh
# Add source from <./myFooBarSource.txt> file to a bottle at path <my/bottle/path>

ace-dt bottle describe --from-file ./myFooBarSource.txt -d my/bottle/path
```

To use best practices, bottle creators should denote sources by a Uniform Resource Identifier (URI) such as the following:

- URL
- Bottle ID

It is important to note that bottles define their sources or parents, not the other way around. A telemetry server discovers and infers the "parent to child" relationship between bottles.

Conceptual Example:

A bottle containing the result of a model training run should include the bottle parts listed below:

- Log file
- Detailed performance metrics
- Model parameters

This bottle’s metadata should denote sources for the following parts:

- Training set URI (URL or bottle reference)
- Test set URI (URL or bottle reference)

These sources allow researchers to locate the original training dataset. Then, using ACE Telemetry, researchers can locate all models generated from that dataset.
