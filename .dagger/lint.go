package main

import (
	"context"
	"dagger/tool/internal/dagger"
)

// Run linters.
func (m *DataTool) Lint() *Lint {
	return &Lint{
		DataTool: m,
	}
}

// Lint organizes linting functions.
type Lint struct {
	*DataTool
}

// Lint YAML files with "yamllint"
// +check
func (l *Lint) Yamllint(
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
func (l *Lint) MarkdownLint(
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
func (l *Lint) Audit(
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
func (l *Lint) GolangciLint(
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
func (l *Lint) Shellcheck(
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
func (l *Lint) Sonar(
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
