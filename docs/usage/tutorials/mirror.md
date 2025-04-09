# Mirror Tutorial

## Intended Audience

This tutorial is written for Data Tool users who want to efficiently transfer OCI images from one registry to another.

## Guided Scenario

In this scenario, a user is working locally and simulating the process used to transfer large data sets between the low-side and high-side of an air gapped environment.

## Prerequisites

It is assumed that readers have familiarity with Data Tools's key concepts and common usage patterns.

> Consult the [Data Tool User Guide](../user-guide.md), if needed

| Task or Tool | Type | Resources |
| ------------ | ---- | --------- |
| Data Tool configuration file established at `~/.config/ace/dt/config.yaml` | Required | [Data Tool Configuration Guide](../../get-started/configuration-guide.md) |
| Podman installed | Required| [Install ASCE Tools](https://gitlab.com/act3-ai/asce/up/-/blob/main/act3-login/README.md#new-user-setup) or follow [Podman Installation Instructions](https://podman.io/docs/installation) |
| Authentication to `reg.git.act3-ace.com` | Required |Run the [ACT3 Login script](https://gitlab.com/act3-ai/asce/up/-/blob/main/act3-login/README.md) or complete manual registry authentication |
| Connected to ACT3 VPN | Required | |

## Workflow Overview

Users who complete this tutorial will:

- Use Podman to establish two OCI repositories to simulate a low-side and high-side of an air gapped environment
- Add the simulated low-side repository to an existing Data Tool configuration file
- Prepare a `sources.list` file to gather sources into the simulated low-side repository
- Execute the `gather` command to populate the low-side repository
- Inspect the contents of the low-side repository
- Execute the `serialize` command to create a local tar file on our machine to simulate writing data to a storage device brought from the low-side to the high-side of an air gapped environment
- Inspect (list) the content of the tar file without extracting it
- Execute the `deserialize` command to move the contents of the `local.tar` file to the simulated high-side repository
- Prepare a `scatter.tmpl` file using Go templating to map locations between the low-side and high-side
- Execute the `scatter` command to distribute the contents of the tar file to the locations designated in the `scatter.tmpl` file
- Clean up resources used during the tutorial

## Step-by-Step Instructions

### Establish Local OCI Repositories

In this step, we first create and work with two local OCI repositories to establish the simulation of the low-side and high-side of an air gapped environment.

Create a repository in a local registry to represent the low-side of the environment:

```sh
podman run -d -p 5001:5000 --name low docker.io/library/registry:2
```

Create a repository in another local registry to represent the high-side of the environment:

```sh
podman run -d -p 5020:5000 --name high docker.io/library/registry:2
```

Add the low-side local registry to Data Tool's configuration file by pasting the following at the bottom of the `~/.config/ace/dt/config.yaml` file:

```yaml
registryConfig:
  registries:
    localhost:5000:
      endpoints:
        - http://localhost:5000
```

You can modify the yaml to specify the endpoint of the local registry because it is `http`. There are a number of configuration options you can set per repository, but you only need to add the low-side entry with its corresponding fully-qualified endpoint.

### Prepare a sources.list File

The `sources.list` file specifies which images Data Tool should use when the `gather` command is run.

The syntax is to list one OCI image reference per line.

Create a `sources.list` file and populate it with the following:

```text
index.docker.io/library/busybox:1.36.1
ghcr.io/act3-ai/data-tool/bottles/mnist:v1.6
```

In this tutorial we are using our `sources.list` file to specify the two images that we want to populate into our low-side repository.

Busybox is a publicly available image from the Docker.io registry.

MNIST is also publicly available dataset. However, we already have an OCI image of MNIST in the ACT3 GitLab registry, so we are using that image instead.

The MNIST data bottle has predecessors. In a later step, we'll see how to view those predecessors when we inspect the contents of the repository.

### Populate the Low-Side Repository

The `gather` command uses the `sources.list` file to populate the low-side repository.

The syntax is:

```sh
ace-dt mirror gather SOURCES-FILE IMAGE [flags]
```

Run the `gather` command and add `sync-1` as the value of a tag:

```sh
ace-dt mirror gather sources.list localhost:5000/gather:sync-1
```

The expected output should be similar to the following:

```sh
ghcr.io/act3-ai/data-tool/bottles/mnist:v1.6 ↦ sending
index.docker.io/library/busybox:1.36.1 ↦ sending
index.docker.io/library/busybox:1.36.1 ↦ Completed in 2.996s
ghcr.io/act3-ai/data-tool/bottles/mnist:v1.6 ↦ Completed in 3.562s
↦ Gather operation complete. Image location: localhost:5000/gather:sync-1
```

### Inspect the Contents of the Local Repository

The `oci tree` command lets you view the contents of the gather repository.

Since we want to see the MNIST predecessors, we will use the an option show those in the tree view.

The syntax is:

```sh
ace-dt oci tree [-i IMAGE | -d DIR] [flags]
```

Run the `oci tree` command and use the `-p=recursive` option:

```sh
ace-dt oci tree --image localhost:5000/gather:sync-1 -p=recursive
```

The expected output should be similar to the following:

```sh
🗂 index sha256:ef2c10c74db0144d1c04540c94956f76eedc5181f4c38adaaafc5f1c7f3ec033
│   
├── 🗂 index sha256:3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79
│   │   
│   ├── [2.2 MB] 📷 image  sha256:023917ec6a886d0e8e15f28fb543515a5fcd8d938edb091e8147db4efed388ee
│   │   ├── [1.5 kB] ⚙ config application/vnd.docker.container.image.v1+json sha256:a416a98b71e224a31ee99cff8e16063554498227d2b696152a9c3e0aa65e5824
│   │   └── [2.2 MB] 🥞 layer application/vnd.docker.image.rootfs.diff.tar.gzip sha256:3f4d90098f5b5a6f6a76e9d217da85aa39b2081e30fa1f7d287138d6e7bf0ad7
│   │   
│   ├── [1.9 MB] 📷 image  sha256:5cd228af7cde277502487da780b34ba111b8fcdcf37ca518d68c5ba565002b36
│   │   ├── [1.5 kB] ⚙ config application/vnd.docker.container.image.v1+json sha256:3010a01e6ddbec8b36101553aa0fb12bc24c076beb64bd4035cccd06bf58af68
│   │   └── [1.9 MB] 🥞 layer application/vnd.docker.image.rootfs.diff.tar.gzip sha256:d590d3b2456f0de3029267d070954c2a77a5d380727a96b2f3919fa58e50d11f
│   │   
│   ├── [968 kB] 📷 image  sha256:ee899917ce6be185380c8404efb61aa683b649ab2d6a81857887fd746404edbf
│   │   ├── [1.5 kB] ⚙ config application/vnd.docker.container.image.v1+json sha256:9f28bca8fad0857e89b180214699f9a438270ccdbd0658931efd23acdc51f9fd
│   │   └── [968 kB] 🥞 layer application/vnd.docker.image.rootfs.diff.tar.gzip sha256:60966b1a2cb720fd8e684985ff16d257d5356840b68b73cebcaca7beded1eacc
│   │   
│   ├── [1.6 MB] 📷 image  sha256:064a9f60d69ca91b86fbc49a700c3e8971d66939a6832d95afe082722af637cc
│   │   ├── [1.5 kB] ⚙ config application/vnd.docker.container.image.v1+json sha256:87837ce2bf82708058189a7064370555222aea21f077cc51793f9ef4393f4f92
│   │   └── [1.6 MB] 🥞 layer application/vnd.docker.image.rootfs.diff.tar.gzip sha256:0983f321071feac207dd8453ebf1e0749d2b4ccd3b9b2d37820c3e3cd4cff952
│   │   
│   ├── [1.9 MB] 📷 image  sha256:1fa89c01cd0473cedbd1a470abb8c139eeb80920edf1bc55de87851bfb63ea11
│   │   ├── [1.5 kB] ⚙ config application/vnd.docker.container.image.v1+json sha256:fc9db2894f4e4b8c296b8c9dab7e18a6e78de700d21bc0cfaf5c78484226db9c
│   │   └── [1.9 MB] 🥞 layer application/vnd.docker.image.rootfs.diff.tar.gzip sha256:8a0af25e8c2e5dc07c14df3b857877f58bf10c944685cb717b81c5a90974a5ee
│   │   
│   ├── [2.3 MB] 📷 image  sha256:4c6415d8307ac0555e20a047b83678d99063c0e9e89355541e8676d1d98f66a7
│   │   ├── [1.5 kB] ⚙ config application/vnd.docker.container.image.v1+json sha256:44ddfaac8324c393489db23dbbb2a2f4ae18d36a86f1234a4c6bb16e459b5ca0
│   │   └── [2.3 MB] 🥞 layer application/vnd.docker.image.rootfs.diff.tar.gzip sha256:494467eb786caa44e77496badb070b5cf4350de34c72e5b1274bcf628603947e
│   │   
│   ├── [2.1 MB] 📷 image  sha256:a7cf3b49df51803ce6168cb56dd786055e92aacbb4f503f8aa2842e9069344b2
│   │   ├── [1.5 kB] ⚙ config application/vnd.docker.container.image.v1+json sha256:f00beaa03019019506b7b202dd1ea2b4af72830daf1681c266c82c5078f804f0
│   │   └── [2.1 MB] 🥞 layer application/vnd.docker.image.rootfs.diff.tar.gzip sha256:abd7b9dd25e08de7ee13c6b0a5621e8e137b15b27254fd4a43a97824afc0c945
│   │   
│   ├── [2.5 MB] 📷 image  sha256:87c45b26a9c5a7aa69d9c145ecb9722bff6a1592cf8de7001e3b86ca33566587
│   │   ├── [1.5 kB] ⚙ config application/vnd.docker.container.image.v1+json sha256:baefdf18d0dee1b2a81875425f67136da27ed45afa427bdd84e466603cb27c62
│   │   └── [2.5 MB] 🥞 layer application/vnd.docker.image.rootfs.diff.tar.gzip sha256:f83bd7e3defd32ffd3efeffe28741d886821784023a767d91e4a754768dcbcd8
│   │   
│   ├── [907 kB] 📷 image  sha256:1411f4a8c78f5fadafa8f733e71f6ff01dfd637263ae090d68511a6e152451e3
│   │   ├── [1.5 kB] ⚙ config application/vnd.docker.container.image.v1+json sha256:a58323eebc13bd1a9d3ca1dd0840d04a45ef0fc58e2c9516d533672f42fa36e1
│   │   └── [907 kB] 🥞 layer application/vnd.docker.image.rootfs.diff.tar.gzip sha256:72ed4adcd9404b88ec314167db591da57059bedf2e5f601303b0eb4a9ab30b9c
│   │   
│   └── [1.9 MB] 📷 image  sha256:7d602b12b1d9c1bdbf4c9255c0ba276ac0d9e0cd781a7c13461e4875cfcae509
│       ├── [1.5 kB] ⚙ config application/vnd.docker.container.image.v1+json sha256:9e8eec71a0033c211daa85cf5557b161e90d761ddbb969dd2db6191bf42b6879
│       └── [1.9 MB] 🥞 layer application/vnd.docker.image.rootfs.diff.tar.gzip sha256:c9ce7e59cf2a387f3495e850aca1229787b67873a9fde8675db605cd03a8e1d4
│   18 MB index total
│   
└── [ 24 MB] 📷 image  sha256:c04a3b5bc5b0513a6c357171d1ebe24d16f490f7a066811f2975ad307336c0e6
    ├── [1.6 kB] ⚙ config application/vnd.act3-ace.bottle.config.v1+json sha256:caea335922ea621bb684fd3bc9db1d6d36683cf8191e235dbd025f8d3353b922
    ├── [ 12 MB] 🥞 layer application/vnd.act3-ace.bottle.layer.v1.tar+zstd sha256:2e742b1e2f6f807ad61452da5c76d2fc261e3c3a5b9d7f7c19c19862176ddee4
    ├── [1.7 MB] 🥞 layer application/vnd.act3-ace.bottle.layer.v1+zstd sha256:8f6220977ae1e1b34cf44941f5a2679a452a03576c6c93392cc093031975f67a
    ├── [4.0 kB] 🥞 layer application/vnd.act3-ace.bottle.layer.v1+zstd sha256:dab3320aa5a5b7a6bac50fbba6796577085b1a53bbcab57dd606e3f7aeccf1fb
    ├── [ 10 MB] 🥞 layer application/vnd.act3-ace.bottle.layer.v1+zstd sha256:0da9e73ffe1481abd3828c8495813d5798dbbb3b90e010622a09f69724a7f000
    ├── [ 27 kB] 🥞 layer application/vnd.act3-ace.bottle.layer.v1+zstd sha256:717713c7efb6c8013d9a4fae72af7d8ee3dc668677ea57283fc4915f9e71d80a
    └── [236 kB] 🥞 layer application/vnd.act3-ace.bottle.layer.v1+zstd sha256:a7bf576b9fe02aed34df729c4e954395d478aaaef76a9d064305299f7f8e6be8
42 MB index total
0 B deduplicated total (predecessors)
42 MB deduplicated total
```

### Serialize Data for Transfer

Typically, the `serialize` command is directed to a tape drive with custom buffer and block size flags. However, we can also run `serialize` to create a local tar file on our machine.

The syntax is:

```sh
ace-dt mirror serialize IMAGE DEST [EXISTING-IMAGE...] [flags]
```

Run the `serialize` command on the gather repository `localhost:5000/gather:sync-1` and direct the output as a local tar file:

```sh
ace-dt mirror serialize localhost:5000/gather:sync-1 local.tar
```

The expected output should be similar to the following:

```sh
Existing ↦ Completed in 0s
Manifest sha256:ef2c1... ↦ Processing manifest sha256:ef2c10c74db0144d1c04540c94956f76eedc5181f4c38adaaafc5f1c7f3ec033
Manifest sha256:ef2c1...|Manifest sha256:3fbc6... ↦ Processing manifest sha256:3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:02391... ↦ Processing manifest sha256:023917ec6a886d0e8e15f28fb543515a5fcd8d938edb091e8147db4efed388ee
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:02391...|Blob sha256:a416a... ↦ Writing blob (1457 B) sha256:a416a98b71e224a31ee99cff8e16063554498227d2b696152a9c3e0aa65e5824
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:02391...|Blob sha256:a416a... ↦ Completed in 3ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:02391...|Blob sha256:3f4d9... ↦ Writing blob (2219949 B) sha256:3f4d90098f5b5a6f6a76e9d217da85aa39b2081e30fa1f7d287138d6e7bf0ad7
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:02391...|Blob sha256:3f4d9... ↦ Completed in 10ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:02391... ↦ Completed [2] in 15ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:5cd22... ↦ Processing manifest sha256:5cd228af7cde277502487da780b34ba111b8fcdcf37ca518d68c5ba565002b36
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:5cd22...|Blob sha256:3010a... ↦ Writing blob (1470 B) sha256:3010a01e6ddbec8b36101553aa0fb12bc24c076beb64bd4035cccd06bf58af68
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:5cd22...|Blob sha256:3010a... ↦ Completed in 2ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:5cd22...|Blob sha256:d590d... ↦ Writing blob (1855500 B) sha256:d590d3b2456f0de3029267d070954c2a77a5d380727a96b2f3919fa58e50d11f
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:5cd22...|Blob sha256:d590d... ↦ Completed in 8ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:5cd22... ↦ Completed [2] in 12ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:ee899... ↦ Processing manifest sha256:ee899917ce6be185380c8404efb61aa683b649ab2d6a81857887fd746404edbf
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:ee899...|Blob sha256:9f28b... ↦ Writing blob (1470 B) sha256:9f28bca8fad0857e89b180214699f9a438270ccdbd0658931efd23acdc51f9fd
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:ee899...|Blob sha256:9f28b... ↦ Completed in 1ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:ee899...|Blob sha256:60966... ↦ Writing blob (967926 B) sha256:60966b1a2cb720fd8e684985ff16d257d5356840b68b73cebcaca7beded1eacc
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:ee899...|Blob sha256:60966... ↦ Completed in 4ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:ee899... ↦ Completed [2] in 8ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:064a9... ↦ Processing manifest sha256:064a9f60d69ca91b86fbc49a700c3e8971d66939a6832d95afe082722af637cc
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:064a9...|Blob sha256:87837... ↦ Writing blob (1469 B) sha256:87837ce2bf82708058189a7064370555222aea21f077cc51793f9ef4393f4f92
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:064a9...|Blob sha256:87837... ↦ Completed in 2ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:064a9...|Blob sha256:0983f... ↦ Writing blob (1606683 B) sha256:0983f321071feac207dd8453ebf1e0749d2b4ccd3b9b2d37820c3e3cd4cff952
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:064a9...|Blob sha256:0983f... ↦ Completed in 7ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:064a9... ↦ Completed [2] in 11ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1fa89... ↦ Processing manifest sha256:1fa89c01cd0473cedbd1a470abb8c139eeb80920edf1bc55de87851bfb63ea11
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1fa89...|Blob sha256:fc9db... ↦ Writing blob (1472 B) sha256:fc9db2894f4e4b8c296b8c9dab7e18a6e78de700d21bc0cfaf5c78484226db9c
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1fa89...|Blob sha256:fc9db... ↦ Completed in 1ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1fa89...|Blob sha256:8a0af... ↦ Writing blob (1916632 B) sha256:8a0af25e8c2e5dc07c14df3b857877f58bf10c944685cb717b81c5a90974a5ee
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1fa89...|Blob sha256:8a0af... ↦ Completed in 8ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1fa89... ↦ Completed [2] in 12ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:4c641... ↦ Processing manifest sha256:4c6415d8307ac0555e20a047b83678d99063c0e9e89355541e8676d1d98f66a7
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:4c641...|Blob sha256:44ddf... ↦ Writing blob (1455 B) sha256:44ddfaac8324c393489db23dbbb2a2f4ae18d36a86f1234a4c6bb16e459b5ca0
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:4c641...|Blob sha256:44ddf... ↦ Completed in 2ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:4c641...|Blob sha256:49446... ↦ Writing blob (2268288 B) sha256:494467eb786caa44e77496badb070b5cf4350de34c72e5b1274bcf628603947e
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:4c641...|Blob sha256:49446... ↦ Completed in 8ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:4c641... ↦ Completed [2] in 12ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:a7cf3... ↦ Processing manifest sha256:a7cf3b49df51803ce6168cb56dd786055e92aacbb4f503f8aa2842e9069344b2
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:a7cf3...|Blob sha256:f00be... ↦ Writing blob (1458 B) sha256:f00beaa03019019506b7b202dd1ea2b4af72830daf1681c266c82c5078f804f0
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:a7cf3...|Blob sha256:f00be... ↦ Completed in 2ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:a7cf3...|Blob sha256:abd7b... ↦ Writing blob (2123242 B) sha256:abd7b9dd25e08de7ee13c6b0a5621e8e137b15b27254fd4a43a97824afc0c945
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:a7cf3...|Blob sha256:abd7b... ↦ Completed in 8ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:a7cf3... ↦ Completed [2] in 13ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:87c45... ↦ Processing manifest sha256:87c45b26a9c5a7aa69d9c145ecb9722bff6a1592cf8de7001e3b86ca33566587
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:87c45...|Blob sha256:baefd... ↦ Writing blob (1459 B) sha256:baefdf18d0dee1b2a81875425f67136da27ed45afa427bdd84e466603cb27c62
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:87c45...|Blob sha256:baefd... ↦ Completed in 1ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:87c45...|Blob sha256:f83bd... ↦ Writing blob (2528595 B) sha256:f83bd7e3defd32ffd3efeffe28741d886821784023a767d91e4a754768dcbcd8
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:87c45...|Blob sha256:f83bd... ↦ Completed in 13ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:87c45... ↦ Completed [2] in 17ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1411f... ↦ Processing manifest sha256:1411f4a8c78f5fadafa8f733e71f6ff01dfd637263ae090d68511a6e152451e3
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1411f...|Blob sha256:a5832... ↦ Writing blob (1459 B) sha256:a58323eebc13bd1a9d3ca1dd0840d04a45ef0fc58e2c9516d533672f42fa36e1
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1411f...|Blob sha256:a5832... ↦ Completed in 2ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1411f...|Blob sha256:72ed4... ↦ Writing blob (907206 B) sha256:72ed4adcd9404b88ec314167db591da57059bedf2e5f601303b0eb4a9ab30b9c
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1411f...|Blob sha256:72ed4... ↦ Completed in 5ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:1411f... ↦ Completed [2] in 9ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:7d602... ↦ Processing manifest sha256:7d602b12b1d9c1bdbf4c9255c0ba276ac0d9e0cd781a7c13461e4875cfcae509
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:7d602...|Blob sha256:9e8ee... ↦ Writing blob (1457 B) sha256:9e8eec71a0033c211daa85cf5557b161e90d761ddbb969dd2db6191bf42b6879
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:7d602...|Blob sha256:9e8ee... ↦ Completed in 1ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:7d602...|Blob sha256:c9ce7... ↦ Writing blob (1927220 B) sha256:c9ce7e59cf2a387f3495e850aca1229787b67873a9fde8675db605cd03a8e1d4
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:7d602...|Blob sha256:c9ce7... ↦ Completed in 7ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6...|Manifest sha256:7d602... ↦ Completed [2] in 12ms
Manifest sha256:ef2c1...|Manifest sha256:3fbc6... ↦ Completed [10] in 127ms
Manifest sha256:ef2c1...|Manifest sha256:c04a3... ↦ Processing manifest sha256:c04a3b5bc5b0513a6c357171d1ebe24d16f490f7a066811f2975ad307336c0e6
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:caea3... ↦ Writing blob (1583 B) sha256:caea335922ea621bb684fd3bc9db1d6d36683cf8191e235dbd025f8d3353b922
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:caea3... ↦ Completed in 2ms
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:2e742... ↦ Writing blob (11796908 B) sha256:2e742b1e2f6f807ad61452da5c76d2fc261e3c3a5b9d7f7c19c19862176ddee4
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:2e742... ↦ Completed in 36ms
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:8f622... ↦ Writing blob (1676892 B) sha256:8f6220977ae1e1b34cf44941f5a2679a452a03576c6c93392cc093031975f67a
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:8f622... ↦ Completed in 7ms
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:dab33... ↦ Writing blob (4020 B) sha256:dab3320aa5a5b7a6bac50fbba6796577085b1a53bbcab57dd606e3f7aeccf1fb
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:dab33... ↦ Completed in 1ms
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:0da9e... ↦ Writing blob (10088258 B) sha256:0da9e73ffe1481abd3828c8495813d5798dbbb3b90e010622a09f69724a7f000
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:0da9e... ↦ Completed in 39ms
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:71771... ↦ Writing blob (27277 B) sha256:717713c7efb6c8013d9a4fae72af7d8ee3dc668677ea57283fc4915f9e71d80a
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:71771... ↦ Completed in 2ms
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:a7bf5... ↦ Writing blob (235524 B) sha256:a7bf576b9fe02aed34df729c4e954395d478aaaef76a9d064305299f7f8e6be8
Manifest sha256:ef2c1...|Manifest sha256:c04a3...|Blob sha256:a7bf5... ↦ Completed in 3ms
Manifest sha256:ef2c1...|Manifest sha256:c04a3... ↦ Completed [7] in 94ms
Manifest sha256:ef2c1... ↦ Completed [2] in 227ms
Buffer Fill ↦ Completed 274 kB in 233ms (1.2 kB/s)
Byte Savings ↦ Completed 42 MB in 233ms (181 kB/s)
 ↦ Serialize action completed
Buffer Fill ↦ 699 kB of 1.0 MB (66.7%) 350 kB/s ‖ Byte Savings ↦ 32 MB of 42 MB (75.9%) 17 MB/s ‖ Manifest sha256:ef2c1... ↦ [1/2]
```

### Inspect the Content of the Tar File

The `tar` command can be used with the `tvf` flags to list the content of the tar file. This allows us to inspect the contents of the tar file without extracting the compressed files.

Run the `tar` command and add the `tvf` flags to show the contents of the tar file we created in the previous step:

```sh
tar tvf local.tar 
```

The expected output should be similar to the following:

```sh
-rw-rw-rw- 0/0              30 1969-12-31 19:00 oci-layout
-rw-rw-rw- 0/0             491 1969-12-31 19:00 index.json
drwxrwxrwx 0/0               0 1969-12-31 19:00 blobs
drwxrwxrwx 0/0               0 1969-12-31 19:00 blobs/sha256
-rw-rw-rw- 0/0             910 1969-12-31 19:00 blobs/sha256/ef2c10c74db0144d1c04540c94956f76eedc5181f4c38adaaafc5f1c7f3ec033
-rw-rw-rw- 0/0            2295 1969-12-31 19:00 blobs/sha256/3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79
-rw-rw-rw- 0/0             528 1969-12-31 19:00 blobs/sha256/023917ec6a886d0e8e15f28fb543515a5fcd8d938edb091e8147db4efed388ee
-rw-rw-rw- 0/0            1457 1969-12-31 19:00 blobs/sha256/a416a98b71e224a31ee99cff8e16063554498227d2b696152a9c3e0aa65e5824
-rw-rw-rw- 0/0         2219949 1969-12-31 19:00 blobs/sha256/3f4d90098f5b5a6f6a76e9d217da85aa39b2081e30fa1f7d287138d6e7bf0ad7
-rw-rw-rw- 0/0             528 1969-12-31 19:00 blobs/sha256/5cd228af7cde277502487da780b34ba111b8fcdcf37ca518d68c5ba565002b36
-rw-rw-rw- 0/0            1470 1969-12-31 19:00 blobs/sha256/3010a01e6ddbec8b36101553aa0fb12bc24c076beb64bd4035cccd06bf58af68
-rw-rw-rw- 0/0         1855500 1969-12-31 19:00 blobs/sha256/d590d3b2456f0de3029267d070954c2a77a5d380727a96b2f3919fa58e50d11f
-rw-rw-rw- 0/0             527 1969-12-31 19:00 blobs/sha256/ee899917ce6be185380c8404efb61aa683b649ab2d6a81857887fd746404edbf
-rw-rw-rw- 0/0            1470 1969-12-31 19:00 blobs/sha256/9f28bca8fad0857e89b180214699f9a438270ccdbd0658931efd23acdc51f9fd
-rw-rw-rw- 0/0          967926 1969-12-31 19:00 blobs/sha256/60966b1a2cb720fd8e684985ff16d257d5356840b68b73cebcaca7beded1eacc
-rw-rw-rw- 0/0             528 1969-12-31 19:00 blobs/sha256/064a9f60d69ca91b86fbc49a700c3e8971d66939a6832d95afe082722af637cc
-rw-rw-rw- 0/0            1469 1969-12-31 19:00 blobs/sha256/87837ce2bf82708058189a7064370555222aea21f077cc51793f9ef4393f4f92
-rw-rw-rw- 0/0         1606683 1969-12-31 19:00 blobs/sha256/0983f321071feac207dd8453ebf1e0749d2b4ccd3b9b2d37820c3e3cd4cff952
-rw-rw-rw- 0/0             528 1969-12-31 19:00 blobs/sha256/1fa89c01cd0473cedbd1a470abb8c139eeb80920edf1bc55de87851bfb63ea11
-rw-rw-rw- 0/0            1472 1969-12-31 19:00 blobs/sha256/fc9db2894f4e4b8c296b8c9dab7e18a6e78de700d21bc0cfaf5c78484226db9c
-rw-rw-rw- 0/0         1916632 1969-12-31 19:00 blobs/sha256/8a0af25e8c2e5dc07c14df3b857877f58bf10c944685cb717b81c5a90974a5ee
-rw-rw-rw- 0/0             528 1969-12-31 19:00 blobs/sha256/4c6415d8307ac0555e20a047b83678d99063c0e9e89355541e8676d1d98f66a7
-rw-rw-rw- 0/0            1455 1969-12-31 19:00 blobs/sha256/44ddfaac8324c393489db23dbbb2a2f4ae18d36a86f1234a4c6bb16e459b5ca0
-rw-rw-rw- 0/0         2268288 1969-12-31 19:00 blobs/sha256/494467eb786caa44e77496badb070b5cf4350de34c72e5b1274bcf628603947e
-rw-rw-rw- 0/0             528 1969-12-31 19:00 blobs/sha256/a7cf3b49df51803ce6168cb56dd786055e92aacbb4f503f8aa2842e9069344b2
-rw-rw-rw- 0/0            1458 1969-12-31 19:00 blobs/sha256/f00beaa03019019506b7b202dd1ea2b4af72830daf1681c266c82c5078f804f0
-rw-rw-rw- 0/0         2123242 1969-12-31 19:00 blobs/sha256/abd7b9dd25e08de7ee13c6b0a5621e8e137b15b27254fd4a43a97824afc0c945
-rw-rw-rw- 0/0             528 1969-12-31 19:00 blobs/sha256/87c45b26a9c5a7aa69d9c145ecb9722bff6a1592cf8de7001e3b86ca33566587
-rw-rw-rw- 0/0            1459 1969-12-31 19:00 blobs/sha256/baefdf18d0dee1b2a81875425f67136da27ed45afa427bdd84e466603cb27c62
-rw-rw-rw- 0/0         2528595 1969-12-31 19:00 blobs/sha256/f83bd7e3defd32ffd3efeffe28741d886821784023a767d91e4a754768dcbcd8
-rw-rw-rw- 0/0             527 1969-12-31 19:00 blobs/sha256/1411f4a8c78f5fadafa8f733e71f6ff01dfd637263ae090d68511a6e152451e3
-rw-rw-rw- 0/0            1459 1969-12-31 19:00 blobs/sha256/a58323eebc13bd1a9d3ca1dd0840d04a45ef0fc58e2c9516d533672f42fa36e1
-rw-rw-rw- 0/0          907206 1969-12-31 19:00 blobs/sha256/72ed4adcd9404b88ec314167db591da57059bedf2e5f601303b0eb4a9ab30b9c
-rw-rw-rw- 0/0             528 1969-12-31 19:00 blobs/sha256/7d602b12b1d9c1bdbf4c9255c0ba276ac0d9e0cd781a7c13461e4875cfcae509
-rw-rw-rw- 0/0            1457 1969-12-31 19:00 blobs/sha256/9e8eec71a0033c211daa85cf5557b161e90d761ddbb969dd2db6191bf42b6879
-rw-rw-rw- 0/0         1927220 1969-12-31 19:00 blobs/sha256/c9ce7e59cf2a387f3495e850aca1229787b67873a9fde8675db605cd03a8e1d4
-rw-rw-rw- 0/0            1214 1969-12-31 19:00 blobs/sha256/c04a3b5bc5b0513a6c357171d1ebe24d16f490f7a066811f2975ad307336c0e6
-rw-rw-rw- 0/0            1583 1969-12-31 19:00 blobs/sha256/caea335922ea621bb684fd3bc9db1d6d36683cf8191e235dbd025f8d3353b922
-rw-rw-rw- 0/0        11796908 1969-12-31 19:00 blobs/sha256/2e742b1e2f6f807ad61452da5c76d2fc261e3c3a5b9d7f7c19c19862176ddee4
-rw-rw-rw- 0/0         1676892 1969-12-31 19:00 blobs/sha256/8f6220977ae1e1b34cf44941f5a2679a452a03576c6c93392cc093031975f67a
-rw-rw-rw- 0/0            4020 1969-12-31 19:00 blobs/sha256/dab3320aa5a5b7a6bac50fbba6796577085b1a53bbcab57dd606e3f7aeccf1fb
-rw-rw-rw- 0/0        10088258 1969-12-31 19:00 blobs/sha256/0da9e73ffe1481abd3828c8495813d5798dbbb3b90e010622a09f69724a7f000
-rw-rw-rw- 0/0           27277 1969-12-31 19:00 blobs/sha256/717713c7efb6c8013d9a4fae72af7d8ee3dc668677ea57283fc4915f9e71d80a
-rw-rw-rw- 0/0          235524 1969-12-31 19:00 blobs/sha256/a7bf576b9fe02aed34df729c4e954395d478aaaef76a9d064305299f7f8e6be8
-rw-rw-rw- 0/0            4219 1969-12-31 19:00 index.json
```

You will notice that the content of this tar file aligns with the [OCI image layout specification](https://github.com/opencontainers/image-spec/blob/main/image-layout.md).

### Deserialize Data for Distribution

At this point in an actual air gapped environment workflow, a sleigh process would be used to transfer the physical media storage between the low-side and high-side of the air gapped environment. We are only simulating that process in this tutorial, so that aspect is not addressed here.

Typically, the `deserialize` command is directed to a tape drive where the output of the `serialize` command was saved in an earlier step. However, we can also run deserialize from the local tar file on our machine to the simulated high-side repository, which in this tutorial is also on the local machine.

The syntax is:

```sh
ace-dt mirror deserialize SOURCE-FILE IMAGE [flags]
```

Run the `deserialize` command and direct the output to the local repository representing the high-side of the environment:

```sh
ace-dt mirror deserialize local.tar localhost:5000/deserialize:sync-1
```

The expected output should be similar to the following:

```sh
Deserializing ↦ oci-layout
 ↦ bytes read: 824635972152 B
Deserializing ↦ index.json
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/ef2c10c74db0144d1c04540c94956f76eedc5181f4c38adaaafc5f1c7f3ec033
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/023917ec6a886d0e8e15f28fb543515a5fcd8d938edb091e8147db4efed388ee
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/a416a98b71e224a31ee99cff8e16063554498227d2b696152a9c3e0aa65e5824
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/3f4d90098f5b5a6f6a76e9d217da85aa39b2081e30fa1f7d287138d6e7bf0ad7
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/5cd228af7cde277502487da780b34ba111b8fcdcf37ca518d68c5ba565002b36
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/3010a01e6ddbec8b36101553aa0fb12bc24c076beb64bd4035cccd06bf58af68
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/d590d3b2456f0de3029267d070954c2a77a5d380727a96b2f3919fa58e50d11f
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/ee899917ce6be185380c8404efb61aa683b649ab2d6a81857887fd746404edbf
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/9f28bca8fad0857e89b180214699f9a438270ccdbd0658931efd23acdc51f9fd
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/60966b1a2cb720fd8e684985ff16d257d5356840b68b73cebcaca7beded1eacc
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/064a9f60d69ca91b86fbc49a700c3e8971d66939a6832d95afe082722af637cc
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/87837ce2bf82708058189a7064370555222aea21f077cc51793f9ef4393f4f92
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/0983f321071feac207dd8453ebf1e0749d2b4ccd3b9b2d37820c3e3cd4cff952
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/1fa89c01cd0473cedbd1a470abb8c139eeb80920edf1bc55de87851bfb63ea11
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/fc9db2894f4e4b8c296b8c9dab7e18a6e78de700d21bc0cfaf5c78484226db9c
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/8a0af25e8c2e5dc07c14df3b857877f58bf10c944685cb717b81c5a90974a5ee
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/4c6415d8307ac0555e20a047b83678d99063c0e9e89355541e8676d1d98f66a7
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/44ddfaac8324c393489db23dbbb2a2f4ae18d36a86f1234a4c6bb16e459b5ca0
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/494467eb786caa44e77496badb070b5cf4350de34c72e5b1274bcf628603947e
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/a7cf3b49df51803ce6168cb56dd786055e92aacbb4f503f8aa2842e9069344b2
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/f00beaa03019019506b7b202dd1ea2b4af72830daf1681c266c82c5078f804f0
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/abd7b9dd25e08de7ee13c6b0a5621e8e137b15b27254fd4a43a97824afc0c945
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/87c45b26a9c5a7aa69d9c145ecb9722bff6a1592cf8de7001e3b86ca33566587
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/baefdf18d0dee1b2a81875425f67136da27ed45afa427bdd84e466603cb27c62
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/f83bd7e3defd32ffd3efeffe28741d886821784023a767d91e4a754768dcbcd8
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/1411f4a8c78f5fadafa8f733e71f6ff01dfd637263ae090d68511a6e152451e3
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/a58323eebc13bd1a9d3ca1dd0840d04a45ef0fc58e2c9516d533672f42fa36e1
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/72ed4adcd9404b88ec314167db591da57059bedf2e5f601303b0eb4a9ab30b9c
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/7d602b12b1d9c1bdbf4c9255c0ba276ac0d9e0cd781a7c13461e4875cfcae509
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/9e8eec71a0033c211daa85cf5557b161e90d761ddbb969dd2db6191bf42b6879
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/c9ce7e59cf2a387f3495e850aca1229787b67873a9fde8675db605cd03a8e1d4
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/c04a3b5bc5b0513a6c357171d1ebe24d16f490f7a066811f2975ad307336c0e6
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/caea335922ea621bb684fd3bc9db1d6d36683cf8191e235dbd025f8d3353b922
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/2e742b1e2f6f807ad61452da5c76d2fc261e3c3a5b9d7f7c19c19862176ddee4
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/8f6220977ae1e1b34cf44941f5a2679a452a03576c6c93392cc093031975f67a
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/dab3320aa5a5b7a6bac50fbba6796577085b1a53bbcab57dd606e3f7aeccf1fb
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/0da9e73ffe1481abd3828c8495813d5798dbbb3b90e010622a09f69724a7f000
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/717713c7efb6c8013d9a4fae72af7d8ee3dc668677ea57283fc4915f9e71d80a
 ↦ bytes read: 824635972152 B
Deserializing ↦ blobs/sha256/a7bf576b9fe02aed34df729c4e954395d478aaaef76a9d064305299f7f8e6be8
 ↦ bytes read: 824635972152 B
Deserializing ↦ index.json
Tagging root node ↦ Completed in 18ms
Deserializing ↦ Completed in 1.425s
```

### Prepare a scatter.tmpl File

The `scatter.tmpl` file specifies which images Data Tool should use when the `scatter` command is run.

Create a `scatter.tmpl` file

```sh
touch scatter.tmpl
```

Use your editor of choice to copy and paste the following contents into the `scatter.tmpl` file:

```go
{{- $annotation := index .Annotations "vnd.act3-ace.manifest.source" -}}
localhost:5000/{{ $annotation }}
```

Using the above template:

- `index.docker.io/library/busybox:1.36.1` would be mapped to `localhost:5000/index.docker.io/library/busybox:1.36.1`.
- `ghcr.io/act3-ai/data-tool/bottles/mnist:v1.6` would be mapped to `localhost:5000/ghcr.io/act3-ai/data-tool/bottles/mnist:v1.6`.

#### Use of Go Templating

Data Tool's mirror features have the ability to use Go language templating to map the images from the deserialize repository to a designated location in the high-side environment.

The simple template file used in this tutorial will send each image to a repository based on its original location on the low-side.

### Populate the High-Side Repository

The `scatter` command uses the `scatter.tmpl` file to distribute or *scatter* the contents of the tar file to the designated location(s) in the high-side environment.

The syntax is:

```sh
ace-dt mirror scatter SOURCE-FILE IMAGE [flags]
```

Run the scatter command using the `scatter.tmpl` file created in the previous step:

```sh
ace-dt mirror scatter localhost:5000/deserialize:sync-1 go-template=scatter.tmpl
```

The expected output should be similar to the following:

```sh
artifact 1/2|destination 1/1 ↦ sending ghcr.io/act3-ai/data-tool/bottles/mnist:v1.6 to localhost:5000/ghcr.io/act3-ai/data-tool/bottles/mnist:v1.6
artifact 2/2|destination 1/1 ↦ sending index.docker.io/library/busybox:1.36.1 to localhost:5000/index.docker.io/library/busybox:1.36.1
artifact 1/2|destination 1/1 ↦ Completed in 60ms
artifact 1/2 ↦ Completed [1] in 60ms
artifact 2/2|destination 1/1 ↦ Completed in 112ms
artifact 2/2 ↦ Completed [1] in 112ms
```

We have successfully gathered images from a local simulated low-side environment, mirrored them to a tar file, and used Go templating to scatter them to a local simulated high-side environment to approximate the workflow that would be followed when using Data Tool to move OCI images into an air gapped environment.

### Clean Up Resources

To conclude this tutorial, we will clean up the resources we created and used.

Stop and remove the local repositories and registries:

```sh
podman container stop registry && podman container rm -v registry
```

You may also want to clean up the following files used in this tutorial:

- `sources.list`
- `local.tar`
- `scatter.tmpl`
