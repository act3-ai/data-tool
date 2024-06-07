# Developer Guide

## Intended Audience

This documentation is written for ACT3 developers who are creating and maintaining the Data Tool software.

## Logging

Logging is done as JSONL so it is complete but harder for a human to easily parse. To convert them on the fly to colored and formatted text run it with `2> >(jq -j -f log.jq)` at the end. The [log.jq](../../cmd/ace-dt/internal/cli/log.jq) filter can be used with `jq` to pretty print the logs. So run `ace-dt` like so (in bash)

```shell
ace-dt bottle commit 2> >(jq -j -f log.jq)
```

### Style Guidelines: Logging Best Practices

The Data team uses `git.act3-ace.com/ace/go-common/pkg/logger`, which uses [slog](https://pkg.go.dev/log/slog) internally.

- Logs should be JSONL formatted but other formats can be supported. They are not meant to be consumed by a user but rather by a developer or operator.
- The logging system is not part of the UX.
- If returning an error there is no need to call `log.Error()` but you might want to add context with `fmt.Errorf()`. If handling an error or ignoring an error then use `log.Error()`.
- All errors from external packages (i.e., not ACT3 developed) should use `fmt.Errorf()` to wrap the error so that the code point is easily located in **our** code base. The format string should not contain the verbs "error", "failed", "cannot", ... since these are implied. We want concise error messages.
- At the creation of errors, the error should be logged (if a logger is available) with `log.Info()`.
- Use `cmd.Printx()` to output text to the user if the `cobra.Command` is available otherwise pass a `io.Writer` then use it with `fmt.FprintX(out, ...)`. Do not use the log to convey information to the user.
- Do not use `fmt.Printx()` (see the alternatives above).
- We only use structured logging. This means that values are explicit and that the first argument to `log.Info()` must always be a string literal (not a formatted string). The key values pairs follow. The keys should be string literals as well. The values are often variables.
- We use `context.Context` to store the logger however we try to only pass the root logger (for this sub-system) in the context. We do not want many loggers stacked up in the context.
- Logging levels are not predefined, and are considered relative to log level 0. No log is output at log level 0, but an API user may provide a higher base logging level, which means the log levels cannot be considered fixed. As a guideline, ace-dt's logging levels are generally organized in the following way:
  - LevelError = 0
  - LevelWarn = 4
  - LevelInfo = 8
  - LevelDebug = 12
- Logging levels are intended to be relative to a caller, for instance a consumer of the ace-dt code as an API may provide a logger with a higher level than 0. For cases where calls within ace-dt should have relative log levels, it is permissable to increment the log level and pass the updated logger through the context. In order to avoid unnecessary resource usage, this technique is not used commonly within ace-dt (As it causes a growing linked list in the context chain that must be traversed, and creates multiple instances of the logger object in memory)
- Use `log.WithName()` sparingly. It should only be used for creating a new logger for a sub-system.
- Avoid the use of `logr.NewContext()` unless you really need to add a new logger to the context. Instead let the exiting logger be passed on through to the caller.

Example 1:

```go
import (
   "context"

   "git.act3-ace.com/ace/go-common/pkg/logger"
)

func DoWork(ctx context.Context, a string) error {
   log := logger.FromContext(ctx).With("a", a)
}
```

Example 2:

> This example shows logging with name; use sparingly

```go
import (
   "context"

   "git.act3-ace.com/ace/go-common/pkg/logger"
)

func DoWork(ctx context.Context) error {
   log := logger.FromContext(ctx).WithGroup("foxtrot")
}

```

Example 3:

> This example is from the CSI driver code base using the ace/data/tool code base

```go
// logger is the root logger

// csiLog is used in csi-driver
csiLog := logger.WithName("csi-driver")

// aceDTLog has a name is one level less verbose
aceDTLog = logger.WithName("ace-dt").V(1)

// pull func from pkg/pull in ace-dt which takes a context and some pull args
pull.Pull(logr.NewContext(ctx, aceDTLog), ...pullargs)

// use csiLog for CSI related work
```

## Testing

### Unit Tests

To run all the unit tests:

```sh
make test-go
```

To run coverage:

```sh
make cover
```

### Functional Tests

To test `ace-dt` by performing operations on a registry and ASCE Telemetry:

```sh
make test-functional
```

To run functional tests manually:

```sh
make start-services
```

Then, `set -a; . test.env; set +a`.

In the same shell, you will have the necessary environment variables exported to run the tests with:

```sh
go test ./...
```

To enable debugging or running from within VSCode use the `test.env` file by adding the following to `.vscode/settings.json`:

```json
{
    "go.testEnvFile": "${workspaceFolder}/.env.test"
}
```

To run the functional tests manually from the command line (e.g., `go test ./...`) it is helpful to install [direnv](https://github.com/direnv/direnv/blob/master/docs/installation.md) then create a file called `.envrc` with the following:

```txt
dotenv_if_exists .env.test
PATH_bin bin
```

To allow the `.envrc` file to be used, in the project's root run:

```sh
direnv allow
```

Then, run:

```sh
make start-services
```

This updates the `.env.test` and `direnv` and automatically sets the `TEST_REGISTRY` and `TEST_TELEMETRY` in your current shell.

It also puts the projects `bin` directory first on the `PATH` so if you try to use `ace-dt` it will be using `bin/ace-dt`.

### Integration Testing

There are two options for integration testing.

Use make:

```sh
make test-go
```

Use the ACT3 pipeline, which does integration testing for Data Tools interaction with ASCE Telemetry and registry (with auth) in the `telemetry test` job.

Integration testing for `ace-dev` to ACE Telemetry is done in the `ace-dev test` job.

### Use Go Report Card

Go Report Card provides Go language code quality reports. It can be run locally or as a [web application](https://goreportcard.com/).

Developers using this option for code quality reports should use the local option documented in the project's [GitHub repository](https://github.com/gojp/goreportcard#installation).

### Use Local GitLab Runner

Using a local GitLab runner locally is useful for pre-merge testing.

#### Install the GitLab Runner

Follow the [GitLab documentation](https://docs.gitlab.com/runner/install/) corresponding to your operating system to install a local runner.

> The GitLab runner requires an [executor](https://docs.gitlab.com/runner/executors/); Data Tool uses Docker

<!-- The following content appears to be deprecated because the referenced file is the same as the primary file in the default location:
#### Prepare the Repo

The standard `.gitlab-ci.yaml` needs to be updated prior using a local runner.

Replace the `.gitlab-ci.yaml` content within your local repo, with the content of the file listed under the merged tab of this
[page](https://gitlab.com/act3-ai/asce/data/tool/-/ci/editor). -->

#### Run Tests Locally

For pre-merge testing, the following commands/steps are most useful:

- For functional test, use: `gitlab-runner exec docker 'functional test'`
- For unit test, use: `gitlab-runner exec docker 'unit test'`
- For golang-ci lint, use: `gitlab-runner exec docker 'golangci-lint'`

**Tips**:

- The runner may fail if your repository has a shallow clone
  - Avoid errors by running `git fetch --unshallow`
- The runner fails when objects tracked by `git-lfs` are not available locally
  - Avoid errors by installing `git-lfs` and running `git lfs pull`

## Dependency Management

Dependencies can be enforced with [`depguard`](https://github.com/OpenPeeDeeP/depguard).

Example usage:

To list the imports of a package (first-level dependencies) run:

```sh
go list -f '{{ join .Imports "\n" }}' ./pkg/transfer/
```

To list the transitive dependencies of a package (all dependencies) run:

```sh
go list -f '{{ join .Deps "\n" }}' ./pkg/transfer/
```

## Releasing

To initiate a release and create a release tag, run the CI pipeline, setting a variable `DO_RELEASE` to `true`. Ensure that at least one commit message has a `fix` or `feat` prefix, in order to trigger the semantic release process. For major version updates, at least one commit message must have `BREAKING CHANGE` in the commit message.

## See Also

- [OCI library for Rust](https://docs.rs/crate/oci-registry-client/0.1.3)
- [OCI image spec](https://github.com/opencontainers/image-spec/blob/main/spec.md)
- [OCI distribution spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md)
- [Bottle anatomy concept guide](../usage/concepts/bottle-anatomy.md)
