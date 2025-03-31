// Package bottle implements the local bottle handling functionality, providing facilities for:
//
// - loading, storing, configuring, and validating bottle metadata (using external schema definition).
// - managing part information, synchronizing local and remote data as well as tracking part changes.
// - working with bottle and part labels and the local representation of part labels on local storage.
// - managing virtual parts (those that exist in a bottle, but have not been transferred locally due to selector use).
// - facilitating telemetry functionality with bottleID and telemetry events.
package bottle

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	orascontent "oras.land/oras-go/v2/content"
	orasoci "oras.land/oras-go/v2/content/oci"

	bottle "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io"
	cfgdef "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io/v1"
	"gitlab.com/act3-ai/asce/data/schema/pkg/mediatype"
	sutil "gitlab.com/act3-ai/asce/data/schema/pkg/util"

	"gitlab.com/act3-ai/asce/data/tool/internal/bottle/label"
	"gitlab.com/act3-ai/asce/data/tool/internal/cache"
	"gitlab.com/act3-ai/asce/data/tool/internal/oci"
	"gitlab.com/act3-ai/asce/data/tool/internal/util"
)

// Bottle represents a bottle and associated options for managing data
// transfer.  The options are set using BottleOption functional options in
// a call to NewBottle constructor.
type Bottle struct {
	Manifest oci.ManifestHandler

	// Definition is the bottle configuration.
	// Outside of methods of this package, Definition must not be modified
	Definition cfgdef.Bottle

	// configuration data (as JSON) for the bottle
	// cfgData must be cleared by **all** modifications to the bottle
	cfgData []byte

	// OriginalManifest is the manifest that was pulled
	OriginalManifest []byte

	// OriginalConfig is the config that was pulled
	OriginalConfig []byte

	// Parts is the complete list of parts (local and virtual) for this bottle
	Parts []PartTrack

	// localPpath is the local filesystem path for this bottle
	localPath string

	// cachePath is the cache directory
	cachePath string

	cache                orascontent.GraphStorage
	VirtualPartTracker   *VirtualParts
	DisableCreateDestDir bool
	disableCache         bool

	bic cache.BIC
}

// BIC returns the BlobInfoCache.
func (btl *Bottle) BIC() cache.BIC {
	return btl.bic
}

// ScratchPath returns the directory (guaranteed to exist) or cache local temporary files.
func (btl *Bottle) ScratchPath() string {
	return filepath.Join(btl.cachePath, "scratch")
}

// GetPartStatus looks for an existing file entry based on the search file entry
// information, and returns a status bitfield.
func (btl *Bottle) GetPartStatus(ctx context.Context, search PartTrack) PartStatus {
	part := btl.partByName(search.GetName())
	var status PartStatus

	if part == nil {
		return status
	}

	status |= StatusExists
	if exists, _ := btl.cache.Exists(ctx, ocispec.Descriptor{Digest: part.GetLayerDigest()}); exists {
		status |= StatusCached
	}

	if btl.VirtualPartTracker != nil {
		dig := search.GetLayerDigest()
		if btl.VirtualPartTracker.HasContent(dig) {
			status |= StatusVirtual
		}
	}
	if part.GetModTime().Before(search.GetModTime()) {
		status |= StatusChanged
	}
	if part.Size != search.Size {
		status |= StatusChanged
	}
	if search.GetLayerDigest() != "" && search.GetLayerDigest() == part.GetLayerDigest() {
		status |= StatusDigestMatch
	}
	if part.GetContentDigest() == "" {
		status |= StatusChanged
	}

	return status
}

// CalculatePublicArtifactDigest updates the artifact at the specified index with the digest of the corresponding file.
func (btl *Bottle) CalculatePublicArtifactDigest(index int) error {
	if index >= 0 && index >= len(btl.Definition.PublicArtifacts) {
		return fmt.Errorf("specified index %d out of range for CalculatePublicArtifactDigest", index)
	}
	artPath := btl.NativePath(btl.Definition.PublicArtifacts[index].Path)
	dig, err := util.DigestFile(artPath)
	if err != nil {
		return fmt.Errorf("digest calculation failed for public artifact %s: %w", artPath, err)
	}
	btl.Definition.PublicArtifacts[index].Digest = dig
	btl.invalidateConfiguration()

	return nil
}

