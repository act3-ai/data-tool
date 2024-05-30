package bottle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	latest "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io/v1"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle/label"
	"gitlab.com/act3-ai/asce/data/tool/internal/storage"
	"gitlab.com/act3-ai/asce/data/tool/internal/util"
)

// ErrBottleAlreadyInit is the error received when trying to init an inited bottle.
var ErrBottleAlreadyInit = errors.New("bottle already initialized")

// EntryFile returns a bottle configuration file path for a path name.
func EntryFile(path string) string {
	return filepath.Join(path, ".dt", "entry.yaml")
}

// partsFile returns a bottle parts config file path for a path name.
func partsFile(path string) string {
	return filepath.Join(path, ".dt", "parts.json")
}

// bottleidFile returns a bottleID file path for a path name.
func bottleidFile(path string) string {
	return filepath.Join(path, ".dt", "bottleid")
}

// manifestFile returns the manifest of a bottle, either post bottle push or pull.
func manifestFile(path string) string {
	return filepath.Join(path, ".dt", ".manifest.json")
}

// pullCmdFile returns a pull command file path for a path name.
func pullCmdFile(path string) string {
	return filepath.Join(path, ".dt", "pullcmd")
}

// configFile returns a pull command file path for a path name.
func configFile(path string) string {
	return filepath.Join(path, ".dt", ".config.json")
}

// configDir returns a bottle configuration directory from bottle dir.
func configDir(path string) string {
	return filepath.Join(path, ".dt")
}

// SigDir returns a bottle signature directory from bottle dir.
func SigDir(path string) string {
	return filepath.Join(path, ".signature")
}

// LoadBottle attempts to load an existing bottle configuration from the
// provided bottle path.  opts provides a method to override the default
// bottle configuration.  LocalPath and YAML file options will be
// set by the path argument, so will not need to be supplied.
func LoadBottle(path string, opts ...BOption) (*Bottle, error) {
	bottleOptions := []BOption{
		WithLocalPath(path),
		WithVirtualParts,
		fromFile(path),
	}
	bottleOptions = append(bottleOptions, opts...)

	return NewBottle(bottleOptions...)
}

// createLocalPath is a helper function for creating an output directory path
// intended for bottle download destination.
func (btl *Bottle) createLocalPath() error {
	err := os.MkdirAll(btl.localPath, 0o777)
	if err != nil {
		return fmt.Errorf("error creating bottle directory: %w", err)
	}
	return nil
}

// createScratchPath is a helper function for creating an output directory path
// intended as a working area for creating archives.
func (btl *Bottle) createScratchPath() error {
	err := os.MkdirAll(btl.ScratchPath(), 0o777)
	if err != nil {
		return fmt.Errorf("error creating scratch directory: %w", err)
	}
	return nil
}

// Save writes the bottle information to the local disk based on
// internal path information, in yaml format. if saveJson is true
// a json version of the data is saved as well.
func (btl *Bottle) Save() error {
	btl.updateDefinitionParts()

	if err := btl.writeEntryYAML(); err != nil {
		return err
	}
	if err := btl.saveLocalParts(); err != nil {
		return err
	}

	// save label files
	p := label.Provider{} // temporary
	for _, part := range btl.Parts {
		p.SetLabels(part.Name, part.Labels)
	}
	if err := p.Save(btl.localPath); err != nil {
		return err
	}

	if btl.VirtualPartTracker != nil {
		if err := btl.VirtualPartTracker.Save(); err != nil {
			return err
		}
	}
	return nil
}

// saveLocalParts saves local part information for a bottle at the bottle's configuration path.
func (btl *Bottle) saveLocalParts() error {
	data, err := json.MarshalIndent(btl.Parts, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling local part data: %w", err)
	}

	if err := os.WriteFile(partsFile(btl.localPath), data, 0o666); err != nil {
		return fmt.Errorf("error writing local part file: %w", err)
	}

	return nil
}

// updateDefinitionParts copies the bottle Part data into the Definition.Parts.
func (btl *Bottle) updateDefinitionParts() {
	oldParts := btl.Definition.Parts
	btl.Definition.Parts = make([]latest.Part, 0, len(btl.Parts))
	for _, p := range btl.Parts {
		btl.Definition.Parts = append(btl.Definition.Parts, p.Part)
	}

	if !reflect.DeepEqual(oldParts, btl.Definition.Parts) {
		// oldParts can be empty because Configure can be passed the entry.yaml data
		// TODO Eventually we want to actually preserve the parts data as well (maybe in the config JSON file, but the entry.yaml is overlaid on top of the data).
		btl.invalidateConfiguration()
	}
}

// fromFile defines a source for a bottle definition to be
// the provided file path.
func fromFile(path string) BOption {
	return func(btl *Bottle) error {
		btl.localPath = path
		data, err := os.ReadFile(EntryFile(path))
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("bottle configuration file not found (Has the bottle been initialized?): %w", err)
		}
		if err != nil {
			return fmt.Errorf("error reading bottle config file: %w", err)
		}

		err = btl.Configure(data)
		if err != nil {
			return err
		}
		// HACK we invalidate the configuration because the data we are passing in is not the config (it is YAML and missing the part information after all)
		btl.invalidateConfiguration()

		// load the local parts data from the parts file (if it exists)
		return WithLocalPartFile(partsFile(path))(btl)
	}
}

