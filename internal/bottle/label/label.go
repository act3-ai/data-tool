// Package label handles part label files
package label

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"

	"gitlab.com/act3-ai/asce/data/tool/internal/util"
)

const (
	// LabelsFilename is the name of the labels marker file.
	LabelsFilename = ".labels.yaml"
	// LabelsFilenameLegacy is the old (ace-dt version<=0.25.4) name of the labels marker file.
	LabelsFilenameLegacy = ".labels.yml"
)

// partLabels tracks part labels the file name as the key.
type partLabels struct {
	Labels map[string]map[string]string `json:"labels,omitempty"`
}

func newPartLabels() *partLabels {
	return &partLabels{
		Labels: make(map[string]map[string]string),
	}
}

// HasSubparts returns true if the path has subparts, error if there was an IO error
// subparts are indicated by the existence of a labels file.
func HasSubparts(fsys fs.FS, pth string) (bool, error) {
	// we want to check the existence of pth just to make sure we are not given a pth that is totally bogus
	if _, err := fs.Stat(fsys, pth); err != nil {
		return false, fmt.Errorf("part directory does not exist: %w", err)
	}

	_, err := fs.Stat(fsys, path.Join(pth, LabelsFilename))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// ignore
	case err != nil:
		return false, fmt.Errorf("failed to check for labels file: %w", err)
	default:
		return true, nil
	}

	_, err = fs.Stat(fsys, path.Join(pth, LabelsFilenameLegacy))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// ignore
	case err != nil:
		return false, fmt.Errorf("failed to check for legacy labels file: %w", err)
	default:
		return true, nil
	}

	return false, nil
}

// Provider is a label provider, keeping track of file label associations in a nested directory structure.
type Provider struct {
	pathLabels map[string]*partLabels
}

// trackPath initializes a pathLabels entry with the provided path string, creating the empty map.
func (p *Provider) trackPath(pth string) {
	if p.pathLabels == nil {
		p.pathLabels = make(map[string]*partLabels)
	}
	if _, ok := p.pathLabels[pth]; ok {
		return
	}
	p.pathLabels[pth] = newPartLabels()
}

// markEmptyPaths accepts a directory path and makes sure each element in the path contains a path labels marker.
// This is intended to be used as a last step after all labels are loaded from other sources to catch paths that
// contain sub entries that don't themselves have labels. The pathStr is assumed to not include the file name, and
// be relative to basePath.
func (p *Provider) markEmptyPaths(pth string) {
	pth = path.Clean(pth)
	// do nothing if the path is empty
	if pth == "." {
		return
	}
	// look at the parent directories
	for pth != "." && pth != "/" {
		p.trackPath(pth)
		pth = path.Dir(pth)
	}
}

// SetLabels sets all the labels on a part.
func (p *Provider) SetLabels(partName string, lbls map[string]string) {
	pp, pn := splitPartName(partName)
	p.trackPath(pp)
	p.pathLabels[pp].Labels[pn] = lbls
	p.markEmptyPaths(pp)
}

// loadLabelFileLegacy allows the label provider to load an older format for label yaml, which only included labels by file
// name.
func (p *Provider) loadLabelFileLegacy(fsys fs.FS, relPath string) error {
	labelData, err := fs.ReadFile(fsys, path.Join(relPath, LabelsFilenameLegacy))
	if err != nil {
		return fmt.Errorf("error reading legacy labels file: %w", err)
	}

	legacyFileLabels := make(map[string]map[string]string)
	if err := yaml.Unmarshal(labelData, &legacyFileLabels); err != nil {
		return fmt.Errorf("failed to parse legacy labels file: %w", err)
	}

	p.trackPath(relPath)
	p.pathLabels[relPath].Labels = legacyFileLabels

	return nil
}

// loadLabelFile reads structured label information from the provided data, and tracks it using the relPath string,
// which can be any identifier, but is primarily used to represent a relative path to the bottle dir.
func (p *Provider) loadLabelFile(fsys fs.FS, relPath string) error {
	labelData, err := fs.ReadFile(fsys, path.Join(relPath, LabelsFilename))
	if err != nil {
		return fmt.Errorf("error reading labels: %w", err)
	}

	var pl partLabels
	if err := yaml.Unmarshal(labelData, &pl); err != nil {
		return fmt.Errorf("error unmarshalling label file: %w", err)
	}

	p.trackPath(relPath)
	p.pathLabels[relPath] = &pl

	return nil
}

// NewProviderFromFS loads all labels from .labels.yaml files found in the specified basePath and subdirs, recursively.
// if a directory does not contain a ".labels.yaml" file, recursion stops for that path.
func NewProviderFromFS(fsys fs.FS) (*Provider, error) {
	p := &Provider{}
	return p, fs.WalkDir(fsys, ".", func(pth string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// directory, check if it has labels file sand load them if they exist

			// try the regular labels file first
			err = p.loadLabelFile(fsys, pth)
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}

			// fallback to legacy labels file
			err = p.loadLabelFileLegacy(fsys, pth)
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}

			// this directory does not contain any further parts (prune the tree)
			return fs.SkipDir
		} // ignore all files
		return nil
	})
}

// writeLabelFile outputs a fileLabels structure to a yaml file.
func writeLabelFile(dir string, fl *partLabels) error {
	data, err := yaml.Marshal(fl)
	if err != nil {
		return fmt.Errorf("error marshalling file labels: %w", err)
	}

	if err = util.CreatePathForFile(filepath.Join(dir, LabelsFilename)); err != nil {
		return fmt.Errorf("unable to create path for labels file: %w", err)
	}

	if err = os.WriteFile(filepath.Join(dir, LabelsFilename), data, 0o666); err != nil {
		return fmt.Errorf("error writing labels file: %w", err)
	}

	// remove the legacy labels file (this is cleanup and done after the new labels are created)
	err = os.Remove(filepath.Join(dir, LabelsFilenameLegacy))
	switch {
	case errors.Is(err, os.ErrNotExist):
		// ignore
	case err != nil:
		return fmt.Errorf("error removing legacy labels file: %w", err)
	}

	return nil
}

// Save writes the label provider label data to yaml files located at the file paths used as the path keys.
func (p *Provider) Save(basePath string) error {
	// Ensure we always have a top level entry
	p.trackPath(".")

	for k, v := range p.pathLabels {
		if err := writeLabelFile(filepath.Join(basePath, filepath.FromSlash(k)), v); err != nil {
			return err
		}
	}

	return nil
}

// LabelsForPart returns the labels for a given part by name.
func (p *Provider) LabelsForPart(partName string) labels.Set {
	pp, pn := splitPartName(partName)
	pl := p.pathLabels[pp]
	if pl == nil {
		return nil
	}
	return labels.Set(pl.Labels[pn])
}

// splitPartName splits a part name into
// the path portion (directory where it's labels file would be),
// the name portion (name in the labels file, with the trailing slash for directories).
func splitPartName(name string) (p string, n string) {
	// parts that are directories are named with a trailing slash.
	// We remove that slash here if present (but record that we did in storeSlash)
	storeSlash := ""
	if strings.HasSuffix(name, "/") {
		storeSlash = "/"
		name = strings.TrimSuffix(name, "/")
	}

	p = path.Dir(name)

	if strings.Contains(name, "/") {
		n = path.Base(name)
	} else {
		n = name
	}
	n += storeSlash

	return
}