// applyDefinitionPartData copy Bottle Definition.Part data into the Bottle Parts data.
func (btl *Bottle) applyDefinitionPartData() {
	if len(btl.Parts) != len(btl.Definition.Parts) {
		panic("mismatch in local parts vs definition parts")
	}
	for i, p := range btl.Definition.Parts {
		btl.Parts[i].Part = p
	}
}

// GetManifestAnnotations returns the expiration time of a bottle
// that is set during bottle push
// Needed to comply with the LocalData interface.
func (btl *Bottle) GetManifestAnnotations() map[string]string {
	return nil
}

// GetManifestArtifactType returns the artifact type for a bottle.  Currently, this is a constant string based on
// the schema bottle media type.
func (btl *Bottle) GetManifestArtifactType() string {
	return mediatype.MediaTypeBottle
}

// AddPartMetadata adds information about a part to the bottle.
func (btl *Bottle) AddPartMetadata(name string,
	contentSize int64, contentDigest digest.Digest,
	layerSize int64, layerDigest digest.Digest,
	mediaType string,
	modtime time.Time,
) {
	btl.Parts = append(btl.Parts, PartTrack{
		Part: cfgdef.Part{
			Name:   name,
			Size:   contentSize,
			Digest: contentDigest,
			// Labels: ,
		},
		LayerSize:   layerSize,
		LayerDigest: layerDigest,
		MediaType:   mediaType,
		Modified:    modtime,
	})
	btl.invalidateConfiguration()
}

// RemovePartMetadata removes a file record from a Bottle files list.
func (btl *Bottle) RemovePartMetadata(name string) {
	for i := range btl.Parts {
		if btl.Parts[i].Name == name {
			btl.Parts = append(btl.Parts[:i], btl.Parts[i+1:]...)
			break
		}
	}
	btl.invalidateConfiguration()
}

// UpdatePartMetadata changes the metadata for the part identified by name.
func (btl *Bottle) UpdatePartMetadata(name string,
	contentSize int64, contentDigest digest.Digest,
	lbls map[string]string,
	layerSize int64, layerDigest digest.Digest,
	mediaType string,
	modtime *time.Time,
) {
	part := btl.partByName(name)
	if part == nil {
		panic(fmt.Sprintf("part \"%s\" not found", name))
	}
	// name is immutable
	changed := false
	if contentSize >= 0 && part.Size != contentSize {
		part.Size = contentSize
		// if we change the content size we need to re-digest,
		// unless we're provided the new ones which will be updated below.
		part.Digest = ""
		part.LayerSize = 0
		part.LayerDigest = ""
		changed = true
	}
	if contentDigest != "" && part.Digest != contentDigest {
		part.Digest = contentDigest
		changed = true
	}

	if lbls != nil && !equalLabels(part.Labels, lbls) {
		part.Labels = lbls
		changed = true
	}

	if layerSize >= 0 {
		part.LayerSize = layerSize
	}
	if layerDigest != "" {
		part.LayerDigest = layerDigest
	}

	if mediaType != "" {
		part.MediaType = mediaType
	}

	if modtime != nil {
		part.Modified = *modtime
	}

	if changed {
		btl.invalidateConfiguration()
	}
}

// ConfigDescriptor returns an oci descriptor for a bottle config, as it corresponds to the config in a manifest.
func (btl *Bottle) ConfigDescriptor() (ocispec.Descriptor, error) {
	configData, err := btl.GetConfiguration()
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	dgst := digest.FromBytes(configData)
	return ocispec.Descriptor{
		MediaType: mediatype.MediaTypeBottleConfig,
		Digest:    dgst,
		Size:      int64(len(configData)),
	}, nil
}

// PartDescriptors generates a list of oci descriptors for bottle parts, which correspond to parts as they are
// represented in a manifest.
func (btl *Bottle) PartDescriptors() ([]ocispec.Descriptor, error) {
	finfos := btl.GetParts()
	if len(finfos) == 0 {
		return nil, fmt.Errorf("no parts found in bottle for creating part descriptor list")
	}

	fileDescs := make([]ocispec.Descriptor, 0, len(finfos))
	for _, f := range finfos {
		fdesc := ocispec.Descriptor{
			MediaType: f.GetMediaType(),
			Digest:    f.GetLayerDigest(),
			Size:      f.GetLayerSize(),
		}
		fileDescs = append(fileDescs, fdesc)
	}
	return fileDescs, nil
}

