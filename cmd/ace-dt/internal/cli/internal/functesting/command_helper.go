// Package functesting provides helper utilities for facilitating functional and integration tests.
package functesting

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/content/oci"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"git.act3-ace.com/ace/go-common/pkg/test"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// NewCommandHelper returns a new CommandHelper using the given cobra command.
func NewCommandHelper(tb testing.TB, cmd *cobra.Command) *CommandHelper {
	tb.Helper()
	cmd.SilenceUsage = true
	c := &CommandHelper{
		t:       tb,
		rootCmd: cmd,
		r:       rand.NewSource(0), // we do not need to seed this (we want a deterministic stream).
	}
	c.setupTestConfig()
	return c
}

// CommandHelper contains methods to more easily use commands in tests.
type CommandHelper struct {
	t          testing.TB
	rootCmd    *cobra.Command
	bottleDir  string
	configFile string
	config     v1alpha1.Configuration
	r          rand.Source
}

func (c *CommandHelper) setupTestConfig() {
	c.t.Helper()

	c.config.CachePath = c.t.TempDir()
	c.config.TelemetryUserName = "test-user"

	testReg := os.Getenv("TEST_REGISTRY")
	c.config.RegistryConfig.Configs = make(map[string]v1alpha1.Registry)
	c.config.RegistryConfig.Configs[testReg] = v1alpha1.Registry{Endpoints: []string{"http://" + testReg}}

	c.configFile = filepath.Join(c.t.TempDir(), "ace-dt-config.yaml")
	// We write a config file (with the cachePath set) and use --config=... to read in the config for actual tested command.
	data, err := json.Marshal(c.config)
	require.NoError(c.t, err)
	require.NoError(c.t, os.WriteFile(c.configFile, data, 0o600))
}

// Context provides access to the command's context.
func (c *CommandHelper) Context() context.Context {
	return c.rootCmd.Context()
}

// PruneCache clears out the file cache.
func (c *CommandHelper) PruneCache() {
	c.t.Helper()
	require.NoError(c.t, os.RemoveAll(c.config.CachePath))
}

// PopulateCache dumps files into the file cache.
func (c *CommandHelper) PopulateCache(num int) {
	c.t.Helper()
	cache, err := oci.NewStorage(c.config.CachePath)
	require.NoError(c.t, err)

	rng := rand.New(c.r)

	ctx := context.TODO() // Push throws away the context
	data := make([]byte, 1000)
	for i := 0; i < num; i++ {
		_, err := rng.Read(data)
		require.NoError(c.t, err)

		desc := ocispec.Descriptor{
			Digest:    digest.FromBytes(data),
			MediaType: "application/octet-stream", // irrelevant
			Size:      int64(len(data)),
		}

		err = cache.Push(ctx, desc, bytes.NewReader(data))
		require.NoError(c.t, err)
	}
}

// SetBottleDir will assign the given bottle directory to the command helper
// This will automatically add the "--bottle-dir bottleDir" flag to the commands run.
func (c *CommandHelper) SetBottleDir(bottleDir string) {
	c.t.Logf("Using bottle at %s", bottleDir)
	c.bottleDir = bottleDir
}

// SetConfig sets the CommandHelper's test config location, which will be used during code execution.
func (c *CommandHelper) SetConfig(configPath string) {
	c.configFile = configPath
}

// GetConfigFile returns the config file location for the given CommandHelper.
func (c *CommandHelper) GetConfigFile() string {
	return c.configFile
}

// GetConfig returns the config for the given CommandHelper.
func (c *CommandHelper) GetConfig() v1alpha1.Configuration {
	return c.config
}

// RunCommand is used to run a list of args on the FuncTest's root command.
// Simulates command line behavior.
func (c *CommandHelper) RunCommand(command ...string) {
	c.t.Helper()
	err := c.RunCommandWithError(command...)
	require.NoError(c.t, err, "Command failed to execute")
}

// RunCommandWithError is used to run a list of args on the FuncTest's root command.
// Simulates command line behavior.
func (c *CommandHelper) RunCommandWithError(command ...string) error {
	if c.bottleDir != "" {
		command = append(command, "--bottle-dir", c.bottleDir)
	}
	if c.configFile != "" {
		command = append(command, "--config", c.configFile)
	}
	// toggle on verbosity
	// command.Args = append(command.Args, "-v=2")

	c.rootCmd.SetArgs(command)

	log := test.Logger(c.t, 0)
	ctx := logger.NewContext(context.Background(), slog.New(log.Handler()))

	// Capture the commands output to a temporary buffer
	writer := new(bytes.Buffer)
	defer func() {
		b := writer.Bytes()
		if len(b) > 0 {
			c.t.Logf("\n=====command output=====\n%s\n=====end command output=====", b)
		}
	}()
	c.rootCmd.SetOut(writer)

	c.t.Logf("Running command %v", command)
	return c.rootCmd.ExecuteContext(ctx)
}