// WithLocalPartFile specifies a local part file path for reading local part metadata, this function does not
// update the bottle definition with the part information, use SyncDefinitionParts after calling this if needed.
func WithLocalPartFile(jsonPartFile string) BOption {
	return func(btl *Bottle) error {
		data, err := os.ReadFile(jsonPartFile)
		if err != nil {
			// We simply do not use the parts file if we cannot read it
			return nil
		}
		err = json.Unmarshal(data, &btl.Parts)
		if err != nil {
			// We simply do not use the parts file if we cannot decode it
			return nil
		}
		return nil
	}
}

// CreateBottle creates a bottle configuration yaml, including creating
// the path to the yaml file (as determined by the bottle library).  It
// then determines if the set can be initialized, and if so, creates a new
// bottle config file.
func CreateBottle(bottlePath string, overwrite bool) error {
	if err := CheckIfCanInitialize(bottlePath, overwrite); err != nil {
		return err
	}

	if err := os.MkdirAll(configDir(bottlePath), 0o777); err != nil {
		return fmt.Errorf("error creating bottle: %w", err)
	}

	return nil
}

// CheckIfCanInitialize examines the provided path and determines if the
// bottle is already initialized (allowing reinit if force is on), or if
// there is an error initializing.
func CheckIfCanInitialize(bottlePath string, force bool) error {
	_, err := os.Stat(configDir(bottlePath))
	switch {
	case err == nil:
		if force {
			return nil
		}
		return ErrBottleAlreadyInit
	case errors.Is(err, fs.ErrNotExist):
		return nil
	default:
		return fmt.Errorf("bottle initialization error: %w", err)
	}
	// TODO this function looks like it solves nearly the same problem as isBottleDir()
}

// PartSelectorFunc is a function that returns true if the part should be included in the download.
type PartSelectorFunc func(storage.PartInfo) bool

var errNoRootBottleFound = errors.New("root bottle directory was not found")

// getRootBottlePath checks if input path is nested within a bottle directory
// if it is, then the absolute path of the bottle root directory is returned
// if no root directory is found, then it returns an empty string, and an error.
func getRootBottlePath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path: %w", err)
	}

	for {
		isBottle, err := isBottleDir(path)
		if err != nil {
			return "", err
		}

		if isBottle {
			return path, nil
		}

		// go up to the parent
		parent := filepath.Dir(path)
		if parent == path {
			// we are at the top
			break
		}
		path = parent
	}

	return "", errNoRootBottleFound
}

// FindBottleRootDir checks if the current path is nested within a bottle, and returns the root directory of a bottle.
// If root bottle is not found, the function return the input path, along with an error.
func FindBottleRootDir(inputPath string) (string, error) {
	// convert the input path to an absolute one, and fail if we can't.
	inputPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", fmt.Errorf("resolving bottle path to absolute path: %w", err)
	}
	// check if supplied path has is a potential bottle path by checking for entry.yaml
	hasEntryYaml, err := isBottleDir(inputPath)
	if err != nil {
		return "", err
	}
	if hasEntryYaml {
		return inputPath, nil
	}

	// if not found, we check if the supplied path is nested within a bottle by finding the root bottle
	rootPath, err := getRootBottlePath(inputPath)
	if err != nil {
		return inputPath, err
	}
	return rootPath, nil
}

// ErrDirNestedInBottle indicates that the destination of the bottle already has a bottle.
var ErrDirNestedInBottle = errors.New("path is nested within bottle directory")

// VerifyPullDir performs checks on the bottle destination path to ensure that it meets the expected criteria.
func VerifyPullDir(destdir string, desc ocispec.Descriptor) error {
	// Before initiating pull options, we want to check if bottle output directory
	// Expected behavior
	//		if  root (parent) bottle is detected -> ErrDirNestedInBottle
	//
	//		if 	destdir is not empty  -> ErrDirNotEmpty
	//		if destdir empty 		-> destdir, nil

	rootBtlDir, err := getRootBottlePath(destdir)
	if err != nil {
		if !errors.Is(err, errNoRootBottleFound) {
			return err
		}
	}

	// test if the manifest digest of the bottle to be pulled matches what exists
	// TODO: optionally test the bottle dir more rigirously, i.e.
	// are we simply missing a few part?
	if rootBtlDir != "" {
		manBytes, err := os.ReadFile(manifestFile(rootBtlDir))
		switch {
		case errors.Is(err, os.ErrNotExist):
			// unlikely case where the .dt dir exists, but not .manifest.json
			return fmt.Errorf("bottle manifest file not found: %w", err)
		case err != nil:
			return fmt.Errorf("loading bottle manifest: %w", err)
		default:
			d := digest.FromBytes(manBytes)
			if d != desc.Digest {
				return fmt.Errorf("existing bottle does not match bottle to be pulled, have digest '%s', pulling digest '%s'", d, desc.Digest)
			}
			return nil // assume no changes to the bottle have been made since last pull
		}
	}

	empty, err := util.IsDirEmpty(destdir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}

	if !empty {
		return fmt.Errorf("checking directory \"%s\": %w", destdir, ErrDirNotEmpty)
	}

	return nil
}

// isBottleDir returns true if the directory is a bottle
// errors are reported as well
// if dir has a .dt directory then it is considered a bottle.
func isBottleDir(dir string) (bool, error) {
	path := configDir(dir)
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("unable to check for .dt directory: %w", err)
	}
	return info.IsDir(), nil
}