// ConstructManifest creates a ManifestHandler that corresponds to the data within a bottle, including configuration,
// part descriptors (layers), and annotations.  This is a representation from an oci level, so part descriptors are
// at the oci layer level (archived/compressed) and not at the bottle level.  The manifest handler that is stored
// within the bottle.
func (btl *Bottle) ConstructManifest() error {
	// get configuration descriptor
	confFileDesc, err := btl.ConfigDescriptor()
	if err != nil {
		return err
	}

	// get manifest level part descriptors (layers)
	fileDescs, err := btl.PartDescriptors()
	if err != nil {
		return err
	}

	// get manifest level annotations
	manAnnotations := btl.GetManifestAnnotations()

	// use bottle media type as a manifest artifact type
	manMediaType := btl.GetManifestArtifactType()

	// Build manifest data based on config, files, and annotations
	manifestData, err := oci.MakeManifest(confFileDesc, fileDescs, manMediaType, manAnnotations)
	if err != nil {
		return fmt.Errorf("failed to encode manifest data: %w", err)
	}
	btl.Manifest = oci.ManifestFromData(ocispec.MediaTypeImageManifest, manifestData)
	return nil
}

func equalLabels(a, b map[string]string) bool {
	// we are avoiding using reflect here
	// return reflect.DeepEqual(a, b)
	// TODO improve this further (make the comparison more efficient by short circuiting)
	return labels.Set(a).String() == labels.Set(b).String()
}

// writeEntryYaml outputs a Bottle definition yaml in pretty printed format.
// This does not include the part information.
func (btl *Bottle) writeEntryYAML() error {
	data, err := btl.Definition.ToDocumentedYAML()
	if err != nil {
		return fmt.Errorf("error adding comments to yaml data: %w", err)
	}
	if err := os.WriteFile(EntryFile(btl.localPath), data, 0o666); err != nil {
		return fmt.Errorf("error writing bottle config yaml: %w", err)
	}
	return nil
}

// LocalData Interface

// Configure configures a Bottle based on cfgData provided as a slice of bytes.
// cfgData should be the full bottle configuration as JSON.
// if it is not then you ust call btl.invalidateConfiguration() after calling Configure().
func (btl *Bottle) Configure(cfgData []byte) error {
	scheme := runtime.NewScheme()
	err := bottle.AddToScheme(scheme)
	if err != nil {
		return fmt.Errorf("error adding type data to conversion scheme: %w", err)
	}

	codecs := serializer.NewCodecFactory(scheme, serializer.EnableStrict)

	bottleOriginal, err := runtime.Decode(codecs.UniversalDeserializer(), cfgData)
	if err != nil {
		return fmt.Errorf("error decoding config data: %w", err)
	}

	if bottleOriginal.GetObjectKind().GroupVersionKind().GroupVersion() == cfgdef.GroupVersion {
		bottleOriginal.(*cfgdef.Bottle).DeepCopyInto(&btl.Definition)
		btl.cfgData = cfgData
	} else {
		// Upgrade bottle config version
		var ociMan *ocispec.Manifest
		if btl.Manifest != nil {
			manifest := btl.Manifest.GetManifestData()
			ociMan = &manifest
		}
		err = scheme.Convert(bottleOriginal, &btl.Definition, ociMan)
		if err != nil {
			return fmt.Errorf("error converting config data: %w", err)
		}
		// we will not have any part data in the btl.Definition if this is a local bottle (e.g., no manifest)
		btl.invalidateConfiguration()
	}

	btl.applyDefinitionPartData()

	return nil
}

// SetManifest sets a bottle's manifest to the provided ManifestHandler.
func (btl *Bottle) SetManifest(manifest oci.ManifestHandler) {
	btl.Manifest = manifest
	contentList := manifest.GetLayerDescriptors()

	// only want to reallocate if the layer count doesn't match
	count := len(contentList)
	if len(btl.Parts) != count {
		btl.Parts = make([]PartTrack, count)
	}
	for i, desc := range contentList {
		btl.Parts[i].LayerDigest = desc.Digest
		btl.Parts[i].LayerSize = desc.Size
		btl.Parts[i].MediaType = desc.MediaType
		// mod time is updated later as it must align with the
		// mod time of the file itself, otherwise all future evaluations
		// would be false positives.
	}
}

// SetRawManifest retains the provided raw manifest data in the bottle for future reference.
func (btl *Bottle) SetRawManifest(inData []byte) {
	btl.Manifest = oci.ManifestFromData(ocispec.MediaTypeImageManifest, inData)
	// btl.ManifestData = inData
	// btl.ManifestDigest = digest.FromBytes(inData)
}

