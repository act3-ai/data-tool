# ace-dt git

## Overview

`ace-dt git` is a command group for performing Git to OCI or OCI to Git operations. Such operations allow a user to
create a copy of a Git repository, storing it in an OCI compliant registry of their choice. The copy in OCI format is referred to as a "Commit Manifest" as it contains Git commits, as well as tag/head references.

## Across Air-Gapped Networks

`ace-dt git` works with `ace-dt mirror` to create or update copies of git repositories within air-gapped network.

Process:

1. Create or update a Commit Manifest containing a copy of a Git repository with `ace-dt git to-oci`.
2. Transfer the Commit Manifest to the air-gapped network with `ace-dt mirror`.
3. Create or update a Git repository within the air-gapped network with `ace-dt git from-oci`.

## Example Workflow

### Create a New Commit Manifest

#### If the tag reference does not exist

```console
$ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/examples/libgit2:sync v1.6.1
 ↦ Manifest digest: sha256:e6075ad2a2752a546753a056afb4e61a32a391bac8658e75450471f6a4862f65
 ↦ Commit Manifest update complete.
```

#### If the tag reference already exists, you may start clean and overwrite it

```console
$ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/examples/libgit2:sync v1.6.1 --clean
 ↦ Manifest digest: sha256:e6075ad2a2752a546753a056afb4e61a32a391bac8658e75450471f6a4862f65
 ↦ Commit Manifest update complete.
```

### Create a New Git Repository from the Commit Manifest

```console
$ ace-dt git from-oci reg.example.com/examples/libgit2:sync ~/destinationRepository/example/
 ↦ Git repository update complete. The following references have been updated:

 ↦ 8a871d13b7f4e186b8ad943ae5a7fcf30be52e67 refs/tags/v1.6.1
```

#### All Repository References

```console
$ git show-ref
8a871d13b7f4e186b8ad943ae5a7fcf30be52e67 refs/tags/v1.6.1
```

OR

```console
$ ace-dt git list-refs reg.example.com/examples/libgit2:sync
 ↦ Digest of reg.example.com/examples/libgit2:sync2: sha256:e6075ad2a2752a546753a056afb4e61a32a391bac8658e75450471f6a4862f65
 ↦ References:
 ↦ 8a871d13b7f4e186b8ad943ae5a7fcf30be52e67 refs/tags/v1.6.1
```

### Update an Existing Commit Manifest

```console
$ ace-dt git to-oci https://github.com/libgit2/libgit2 reg.example.com/examples/libgit2:sync v1.6.2
 ↦ Manifest digest: sha256:957b0ea2506585fd86d4796557ac080b0ab8a79808c490458a22b44a185adf34
 ↦ Commit Manifest update complete.
```

### Update a Git Repository from the Updated Commit Manifest

```console
$ ace-dt git from-oci reg.example.com/examples/libgit2:sync ~/destinationRepository/example/ 
 ↦ Git repository update complete. The following references have been updated:

 ↦ 25ec37379ed07b10c4ecc6143cf6018cabc8f857 refs/tags/v1.6.2
```

#### All Repository References Again

```console
$ git show-ref
8a871d13b7f4e186b8ad943ae5a7fcf30be52e67 refs/tags/v1.6.1
25ec37379ed07b10c4ecc6143cf6018cabc8f857 refs/tags/v1.6.2
```

OR

```console
$ ace-dt git list-refs reg.example.com/examples/libgit2:sync
 ↦ Digest of reg.example.com/examples/libgit2:sync2: sha256:e6075ad2a2752a546753a056afb4e61a32a391bac8658e75450471f6a4862f65
 ↦ References:
 ↦ 8a871d13b7f4e186b8ad943ae5a7fcf30be52e67 refs/tags/v1.6.1
 ↦ 25ec37379ed07b10c4ecc6143cf6018cabc8f857 refs/tags/v1.6.2
```
