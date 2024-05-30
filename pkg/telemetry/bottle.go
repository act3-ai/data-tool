// Package telemetry provides utilities for telemetry bottle resolution and event updating.
package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/errdef"

	latest "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io/v1"
	"gitlab.com/act3-ai/asce/data/schema/pkg/mediatype"
	telemv1alpha1 "gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/client"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/types"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
	reg "gitlab.com/act3-ai/asce/data/tool/pkg/registry"
	tbottle "gitlab.com/act3-ai/asce/data/tool/pkg/transfer/bottle"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// ErrTelemetrySend indicates that a telemetry send event was not properly
// reported.
var ErrTelemetrySend = errors.New("telemetry reporting delegate failed")

// Adapter is a lightweight wrapper around the telemetry client that speaks Bottles.
// TODO: revise name?
// TODO: We should consider updating the adapter to handle "batch processing" of bottles, i.e.
// a workflow with multiple calls to ResolveWithTelemetry finishing with a single NotifyTelemetryWithAll
// method which calls NotifyTelemetry for each resolved bottle. Memory store should be updated to handle this.
type Adapter struct {
	client   client.Client
	urls     []string
	userName string

	// cache holds bottle manifests and configs
	cache *memory.Store
}

// NewAdapter creates the MultiClient for telemetry from the global configuration.
func NewAdapter(hosts []telemv1alpha1.Location, userName string) *Adapter {
	mc := client.NewMultiClientConfig(hosts)
	urls := make([]string, len(hosts))
	for i, loc := range hosts {
		urls[i] = string(loc.URL)
	}
	return &Adapter{
		client:   mc,
		urls:     urls,
		userName: userName,
		cache:    memory.New(),
	}
}

// ResolveWithTelemetry resolves a bottle reference to an oras.GraphTarget and an OCI descriptor.
// A bottle reference takes the scheme "bottle://<digest>" where the digest is an OCI config digest.
// Upon resolving a bottle reference, the bottle's version is validated for security purposes.
// It is safe to provide a regular OCI reference, although telemetry will not be used.
func (a *Adapter) ResolveWithTelemetry(ctx context.Context, reference string,
	sourceTargeter reg.ReadOnlyGraphTargeter, transferOpts tbottle.TransferOptions,
) (oras.ReadOnlyGraphTarget, ocispec.Descriptor, types.Event, error) {
	// validate reference
	r, err := ref.FromString(reference)
	if err != nil {
		return nil, ocispec.Descriptor{}, types.Event{}, fmt.Errorf("invalid input argument %s: %w", reference, err)
	}

	// TODO: look into telemetry pkg for it's own bottle ref parser
	refList := []ref.Ref{r}
	var hasSchemeBottle bool          // do we need to ensureBottleVersion later?
	if r.Scheme == ref.SchemeBottle { // do we have a "bottle:<digest>" scheme" ?
		hasSchemeBottle = true
		tempRefList, err := a.FindBottle(ctx, r)
		if err != nil {
			return nil, ocispec.Descriptor{}, types.Event{}, fmt.Errorf("unable to find bottle using telemetry server(s): %w", err)
		}
		refList = tempRefList
	}

	if len(refList) == 0 {
		return nil, ocispec.Descriptor{}, types.Event{}, errdef.ErrNotFound
	}

	// we'll accept the first discovered bottle, that is valid
	var errs []error
	for _, r := range refList { // refList contains only OCI references
		src, desc, err := a.resolveAndValidate(ctx, r, sourceTargeter, transferOpts, hasSchemeBottle)
		if err != nil {
			errs = append(errs, fmt.Errorf("resolving reference '%s': %w", r.String(), err))
			continue
		}
		// correct bottle discovered

		// build the telemetry event
		manBytes, err := content.FetchAll(ctx, a.cache, desc)
		if err != nil {
			errs = append(errs, fmt.Errorf("retrieving manifest from cache: %w", err))
			continue
		}
		event := a.NewEvent(r.String(), manBytes, types.EventPull)

		return src, desc, event, nil
	}

	return nil, ocispec.Descriptor{}, types.Event{}, errors.Join(errs...)
}

