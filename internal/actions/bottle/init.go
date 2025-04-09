package bottle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/act3-ai/go-common/pkg/logger"

	"github.com/act3-ai/data-tool/internal/bottle"
)

// Init represents the init action.
type Init struct {
	*Action

	Force bool // Recreate the initialization metadata even if it already exists
}

// Run runs the bottle init action.
func (action *Init) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "bottle init command activated")

	bottlePath, err := createOrVerifyPath(ctx, action.Dir)
	if err != nil {
		return fmt.Errorf("could not make or verify bottle path %s: %w", action.Dir, err)
	}
	err = bottle.CreateBottle(bottlePath, action.Force)
	if err != nil {
		return fmt.Errorf("could not create bottle config at path %s: %w", bottlePath, err)
	}

	opts := []bottle.BOption{
		bottle.WithLocalPath(bottlePath),
		bottle.DisableCache(true),
	}
	btl, err := bottle.NewBottle(opts...)
	if err != nil {
		return err
	}

	// add the parts to the bottle
	_, _, err = bottle.InspectBottleFiles(ctx, btl, bottle.Options{Visitor: bottle.PrepareUpdatedParts(ctx, btl)})
	if err != nil {
		return fmt.Errorf("failed while checking for updated bottle parts: %w", err)
	}

	// now we read the labels after we have all the parts
	if err := btl.LoadLocalLabels(); err != nil {
		return err
	}

	// Update bottle object by processing files.  Save is disabled because we already have the config file open,
	// and init does not perform archive/caching processes on files, in favor of commit and push performing those
	// functions.
	if err := bottle.SaveUpdatesToSet(ctx, btl, bottle.SaveOptions{
		NoArchive: true,
		NoCommit:  true,
		NoDigest:  true,
	}); err != nil {
		return err
	}

	_, err = fmt.Fprintln(out, "Bottle initialization complete.")
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "Init command completed")
	return nil
}

// createOrVerifyPath ensures that a local bottle directory exists, whether that
// means creating a new one, or ensuring one is available.
func createOrVerifyPath(ctx context.Context, bottlePath string) (string, error) {
	log := logger.FromContext(ctx)
	bottlePath, err := filepath.Abs(bottlePath)
	if err != nil {
		return "", fmt.Errorf("error converting bottle directory into absolute path: %w", err)
	}

	log.InfoContext(ctx, "Ensuring the bottle directory exists at path", "path", bottlePath)
	fileInfo, err := os.Stat(bottlePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.InfoContext(ctx, "Path does not exist, creating it", "path", bottlePath)
			err = os.Mkdir(bottlePath, 0o777)
			if err != nil {
				return "", fmt.Errorf("error creating bottle directory: %w", err)
			}
			// successfully create directory
			return bottlePath, nil
		}
		return "", fmt.Errorf("error stat-ing bottle: %w", err)
	}

	if !fileInfo.IsDir() {
		// it is a file
		return "", fmt.Errorf("bottle path should be a directory but instead it is a points to a file (or similar)")
	}

	log.InfoContext(ctx, "Path exists", "path", bottlePath)
	return bottlePath, nil
}
