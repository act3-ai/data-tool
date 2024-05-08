# HACKS

## Oras Fork

We are currently using an ORAS fork to support mounting.  We hope it will be merged upstream and released soon..

To update the branch we use run:

```shell
go mod edit -replace oras.land/oras-go/v2=github.com/ktarplee/oras-go/v2@tool
go mod tidy
```
