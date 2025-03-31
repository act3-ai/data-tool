package main

import (
	"context"
	"dagger/tool/internal/dagger"
	"fmt"
	"path/filepath"
)

const (
	pkgPath  = "pkg/apis/config.dt.act3-ace.io"
	docsPath = "docs/apis/config.dt.act3-ace.io"
)

// Generate CLI documentation.
func (t *Tool) CLIDocs(ctx context.Context) *dagger.Directory {
	acedt := t.Build(ctx, "linux/amd64", "", "")

	cliDocsPath := "docs/cli"
	return dag.Go().
		WithSource(t.Source).
		Container().
		WithFile("/usr/local/bin/ace-dt", acedt).
		WithExec([]string{"ace-dt", "gendocs", "md", "--only-commands", cliDocsPath}).
		Directory(cliDocsPath)
}

// Generate API documentation.
func (t *Tool) APIDocs() (*dagger.Directory, error) {

	ctr := dag.Go().
		WithSource(t.Source).
		Exec([]string{"go", "install", goCrdRefDocs})

	ctr = ctr.WithExec([]string{"crd-ref-docs", "--config=apidocs.yaml", "--renderer=markdown",
		fmt.Sprintf("--source-path=%s/", pkgPath),
		fmt.Sprintf("--output-path=%s/", docsPath),
	})

	return ctr.WithoutFile(filepath.Join(docsPath, "out.md")).
		Directory(docsPath), nil
}

// Generate pkg/apis with controller-gen.
func (t *Tool) Generate() *dagger.Directory {
	ctr := dag.Go().
		WithSource(t.Source).
		WithEnvVariable("GOBIN", "/work/src/tool").
		Exec([]string{"go", "install", goControllerGen}).
		WithExec([]string{"go", "generate", "./..."})

	return ctr.Directory(pkgPath)
}
