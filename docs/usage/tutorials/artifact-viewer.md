# Artifact Viewer Tutorial

## Intended Audience

This tutorial is written for Data Tool users who want to create an artifact viewer to display an artifact in a telemetry server.

## Guided Scenario

In this scenario, a user has created a data bottle and associated artifact metadata with the bottle. The media type of the associated artifact is one that is not automatically recognized and visualized in a telemetry server (i.e. not listed in the [Metadata Concept Guide](../concepts/bottle-metadata.md#artifact) section). The user writes a specific application to parse and interpret the artifact so it can be incorporated into ASCE Hub and into the bottle.

<!-- ## Workflow Overview -->

## Step-by-Step Instructions

Below is an example of modifying the `.dt/entry.yaml` file to define a "viewer" for [TensorBoard](https://www.tensorflow.org/tensorboard) logs.

```yaml
annotations:
    viewer.data.act3-ace.io/tensorboard: |-
    {
        "accept": "application/x.tensorboard.logs", 
        "acehub": {
            "image": "zot.lion.act3-ace.ai/tmp/tbv", 
            "resources": {"cpu": "1", "memory": "1Gi"}, 
            "proxyType": "normal"
        }
    }

publicArtifacts:
- mediaType: application/x.tensorboard.logs
  name: Tensorboard logs
  path: logs.tb
  digest: sha256:98f38f12db221a8cf8ca7aadfdcd759b01d52eb4ebb3eedbb2d97e92805c6960
```

Where `zot.lion.act3-ace.ai/tmp/tbv` is a Docker image to be used by the viewer.  The viewer image can be create from this Dockerfile:

```Dockerfile
FROM docker.io/tensorflow/tensorflow 

COPY start.sh /start.sh
ENTRYPOINT [ "/start.sh" ]
```

Where `start.sh` is:

```sh
#!/bin/bash

logdir=$(dirname $ACE_OPEN_PATH)/$(cat "$ACE_OPEN_PATH")
echo "Using logdir $logdir"
tensorboard --logdir "$logdir" --bind_all --port=8888
```

For this example, assume the [TensorFlow](https://www.tensorflow.org) training run wrote the TensorBoard logs to the directory `logs`. Since artifacts are single files and not directories full of files we create a file called `logs.tb` that contains the path to the logs (e.g., `logs` in this example).  In the example, an artifact is defined for `logs.tb` (not the actual `logs` directory) using a custom media type named `application/x.tensorboard.logs`.

We also define a new *viewer* by adding the annotation `viewer.data.act3-ace.io/tensorboard`, where `tensorboard` is the name of the new viewer. The value of the annotation is a small JSON document that has two fields.

The `accept` field follows the [HTTP Accept header syntax](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept) for defining the media types that this viewer "understands". In this case, we only allow the viewer to accept one specific media type for TensorBoard logs. The `acehub` field is a *cluster template*. The cluster template is used to automate the creation of an environment accessed from the ASCE Hub GUI. Each template can define a specific environment configuration including its:

- Image
- Ports
- Script
- Environment Variables
- Resources
- CPUs
- Memory
- GPUs
- Data Bottles

In the example above, a bottle launched in the Hub GUI from a telemetry server will have the `ACE_OPEN_PATH` variable set to the absolute path of the artifact file.  When users open this Hub environment from the GUI, they will be taken directly to TensorBoard's interactive visualization.

The TensorBoard example source code is available [here](https://git.act3-ace.com/ace/examples/ace-hub-demo/-/tree/master/tb).
