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

// TODO: revise caching

// Generate CLI documentation.
// +generate
func (m *DataTool) CLIDocs(
	ctx context.Context,
	// Top level source code directory
	// +defaultPath="/"
	// +ignore=["**", "!go.*", "!pkg", "!internal", "!cmd", "!docs", "!apidocs.yaml", "!.git", "!.goreleaser.yaml"]
	src *dagger.Directory,
) *dagger.Changeset {
	acedt := m.Build(ctx, src, "linux/amd64", true)

	return dag.Container().
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithFile("/usr/local/bin/ace-dt", acedt).
		WithExec([]string{"ace-dt", "gendocs", "md", "--only-commands", cliDocsPath}).
		Directory("/src").
		Changes(src)
}

// Generate API documentation.
// +generate
func (m *DataTool) APIDocs(
	// Top level source code directory
	// +defaultPath="/"
	// +ignore=["**", "!go.*", "!pkg", "!internal", "!cmd", "!docs", "!apidocs.yaml"]
	src *dagger.Directory,
) *dagger.Changeset {
	return dag.Go(dagger.GoOpts{
		Source: src,
	}).
		Env().
		WithExec([]string{"go", "tool", "crd-ref-docs", "--config=apidocs.yaml", "--renderer=markdown",
			fmt.Sprintf("--source-path=%s/", pkgPath),
			fmt.Sprintf("--output-path=%s/", apiDocsPath),
		}).
		WithoutFile(filepath.Join(apiDocsPath, "out.md")). // TODO: Necessary?
		Directory("/app").
		Changes(src)
}

// Generate pkg/apis with controller-gen.
// +generate
func (m *DataTool) GoGenerate(
	// Top level source code directory
	// +defaultPath="/"
	// +ignore=["**", "!go.*", "!pkg", "!internal", "!cmd", "!docs"]
	src *dagger.Directory,
) *dagger.Changeset {
	return dag.Go(dagger.GoOpts{
		Source: src,
	}).
		Env().
		WithExec([]string{"go", "generate", "./..."}).
		Directory("/app").
		Changes(src)
}
