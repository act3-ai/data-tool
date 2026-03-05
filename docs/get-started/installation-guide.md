# Data Tool Installation Guide

## Intended Audience

This documentation is written for Data Tool users who have not run the ACT3 Login script and want to use an alternative installation method.

## Alternate Installation Methods

Several alternate installation methods are available:

- Homebrew Tap
- OCI Image
- Install with Go
- Prebuilt binary
- Build from source

### Homebrew Tap

A pre-built binary is available on the act3-ai homebrew tap.

- `brew tap act3-ai/tap`
- `brew install ace-dt`

### OCI Image (Docker-like Image)

Container runtime images for supported platforms are available in the [Github package registry](https://github.com/act3-ai/data-tool/pkgs/container/data-tool).

### Install with Go

First, install and configure Go, using any of the following options:

- [Official Go installer](https://go.dev/doc/install)
- `brew install go`
- `snap install go --classic`

Installing Data Tool using the Go approach works on all platforms supported by Go. When Data Tool is installed with Go, the executable is built on the system and it is therefore not subject to security restrictions.

#### Create or Update `~/.netrc` File

Make sure that you have a NETRC file. Create one if you do not have one.

On UNIX/macOS the file is at `~/.netrc` and on Windows it is `~/_netrc`. You may also set the `NETRC` environment variable to a path containing your NETRC credentials.

The correct structure for the file is shown below:

```txt
machine github.com
    login your-username
    password your-github-personal-access-token
```

Install `ace-dt` with

```sh
go install github.com/act3-ai/data-tool/cmd/ace-dt@latest
```

You may replace `<latest>` with a tag you would like to install.

`go install` adds the executable to your `$GOPATH/bin`, so make sure it is on your `$PATH`.

If you installed Data Tool with Go but are unable to run `ace-dt` via the CLI, add the following to your `~/.bashrc` file then run `source ~/.bashrc`

export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH=$PATH:$GOPATH/bin

Depending on the installation method used, you may need to manually install autocompletion. See `ace-dt completion --help`.

### Prebuilt Binary
  
A list of archived prebuilt binaries is available on Data Tool's [GitHub Release page](https://github.com/act3-ai/data-tool/releases). Options are available for 64-bit Linux, macOS, and Windows.

Archives include an executable, LICENSE, and release notes for the downloaded version.

Download the respective archive for your system, extract it, and move `ace-dt` where desired.

These are unsigned binaries.

**macOS users** need to complete additional steps to trust the unsigned binary by following **one** of the following options:

- In System Preferences -> Security and Privacy panel trust the program.
- Delete the "quarantine attribute" from the binary file with `xattr -d com.apple.quarantine ~/bin/act3-pt`

> Updating the quarantine attribute will place the output ace-dt executable in the ./bin directory when the build is complete

### Build from Source

Clone the `tool` repository to your local working directory:

```sh
git clone git@github.com:act3-ai/data-tool.git
```

Then change into the the root of the cloned repository.

#### Build with Dagger

Dagger is a tool we use to build reusable pipelines, utilized in both CI and local dev environments. You can utilize our pipeline build process to build from source. See [dagger docs](https://docs.dagger.io/).

`dagger call build --platform linux/amd64 export --path path/to/destination`

#### Build with Go

`go build -o path/to/destination ./cmd/ace-dt`