// invalidateConfiguration invalidates the bottle configuration data so it will have to be re-generated when needed.
func (btl *Bottle) invalidateConfiguration() {
	btl.cfgData = nil
}

// GetConfiguration returns a byte slice containing json
// formatted configuration data for the bottle.
func (btl *Bottle) GetConfiguration() ([]byte, error) {
	// We prefer to use the cfgData if it exists (to avoid insignificant, but bottle ID changing, re-marshalling changes)
	if btl.cfgData == nil {
		btl.updateDefinitionParts()
		d, err := json.Marshal(btl.Definition)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize bottle configuration: %w", err)
		}
		// cache it, until invalidated again
		btl.cfgData = d
	}
	return btl.cfgData, nil
}

// GetPath returns a string representing the local path for the
// Bottle.
func (btl *Bottle) GetPath() string {
	return btl.localPath
}

// GetManifestPath returns a path to the bottle's OCI manifest,
// i.e. '.manifest.json'.
func (btl *Bottle) GetManifestPath() string {
	return manifestFile(btl.GetPath())
}

// GetConfigPath returns a path to the bottle's OCI config,
// i.e. '.config.json'. To get the path for the configuration
// in entry.yaml use EntryFile().
func (btl *Bottle) GetConfigPath() string {
	return configFile(btl.GetPath())
}

// GetCache returns a oras content.GraphStorage interface for working with local
// cache data.
func (btl *Bottle) GetCache() orascontent.GraphStorage {
	return btl.cache
}

// GetBottleID returns a digest for the bottle configuration.
func (btl *Bottle) GetBottleID() digest.Digest {
	cfgData, err := btl.GetConfiguration()
	if err != nil {
		return ""
	}
	return digest.FromBytes(cfgData)
}

// GetBottleManifest returns a digest for the bottle configuration.
func (btl *Bottle) GetBottleManifest() ([]byte, error) {
	if btl.Manifest == nil {
		return nil, nil
	}
	raw, err := btl.Manifest.GetManifestRaw()
	if err != nil {
		return nil, fmt.Errorf("getting JSON encoded bottle manifest: %w", err)
	}
	return raw, nil
}

// NumParts returns the number of parts in the bottle.
func (btl *Bottle) NumParts() int {
	return len(btl.Parts)
}

// GetParts returns a slice of PartTrack structures presented as a slice of FileInfo interfaces.
func (btl *Bottle) GetParts() []PartInfo {
	parts := make([]PartInfo, len(btl.Parts))
	for i, v := range btl.Parts {
		parts[i] = &v
	}
	return parts
}

// GetPartByLayerDescriptor returns a PartInfo for a part based on a descriptor from a manifest, matching by digest.
func (btl *Bottle) GetPartByLayerDescriptor(descriptor ocispec.Descriptor) PartInfo {
	for _, p := range btl.Parts {
		if p.LayerDigest == descriptor.Digest {
			return &p
		}
	}
	return nil
}

// GetPartByName returns a single FileInfo matching the provided name.
func (btl *Bottle) GetPartByName(partName string) PartInfo {
	part := btl.partByName(partName)
	if part == nil {
		return nil
	}
	return part
}

// partByName returns a pointer to a PartTrack matching the provided name.
// returns nil if not found.
func (btl *Bottle) partByName(partName string) *PartTrack {
	// TODO this is inefficient.  store an index by name as a map[string]*PartTrack
	for i, v := range btl.Parts {
		if v.GetName() == partName {
			return &btl.Parts[i]
		}
		// try the slash at the end as well.
		if v.GetName() == partName+"/" {
			return &btl.Parts[i]
		}
	}
	return nil
}

// AddPartLabel adds a label to a part listed in the bottle. Part labels are
// added to LocalPart labels map, then synced to the bottle definitions.
func (btl *Bottle) AddPartLabel(ctx context.Context, key, value, partPath string) error {
	partName, err := btl.partName(partPath)
	if err != nil {
		return err
	}

	part := btl.partByName(partName)
	if part == nil {
		return fmt.Errorf("failed to add part label because part \"%s\" not found", partName)
	}

	if part.Labels == nil {
		part.Labels = map[string]string{}
	}
	part.Labels[key] = value

	return nil
}

