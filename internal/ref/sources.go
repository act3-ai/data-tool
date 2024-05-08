package ref

import (
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// AnnotationSrcRef is an annotation key for a record containing the original source reference for a manifest.
	// This is what the user requested verbatim.
	AnnotationSrcRef = "vnd.act3-ace.manifest.source" // There is no OCI Spec annotation for this

	// AnnotationRefName is an annotation key for denoting the fully qualified name of an image.
	AnnotationRefName = ocispec.AnnotationRefName
)
