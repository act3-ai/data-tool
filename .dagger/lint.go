package main

import (
	"context"
	"dagger/tool/internal/dagger"
)

// Lint YAML files with "yamllint"
// +check
func (m *DataTool) Yamllint(
	ctx context.Context,
	// Source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
) *dagger.Container {
	return dag.Yamllint(src).WithStrict().Lint()
}

// Lint markdown files with "markdownlint-cli2"
// +check
func (m *DataTool) MarkdownLint(
	ctx context.Context,
	// Source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
) *dagger.Container {
	return dag.Markdownlint(src).Lint()
}

// Audit Go dependencies with "govulncheck"
// +check
func (m *DataTool) Audit(
	ctx context.Context,
	// top level source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
) (string, error) {
	return dag.Govulncheck().ScanSource(ctx, src)
}

// Lint Go code with "golangci-lint"
// +check
func (m *DataTool) GolangciLint(
	ctx context.Context,
	// top level source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
) *dagger.Container {
	return dag.
		GolangciLint().
		Run(src, dagger.GolangciLintRunOpts{
			Timeout: "5m",
		})
}

// Lint shell scripts with "shellcheck"
// +check
func (m *DataTool) Shellcheck(
	ctx context.Context,
	// top level source code directory
	// +defaultPath="/"
	// +ignore=["**", "!**/*.sh", "!.shellcheckrc"]
	src *dagger.Directory,
) *dagger.Container {
	srcFiltered := src.
		Filter(dagger.DirectoryFilterOpts{
			Include:   []string{"**/*.sh", ".shellcheckrc"},
			Gitignore: true,
		})

	return dag.
		Container().
		From("koalaman/shellcheck-alpine:latest").
		WithMountedDirectory("/src", srcFiltered).
		WithWorkdir("/src").
		WithExec([]string{"sh", "-c", `shellcheck bin/*.sh`})
}

// Run sonar scanner
func (m *DataTool) Sonar(
	ctx context.Context,
	// top level source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
) (string, error) {
	return dag.
		Sonarqube(dagger.SonarqubeOpts{Base: "sonarqube:26.5.0.122743-community"}).
		Scan(ctx, src)
}
