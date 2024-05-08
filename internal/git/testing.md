# Results of test repository creation functions

## Result of createGitRepo

```mermaid
---
title: Generated Git Repo
---
%%{init: { 'logLevel': 'debug', 'theme': 'dark', 'gitGraph': {'rotateCommitLabel': true}} }%%

gitGraph LR:
   commit tag: "v1.0.0"
   commit tag: "v1.0.1"
   branch Feature1
   checkout Feature1
   commit
   commit
   commit id: "head" tag: "v1.0.2"
   checkout main
   branch Feature2
   checkout Feature2
   commit
   commit
   commit id: "Head" tag: "v1.0.3"
   checkout main
   merge Feature1
   checkout main
   commit
   commit
   commit tag: "v1.2.0"
   commit
   commit id: "Main Head"
```

## Base Test Cases

Note: The location of changes are highlighted (a square).

### Base Layer

```mermaid
gitGraph
   commit
   commit tag: "v1.0.1" type: HIGHLIGHT

```

### Add Tag Ref to New Commit

```mermaid
gitGraph
   commit
   commit tag: "v1.0.1"
   branch Feature1
   checkout Feature1
   commit
   commit
   commit tag: "v1.0.2" type: HIGHLIGHT
```

This test does not contain the branch head ref, only the tag on the branch.

### Add Head Ref to New Commit

```mermaid
gitGraph
   commit
   commit tag: "v1.0.1"
   branch Feature1
   checkout Feature1
   commit
   commit
   commit tag: "v1.0.2"
   checkout main
   branch Feature2
   commit
   commit
   commit id: "Head" type: HIGHLIGHT
```

This test does not contain the v1.0.3 tag, but does contain the head reference.

### Add Tag Ref to Existing Branch Head Ref

```mermaid
gitGraph
   commit
   commit tag: "v1.0.1"
   branch Feature1
   checkout Feature1
   commit
   commit
   commit tag: "v1.0.2"
   checkout main
   branch Feature2
   commit
   commit
   commit id: "Head" tag: "v1.0.3" type: HIGHLIGHT
```

This tests adds the v1.0.3 tag to the existing head reference.

### Add Branch Head Ref to Existing Tag Ref

```mermaid
gitGraph
   commit
   commit tag: "v1.0.1"
   branch Feature1
   checkout Feature1
   commit
   commit
   commit id: "head" tag: "v1.0.2" type: HIGHLIGHT
   checkout main
   branch Feature2
   commit
   commit
   commit id: "Head" tag: "v1.0.3"
```

### Add Tag Ref to Existing Commit

```mermaid
gitGraph
   commit tag: "v1.0.0" type: HIGHLIGHT
   commit tag: "v1.0.1"
   branch Feature1
   checkout Feature1
   commit
   commit
   commit id: "head" tag: "v1.0.2"
   checkout main
   branch Feature2
   commit
   commit
   commit id: "Head" tag: "v1.0.3"
```

### Add Tag Ref to New Commit - For Commit Tips Min Set

```mermaid
gitGraph
   commit tag: "v1.0.0"
   commit tag: "v1.0.1"
   branch Feature1
   checkout Feature1
   commit
   commit
   commit id: "head" tag: "v1.0.2"
   checkout main
   branch Feature2
   commit
   commit
   commit id: "Head" tag: "v1.0.3"
   checkout main
   merge Feature1
   commit
   commit
   commit tag: "v1.2.0" type: HIGHLIGHT
```

This test is mainly for testing commit tip updating.

## Result of updateGitRepo

```mermaid
---
title: Updated Git Repo
---
gitGraph
   commit tag: "v1.0.0"
   commit tag: "v1.0.1"
   branch Feature1
   checkout Feature1
   commit
   commit
   commit id: "head" tag: "v1.0.2"
   checkout main
   branch Feature2
   checkout Feature2
   commit
   commit
   commit tag: "v1.0.3"
   checkout main
   merge Feature1
   checkout main
   commit
   commit
   commit
   commit
   commit id: "Main Head" tag: "v1.2.0"

   checkout Feature2
   commit id: "Head"
```

## Test Cases - that require an update

### Update Branch Head Ref

```mermaid
gitGraph
   commit tag: "v1.0.0"
   commit tag: "v1.0.1"
   branch Feature1
   checkout Feature1
   commit
   commit
   commit id: "head" tag: "v1.0.2"
   checkout main
   branch Feature2
   checkout Feature2
   commit
   commit
   commit tag: "v1.0.3"
   checkout main
   merge Feature1
   checkout main
   commit
   commit
   commit tag: "v1.2.0"
   commit
   commit id: "Main Head"

   checkout Feature2
   commit id: "Head" type: HIGHLIGHT
```

### Update Tag Ref

```mermaid
gitGraph
   commit tag: "v1.0.0"
   commit tag: "v1.0.1"
   branch Feature1
   checkout Feature1
   commit
   commit
   commit id: "head" tag: "v1.0.2"
   checkout main
   branch Feature2
   checkout Feature2
   commit
   commit
   commit tag: "v1.0.3"
   checkout main
   merge Feature1
   checkout main
   commit
   commit
   commit
   commit
   commit id: "Main Head" tag: "v1.2.0" type: HIGHLIGHT

   checkout Feature2
   commit id: "Head"
```

## createLFSRepo

```mermaid
gitGraph
   commit id: "lfsFile.txt"
   branch Feature1
   checkout Feature1
   commit id: "lfsFile1.txt"
   checkout main
   branch Feature2
   commit id: "lfsFile2.txt"
```
