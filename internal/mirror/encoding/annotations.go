package encoding

const (
	// AnnotationGatherVersion is the key for the annotation to denote the ace-dt version used during gather.
	AnnotationGatherVersion = "vnd.act3-ace.data.version"

	// AnnotationSerializationVersion is the key for the annotaion to denote the serialization format version.
	AnnotationSerializationVersion = "vnd.act3-ace.data.serialization.version"

	// AnnotationExtraManifests is the annotation key used for index of index descriptors.  The value is a JSON encoded array of descriptors.
	AnnotationExtraManifests = "vnd.act3-ace.data.extra-manifests"

	// AnnotationLayerSizeTotal is the size (in bytes) of the layers.
	AnnotationLayerSizeTotal = "vnd.act3-ace.data.layer.size.total"

	// AnnotationLayerSizeDeduplicated is the size (in bytes) of the layers with duplicated removed.
	// This is the size that would need to be transferred if there was no data on the receiving side.
	AnnotationLayerSizeDeduplicated = "vnd.act3-ace.data.layer.size.deduplicated"

	// AnnotationArchiveOffset after this descriptor is written to the tar archive.  This value is the number of bytes written to the tar archive.  In other words it is the minimum number of bytes necessary (of the tar archive) needed to recover this descriptor.
	AnnotationArchiveOffset = "vnd.act3-ace.data.offset"

	// AnnotationLabels is the JSON encoded map of labels.
	AnnotationLabels = "data.act3-ace.io/labels"

	// AnnotationSrcIndex is the string source index of a manifest (sourced from a multi-architecture index). Its digest can be computed to get the original manifest digest/ID.
	AnnotationSrcIndex = "data.act3-ace.io/source-index"
)