// RemovePartLabel removes a label from a part using the provided key.
func (btl *Bottle) RemovePartLabel(ctx context.Context, key, partPath string) error {
	partName, err := btl.partName(partPath)
	if err != nil {
		return err
	}

	part := btl.partByName(partName)
	if part == nil {
		return fmt.Errorf("failed to remove label because part \"%s\" not found", partName)
	}

	delete(part.Labels, key)
	return nil
}

// AddLabel adds a label to the bottle.
func (btl *Bottle) AddLabel(k, v string) {
	if btl.Definition.Labels == nil {
		btl.Definition.Labels = map[string]string{
			k: v,
		}
	} else {
		btl.Definition.Labels[k] = v
	}
	btl.invalidateConfiguration()
}

// RemoveLabel removes the label associated with the input key. Returns an error if the key is not found in labels map
// of bottle schema.
func (btl *Bottle) RemoveLabel(k string) error {
	if _, ok := btl.Definition.Labels[k]; !ok {
		return fmt.Errorf("label with key \"%s\" was not found on bottle", k)
	}
	delete(btl.Definition.Labels, k)
	btl.invalidateConfiguration()
	return nil
}

// DeprecateBottleID adds given bottleID to the bottle's deprecates slice.
func (btl *Bottle) DeprecateBottleID(bottleID digest.Digest) {
	// do not add any duplicates to the dpecrecates slice
	for _, v := range btl.Definition.Deprecates {
		if v == bottleID {
			return
		}
	}
	btl.Definition.Deprecates = append(btl.Definition.Deprecates, bottleID)
	btl.invalidateConfiguration()
}

// AddAnnotation adds input key, and value to a bottle annotations map in the bottle schema.
func (btl *Bottle) AddAnnotation(k, v string) {
	if btl.Definition.Annotations == nil {
		annot := make(map[string]string)
		annot[k] = v
		btl.Definition.Annotations = annot
	} else {
		btl.Definition.Annotations[k] = v
	}
	btl.invalidateConfiguration()
}

// RemoveAnnotation removes the annotation associated with the input key. Returns an error if the key is not found in
// annotation map of bottle schema.
func (btl *Bottle) RemoveAnnotation(k string) error {
	_, ok := btl.Definition.Annotations[k]
	if !ok {
		return fmt.Errorf("annotation with key \"%s\" was not found on bottle", k)
	}
	delete(btl.Definition.Annotations, k)
	btl.invalidateConfiguration()

	return nil
}

// AddSourceInfo adds a source to the bottle's list of sources
// If the name or the uri matches an already available source,
// then the item is updated.
// Return an error if the input source fails validation.
func (btl *Bottle) AddSourceInfo(sourceIn cfgdef.Source) error {
	// validate input source
	if err := sourceIn.Validate(); err != nil {
		return err
	}

	// update source if name or URI matches an already present source
	for idx, item := range btl.Definition.Sources {
		if item.Name == sourceIn.Name {
			btl.Definition.Sources[idx] = sourceIn
			return nil
		}
	}

	btl.Definition.Sources = append(btl.Definition.Sources, sourceIn)
	btl.invalidateConfiguration()

	return nil
}

// RemoveSourceInfo removes a source from the bottle's list of sources
// if the name is not found, then we return an error to the caller.
func (btl *Bottle) RemoveSourceInfo(name string) error {
	currentSources := btl.Definition.Sources
	updatedSources := make([]cfgdef.Source, 0, len(currentSources))
	found := false

	for _, source := range currentSources {
		if strings.EqualFold(source.Name, strings.TrimSpace(name)) {
			found = true
			continue // if name is found, then exit the loop
		}
		updatedSources = append(updatedSources, source)
	}

	if !found {
		return fmt.Errorf("source with name \"%s\" was not found", name)
	}

	// update the bottle authors
	btl.Definition.Sources = updatedSources
	btl.invalidateConfiguration()

	return nil
}

// AddMetricInfo adds a metric to the bottle's list of metrics.
// If the name of the input metric matches an already present
// metrics, then the metric is updated.
// Return an error if the input metric fails validation.
func (btl *Bottle) AddMetricInfo(metricIn cfgdef.Metric) error {
	// validate input metric
	if err := metricIn.Validate(); err != nil {
		return fmt.Errorf("failed to validate metric: %w", err)
	}

	// update metric if input name matches
	for idx, item := range btl.Definition.Metrics {
		if item.Name == metricIn.Name {
			btl.Definition.Metrics[idx] = metricIn
			return nil
		}
	}

	btl.Definition.Metrics = append(btl.Definition.Metrics, metricIn)
	btl.invalidateConfiguration()

	return nil
}

