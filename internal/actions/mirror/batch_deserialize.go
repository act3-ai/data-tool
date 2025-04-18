package mirror

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/act3-ai/go-common/pkg/logger"

	"github.com/act3-ai/data-tool/internal/mirror"
	"github.com/act3-ai/data-tool/internal/ui"
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

	existingSyncs := map[string]string{}
	for {
		syncedFile, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("reading successful sync file: %w", err)
		}
		existingSyncs[syncedFile[0]] = "" // we really only need to store the filename for fast retrieval.
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
			destinationWithReference := fmt.Sprintf("%s:%s", destination, strings.Split(splitName[1], ".")[0])
			gt, err := action.Config.GraphTarget(ctx, destinationWithReference)
			if err != nil {
				return err
			}

			// parse with endpoint resolution
			destRef, err := action.Config.ParseEndpointReference(destinationWithReference)
			if err != nil {
				return fmt.Errorf("parsing destination reference: %w", err)
			}
			// build the deserialize options
			opts := mirror.DeserializeOptions{
				DestStorage:         gt,
				DestTargetReference: destRef,
				SourceFile:          filepath.Join(syncDir, entry.Name()),
				BufferSize:          0,
				DryRun:              false,
				RootUI:              rootUI,
				Strict:              false,
				Log:                 log,
			}
			// deserialize each tar file to the destination directory and tag with the image name.
			// e.g., registry.example.com/foo:image1, registry.example.com/foo:image2, etc...
			if _, err := mirror.Deserialize(ctx, opts); err != nil {
				return err
			}
			successfulSyncs = append(successfulSyncs, []string{entry.Name(), destinationWithReference, time.Now().String()})
		}
	}
	// write out to successful_syncs.txt the tar file name and the date/timestamp.
	w := csv.NewWriter(file)
	if len(existingSyncs) == 0 {
		if err := w.Write([]string{"filename", "artifact", "timestamp"}); err != nil {
			return fmt.Errorf("writing successful sync file header: %w", err)
		}
	}
	if err := w.WriteAll(successfulSyncs); err != nil {
		return fmt.Errorf("writing to successful syncs file: %w", err)
	}
	return nil
}

func isSynced(filename string, existingSyncs map[string]string) bool {
	if _, ok := existingSyncs[filename]; ok {
		return true
	}
	return false
}
