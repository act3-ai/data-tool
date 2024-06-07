/*
Package cli contains all the subcommands for the CLI

Root level ace-dt CLI handler, based on Cobra command handler library.

Individual commands are implemented as subcommands similar to git, kubectl, docker, etc
Each of these subcommands are managed by a separate command .go file, and provide their
own options and argument validation as appropriate.  In most cases, the work is then
farmed out to an appropriate package (in pkg).  This enables the pkg code to be used as
a library by other tools if desired, and modified independently from the rest of the app
*/
package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"

	telemv1alpha1 "git.act3-ace.com/ace/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
	"git.act3-ace.com/ace/go-common/pkg/config"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"git.act3-ace.com/ace/go-common/pkg/redact"
	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/bottle"
	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/git"
	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/mirror"
	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/oci"
	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/pypi"
	"gitlab.com/act3-ai/asce/data/tool/cmd/ace-dt/internal/cli/security"
	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// NewToolCmd returns a command that represents the root command ace-dt.
func NewToolCmd(version string) *cobra.Command {
	action := actions.NewTool(version)

	// Add environment variable configuration overrides
	action.Config.AddConfigOverride(envOverrides)

	// cmd represents the base command when called without any subcommands. There is
	//  no default behavior, apart from outputting help.
	cmd := &cobra.Command{
		Use:   "ace-dt",
		Short: "data management tool for bottles and artifacts",
		Long:  `Provides data transfer facilities for obtaining data bottles based on a data registry, as well as capabilities for adding new data bottles to a data registry.`,
	}

	cmd.PersistentFlags().StringArrayVar(&action.Config.ConfigFiles, "config",
		config.EnvPathOr("ACE_DT_CONFIG", config.DefaultConfigSearchPath("ace", "dt", "config.yaml")),
		`configuration file location (setable with env "ACE_DT_CONFIG").
The first configuration file present is used.  Others are ignored.
`)

	cmd.AddGroup(
		&cobra.Group{
			ID:    "setup",
			Title: "Setup commands",
		},
		&cobra.Group{
			ID:    "core",
			Title: "Core commands",
		},
	)

	cmd.AddCommand(
		// AuthCreds commands
		newLoginCmd(action),
		newLogoutCmd(action),
		newConfigCmd(action),
		newCompletionCmd(),
		bottle.NewBottleCmd(action),
		newUtilCmd(action),
		mirror.NewMirrorCmd(action),
		pypi.NewPypiCmd(action),
		oci.NewOciCmd(action),
		git.NewGitCmd(action),
		security.NewSecurityCmd(action),
		newRunRecipe(),
	)

	return cmd
}

// envOverrides reads environment variables to override the configuration c.
func envOverrides(ctx context.Context, c *v1alpha1.Configuration) error {
	log := logger.V(logger.FromContext(ctx), 2)

	name := "ACE_DT_CACHE_PRUNE_MAX"
	if value, exists := os.LookupEnv(name); exists {
		q, err := resource.ParseQuantity(value)
		if err != nil {
			return fmt.Errorf("invalid environment variable \"%s\"=%s : %w", name, value, err)
		}
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.CachePruneMax = &q
	}

	name = "ACE_DT_CACHE_PATH"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.CachePath = value
	}

	name = "ACE_DT_CHUNK_SIZE"
	if value, exists := os.LookupEnv(name); exists {

		q, err := resource.ParseQuantity(value)
		if err != nil {
			return fmt.Errorf("invalid environment variable \"%s\"=%s : %w", name, value, err)
		}
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.ChunkSize = &q
	}

	name = "ACE_DT_CONCURRENT_HTTP"
	if value, exists := os.LookupEnv(name); exists {
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid environment variable \"%s\"=%s : %w", name, value, err)
		}
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.ConcurrentHTTP = int(v)
	}

	// Overrides for RegistryAuthFile
	name = "DOCKER_CONFIG"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.RegistryAuthFile = value
	}

	name = "REGISTRY_AUTH_FILE"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.RegistryAuthFile = value
	}

	// higher precedence than DOCKER_AUTH and REGISTRY_AUTH_FILE
	name = "ACE_DT_REGISTRY_AUTH_FILE"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.RegistryAuthFile = value
	}

	name = "ACE_DT_EDITOR"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.Editor = value
	}

	name = "ACE_DT_TELEMETRY_URL"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.Telemetry = []telemv1alpha1.Location{
			{URL: redact.SecretURL(value)},
		}
	}

	name = "ACE_DT_TELEMETRY_USERNAME"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.TelemetryUserName = value
	}

	// TODO make this a reasonable default based on operating system

	// Overrides for Editor
	name = "EDITOR"
	if value, exists := os.LookupEnv("EDITOR"); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.Editor = value
	}

	// This has higher precedence
	name = "VISUAL"
	if value, exists := os.LookupEnv("VISUAL"); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.Editor = value
	}

	return nil
}
