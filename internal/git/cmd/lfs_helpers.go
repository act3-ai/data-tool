package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
)

// LFSObjsPath is the relative path to LFS objects.
//
// In a bare repository this is the full path from the top-level
// directory to the LFS objects sub directory. In a normal repository
// the '.git' prefix must be added.
const LFSObjsPath = "lfs/objects" // ignore '.git' dir as we clone with --bare

// ResolveLFSOIDPath resolves a relative path to an lfs object: `lfs/objects/ab/bc/abcdef...`
//
// TODO: Do we need to support an alternative lfs objects path?
func (c *Helper) ResolveLFSOIDPath(oid string) string {
	return filepath.Join(LFSObjsPath, oid[0:2], oid[2:4], oid) // oid "abcdef" -> ab/cd/abcdef
}

// ListReachableLFSFiles lists all git lfs tracked files reachable from argRevList, returning
// a slice of lfs OIDs.
//
// TODO: This is a very expensive operation.
func (c *Helper) ListReachableLFSFiles(argRevList ...string) ([]string, error) {
	resolver := make(map[string]struct{}) // unlikely that we can predict how many new lfs files
	reachable := make([]string, 0)

	// must do one at a time, see `man git-lfs-ls-files`
	for _, ref := range argRevList {
		newFiles, err := c.LSFiles(ref, "--long", "--deleted")
		if err != nil {
			return nil, fmt.Errorf("retrieving new git lfs tracked files for ref %s: %w", ref, err)
		}
		// add the OID paths
		for _, file := range newFiles {
			split := strings.Fields(file) // split[0] = OID; split[1] = -/*; split[2] = file name
			if _, ok := resolver[split[0]]; !ok {
				resolver[split[0]] = struct{}{}

				reachable = append(reachable, split[0])
			}
		}
	}

	return reachable, nil
}

// ConfigureLFS configures the git repository lfs options.
func (c *Helper) ConfigureLFS() error {
	if c.ServerURL != "" {
		if err := c.Config("lfs.url", c.ServerURL); err != nil {
			return fmt.Errorf("setting lfs.url to %s in git config: %w", c.ServerURL, err)
		}
	}
	return nil
}
