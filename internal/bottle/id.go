package bottle

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"github.com/opencontainers/go-digest"
)

// SaveExtraBottleInfo saves the bottle ID and config JSON and the pull command used
// This data is never read by ace-dt and is only saved as a convenience to the user.
func SaveExtraBottleInfo(ctx context.Context, btl *Bottle) error {
	log := logger.FromContext(ctx)

	// this must always write to the .dt/bottleid
	log.InfoContext(ctx, "saving bottleID")
	if err := SaveBottleID(btl); err != nil {
		return err
	}

	log.InfoContext(ctx, "savingg bottle config")
	if err := SaveBottleConfig(btl); err != nil {
		return err
	}

	log.InfoContext(ctx, "saving bottle manifest")
	if err := SaveBottleManifest(btl); err != nil {
		return err
	}

	return SavePullCmd(ctx, btl)
}

// SaveBottleManifest writes a bottle's manifest as raw json to the bottle's manifest file (.dt/manifest.json).
func SaveBottleManifest(btl *Bottle) error {
	raw, err := btl.GetBottleManifest()
	if err != nil {
		return fmt.Errorf("saving bottle manifest: %w", err)
	}
	if err := os.WriteFile(manifestFile(btl.localPath), raw, 0o666); err != nil {
		return fmt.Errorf("unable to write bottle manifest: %w", err)
	}
	return nil
}

// SaveBottleID writes a bottle id as "<algorithm>:<digest>" to the bottle's bottleid file.
func SaveBottleID(btl *Bottle) error {
	return SaveBottleIDFile(btl, bottleidFile(btl.localPath))
}

// SaveBottleIDFile writes a bottle id as "<algorithm>:<digest>" to a given file.
func SaveBottleIDFile(btl *Bottle, path string) error {
	if err := os.WriteFile(path, []byte(btl.GetBottleID()), 0o666); err != nil {
		return fmt.Errorf("unable to save bottleID: %w", err)
	}

	return nil
}

// SavePullCmd writes a pull command using the bottle digest as source, and including part selectors if there are any
// virtual parts. This function is safe to call in any bottle command, and will log any errors instead of returning them.
func SavePullCmd(ctx context.Context, btl *Bottle) error {
	log := logger.FromContext(ctx)

	bottleID := btl.GetBottleID().String()

	partsel := ""
	if btl.VirtualPartTracker != nil && len(btl.VirtualPartTracker.VirtRecords) != 0 {
		for _, p := range btl.Parts {
			if !btl.VirtualPartTracker.HasContent(p.Digest) {
				partsel += " -p \"" + p.Name + "\""
			}
		}
	}

	content := fmt.Sprintf("ace-dt bottle pull bottle:%s %s\n", bottleID, partsel)

	path := pullCmdFile(btl.localPath)
	logger.V(log, 1).InfoContext(ctx, "Saving pull command", "command", content, "path", path)

	if err := os.WriteFile(path, []byte(content), 0o666); err != nil {
		return fmt.Errorf("unable to create file for pullcmd: %w", err)
	}

	return nil
}

// SaveBottleConfig saves a json representation of the bottle to the local bottle configuration directory.
func SaveBottleConfig(btl *Bottle) error {
	bottleData, err := btl.GetConfiguration()
	if err != nil {
		return fmt.Errorf("unable to get bottle configuration data for json output: %w", err)
	}

	if err := os.WriteFile(configFile(btl.localPath), bottleData, 0o666); err != nil {
		return fmt.Errorf("unable to save bottle config JSON: %w", err)
	}

	return nil
}

// ReadBottleIDFile reads and parses the bottle id file of the bottle.
func ReadBottleIDFile(bottlePath string) (digest.Digest, error) {
	data, err := os.ReadFile(bottleidFile(bottlePath))
	if err != nil {
		return "", fmt.Errorf("unable to read bottle ID file: %w", err)
	}

	dgst, err := digest.Parse(string(data))
	if err != nil {
		return "", fmt.Errorf("invalid bottle ID: %w", err)
	}

	return dgst, nil
}
