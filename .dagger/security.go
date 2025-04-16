package main

import (
	"context"
	"dagger/tool/internal/dagger"
)

// Run govulncheck.
func (t *Tool) VulnCheck(ctx context.Context) (string, error) {
	return dag.Go().
		WithSource(t.Source).
		WithCgoDisabled().
		Exec([]string{"go", "install", goVulnCheck}).
		WithExec([]string{"govulncheck", "./..."}).
		Stdout(ctx)
}

// Use ace-dt to perform a vulnerability scan on a list of OCI artifacts.
func (t *Tool) Scan(ctx context.Context,
	// Path to OCI artifact list
	sources *dagger.File,
) (string, error) {
	grype := dag.Container().
		From(imageGrype).
		File("/grype")

	grypeDB := t.GrypeDB(ctx)

	syft := dag.Container().
		From(imageSyft).
		File("/syft")

	const cachePath = "/cache/grype"

	sourcePath := "artifacts.txt"
	return dag.Container().
		WithMountedSecret("/root/.docker/config.json", t.RegistryConfig.Secret()).
		From("cgr.dev/chainguard/bash").
		WithFile("/usr/local/bin/ace-dt", build(ctx, t.Source, "linux/amd64", false)).
		WithFile("/usr/local/bin/grype", grype).
		WithFile("/usr/local/bin/syft", syft).
		WithFile(sourcePath, sources).
		WithDirectory(cachePath, grypeDB).
		WithEnvVariable("GRYPE_DB_CACHE_DIR", cachePath).
		WithUser("0").
		WithExec([]string{"grype", "db", "update"}).
		WithExec([]string{"ace-dt", "security", "scan", "-o=table",
			"--source-file", sourcePath, "--push-reports"}).
		Stdout(ctx)
}

// Download the Grype vulnerability database
func (t *Tool) GrypeDB(ctx context.Context) *dagger.Directory {
	const cachePath = "/cache/grype"

	return dag.Container().
		From(imageGrype).
		// WithUser(owner).
		// WithMountedCache(cachePath, dag.CacheVolume("grype-db-cache"), dagger.ContainerWithMountedCacheOpts{Owner: owner}).
		// comment out the line below to see the cached date output
		// WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithEnvVariable("GRYPE_DB_CACHE_DIR", cachePath).
		WithExec([]string{"/grype", "db", "update"}).
		Directory(cachePath)
}
