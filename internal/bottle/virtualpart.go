package bottle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/opencontainers/go-digest"

	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
)

// VirtTrackingFile is the filename for local virtual parts tracking json file.
const VirtTrackingFile = "vlayers.json"

// A VirtRecord is a structure for tracking data for an individual virtual part, containing LayerID, ContentID, and
// a Location reference where the Layer can be found or retrieved.  Content and Layer IDs are stored in digest format;
// with the algorithm included.
type VirtRecord struct {
	LayerID   digest.Digest `json:"layer-id"`
	ContentID digest.Digest `json:"content-id"`
	Location  string        `json:"location"`
}

// VirtualParts represents a set of local bottle part metadata that tracks parts that are not stored in a local copy of
// a bottle.  Virtual parts are added to this collection when a pull uses a selector that excludes parts.  Each virtual
// part is identified by a content ID, layer ID, and a reference where the bottle can be retrieved.
type VirtualParts struct {
	filePath    string
	VirtRecords []VirtRecord `json:"virt-records"`
}

// IsMountable implements the MountableLayers interface for discovering sources for cross repo blob mounting.
func (vp *VirtualParts) IsMountable(layerID digest.Digest, destMount ref.Ref) bool {
	for _, v := range vp.VirtRecords {
		if v.LayerID == layerID {
			s, err := ref.FromString(v.Location, ref.SkipRefValidation)
			if err != nil {
				continue
			}

			if s.Match(destMount, ref.RefMatchRegRepo) {
				// already in repo
				return false
			}
			if s.Match(destMount, ref.RefMatchReg) {
				// found on registry in a different repo
				return true
			}
		}
	}
	return false
}

// MustMount implements the MountableLayers interface for determining whether a layer must be mounted, which is always
// true for virtual parts (since they don't exist locally for upload).
func (vp *VirtualParts) MustMount(layerID digest.Digest) bool {
	return true
}

// Sources implements the MountableLayers interface for identifying a source for a virtual part layer ID.
func (vp *VirtualParts) Sources(layerID digest.Digest, destMount ref.Ref) []ref.Ref {
	for _, v := range vp.VirtRecords {
		if v.LayerID == layerID {
			s, err := ref.FromString(v.Location, ref.SkipRefValidation)
			if err != nil {
				continue
			}
			return []ref.Ref{s}
		}
	}
	return []ref.Ref{}
}

// GetSources implements the cache SourceProvider interface for locating registry hosts for virtualized parts.
func (vp *VirtualParts) GetSources() map[digest.Digest][]string {
	sources := make(map[digest.Digest][]string)
	for _, v := range vp.VirtRecords {
		sources[v.LayerID] = []string{v.Location}
	}
	return sources
}

// AddSource associates a source ref with a provided layer ID.  This currently overwrites an existing source value with
// a new ref.
func (vp *VirtualParts) AddSource(layerID digest.Digest, source ref.Ref, mustMount bool) {
	for i, v := range vp.VirtRecords {
		if v.LayerID == layerID {
			vp.VirtRecords[i].Location = source.String()
		}
	}
}

// NewVirtualPartTracker initializes a VirtualParts object and loads any existing data from disk.
func NewVirtualPartTracker(basePath string) *VirtualParts {
	vp := &VirtualParts{filePath: filepath.Join(basePath, VirtTrackingFile)}
	err := vp.Load()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return vp // vp is unmodified in the case of the file not existing
	}
	return vp
}

// HasLayer returns true if the virtual parts collection contains the specified layer digest.
func (vp *VirtualParts) HasLayer(digest digest.Digest) bool {
	if _, ok := vp.getRecord(digest, true); ok {
		return true
	}
	return false
}

// HasContent returns true if the virtual parts collection contains the specified content digest.
func (vp *VirtualParts) HasContent(digest digest.Digest) bool {
	if _, ok := vp.getRecord(digest, false); ok {
		return true
	}
	return false
}

// Add adds a layer/content and location reference for a layer to the collection.
func (vp *VirtualParts) Add(layerID digest.Digest, contentID digest.Digest, loc ref.Ref) {
	vp.VirtRecords = append(vp.VirtRecords, VirtRecord{
		layerID,
		contentID,
		loc.String(),
	})
}

// getRecord is a helper function for retrieving a VirtRecord based on layerID or contentID, depending on byLayer. False
// is returned if the layer or content cannot be found.
func (vp *VirtualParts) getRecord(id digest.Digest, byLayer bool) (VirtRecord, bool) {
	for _, r := range vp.VirtRecords {
		if byLayer {
			if r.LayerID == id {
				return r, true
			}
		} else {
			if r.ContentID == id {
				return r, true
			}
		}
	}
	return VirtRecord{}, false
}

// GetLayerLocation returns a ref.Ref that describes a location where the layer can be found.
func (vp *VirtualParts) GetLayerLocation(layerID digest.Digest) (ref.Ref, error) {
	if r, ok := vp.getRecord(layerID, true); ok {
		return ref.FromString(r.Location, ref.SkipRefValidation)
	}
	return ref.Ref{}, fmt.Errorf("layer ID not found in virtual parts")
}

// GetContentLocation returns a ref.Ref that describes a location where the layer can be found.
func (vp *VirtualParts) GetContentLocation(layerID digest.Digest) (ref.Ref, error) {
	if r, ok := vp.getRecord(layerID, false); ok {
		return ref.FromString(r.Location, ref.SkipRefValidation)
	}
	return ref.Ref{}, fmt.Errorf("content ID not found in virtual parts")
}

// Load loads virtual part tracking data from disk.
func (vp *VirtualParts) Load() error {
	dat, err := os.ReadFile(vp.filePath)
	if err != nil {
		return fmt.Errorf("error loading virtual part data: %w", err)
	}
	return json.Unmarshal(dat, vp)
}

// Save writes virtual file data to disk,.
func (vp *VirtualParts) Save() error {
	dat, err := json.MarshalIndent(vp, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling virtual part data: %w", err)
	}

	if err := os.WriteFile(vp.filePath, dat, 0o666); err != nil {
		return fmt.Errorf("error creating virtual part file: %w", err)
	}

	return nil
}
