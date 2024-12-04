// Package conf provides configuration management for ace-dt and ace-dt library consumers.  The package allows loading
// and access of configuration details, including registry configuration.  Configuration details are validated against
// the current configuration scheme.
package conf

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/credentials"

	telemv1alpha2 "git.act3-ace.com/ace/data/telemetry/v2/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
	"git.act3-ace.com/ace/go-common/pkg/config"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/orasutil"
	dtreg "gitlab.com/act3-ai/asce/data/tool/internal/registry"
	regcache "gitlab.com/act3-ai/asce/data/tool/internal/registry/cache"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
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
	credStore     credentials.Store // in-memory credential store

	// blob caching
	blobCacher *orasutil.BlobCacher
}

// New returns a validated empty configuration object.  Configuration files should be defined using
// AddConfigFiles.  To initialize an empty default configuration, use New() followed by GetSafe().
func New(credOpts ...Option) *Configuration {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	cfg := &Configuration{
		scheme:    scheme,
		userAgent: "ace-dt", // default
	}

	for _, o := range credOpts {
		o(cfg)
	}

	// avoid unnecessary sync.Map alloc
	if cfg.credStore == nil {
		cfg.credStore = credentials.NewMemoryStore()
	}

	return cfg
}

// loadConfig loads configuration details from defined configuration files and override functions.
func (cfg *Configuration) loadConfig(ctx context.Context) error {
	log := logger.V(logger.FromContext(ctx), 1)

	// create the registry cache
	cfg.registryCache = regcache.NewRegistryCache()

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

	if cfg.config.CachePath != "" {
		cfg.blobCacher, err = orasutil.NewBlobCacher(cfg.config.CachePath)
		if err != nil {
			log.ErrorContext(ctx, "failed to initialize blob cache", "error", err)
		}
	}

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
		log.InfoContext(ctx, "Successfully loaded configuration")
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

// Repository sets up a repository target based on a reference string, making use of registry configuration
// settings. Since we return the Repository structure, it's troublesome to wrap the interface. In cases
// where alternative endpoints are used, it may be necessary to use an EndpointResolver.
func (cfg *Configuration) Repository(ctx context.Context, ref string) (*remote.Repository, error) {
	log := logger.V(logger.FromContext(ctx), 1)
	// if the config is not loaded, we should load it
	if cfg.config == nil {
		if err := cfg.loadConfig(ctx); err != nil {
			return nil, err
		}
	}
	rcfg := cfg.config.RegistryConfig

	log.InfoContext(ctx, "Creating repository", "registryConfig", ref)
	regTarget, err := dtreg.CreateRepoWithCustomConfig(ctx, &rcfg, ref, cfg.registryCache, cfg.userAgent, cfg.credStore)
	if err != nil {
		return nil, fmt.Errorf("creating repository %q: %w", ref, err)
	}
	repo, ok := regTarget.(*remote.Repository)
	if !ok {
		return nil, fmt.Errorf("error creating registry repository: %s", ref)
	}
	// Log warnings for now
	repo.HandleWarning = func(warning remote.Warning) {
		log.InfoContext(ctx, "Warning from remote registry", "text", warning.Text, "agent", warning.Agent, "code", warning.Code)
		// TODO this should be written to STDERR
	}
	return repo, nil
}

// GraphTarget sets up a repository target based on a reference string, making use of registry configuration
// settings. Implements GraphTargeter.
func (cfg *Configuration) GraphTarget(ctx context.Context, ref string) (oras.GraphTarget, error) {
	repo, err := cfg.Repository(ctx, ref)
	if err != nil {
		return nil, err
	}
	gt := oras.GraphTarget(repo)

	endpointURL, err := dtreg.ResolveEndpoint(&cfg.config.RegistryConfig, repo.Reference)
	if err != nil {
		return nil, fmt.Errorf("resolving alternate registry endpoint '%s': %w", repo.Reference.String(), err)
	}

	if endpointURL != nil {
		gt = dtreg.NewEndpointResolver(gt, endpointURL.Host)
	}

	if cfg.blobCacher != nil {
		return cfg.blobCacher.GraphTarget(gt), nil
	}
	return gt, nil
}

// ReadOnlyGraphTarget sets up a read-only repository target based on a reference string making
// use of registry configuration settings. Implements ReadOnlyGraphTargeter.
func (cfg *Configuration) ReadOnlyGraphTarget(ctx context.Context, ref string) (oras.ReadOnlyGraphTarget, error) {
	return cfg.GraphTarget(ctx, ref)
}

// ParseEndpointReference is the same as the oras registry.ParseReference, except that it resolves
// endpoints replacing the registry portion of the reference as defined by the configuration.
// Implements EndpointReferenceParser.
func (cfg *Configuration) ParseEndpointReference(reference string) (registry.Reference, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return registry.Reference{}, fmt.Errorf("parsing reference '%s': %w", reference, err)
	}

	endpointURL, err := dtreg.ResolveEndpoint(&cfg.config.RegistryConfig, ref)
	if err != nil {
		return registry.Reference{}, fmt.Errorf("resolving alternate registry endpoint '%s': %w", ref.String(), err)
	}
	ref.Registry = endpointURL.Host

	return ref, nil
}

