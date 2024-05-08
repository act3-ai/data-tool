# Data Tool Installation Guide

## Intended Audience

This documentation is written for Data Tool users who have not run the ACT3 Login script and want to use an alternative installation method.

## Alternate Installation Methods

Several alternate installation methods are available:

- Docker
- Install with Go
- Prebuilt binary
- Build from source

### Docker

Data Tool is included in the `reg.git.act3-ace.com/devsecops/dev-tools` image.

The [dev-tools project](https://git.act3-ace.com/ace/dev-tools) also packages Data Tool in an image by itself. This allows developers to pull the Data Tool image into another image.  

See the example [Dockerfile](https://gitlab.com/act3-ai/asce/data/tool/-/blob/main/sample/Dockerfile?ref_type=heads) for example usage.

### Install with Go

First, install and configure Go, using any of the following options:

- [Official Go installer](https://go.dev/doc/install)
- `brew install go`
- `snap install go --classic`

Then run:

```shell
go env -w GOPRIVATE=git.act3-ace.com
```

Installing Data Tool using the Go approach works on all platforms supported by Go. When Data Tool is installed with Go, the executable is built on the system and it is therefore not subject to security restrictions.

#### Create or Update `~/.netrc` File

Make sure that you have a NETRC file. Create one if you do not have one.

On UNIX/macOS the file is at `~/.netrc` and on Windows it is `~/_netrc`. You may also set the `NETRC` environment variable to a path containing your NETRC credentials.

The correct structure for the file is shown below:

```txt
machine git.act3-ace.com
    login your-username
    password your-gitlab-personal-access-token
```

> See [this post](https://seankhliao.com/blog/12021-04-29-go-private-modules-in-gitlab/) for more information and troubleshooting related to private GitLab repositories with Go.

Install `ace-dt` with

```sh
go install gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt@latest
```

You may replace `<latest>` with a tag you would like to install.  

`go install` adds the executable to your `$GOPATH/bin`, so make sure it is on your `$PATH`.

If you installed Data Tool with Go but are unable to run `ace-dt` via the CLI, add the following to your `~/.bashrc` file then run `source ~/.bashrc`

export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH=$PATH:$GOPATH/bin

Depending on the installation method used, you may need to manually install autocompletion. See `ace-dt completion --help`.

### Prebuilt Binary
  
A list of prebuilt binaries is available on Data Tool's [GitLab Release page](https://gitlab.com/act3-ai/asce/data/tool/-/releases). Options are available for 64-bit Linux, macOS, and Windows.

These are unsigned binaries.

**macOS users** need to complete additional steps to trust the unsigned binary by following **one** of the following options:

- In System Preferences -> Security and Privacy panel trust the program.
- Delete the "quarantine attribute" from the binary file with `xattr -d com.apple.quarantine ~/bin/act3-pt`

> Updating the quarantine attribute will place the output ace-dt executable in the ./bin directory when the build is complete

**Linux users** need to:

- Rename the downloaded file to `ace-dt`
- Change the permissions of the downloaded file to make it executable:

```sh
chmod +x <filename>
```

**Windows users** need to:

- Rename the downloaded file to `ace-dt.exe`
- No permissions need to be changed

### Build from Source

Clone the `tool` repository to your local working directory:

```sh
git clone git@git.act3-ace.com:ace/data/tool.git
```

Change into the the root of the cloned repository, then run:

```sh
make
```