// RemoveMetricInfo removes the metric with specified name from the bottle.
// If the metric is not found, then an error is returned to the caller.
func (btl *Bottle) RemoveMetricInfo(metricName string) error {
	currentMetrics := btl.Definition.Metrics
	updatedMetrics := make([]cfgdef.Metric, 0, len(currentMetrics))
	found := false

	for _, metric := range currentMetrics {
		if strings.EqualFold(metric.Name, strings.TrimSpace(metricName)) {
			found = true
			continue // if name is found, then exit the loop
		}
		updatedMetrics = append(updatedMetrics, metric)
	}

	if !found {
		return fmt.Errorf("metric with \"%s\" was not found", metricName)
	}

	// update the bottle authors
	btl.Definition.Metrics = updatedMetrics
	btl.invalidateConfiguration()

	return nil
}

// AddAuthorInfo adds an authors to the bottle's list of authors
// If the name or the email being added matches an existing author
// that author entry is updated
// Returns an error when input is invalid.
func (btl *Bottle) AddAuthorInfo(authorIn cfgdef.Author) error {
	if err := authorIn.Validate(); err != nil {
		return fmt.Errorf("ozzo-validation failed to validate author: %w", err)
	}

	// update metric if input name matches
	for idx, item := range btl.Definition.Authors {
		if item.Name == authorIn.Name || item.Email == authorIn.Email {
			btl.Definition.Authors[idx] = authorIn
			return nil
		}
	}

	btl.Definition.Authors = append(btl.Definition.Authors, authorIn)
	btl.invalidateConfiguration()

	return nil
}

// RemoveAuthorInfo removes author information from the bottle's definition
// using author's name as a key. If name is not found, then an error is returned.
func (btl *Bottle) RemoveAuthorInfo(name string) error {
	currentAuthors := btl.Definition.Authors
	updatedAuthors := make([]cfgdef.Author, 0, len(currentAuthors))

	found := false
	for _, author := range currentAuthors {
		if strings.EqualFold(author.Name, strings.TrimSpace(name)) {
			found = true
			continue // if name is found, then skip the loop
		}
		updatedAuthors = append(updatedAuthors, author)
	}

	if !found {
		return fmt.Errorf("author with name \"%s\" was not found", name)
	}

	// update the bottle authors
	btl.Definition.Authors = updatedAuthors
	btl.invalidateConfiguration()

	return nil
}

// SetDescription add or update the description of a bottle.
func (btl *Bottle) SetDescription(content string) {
	btl.Definition.Description = content
	btl.invalidateConfiguration()
}

// partName computes the UNIX style path that is relative to the bottle
// It also validates that it contains valid characters.
// This relative path is also known as a the part's name.
// pth is a native OS path to a file/dir within a bottle.
func (btl *Bottle) partName(pth string) (string, error) {
	// need the absolute path so we can later make it bottle relative
	absPath, err := filepath.Abs(pth)
	if err != nil {
		return "", fmt.Errorf("error resolving artifact path: %w", err)
	}

	// Get bottle relative path of part
	relPath, err := filepath.Rel(btl.localPath, absPath)
	if err != nil {
		return "", fmt.Errorf("artifact path %s is not relative to bottle path %s: %w", absPath, btl.localPath, err)
	}

	relPath = filepath.ToSlash(relPath)

	// relative path returns directory navigation if a relative path can be found by navigating up to a common base dir
	// this occurs for instance if cwd is not within the bottle dir, and the path to the part is provided relative to
	// the bottle dir.  e.g. labeling part.txt located at /path/to/bottle/part.txt while cwd= /path/to/another/dir
	// We can probably automatically fix this for the user here by joining the provided path with the bottle dir, but
	// there might be corner cases with that approach... instead just give an error so the user can clarify intention.
	if strings.HasPrefix(relPath, "../") || relPath == ".." {
		return "", fmt.Errorf("artifact path \"%s\" is not relative to bottle directory", absPath)
	}

	if !sutil.IsPortablePath(relPath) {
		return "", fmt.Errorf("path \"%s\" is not portable.  Derived from \"%s\"", relPath, pth)
	}

	return relPath, nil
}

// NativePath converts the bottle relative UNIX path to a host native path
// This is the inverse of partName().  e.g., nativePath(partName("/foo/bar")) == "/foo/bar".
func (btl *Bottle) NativePath(pth string) string {
	return filepath.Join(btl.localPath, filepath.FromSlash(pth))
}

