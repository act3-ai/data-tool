package main

import (
	"context"
	"dagger/tool/internal/dagger"
	"fmt"
	"path/filepath"
)

const (
	pkgPath     = "pkg/apis/config.dt.act3-ace.io"
	apiDocsPath = "docs/apis/config.dt.act3-ace.io"
	cliDocsPath = "docs/cli"
)

// Run all auto generators: CLI docs, API docs, and go generate
func (t *Tool) GenAll(ctx context.Context) *dagger.Directory {
	return dag.Directory().
		WithDirectory(cliDocsPath, t.CLIDocs(ctx)).
		WithDirectory(apiDocsPath, t.APIDocs()).
		WithDirectory(pkgPath, t.Generate())
}

// Generate CLI documentation.
func (t *Tool) CLIDocs(ctx context.Context) *dagger.Directory {
	acedt := t.Build(ctx, "linux/amd64", false)

	return dag.Go().
		WithSource(t.Source).
		Container().
		WithFile("/usr/local/bin/ace-dt", acedt).
		WithExec([]string{"ace-dt", "gendocs", "md", "--only-commands", cliDocsPath}).
		Directory(cliDocsPath)
}

// Generate API documentation.
func (t *Tool) APIDocs() *dagger.Directory {
	return dag.Go().
		WithSource(t.Source).
		Exec([]string{"go", "install", goCrdRefDocs}).
		WithExec([]string{"crd-ref-docs", "--config=apidocs.yaml", "--renderer=markdown",
			fmt.Sprintf("--source-path=%s/", pkgPath),
			fmt.Sprintf("--output-path=%s/", apiDocsPath),
		}).
		WithoutFile(filepath.Join(apiDocsPath, "out.md")). // TODO: Necessary?
		Directory(apiDocsPath)
}

// Generate pkg/apis with controller-gen.
func (t *Tool) Generate() *dagger.Directory {
	return dag.Go().
		WithSource(t.Source).
		WithEnvVariable("GOBIN", "/work/src/tool").
		Exec([]string{"go", "install", goControllerGen}).
		WithExec([]string{"go", "generate", "./..."}).
		Directory(pkgPath)
}
