package encoding

import (
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// MediaTypeGather is the artifact type used for the gathered imaged.
	MediaTypeGather = "application/vnd.act3-ace.data.gather+json"
)

// Docker compatible media types.
const (
	MediaTypeDockerManifestList = "application/vnd.docker.distribution.manifest.list.v2+json"
	MediaTypeDockerManifest     = "application/vnd.docker.distribution.manifest.v2+json"
)

// TODO the above definition are also in pkg/oci/manifesthandler.go

// IsIndex returns true if mt is a OCI index compatible media type.
func IsIndex(mt string) bool {
	return mt == ocispec.MediaTypeImageIndex || mt == MediaTypeDockerManifestList
}

// IsImage returns true if mt is a OCI image manifest compatible media type.
func IsImage(mt string) bool {
	return mt == ocispec.MediaTypeImageManifest || mt == MediaTypeDockerManifest
}

// IsManifest returns true if this is any form of manifest/taggable (index or image).
func IsManifest(mt string) bool {
	return IsImage(mt) || IsIndex(mt)
}
