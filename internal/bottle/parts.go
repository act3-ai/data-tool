package bottle

import (
	"time"

	"github.com/opencontainers/go-digest"
	"k8s.io/apimachinery/pkg/labels"

	cfgdef "git.act3-ace.com/ace/data/schema/pkg/apis/data.act3-ace.io/v1"
)

// PartTrack is a superset of the bottle definition part entry, and provides additional fields for tracking information
// and changes to parts.  PartTrack structures implement the oci.FileInfo interface for inspecting part information.
type PartTrack struct {
	cfgdef.Part
	LayerSize   int64         `json:"layerSize"`
	LayerDigest digest.Digest `json:"layerDigest"`
	MediaType   string        `json:"mediaType"`

	Modified time.Time `json:"modified"`
}

// GetName returns the name of the part
// implements oci.FileInfo interface.
func (p *PartTrack) GetName() string {
	return p.Name
}

// GetContentSize for PartTrack returns the size from the part track's part definition. This size is the original part size.
func (p *PartTrack) GetContentSize() int64 {
	return p.Size
}

// GetContentDigest returns the digest of the content (tar file, uncompressed) of the part
// implements oci.FileInfo interface.
func (p *PartTrack) GetContentDigest() digest.Digest {
	return p.Digest
}

// GetLabels implements oci.FileInfo.
func (p *PartTrack) GetLabels() labels.Set {
	return p.Labels
}

// GetLayerDigest for PartEntry implements storage.FileInfo interface, returning the digest for the
// part as represented by a layer when transformed to the part format.
func (p *PartTrack) GetLayerDigest() digest.Digest {
	return p.LayerDigest
}

// GetMediaType for PartTrack implements the GetFormat portion of the oci.FileInfo interface, returning the file entry
// format. This format is the "network format" for the part, which is a potentially tarred/compressed/encrypted
// manifestation of the part, and not the on-disk format of the part files.
func (p *PartTrack) GetMediaType() string {
	if p.MediaType == "" {
		panic("getting empty media type from PartTrack")
	}
	return p.MediaType
}

// GetModTime for PartTrack implements GetModified portion of the oci.FileInfo interface, returning the modified
// time for a part.
func (p *PartTrack) GetModTime() time.Time {
	return p.Modified
}

// SetMediaType sets the media type string for a part.
func (p *PartTrack) SetMediaType(mt string) {
	if mt == "" {
		panic("media type must be non-empty")
	}
	p.MediaType = mt
}

// SetLayerDigest sets a layer digest based on string format for a part.
func (p *PartTrack) SetLayerDigest(dgst digest.Digest) {
	p.LayerDigest = dgst
}

// GetLayerSize returns the size of the layer (compresses size).
func (p *PartTrack) GetLayerSize() int64 {
	return p.LayerSize
}

// SetLayerSize sets the layer size information for a part.
func (p *PartTrack) SetLayerSize(sz int64) {
	if sz < 0 {
		panic("File size must not be negative.")
	}
	p.LayerSize = sz
}

// SetModTime sets modified time information for a part.
func (p *PartTrack) SetModTime(mtime time.Time) {
	p.Modified = mtime
}
