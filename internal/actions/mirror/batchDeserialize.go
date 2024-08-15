package mirror

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.act3-ace.com/ace/data/tool/internal/mirror"
	"git.act3-ace.com/ace/data/tool/internal/ui"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// BatchDeserialize represents the mirror batch-deserialize action.
type BatchDeserialize struct {
	*Action
	SuccessfulSyncFile string
}

// Run runs the mirror batch-deserialize action.
func (action *BatchDeserialize) Run(ctx context.Context, syncDir, destination string) error {
	rootUI := ui.FromContextOrNoop(ctx)
	log := logger.FromContext(ctx)

	var successfulSyncs [][]string
	file, err := os.OpenFile(filepath.Join(syncDir, action.SuccessfulSyncFile), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("opening successful sync file: %w", err)
	}
	defer file.Close()
	r := csv.NewReader(file)
	existingSyncs, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("reading successful sync file: %w", err)
	}
	// all of the tar files live in syncDir/data directory.
	// get all of the files within syncDir/data, ignore the trackerfile.
	entries, err := os.ReadDir(syncDir)
	if err != nil {
		return fmt.Errorf("reading the directory: %w", err)
	}
	for _, entry := range entries {
		switch {
		case filepath.Ext(entry.Name()) == ".csv":
			// ignore trackerfile and sync file
			continue
		case entry.IsDir():
			continue

		case isSynced(entry.Name(), existingSyncs):
			log.InfoContext(ctx, "File has previously been synced. Skipping...", "filename", entry.Name())
			// file has previously been synced so we don't want to push those blobs again.
			continue

		default:
			// parse the file to get the image tag.
			splitName := strings.Split(entry.Name(), "-")
			if len(splitName) != 2 {
				return fmt.Errorf("unexpected file name %s. Batch Deserialize expects tar files with an int-name format, e.g., 0-image1, 1-image2, etc", entry.Name())
			}
			// create the destination target
			repo, err := action.Config.Repository(ctx, strings.Join([]string{destination, strings.Split(splitName[1], ".")[0]}, ":"))
			if err != nil {
				return err
			}
			// build the deserialize options
			opts := mirror.DeserializeOptions{
				DestTarget:          repo,
				DestTargetReference: repo.Reference,
				SourceFile:          filepath.Join(syncDir, entry.Name()),
				BufferSize:          0,
				DryRun:              false,
				RootUI:              rootUI,
				Strict:              false,
				Log:                 log,
			}
			// deserialize each tar file to the destination directory and tag with the image name.
			// e.g., registry.example.com/foo:image1, registry.example.com/foo:image2, etc...
			if err := mirror.Deserialize(ctx, opts); err != nil {
				return err
			}
			successfulSyncs = append(successfulSyncs, []string{entry.Name(), time.Now().String()})
		}
	}
	// write out to successful_syncs.txt the tar file name and the date/timestamp.
	w := csv.NewWriter(file)
	if err := w.WriteAll(successfulSyncs); err != nil {
		return fmt.Errorf("writing to successful syncs file: %w", err)
	}
	return nil
}

func isSynced(filename string, existingSyncs [][]string) bool {
	for _, line := range existingSyncs {
		if filename == line[0] {
			return true
		}
	}
	return false
}