// NewRegistry creates a ORAS registry using the registry configuration.
func (cfg *Configuration) NewRegistry(ctx context.Context, reg string) (*remote.Registry, error) {
	// FIXME  This is backwards.  The creation of the registry should come first.
	repo, err := cfg.Repository(ctx, reg+"/bogus:v1")
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

// CredStore returns a docker credentials store.
func (cfg *Configuration) CredStore() credentials.Store {
	return cfg.credStore
}

// WithRegistryConfig overwrites the loaded registry configuration, appending new
// registry configurations if they do not already exist.
func WithRegistryConfig(regCfg v1alpha1.RegistryConfig) ConfigOverrideFunction {
	return func(ctx context.Context, c *v1alpha1.Configuration) error {
		// sanity checks
		if c.ConfigurationSpec.RegistryConfig.Configs == nil {
			c.ConfigurationSpec.RegistryConfig.Configs = make(map[string]v1alpha1.Registry)
		}
		if c.ConfigurationSpec.RegistryConfig.EndpointConfig == nil {
			c.ConfigurationSpec.RegistryConfig.EndpointConfig = make(map[string]v1alpha1.EndpointConfig)
		}

		// overwrites an existing entry
		for k, v := range regCfg.Configs {
			c.ConfigurationSpec.RegistryConfig.Configs[k] = v
		}
		for k, v := range regCfg.EndpointConfig {
			c.ConfigurationSpec.RegistryConfig.EndpointConfig[k] = v
		}
		return nil
	}
}

// WithConcurrency overwrites the loaded concurrency configuration.
func WithConcurrency(conc int) ConfigOverrideFunction {
	return func(ctx context.Context, c *v1alpha1.Configuration) error {
		// only overwrite with a valid value
		if conc > 0 {
			c.ConcurrentHTTP = conc
		}
		return nil
	}
}

// WithCachePath overwrites the loaded blob cache path.
func WithCachePath(path string) ConfigOverrideFunction {
	return func(ctx context.Context, c *v1alpha1.Configuration) error {
		c.CachePath = path
		return nil
	}
}

// WithTelemetry overwrites the telemetry username while appending telemetry hosts to the
// loaded telemetry configuration.
func WithTelemetry(hosts []telemv1alpha2.Location, userName string) ConfigOverrideFunction {
	return func(ctx context.Context, c *v1alpha1.Configuration) error {
		// sanity check
		if c.Telemetry == nil {
			c.Telemetry = make([]telemv1alpha2.Location, 1)
		}
		c.Telemetry = append(c.Telemetry, hosts...)

		// don't erase and not replace
		if userName != "" {
			c.TelemetryUserName = userName
		}
		return nil
	}
}

// Option provides additional optional configuration.
type Option func(*Configuration)

// WithUserAgent overrides the default user-agent string used for http requests.
func WithUserAgent(userAgent string) Option {
	return func(c *Configuration) {
		c.userAgent = userAgent
	}
}

// WithCredentialStore sets the credential store.
func WithCredentialStore(store credentials.Store) Option {
	return func(c *Configuration) {
		c.credStore = store
	}
}
