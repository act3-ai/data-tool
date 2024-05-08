package bottle

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/opencontainers/go-digest"

	telemv1alpha1 "gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/client"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/types"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
)

// TelemetryAdapter is a lightweight wrapper around the telemetry client that speaks Bottles.
type TelemetryAdapter struct {
	client client.Client
	urls   []string

	userName string
}

// NewTelemetryAdapter creates the MultiClient for telemetry from the global configuration.
func NewTelemetryAdapter(hosts []telemv1alpha1.Location, userName string) *TelemetryAdapter {
	mc := client.NewMultiClientConfig(hosts)
	urls := make([]string, len(hosts))
	for i, loc := range hosts {
		urls[i] = string(loc.URL)
	}
	return &TelemetryAdapter{mc, urls, userName}
}

// SendTelemetry sends bottle data to a telemetry host, if a "telemetry" config value is defined.
// Returns the URLS that can be used to view the bottle, and any error.
func (ta *TelemetryAdapter) SendTelemetry(ctx context.Context, btl *Bottle, r ref.Ref, action types.EventAction) ([]string, error) {
	if ta == nil {
		return nil, nil
	}

	evt := newEventFromBottle(btl, r, action, ta.userName)

	eventJSON, err := json.Marshal(evt)
	if err != nil {
		return nil, fmt.Errorf("error marshalling event: %w", err)
	}

	getArtifact := func(artifactDigest digest.Digest) ([]byte, error) {
		for _, art := range btl.Definition.PublicArtifacts {
			if art.Digest == artifactDigest {
				artifactPath := filepath.Join(btl.localPath, art.Path)
				var b []byte
				if b, err = os.ReadFile(artifactPath); err != nil {
					return b, fmt.Errorf("error reading artifact: %w", err)
				}
				return b, nil
			}
		}
		return nil, fmt.Errorf("artifact not found by digest %s", artifactDigest.String())
	}

	// It is possible for manifest and config data to change on a pull if there is a schema upgrade.  This can cause a
	// mismatch for a telemetry event, so we send the old config/manifest data so tracking is correct.  This is only
	// needed on a pull event
	bottleConfigJSON, err := btl.GetConfiguration()
	if err != nil {
		return nil, err
	}
	if btl.OriginalConfig != nil && action == types.EventPull {
		bottleConfigJSON = btl.OriginalConfig
	}

	bottleManifestJSON, err := btl.Manifest.GetManifestRaw()
	if err != nil {
		return nil, fmt.Errorf("getting bottle manifest JSON: %w", err)
	}
	if btl.OriginalManifest != nil && action == types.EventPull {
		bottleManifestJSON = btl.OriginalManifest
	}

	// We want to limit the time allowed for interacting with telemetry since it is often not critical
	// TODO make the timeout configurable (really this should probably be handled in the caller's ctx that is passed)
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Right now error handling goes like so.  We produce an error if ANY of the telemetry servers is unreachable.
	// Therefore we can assume that if we get to this point then all the telemetry servers were notified.
	// This assumption might change in the future.
	alg := digest.Canonical
	urls := make([]string, len(ta.urls))
	for i, u := range ta.urls {
		uu, err := url.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("parsing telemetry URL from config: %w", err)
		}
		urls[i] = client.BottleDetailURL(*uu, alg.FromBytes(bottleConfigJSON))
	}

	return urls, ta.client.SendEvent(ctxTimeout, alg, eventJSON, bottleManifestJSON, bottleConfigJSON, getArtifact)
}

// SendSignatures sends any existing signatures to telemetry for validation and storage.
func (ta *TelemetryAdapter) SendSignatures(ctx context.Context, summary *types.SignaturesSummary) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	sigSummaryJSON, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("error marshalling signature summary: %w", err)
	}

	return ta.client.SendSignature(ctxTimeout, digest.Canonical, sigSummaryJSON, nil)
}

// FindBottle takes a bottle reference (of type bottle:id) and a telemetry host and returns a
// list of bottle references that bottle is located at (if the telemetry host knows of said bottle).
func (ta *TelemetryAdapter) FindBottle(ctx context.Context, refSpec ref.Ref) ([]ref.Ref, error) {
	if ta == nil {
		return nil, nil
	}

	log := logger.FromContext(ctx)

	locations, err := ta.client.GetLocations(ctx, digest.Digest(refSpec.Digest))
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

// newEventFromBottle creates a telemetry Event based on the information in the provided bottle, and assigns the action
// type provided.
func newEventFromBottle(btl *Bottle, location ref.Ref, action types.EventAction, username string) *types.Event {
	dig := btl.Manifest.GetManifestDescriptor().Digest
	if btl.OriginalManifest != nil && action == types.EventPull {
		dig = digest.FromBytes(btl.OriginalManifest)
	}
	evt := &types.Event{
		ManifestDigest: dig,
		Action:         action,
		Repository:     location.RepoStringWithScheme(),
		Tag:            location.Tag,
		// AuthRequired: false, // FIXME
		// Bandwidth: 0, // FIXME
		Timestamp: time.Now(),
		Username:  username,
	}

	return evt
}