// AddArtifact adds a public artifact to the bottle definition.
// The artifact is keyed on a path. If the artifact passed in is found
// in the definition, the item is updated. If no changes are found, then
// the function returns without making any changes to the bottle definition.
// If a field of the bottle is invalid, an error is returned.
func (btl *Bottle) AddArtifact(name, pth, mediaType string, dgst digest.Digest) error {
	bottleRelPath, err := btl.partName(pth)
	if err != nil {
		return fmt.Errorf("artifact path: %w", err)
	}

	// Create new public artifact information
	artifact := cfgdef.PublicArtifact{
		Name:      name,
		Path:      bottleRelPath,
		MediaType: mediaType,
		Digest:    dgst,
	}

	if err := artifact.Validate(); err != nil {
		return err
	}

	// log.Info("Saving bottle public artifact", "artifact", artifact)

	// check if artifact is duplicate, and create
	for idx, currArt := range btl.Definition.PublicArtifacts {
		if currArt.Path == artifact.Path {
			// if the path is the same, but the other data is different, then we
			// update the artifact and return early
			if artifact != currArt {
				btl.Definition.PublicArtifacts[idx] = artifact
				btl.invalidateConfiguration()
				return nil
			}
			// otherwise, nothing has changed, so we return
			return nil
		}
	}

	// since it was not found in current list of artifacts, we add it to the definition
	btl.Definition.PublicArtifacts = append(btl.Definition.PublicArtifacts, artifact)
	btl.invalidateConfiguration()

	return nil
}

// RemoveArtifact removes a public artifact from the bottle definition.
// It looks for the provided path in the public artifact's definition. If
// the specified path is not found, then an error is returned.
func (btl *Bottle) RemoveArtifact(pth string) error {
	bottleRelPath, err := btl.partName(pth)
	if err != nil {
		return fmt.Errorf("artifact path: %w", err)
	}

	currentPubArt := btl.Definition.PublicArtifacts
	updatedPubArt := make([]cfgdef.PublicArtifact, 0, len(currentPubArt))
	found := false
	for _, pubArt := range currentPubArt {
		if pubArt.Path == bottleRelPath {
			found = true
			continue // if name is found, then skip the loop
		}
		updatedPubArt = append(updatedPubArt, pubArt)
	}

	if !found {
		return fmt.Errorf("artifact at path \"%s\" was not found in bottle (derived from \"%s\")", bottleRelPath, pth)
	}

	// update the bottle authors
	btl.Definition.PublicArtifacts = updatedPubArt
	btl.invalidateConfiguration()

	return nil
}

// Touch marks all file entries with the file system reported time if the part exists and can be accessed.  The modtime
// information is only updated if the existing modified time is earlier than the time reported by the filesystem.
func (btl *Bottle) Touch() {
	for i := range btl.Parts {
		realName := btl.NativePath(btl.Parts[i].Name)
		fi, err := os.Stat(realName)
		if err != nil {
			// either can't access the file or it doesn't exist, either way just skip the modified data
			continue
		}

		// only update modified time if the file on disk is more recent than the data in the Bottle
		if btl.Parts[i].Modified.Before(fi.ModTime()) {
			btl.Parts[i].Modified = fi.ModTime()
		}
	}
}

// VerifyBottleID verifies that the bottle ID matched what is passed in.
// The algorithm in the matchBottleID is used to compute the digest of the bottle config.
func (btl *Bottle) VerifyBottleID(matchBottleID digest.Digest) error {
	configBytes := btl.OriginalConfig
	// We use the algorithm from the match bottle ID when computing the bottle ID

	alg := matchBottleID.Algorithm()
	if !alg.Available() {
		return fmt.Errorf("digest \"%s\" not available", alg)
	}

	bottleID := alg.FromBytes(configBytes)
	if bottleID != matchBottleID {
		return fmt.Errorf("bottleID verification failed! Received %s", bottleID)
	}

	return nil
}

// BOption defines a function that operates on a Bottle to configure options
// A variadic list fo BottleOptions is provided to the Bottle constructor to
// set desired options (and use defaults for the rest).
type BOption func(*Bottle) error

