package main

import (
	"context"
	"dagger/tool/internal/dagger"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	registryPort  = 5000
	telemetryPort = 8100
)

// Run all tests.
func (m *DataTool) TestAll(ctx context.Context,
	// Top level source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
) (string, error) {
	unitResults, unitErr := m.TestUnit(ctx, src)

	funcResults, funcErr := m.TestFunctional(ctx, src)

	intResults, intErr := m.TestIntegration(ctx, src)

	out := "Unit Test Results:\n" + unitResults +
		"\n=====\n\nFunctional Test Results:\n" + funcResults +
		"\n=====\n\nIntegration TestResults:\n" + intResults
	return out, errors.Join(unitErr, funcErr, intErr)
}

// Run unit tests.
// +check
func (m *DataTool) TestUnit(ctx context.Context,
	// Top level source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
) (string, error) {
	return dag.Go(dagger.GoOpts{
		Source: src,
	}).
		Env().
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "-y", "git-lfs"}).
		WithExec([]string{"go", "test", "./..."}).
		Stdout(ctx)
}

// Run functional tests.
// +check
func (m *DataTool) TestFunctional(ctx context.Context,
	// Top level source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
) (string, error) {
	// start registry server
	regService := m.Registry()
	regService, err := regService.Start(ctx)
	if err != nil {
		return "", err
	}
	defer regService.Stop(ctx)

	regEndpoint, err := regService.Endpoint(ctx,
		dagger.ServiceEndpointOpts{
			Scheme: "http",
			Port:   registryPort,
		})
	if err != nil {
		return "", err
	}

	// start telemetry server
	telemService := m.TelemetryWithPostgres(ctx)
	telemService, err = telemService.Start(ctx)
	if err != nil {
		return "", err
	}
	defer telemService.Stop(ctx)

	telemEndpoint, err := telemService.Endpoint(ctx,
		dagger.ServiceEndpointOpts{
			Scheme: "http",
			Port:   telemetryPort,
		})
	if err != nil {
		return "", err
	}

	acedtConfigPath := "ace-dt-config.yaml"
	results, err := dag.Go(dagger.GoOpts{
		Source: src,
		Base:   dag.Container().From("golang:trixie"), // ensure access to apt for installing git-lfs
	}).
		Env().
		// install git-lfs dependency for ace-dt git tests
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "-y", "git-lfs"}).
		// bind to registry and telemetry servers
		WithServiceBinding("telemetry", telemService).
		WithServiceBinding("registry", regService).
		WithEnvVariable("TEST_REGISTRY", strings.TrimPrefix(regEndpoint, "http://")).
		WithEnvVariable("TEST_TELEMETRY", telemEndpoint).
		// allow plain-http for registry
		WithNewFile(acedtConfigPath, insecureRegConfig(regEndpoint)).
		WithEnvVariable("ACE_DT_CONFIG", acedtConfigPath).
		WithExec([]string{"go", "test", "-count=1", "./..."}).
		CombinedOutput(ctx)
	if err != nil {
		return results, err
	}

	// shutdown servers
	_, err = telemService.Stop(ctx)
	if err != nil {
		return "", err
	}
	_, err = regService.Stop(ctx)
	if err != nil {
		return "", err
	}
	return results, err
}