// NotifyTelemetry updates a telemetry host with the provided event. It prioritizes local bottle data, and fetches
// missing data from the remote as necessary.
func (a *Adapter) NotifyTelemetry(ctx context.Context, src content.ReadOnlyGraphStorage, desc ocispec.Descriptor, bottleDir string,
	event types.Event,
) ([]string, error) {
	log := logger.FromContext(ctx)

	// copy and cache bottle metadata
	err := a.copyCacheMetadata(ctx, src, desc)
	if err != nil {
		return nil, err
	}

	var manifest ocispec.Manifest
	var config latest.Bottle
	manBytes, err := content.FetchAll(ctx, a.cache, desc) // fetch manifest from cache
	if err != nil {
		return nil, fmt.Errorf("fetching bottle manifest: %w", err)
	}
	err = json.Unmarshal(manBytes, &manifest)
	if err != nil {
		return nil, fmt.Errorf("decoding manifest content: %w", err)
	}

	// fetch config
	cfgBytes, err := content.FetchAll(ctx, a.cache, manifest.Config) // fetch config from cache
	if err != nil {
		return nil, fmt.Errorf("fetching bottle config: %w", err)
	}
	err = json.Unmarshal(cfgBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("decoding manifest content: %w", err)
	}

	// assumes entire bottle is pulled or atleast public artifacts are available
	// TODO: fail or be nice and get it via virtual parts
	// TODO: consider failing if public artifiacts are not in telemetry
	artifactResolver := make(map[digest.Digest]latest.PublicArtifact, len(config.PublicArtifacts))
	for _, artifact := range config.PublicArtifacts {
		artifactResolver[artifact.Digest] = artifact
	}
	getArtifact := func(artifactDigest digest.Digest) ([]byte, error) {
		art, ok := artifactResolver[artifactDigest]
		if !ok {
			return nil, fmt.Errorf("artifact not found by digest %s", artifactDigest.String())
		}

		artifactPath := filepath.Join(bottleDir, art.Path)
		b, err := os.ReadFile(artifactPath)
		if err != nil {
			return b, fmt.Errorf("error reading artifact: %w", err)
		}
		return b, nil
	}

	// Increase telemetry log verbosity level
	telemURLs, err := a.sendTelemetry(logger.NewContext(ctx, logger.V(log, 1)), desc, getArtifact, event)
	if err != nil {
		log.ErrorContext(ctx, "Failed to send telemetry", "error", err.Error())
		return nil, ErrTelemetrySend
	}

	// fallback to the remote if not found locally
	var summary *types.SignaturesSummary
	switch {
	case bottleDir != "":
		summary, err = newSummaryFromLocal(ctx, bottleDir, desc)
		if err == nil {
			break
		}
		fallthrough
	default:
		summary, err = newSummaryFromRemote(ctx, src, desc, digest.FromBytes(cfgBytes))
		if err != nil {
			log.ErrorContext(ctx, "Failed to generate signature summary message", "error", err.Error())
			return nil, fmt.Errorf("generating signature detail message: %w", err)
		}
	}

	if summary != nil {
		err = a.sendSignatures(logger.NewContext(ctx, logger.V(log, 1)), summary)
		if err != nil {
			return nil, err
		}
	}
	return telemURLs, nil
}

// FindBottle takes a bottle reference (of type bottle:id) and a telemetry host and returns a
// list of bottle references that bottle is located at (if the telemetry host knows of said bottle).
func (a *Adapter) FindBottle(ctx context.Context, refSpec ref.Ref) ([]ref.Ref, error) {
	if a == nil {
		return nil, nil
	}
	log := logger.FromContext(ctx)

	locations, err := a.client.GetLocations(ctx, digest.Digest(refSpec.Digest))
	if err != nil {
		return nil, err
	}

	logger.V(log, 1).InfoContext(ctx, "parsing telemetry results for references")
	refList := make([]ref.Ref, 0, len(locations))
	for _, location := range locations {
		tmpRef, err := ref.FromString(location.Repository)
		if err != nil {
			logger.V(log, 2).InfoContext(ctx, "unable to parse ref from telemetry", "repository", location.Repository)
			continue
		}
		tmpRef.Digest = location.Digest.String()
		logger.V(log, 1).InfoContext(ctx, "found ref from telemetry", "ref", tmpRef.String())

		// copy over query and selector information to each discovered ref from telemetry
		tmpRef.Selectors = refSpec.Selectors
		tmpRef.Query = refSpec.Query

		refList = append(refList, tmpRef)
	}

	return refList, nil
}

// NewEvent creates a telemetry Event based on bottle metdata and an event action type.
func (a *Adapter) NewEvent(location string, rawManifest []byte, action types.EventAction) types.Event {
	loc := ref.RepoFromString(location)
	dig := digest.FromBytes(rawManifest)

	return types.Event{
		ManifestDigest: dig,
		Action:         action,
		Repository:     loc.RepoStringWithScheme(),
		Tag:            loc.Tag,
		// AuthRequired: false, // FIXME
		// Bandwidth: 0, // FIXME
		Timestamp: time.Now(),
		Username:  a.userName,
	}
}

