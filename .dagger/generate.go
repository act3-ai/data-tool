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

// Generate CLI documentation.
// +generate
func (m *DataTool) CLIDocs(ctx context.Context) *dagger.Changeset {
	acedt := m.Build(ctx, "linux/amd64", true)

	return dag.Container().
		WithDirectory("/src", m.Source).
		WithWorkdir("/src").
		WithFile("/usr/local/bin/ace-dt", acedt).
		WithExec([]string{"ace-dt", "gendocs", "md", "--only-commands", cliDocsPath}).
		Directory("/src").
		Changes(m.Source)
}

// Generate API documentation.
// +generate
func (m *DataTool) APIDocs() *dagger.Changeset {
	return dag.Go(dagger.GoOpts{
		Source: m.Source,
	}).
		Env().
		WithExec([]string{"go", "tool", "crd-ref-docs", "--config=apidocs.yaml", "--renderer=markdown",
			fmt.Sprintf("--source-path=%s/", pkgPath),
			fmt.Sprintf("--output-path=%s/", apiDocsPath),
		}).
		WithoutFile(filepath.Join(apiDocsPath, "out.md")). // TODO: Necessary?
		Directory("/app").
		Changes(m.Source)
}

// Generate pkg/apis with controller-gen.
// +generate
func (m *DataTool) Generate() *dagger.Changeset {
	return dag.Go(dagger.GoOpts{
		Source: m.Source,
	}).
		Env().
		WithExec([]string{"go", "generate", "./..."}).
		Directory("/app").
		Changes(m.Source)
}
