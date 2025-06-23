package main

import (
	"context"
	"dagger/tool/internal/dagger"
	"fmt"
	"strings"

	"github.com/sourcegraph/conc/pool"
)

// Run linters.
func (t *Tool) Lint() *Lint {
	return &Lint{
		Source: t.Source,
	}
}

// Lint organizes linting functions.
type Lint struct {
	Source *dagger.Directory
}

// Run all linters; Yaml, Markdown, Golang, and Shell.
func (l *Lint) All(ctx context.Context,
	// Source code directory
	// +defaultPath="/"
	src *dagger.Directory,
) (string, error) {
	p := pool.NewWithResults[string]().WithContext(ctx)

	p.Go(func(ctx context.Context) (string, error) {
		ctx, span := Tracer().Start(ctx, "yamllint")
		defer span.End()
		return l.Yamllint(ctx, l.Source)
	})

	p.Go(func(ctx context.Context) (string, error) {
		ctx, span := Tracer().Start(ctx, "markdownlint")
		defer span.End()
		return l.Markdownlint(ctx, l.Source)
	})

	p.Go(func(ctx context.Context) (string, error) {
		ctx, span := Tracer().Start(ctx, "golangci-lint")
		defer span.End()
		return dag.GolangciLint().
			Run(l.Source, dagger.GolangciLintRunOpts{
				Timeout: "5m",
			}).
			Stdout(ctx)
	})

	p.Go(func(ctx context.Context) (string, error) {
		ctx, span := Tracer().Start(ctx, "govulncheck")
		defer span.End()
		return dag.Govulncheck().
			ScanSource(ctx, l.Source)
	})

	p.Go(func(ctx context.Context) (string, error) {
		ctx, span := Tracer().Start(ctx, "shellcheck")
		defer span.End()
		return l.Shellcheck(ctx, src)
	})

	result, err := p.Wait()
	// TODO maybe we should order the lint result strings
	return strings.Join(result, "\n=====\n"), err
}

// Lint yaml files.
func (l *Lint) Yamllint(ctx context.Context,
	// Source code directory
	// +defaultPath="/"
	src *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("docker.io/cytopia/yamllint:1").
		WithWorkdir("/src").
		WithDirectory("/src", src).
		WithExec([]string{"yamllint", "."}).
		Stdout(ctx)
}

// Lint markdown files.
func (l *Lint) Markdownlint(ctx context.Context,
	// source code directory
	// +defaultPath="/"
	src *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("docker.io/davidanson/markdownlint-cli2:v0.14.0").
		WithWorkdir("/src").
		WithDirectory("/src", src).
		WithExec([]string{"markdownlint-cli2", "."}).
		Stdout(ctx)
}

// Lint shell files.
func (l *Lint) Shellcheck(ctx context.Context,
	// Source code directory
	// +defaultPath="."
	src *dagger.Directory,
) (string, error) {
	// TODO: Consider adding an option for specifying script files that don't have the extension, such as WithShellScripts.
	shEntries, err := src.Glob(ctx, "**/*.sh")
	if err != nil {
		return "", fmt.Errorf("globbing shell scripts with *.sh extension: %w", err)
	}

	bashEntries, err := src.Glob(ctx, "**/*.bash")
	if err != nil {
		return "", fmt.Errorf("globbing shell scripts with *.bash extension: %w", err)
	}

	p := pool.NewWithResults[string]().
		WithMaxGoroutines(4).
		WithErrors().
		WithContext(ctx)

	entries := append(shEntries, bashEntries...)
	for _, entry := range entries {
		p.Go(func(ctx context.Context) (string, error) {
			r, err := dag.Shellcheck().
				Check(src.File(entry)).
				Report(ctx)
			// this is needed because of weird error handling  in shellcheck here:
			// https://github.com/dagger/dagger/blob/0b46ea3c49b5d67509f67747742e5d8b24be9ef7/modules/shellcheck/main.go#L137
			if r != "" {
				return "", fmt.Errorf("results for file %s:\n%s", entry, r)
			}
			// r = fmt.Sprintf("Results for file %s:\n%s", entry, r)
			return r, err
		})
	}

	res, err := p.Wait()
	return strings.Join(res, "\n\n"), err
}

// Lint golang files.
func (l *Lint) Go(ctx context.Context) (string, error) {
	return dag.GolangciLint().
		Run(l.Source, dagger.GolangciLintRunOpts{
			Timeout: "5m",
		}).
		Stdout(ctx)
}