// resolveAndValidate resolves an OCI reference to an oras.ReadOnlyGraphTarget and a descriptor. It also validates the
// bottle's version if the provided ref used to be the bottle scheme before it was resolved to an OCI reference.
func (a *Adapter) resolveAndValidate(ctx context.Context, r ref.Ref, sourceTargeter reg.ReadOnlyGraphTargeter,
	transferOpts tbottle.TransferOptions, validateVersion bool,
) (oras.ReadOnlyGraphTarget, ocispec.Descriptor, error) {
	src, desc, err := tbottle.Resolve(ctx, r.String(), sourceTargeter, transferOpts)
	if err != nil {
		return nil, ocispec.Descriptor{}, fmt.Errorf("resolving reference: %w", err)
	}

	// copy and cache bottle metadata
	err = a.copyCacheMetadata(ctx, src, desc)
	if err != nil {
		return nil, ocispec.Descriptor{}, err
	}

	var manifest ocispec.Manifest
	manBytes, err := content.FetchAll(ctx, a.cache, desc) // fetch manifest from cache
	if err != nil {
		return nil, ocispec.Descriptor{}, fmt.Errorf("fetching bottle manifest: %w", err)
	}
	err = json.Unmarshal(manBytes, &manifest)
	if err != nil {
		return nil, ocispec.Descriptor{}, fmt.Errorf("decoding manifest content: %w", err)
	}

	// validate bottle version
	if validateVersion {
		cfgBytes, err := content.FetchAll(ctx, a.cache, manifest.Config) // fetch config from cache
		if err != nil {
			return nil, ocispec.Descriptor{}, fmt.Errorf("fetching bottle config: %w", err)
		}
		err = ensureBottleVersion(cfgBytes)
		if err != nil {
			return nil, ocispec.Descriptor{}, fmt.Errorf("ensuring bottle version: %w", err)
		}
	}

	return src, desc, nil
}

// sendTelemetry sends bottle data to a telemetry host, if a "telemetry" config value is defined.
// Returns the URLS that can be used to view the bottle, and any error. Requires the manifest
// and config to already be cached.
func (a *Adapter) sendTelemetry(ctx context.Context, desc ocispec.Descriptor, getArtifact client.GetArtifactDataFunc, event types.Event) ([]string, error) {
	if a == nil {
		return nil, nil
	}

	rawManifest, err := content.FetchAll(ctx, a.cache, desc)
	if err != nil {
		return nil, fmt.Errorf("fetching bottle manifest from cache: %w", err)
	}

	var manifest ocispec.Manifest
	err = json.Unmarshal(rawManifest, &manifest)
	if err != nil {
		return nil, fmt.Errorf("decoding bottle manifest: %w", err)
	}

	rawConfig, err := content.FetchAll(ctx, a.cache, manifest.Config)
	if err != nil {
		return nil, fmt.Errorf("fetching bottle config from cache: %w", err)
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("error marshalling event: %w", err)
	}

	// We want to limit the time allowed for interacting with telemetry since it is often not critical
	// TODO make the timeout configurable (really this should probably be handled in the caller's ctx that is passed)
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Right now error handling goes like so.  We produce an error if ANY of the telemetry servers is unreachable.
	// Therefore we can assume that if we get to this point then all the telemetry servers were notified.
	// This assumption might change in the future.
	alg := digest.Canonical
	urls := make([]string, len(a.urls))
	for i, u := range a.urls {
		uu, err := url.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("parsing telemetry URL from config: %w", err)
		}
		urls[i] = client.BottleDetailURL(*uu, alg.FromBytes(rawConfig))
	}

	return urls, a.client.SendEvent(ctxTimeout, alg, eventJSON, rawManifest, rawConfig, getArtifact)
}

// ensureBottleVersion checks the bottle version.
// This is called when the bottle ID is used to pull the bottle (i.e., we don't trust the manifest).
func ensureBottleVersion(rawConfig []byte) error {
	// pulled by bottle ID
	typeMeta := metav1.TypeMeta{}
	if err := json.Unmarshal(rawConfig, &typeMeta); err != nil {
		// returning an insecure archive error here since we don't know for sure if the version is sufficient.  Note
		// this error is unlikely to occur since a decode error will likely have already occurred for this config
		return fmt.Errorf("json decode error, unable to verify bottle version. %w", bottle.ErrPartInsecureArchive)
	}

	// TODO: all cases besides recent (v1) fail, don't use latest
	if typeMeta.GroupVersionKind() != latest.GroupVersion.WithKind("Bottle") { // TODO: latest is not the right answer, certain versions have an issue
		return fmt.Errorf("cannot pull by bottle id, use manifest digest or tag directly.  There is a security concern with pulling by this version of the bottle by bottle ID: %w", bottle.ErrPartInsecureArchive)
	}

	return nil
}

// cacheMetadata fetches remote bottle manifests and configs into the cache, skipping layers.
func (a *Adapter) copyCacheMetadata(ctx context.Context, src content.ReadOnlyStorage, desc ocispec.Descriptor) error {
	copyOpts := oras.CopyGraphOptions{
		PreCopy: preCopyOnlyMetadata,
	}
	err := oras.CopyGraph(ctx, src, a.cache, desc, copyOpts)
	if err != nil {
		return fmt.Errorf("caching bottle metadata: %w", err)
	}
	return nil
}

// preCopyOnlyMetadata strictly copies bottle manifests and configs, skipping layers.
// Use carefully, not all oras storages or registries accept manifests without layers.
func preCopyOnlyMetadata(ctx context.Context, desc ocispec.Descriptor) error {
	switch {
	case desc.MediaType == ocispec.MediaTypeImageManifest:
		// noop, copy bottle manifest
		return nil
	case mediatype.IsBottleConfig(desc.MediaType):
		// noop, copy bottle config
		return nil
	case mediatype.IsLayer(desc.MediaType):
		return oras.SkipNode
	default:
		return fmt.Errorf("unexpected descriptor mediatype '%s'", desc.MediaType)
	}
}