// NewReadOnlyBottle creates a bottle without initializing any features that would produce side effects, such as
// creating directories, initializing authentication, performing local metadata initializations.  Despite the name, a
// read only bottle can be modified after creation, such as configuring with config data or modifying bottle settings
// directly.
func NewReadOnlyBottle() (*Bottle, error) {
	btl := &Bottle{
		Definition:           cfgdef.NewBottle(),
		disableCache:         true,
		DisableCreateDestDir: true,
	}

	return btl, nil
}

// NewBottle creates a new Bottle and prepares it for transfer using the Bottle.Get
// function.  The created bottle is configured with the options provided, which
// are any number of functional options.  For functionality, at least remote source
// and local path should be defined in the option list.   If not disabled, the
// output directory is created during construction.
func NewBottle(options ...BOption) (*Bottle, error) {
	btl := &Bottle{
		Definition: cfgdef.NewBottle(),
	}

	for _, o := range options {
		if err := o(btl); err != nil {
			return nil, err
		}
	}
	if btl.cachePath == "" {
		btl.cachePath = filepath.Join(btl.localPath, ".dt", "cache")
	}
	if !btl.DisableCreateDestDir {
		if err := btl.createLocalPath(); err != nil {
			return nil, err
		}
		if err := btl.createScratchPath(); err != nil {
			return nil, err
		}
	}
	if !btl.disableCache {
		storage, err := orasoci.NewStorage(btl.cachePath)
		if err != nil {
			return nil, fmt.Errorf("initializing cache storage: %w", err)
		}

		// setup optimization
		optStorage, err := cache.NewFileMounter(btl.cachePath, storage)
		if err != nil {
			// non-fatal, but this should never fail
			btl.cache = cache.NewPredecessorCacher(storage)
		} else {
			// support handling of signatures and other predecessors
			btl.cache = cache.NewPredecessorCacher(optStorage)
		}
	} else {
		btl.cache = &cache.NilCache{}
	}

	// if btl.VirtualPartTracker != nil {
	// 	btl.FilterCallback = watchFilteredParts(btl.FilterCallback, btl)
	// }

	return btl, nil
}

// WithLocalPath defines a local path for a Bottle, which can be
// the destination for a download or source for upload.
func WithLocalPath(pth string) BOption {
	return func(btl *Bottle) error {
		btl.localPath = pth
		return nil
	}
}

// WithCachePath defines a file path where Bottle blobs are stored
// during download and upload.
func WithCachePath(pth string) BOption {
	return func(btl *Bottle) error {
		btl.cachePath = pth
		return nil
	}
}

// WithBlobInfoCache defines a file path where a blob info cache is discovered or created, usually this will be
// in the cache path.  If an empty string is provided, a memory based BIC will be used.
func WithBlobInfoCache(bicPath string) BOption {
	return func(btl *Bottle) error {
		// Preserve the ability to set up a memory based BIC for when persistent BIC isn't desired
		if bicPath != "" {
			btl.bic = cache.NewCache(filepath.Join(bicPath, "blobinfocache.boltdb"))
		} else {
			btl.bic = cache.NewCache("")
		}
		return nil
	}
}

// WithLocalLabels populates the bottle label provider with the labels found in .labels.yaml files on the filesystem,
// located in the provided path.
func WithLocalLabels() BOption {
	return func(btl *Bottle) error {
		return btl.LoadLocalLabels()
	}
}

// LoadLocalLabels loads labels files.
func (btl *Bottle) LoadLocalLabels() error {
	p, err := label.NewProviderFromFS(os.DirFS(btl.localPath))
	if err != nil {
		return err
	}

	// save labels to btl.Parts
	for i, part := range btl.Parts {
		btl.Parts[i].Labels = p.LabelsForPart(part.Name)
	}

	return nil
}

// WithVirtualParts populates a virtual part tracker with virtual part information found in a bottle config directory.
func WithVirtualParts(btl *Bottle) error {
	btl.VirtualPartTracker = NewVirtualPartTracker(configDir(btl.localPath))
	return nil
}

// DisableDestinationCreate sets or clears a flag that optionally
// disables the creation of an output path before transfer. This
// only pertains to Bottle downloads.
func DisableDestinationCreate(disable bool) BOption {
	return func(btl *Bottle) error {
		btl.DisableCreateDestDir = disable
		return nil
	}
}

// DisableCache sets or clears a flag that optionally disables the
// usage of cache storage. This is primarily intended for use when
// no data transfer is expected for bottle items (such as when
// performing a remote query).
func DisableCache(disable bool) BOption {
	return func(btl *Bottle) error {
		btl.disableCache = disable
		return nil
	}
}