// Run integration tests.
// +check
func (m *DataTool) TestIntegration(ctx context.Context,
	// Top level source code directory
	// +defaultPath="/"
	// +ignore=[".git"]
	src *dagger.Directory,
) (string, error) {
	// start registry server
	regService := m.Registry()
	regService, err := regService.Start(ctx)
	if err != nil {
		return "", err
	}
	defer regService.Stop(ctx)

	regEndpoint, err := regService.Endpoint(ctx, dagger.ServiceEndpointOpts{Scheme: "http"})
	if err != nil {
		return "", err
	}
	regHost := strings.TrimPrefix(regEndpoint, "http://")

	// start telemetry server
	telemService := m.TelemetryWithPostgres(ctx)
	telemService, err = telemService.Start(ctx)
	if err != nil {
		return "", err
	}
	defer telemService.Stop(ctx)

	telemEndpoint, err := telemService.Endpoint(ctx, dagger.ServiceEndpointOpts{Scheme: "http"})
	if err != nil {
		return "", err
	}

	acedt := build(src, "linux/amd64", true)

	originalBottleRef := "ghcr.io/act3-ai/data-tool/bottles/mnist:v1.6"
	bottleID, err := dag.Wolfi().
		Container().
		WithFile("/usr/local/bin/ace-dt", acedt).
		// WithMountedSecret("/root/.docker/config.json", t.RegistryConfig.Secret()).
		WithExec([]string{"ace-dt", "bottle", "pull", originalBottleRef, "-d", "bottle"}).
		File("bottle/.dt/bottleid").
		Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("resolving bottleID: %w", err)
	}

	acedtConfigPath := "ace-dt-config.yaml"
	_, err = dag.Wolfi().
		Container().
		WithFile("/usr/local/bin/ace-dt", acedt).
		// WithMountedSecret("/root/.docker/config.json", t.RegistryConfig.Secret()).
		// bind to registry and telemetry servers
		WithServiceBinding("server", telemService).
		WithServiceBinding("registry", regService).
		WithEnvVariable("ACE_DT_TELEMETRY_URL", telemEndpoint).
		WithEnvVariable("ACE_DT_TELEMETRY_USERNAME", "ci-test-user").
		// allow plain-http for test registry
		WithNewFile(acedtConfigPath, insecureRegConfig(regEndpoint)).
		WithEnvVariable("ACE_DT_CONFIG", acedtConfigPath).
		WithExec([]string{"ace-dt", "bottle", "pull", originalBottleRef, "-d", "bottle"}).
		WithExec([]string{"ace-dt", "bottle", "push", fmt.Sprintf("%s/bottle/mnist:v1.6", regHost), "-d", "bottle"}).
		WithExec([]string{"ace-dt", "bottle", "pull", fmt.Sprintf("bottle:%s", bottleID), "-d", "bottle-pull2"}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("running integration test: %w", err)
	}

	// ensure telemetry reports the original location and the new one
	resp, err := http.DefaultClient.Get(fmt.Sprintf("%s/api/location?bottle_digest=%s", telemEndpoint, bottleID))
	if err != nil {
		return "", fmt.Errorf("fetching bottle locations from telemetry: %w", err)
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return "", fmt.Errorf("closing response body: %w", err)
	}
	result := string(res)
	if !strings.Contains(result, regHost) {
		return "", fmt.Errorf("telemetry did not report new bottle location, body: %s", result)
	}
	if !strings.Contains(result, strings.TrimSuffix(originalBottleRef, ":v1.6")) {
		return "", fmt.Errorf("telemetry did not report original bottle location, body: %s", result)
	}

	// shutdown servers
	_, err = telemService.Stop(ctx)
	if err != nil {
		return "", err
	}
	_, err = regService.Stop(ctx)
	if err != nil {
		return "", err
	}

	return "Successful", nil // TODO: better output?
}

// Start a registry.
func (m *DataTool) Registry() *dagger.Service {
	return dag.Container().
		From(imageRegistry).
		WithExposedPort(registryPort).
		AsService()
}

// ServerWithPostgres starts a telemetry service with postgres.
//
// This function makes it easier to expose a telemetry server to the host
// with less hassle connecting it to postgres.
func (m *DataTool) TelemetryWithPostgres(ctx context.Context) *dagger.Service {
	return m.Telemetry(m.Postgres())
}

// Start a telemetry server.
func (m *DataTool) Telemetry(
	// postgres service
	postgres *dagger.Service,
) *dagger.Service {
	telem := dag.Container().
		From(imageTelemetry).
		File("/usr/local/bin/telemetry")

	return dag.Wolfi().
		Container().
		WithFile("/usr/local/bin/telemetry", telem).
		WithEnvVariable("ACE_TELEMETRY_DSN", "postgres://testUser:testPassword@postgres/testdb").
		WithServiceBinding("postgres", postgres).
		WithExposedPort(telemetryPort).
		AsService(dagger.ContainerAsServiceOpts{
			Args: []string{"telemetry", "serve", "--listen", fmt.Sprintf(":%d", telemetryPort)},
		})
}

// Start postgres as a service.
func (tm *DataTool) Postgres() *dagger.Service {
	return dag.Container().
		From(imagePostgres).
		// Notice: changes to these env vars must be reflected in uses of
		// ACE_TELEMETRY_DSN
		WithEnvVariable("POSTGRES_DB", "testdb").
		WithEnvVariable("POSTGRES_USER", "testUser").
		WithEnvVariable("POSTGRES_PASSWORD", "testPassword").
		WithEnvVariable("POSTGRES_HOST_AUTH_METHOD", "trust").
		WithExposedPort(5432).
		AsService(dagger.ContainerAsServiceOpts{UseEntrypoint: true})
}

// insecureRegConfig creates ace-dt config file contents that enables insecure registry communication.
func insecureRegConfig(testRegEndpoint string) string {
	contents := strings.Builder{}
	contents.WriteString("apiVersion: config.dt.act3-ace.io/v1alpha1")
	contents.WriteString("\nkind: Configuration")
	contents.WriteString("\n\nregistryConfig:")
	contents.WriteString("\n registries:")
	contents.WriteString(fmt.Sprintf("\n  %s:", strings.TrimPrefix(testRegEndpoint, "http://")))
	contents.WriteString("\n   endpoints:")
	contents.WriteString(fmt.Sprintf("\n    - %s\n\n", testRegEndpoint))

	return contents.String()
}
