// Package conf provides configuration management for ace-dt and ace-dt library consumers.  The package allows loading
// and access of configuration details, including registry configuration.  Configuration details are validated against
// the current configuration scheme.
package conf

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	dtreg "gitlab.com/act3-ai/asce/data/tool/internal/registry"
	regcache "gitlab.com/act3-ai/asce/data/tool/internal/registry/cache"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/go-common/pkg/config"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// ConfigOverrideFunction defines a function used to override config values after loading.
type ConfigOverrideFunction func(ctx context.Context, c *v1alpha1.Configuration) error

// Configuration stores and manages the ace-dt configuration file, based on the scheme and version defined in the apis.
// This data includes application settings, as well as telemetry and registry connection settings.
type Configuration struct {
	scheme *runtime.Scheme
	config *v1alpha1.Configuration

	// ConfigFiles stores the search locations for the config file in ascending priority order.  This field is exported
	// in order to allow indirect setting through command line processing in cobra.
	ConfigFiles []string

	// Handles overrides for configuration
	configOverrideFunctions []ConfigOverrideFunction

	// userAgent is a string used to identify the client during communication with registries
	userAgent string
	// Stores Registry Information
	registryCache *regcache.RegistryCache
}

// NewConfiguration returns a validated empty configuration object.  Configuration files should be defined using
// AddConfigFiles.  To initialize an empty default configuration, use NewConfiguration() followed by GetSafe().
func NewConfiguration(userAgent string) *Configuration {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	return &Configuration{scheme: scheme, userAgent: userAgent}
}

// loadConfig loads configuration details from defined configuration files and override functions.
func (cfg *Configuration) loadConfig(ctx context.Context) error {
	log := logger.V(logger.FromContext(ctx), 1)

	log.InfoContext(ctx, "Loading configuration", "configFiles", cfg.ConfigFiles)
	cfg.config = &v1alpha1.Configuration{}
	err := config.Load(logger.FromContext(ctx), cfg.GetConfigScheme(), cfg.config, cfg.ConfigFiles)
	if err != nil {
		return fmt.Errorf("failed to load config from config files %v: %w", cfg.ConfigFiles, err)
	}

	// Loop through override functions, applying each to the configuration
	for _, overrideFunction := range cfg.configOverrideFunctions {
		err = overrideFunction(ctx, cfg.config)
		if err != nil {
			return fmt.Errorf("config override function failed: %w", err)
		}
	}

	// create the registry cache
	cfg.registryCache = regcache.NewRegistryCache()

	log.InfoContext(ctx, "Using configuration", "config", cfg.config)
	return nil
}

// Get returns a configuration settings structure. The settings are loaded if they have not yet been loaded by a
// previous operation.
func (cfg *Configuration) Get(ctx context.Context) *v1alpha1.Configuration {
	log := logger.V(logger.FromContext(ctx), 1)
	if cfg.config == nil {
		if err := cfg.loadConfig(ctx); err != nil {
			log.InfoContext(ctx, "Unable to load existing configuration, using defaults")
			cfg.config = &v1alpha1.Configuration{}
			v1alpha1.ConfigurationDefault(cfg.config)
		}
	} else {
		log.InfoContext(ctx, "Using already loaded configuration")
	}
	return cfg.config
}

// GetConfigScheme returns the runtime scheme used for configuration file loading.
func (cfg *Configuration) GetConfigScheme() *runtime.Scheme {
	return cfg.scheme
}

// AddConfigFiles adds config files to the list of files parsed for configuration details.
func (cfg *Configuration) AddConfigFiles(files []string) {
	cfg.ConfigFiles = append(cfg.ConfigFiles, files...)
}

// AddConfigOverride adds an overrideFunction that will be passed to config.Load to edit config.
func (cfg *Configuration) AddConfigOverride(overrideFunction ...ConfigOverrideFunction) {
	if cfg.configOverrideFunctions == nil {
		cfg.configOverrideFunctions = []ConfigOverrideFunction{}
	}
	cfg.configOverrideFunctions = append(cfg.configOverrideFunctions, overrideFunction...)
}

// ConfigureRepository sets up a repository target based on a reference string, making use of registry configuration
// settings.
func (cfg *Configuration) ConfigureRepository(ctx context.Context, ref string) (*remote.Repository, error) {
	log := logger.V(logger.FromContext(ctx), 1)
	// if the config is not loaded, we should load it
	if cfg.config == nil {
		if err := cfg.loadConfig(ctx); err != nil {
			return nil, err
		}
	}
	rcfg := cfg.config.RegistryConfig
	log.InfoContext(ctx, "Creating repository", "registryConfig", ref)
	regTarget, err := dtreg.CreateRepoWithCustomConfig(ctx, &rcfg, ref, cfg.registryCache, cfg.userAgent)
	if err != nil {
		return nil, fmt.Errorf("creating repository %q: %w", ref, err)
	}
	repo, ok := regTarget.(*remote.Repository)
	if !ok {
		return nil, fmt.Errorf("error creating registry repository: %s", ref)
	}
	return repo, nil
}

// NewRegistry creates a ORAS registry using the registry configuration.
func (cfg *Configuration) NewRegistry(ctx context.Context, reg string) (*remote.Registry, error) {
	// FIXME  This is backwards.  The creation of the registry should come first.
	repo, err := cfg.ConfigureRepository(ctx, reg+"/bogus:v1")
	if err != nil {
		return nil, err
	}

	return &remote.Registry{
		// We need a the Repository.clone() method here.
		RepositoryOptions: remote.RepositoryOptions{
			Client:          repo.Client,
			Reference:       registry.Reference{Registry: repo.Reference.Registry},
			PlainHTTP:       repo.PlainHTTP,
			SkipReferrersGC: repo.SkipReferrersGC,
		},
	}, nil
}

// UserAgent returns the configured userAgent string.
func (cfg *Configuration) UserAgent() string {
	return cfg.userAgent
}
