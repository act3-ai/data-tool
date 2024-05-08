// Package descriptor implements a comparable descriptor and related de-duplication utilities
package descriptor

import (
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// The below is from ORAS https://github.com/oras-project/oras-go/blob/86176e8c5e8c63f418ed2f71bead3abe0b5f2ccb/internal/descriptor/descriptor.go#L31C13-L31C14

// Descriptor contains the minimum information to describe the disposition of
// targeted content.
// Since it only has strings and integers, Descriptor is a comparable struct.
type Descriptor struct {
	// MediaType is the media type of the object this schema refers to.
	MediaType string `json:"mediaType,omitempty"`

	// Digest is the digest of the targeted content.
	Digest digest.Digest `json:"digest"`

	// Size specifies the size in bytes of the blob.
	Size int64 `json:"size"`
}

// FromOCI shrinks the OCI descriptor to the minimum.
func FromOCI(desc ocispec.Descriptor) Descriptor {
	return Descriptor{
		MediaType: desc.MediaType,
		Digest:    desc.Digest,
		Size:      desc.Size,
	}
}

// ToOCI converts the minimal descriptor to a regular OCI descriptor.
func ToOCI(desc Descriptor) ocispec.Descriptor {
	return ocispec.Descriptor{
		MediaType: desc.MediaType,
		Digest:    desc.Digest,
		Size:      desc.Size,
	}
}
