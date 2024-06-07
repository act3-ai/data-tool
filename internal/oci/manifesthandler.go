// Package oci implements oci specific functionality, primarily related to interpreting and serializing manifests and
// manifest lists.  Bottles are stored as oci objects with standard manifest formats, but have special handling for the
// purposes of working with configuration blobs.
package oci

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/opencontainers/image-spec/specs-go"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"git.act3-ace.com/ace/data/schema/pkg/validation"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
)

const (
	// ManifestUnsupportedType indicates that the server returned an unsupported media type for a manifest.
	ManifestUnsupportedType = 1
	// ManifestBadFormat indicates a badly formatted manifest was returned from the server.
	ManifestBadFormat = 2
	// ManifestUnsupportedArchitecture indicates a failure to identify a valid architecture for an image.
	ManifestUnsupportedArchitecture = 3
	// ManifestOK indicates that a manifest pull occurred successfully.
	ManifestOK = 200
	// ManifestBadRequest is a manifest pull status indicating a malformed repository Name.
	ManifestBadRequest = 400
	// ManifestAuthRequired is a manifest pull status indicating authentication needs to be performed.
	ManifestAuthRequired = 401
	// ManifestAccessDenied is a manifest pull status indicating user doesn't have access to repository.
	ManifestAccessDenied = 403
	// ManifestNotFound is a manifest pull status indicating that the requested manifest is not found.
	ManifestNotFound = 404
)

// CtxPlatformKey is a context key type for specifying context platform values for image selection.
type CtxPlatformKey string

const (
	// CtxPlatform specifies a combined os/arch platform selector.
	CtxPlatform CtxPlatformKey = "platform"
	// CtxPlatformOs specifies an os platform selector.
	CtxPlatformOs CtxPlatformKey = "platform_os"
	// CtxPlatformArch specifies an architecture platform selector.
	CtxPlatformArch CtxPlatformKey = "platform_arch"
	// CtxPlatformVariant specifies an architecture variant platform selector.
	CtxPlatformVariant CtxPlatformKey = "platform_variant"
)

// reportStatus logs results of manifest retrieval and returns true if the transfer was successful
// func reportStatus(ctx context.Context, srcInfo ref.Ref, manifest ManifestHandler) bool {
// 	log := logger.FromContext(ctx)
// 	ret := false
// 	switch manifest.GetStatus().Status {
// 	case ManifestOK:
// 		log.V(1).Info("Manifest pull successful", "ref", srcInfo.String())
// 		ret = true
// 	case ManifestUnsupportedType:
// 		log.V(1).Info("Unsupported manifest format", "ref", srcInfo.String())
// 	case ManifestBadFormat:
// 		log.V(1).Info("Failed to parse manifest parse", "ref", srcInfo.String(), "status", manifest.GetStatus().StatusInfo)
// 	case ManifestBadRequest:
// 		log.V(1).Info("Bad reference request", "ref", srcInfo.String(), "status", 400)
// 	case ManifestAuthRequired:
// 		log.V(1).Info("Authorization required, or access denied", "ref", srcInfo.String(), "status", 401)
// 	case ManifestAccessDenied:
// 		log.V(1).Info("Access to reference denied", "ref", srcInfo.String(), "status", 403)
// 	case ManifestNotFound:
// 		log.V(1).Info("Reference not found", "ref", srcInfo.String(), "status", 404)
// 	default:
// 		log.V(1).Info("Unexpected manifest status", "ref", srcInfo.String(), "status", manifest.GetStatus().Status)
// 	}
// 	return ret
// }

const (
	// MediaTypeImageManifestDocker represents the media type for docker's v2 manifest, which matches the oci manifest
	// spec, but must be added as an acceptable media type.
	MediaTypeImageManifestDocker string = "application/vnd.docker.distribution.manifest.v2+json"
	// MediaTypeImageManifestDockerList represents the media type for manifest list content, which is not supported by
	// oci, but provides a redirection to alternate platform versions (or sub-manifests?)
	MediaTypeImageManifestDockerList string = "application/vnd.docker.distribution.manifest.list.v2+json"
)

// ManifestHandler defines an interface for unifying layer descriptor extraction from manifests, different manifest
// formats store layer descriptors differently.
type ManifestHandler interface {
	// GetManifestDescriptor returns the OCI descriptor for the manifest data
	GetManifestDescriptor() ocispec.Descriptor
	// GetConfigDescriptor returns the OCI descriptor for the config record within the manifest data
	GetConfigDescriptor() ocispec.Descriptor
	// GetLayerDescriptors returns a list of OCI descriptors for each layer specified in a manifest
	GetLayerDescriptors() []ocispec.Descriptor
	// GetContentDescriptors returns a list of OCI descriptors for each layer and config specified in a manifest
	GetContentDescriptors() []ocispec.Descriptor
	// GetManifestData returns an OCI manifest structure filled with manifest data
	GetManifestData() ocispec.Manifest
	// GetManifestRaw returns a json formatted slice of bytes representing the manifest data
	GetManifestRaw() ([]byte, error)
	// GetStatus returns information about the transfer of a manifest, including http and internal error information
	GetStatus() ManifestStatusInfo
	// AddAnnotation adds a key value annotation to the manifest
	AddAnnotation(key string, value string)
	// GetAnnotation returns a annotation value based on a given key, or empty if the key is not found
	GetAnnotation(key string) string
	// Copy duplicates a manifest handler, changes to be made without affecting the original. Notably, this facilitates
	// tracking the transfer of a manifest to multiple destinations asynchronously (due to transfer status information
	// for GetStatus needing to be tracked separately)
	Copy() ManifestHandler
}

// ManifestsOption allows configuration of an oci manifests structure on creation, such as assigning a transferworker.
type ManifestsOption func(ocim *Manifests)

// Manifests is an object that implements retrieval of manifests specified in source set.  A destination store
// provides a location where manifest blobs can be saved, and prior existence of objects can be checked with an
// Exist set.
type Manifests struct {
	SourceSet       ref.Set
	Client          *http.Client
	InsecureAllowed bool
	manifests       []ManifestHandler
	Concurrency     int
}

// ManifestFromData generates a ManifestHandler from json formatted data bytes and a content type hint.
func ManifestFromData(contentType string, data []byte) ManifestHandler {
	switch contentType {

	// TODO: Reevaluate the below comment;  this does not seem to be the case, the content types are not fully compatible

	// using the content type from the request data, we determine what media type
	// to mark the image as. Since "application/vnd.oci.image.manifest.v1+json" and
	// "application/vnd.docker.distribution.manifest.v2+json" are compatible,
	// we elect to choose the first one as the media type, as it is matches
	// oci specification
	// reference: https://github.com/opencontainers/image-spec/blob/main/media-types.md#compatibility-matrix=
	case ocispec.MediaTypeImageManifest, MediaTypeImageManifestDocker:
		var ociman ocispec.Manifest
		status := ManifestOK
		statusInfo := ""
		err := json.Unmarshal(data, &ociman)
		if err != nil {
			status = ManifestBadFormat
			statusInfo = "Failed to parse manifest info from remote: " + err.Error()
			err = fmt.Errorf("failed to parse manifest info from remote: %w", err)
		}
		return &BottleManifest{
			ocispec.Descriptor{
				Digest:    digest.FromBytes(data),
				Size:      int64(len(data)),
				MediaType: contentType,
			},
			ociman,
			data,
			ManifestStatusInfo{status, statusInfo, err},
		}
	case MediaTypeImageManifestDockerList:
		var manlist ocispec.Index
		status := ManifestOK
		statusInfo := ""
		err := json.Unmarshal(data, &manlist)
		if err != nil {
			status = ManifestBadFormat
			statusInfo = "Failed to parse manifest list from remote: " + err.Error()
			err = fmt.Errorf("failed to parse manifest list from remote: %w", err)
		}
		return &ManifestListHandler{
			ocispec.Descriptor{
				Digest:    digest.FromBytes(data),
				Size:      int64(len(data)),
				MediaType: contentType,
			},
			manlist,
			data,
			ManifestStatusInfo{status, statusInfo, err},
		}
	default:
		return &BottleManifest{
			ocispec.Descriptor{},
			ocispec.Manifest{},
			[]byte{},
			ManifestStatusInfo{
				ManifestUnsupportedType,
				"Unsupported manifest media type: " + contentType,
				errors.New("Unsupported manifest media type " + contentType),
			},
		}
	}
}

// MakeManifest creates a manifest object and encodes it to json, including config descriptor, layers, artifactType and
// annotations.  If an artifact type is not known, an empty string can be used to avoid adding it to the manifest.
func MakeManifest(config ocispec.Descriptor, layers []ocispec.Descriptor, artifactType string, annos map[string]string) ([]byte, error) {
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2, // historical value. does not pertain to OCI or docker version
		},
		MediaType:   ocispec.MediaTypeImageManifest,
		Config:      config,
		Layers:      layers,
		Annotations: annos,
	}
	if artifactType != "" {
		manifest.ArtifactType = artifactType
	}
	if err := validation.ValidateManifest(manifest); err != nil {
		return nil, err
	}
	return json.Marshal(manifest)
}

// GetNumSources returns the number of manifests in the cache if there are any, or the number of source refs defined
// in the internal SourceSet.
func (om Manifests) GetNumSources() int {
	if len(om.manifests) != 0 {
		return len(om.manifests)
	}
	return len(om.SourceSet.Refs())
}

// ManifestStatusInfo is a structure for recording status information encountered while pulling a manifest.
type ManifestStatusInfo struct {
	Status     int
	StatusInfo string
	Error      error
}

//// ManifestHandlers for each type of manifest;  generic oci (bottles), Docker, helm

// BottleManifest is an alias for ocispec.Manifest which implements the manifest handler interface. It also contains
// information about the manifest descriptor itself (notably, size, Digest, etc), and a status structure used during
// transfer. This structure implements the ManifestHandler interface.
type BottleManifest struct {
	Descriptor  ocispec.Descriptor
	Manifest    ocispec.Manifest
	rawManifest []byte
	ManifestStatusInfo
}

// GetManifestDescriptor returns the oci descriptor for the manifest itself, implementing the ManifestHandler interface.
func (b *BottleManifest) GetManifestDescriptor() ocispec.Descriptor {
	return b.Descriptor
}

// GetManifestData returns the full manifest data object.
func (b *BottleManifest) GetManifestData() ocispec.Manifest {
	return b.Manifest
}

// GetManifestRaw returns the original raw data for the manifest, useful to avoid errors that occur when roundtripping
// decode/encode for non-bottle manifests.
func (b *BottleManifest) GetManifestRaw() ([]byte, error) {
	manifestBytes, err := json.Marshal(b.Manifest)
	if err != nil {
		return nil, err
	}
	return manifestBytes, nil
}

// GetConfigDescriptor returns the configuration descriptor contained within the manifest.  Note this does not include
// any config data apart from annotations, a separate transfer is necessary to retrieve the required data.
func (b *BottleManifest) GetConfigDescriptor() ocispec.Descriptor {
	return b.Manifest.Config
}

// GetLayerDescriptors implements the ManifestHandler interface.
func (b *BottleManifest) GetLayerDescriptors() []ocispec.Descriptor {
	return b.Manifest.Layers
}

// GetContentDescriptors returns a collection of content descriptors, with the config blob first, implementing the
// ManifestHandler interface.
func (b *BottleManifest) GetContentDescriptors() []ocispec.Descriptor {
	descs := make([]ocispec.Descriptor, 0, len(b.Manifest.Layers)+1)
	descs = append(descs, b.Manifest.Config)
	descs = append(descs, b.Manifest.Layers...)

	return descs
}

// GetStatus returns the status information after a transfer for a manifest, implementing the ManifestHandler interface.
func (b *BottleManifest) GetStatus() ManifestStatusInfo {
	return b.ManifestStatusInfo
}

// AddAnnotation adds the provided key/value annotation to the manifest descriptor, allowing custom information to be
// assigned to individual manifests, and implementing the ManifestHandler interface.
func (b *BottleManifest) AddAnnotation(key string, value string) {
	if b.Descriptor.Annotations == nil {
		b.Descriptor.Annotations = make(map[string]string)
	}
	b.Descriptor.Annotations[key] = value
}

// GetAnnotation returns the annotation for the manifest denoted by key, or an empty string.
func (b *BottleManifest) GetAnnotation(key string) string {
	if b.Descriptor.Annotations == nil {
		return ""
	}
	if s, ok := b.Descriptor.Annotations[key]; ok {
		return s
	}
	return ""
}

// Copy creates a fresh manifest handler based on the data contained in the current BottleManifest. This allows a
// duplicate handler to be modified (such as adding annotations) without updating the originating handler.
func (b *BottleManifest) Copy() ManifestHandler {
	return ManifestFromData(b.Descriptor.MediaType, b.rawManifest)
}

// ManifestListHandler is a manifest handler representing a list of manifests as provided by the MediaTypeManifestDockerList.
type ManifestListHandler struct {
	Descriptor   ocispec.Descriptor
	ManifestList ocispec.Index
	rawManifest  []byte
	ManifestStatusInfo
}

// GetManifestDescriptor for ManifestListHandler returns the oci descriptor for the manifest itself, implementing
// the ManifestHandler interface.
func (mlh *ManifestListHandler) GetManifestDescriptor() ocispec.Descriptor {
	return mlh.Descriptor
}

// GetManifestData for ManifestListHandler returns an empty Manifest object.
func (mlh *ManifestListHandler) GetManifestData() ocispec.Manifest {
	return ocispec.Manifest{}
}

// GetManifestRaw returns the original raw data for the manifest list.
func (mlh *ManifestListHandler) GetManifestRaw() ([]byte, error) {
	manifestBytes, err := json.Marshal(mlh.ManifestList)
	if err != nil {
		return nil, err
	}
	return manifestBytes, nil
}

// GetContentDescriptors returns a collection of manifest descriptors from the manifest list.
func (mlh *ManifestListHandler) GetContentDescriptors() []ocispec.Descriptor {
	return mlh.ManifestList.Manifests
}

// GetLayerDescriptors implements ManifestHandler.
func (mlh *ManifestListHandler) GetLayerDescriptors() []ocispec.Descriptor {
	return []ocispec.Descriptor{}
}

// GetConfigDescriptor returns an empty descriptor, as manifest lists do not include config information.
func (mlh *ManifestListHandler) GetConfigDescriptor() ocispec.Descriptor {
	return ocispec.Descriptor{}
}

// GetStatus returns the status information after a transfer for a manifest list, implementing the ManifestHandler
// interface.
func (mlh *ManifestListHandler) GetStatus() ManifestStatusInfo {
	return mlh.ManifestStatusInfo
}

// AddAnnotation adds the provided key/value annotation to the manifest descriptor, allowing custom information to be
// assigned to individual manifests, and implementing the ManifestHandler interface.
func (mlh *ManifestListHandler) AddAnnotation(key string, value string) {
	if mlh.Descriptor.Annotations == nil {
		mlh.Descriptor.Annotations = make(map[string]string)
	}
	mlh.Descriptor.Annotations[key] = value
}

// GetAnnotation returns the annotation for the manifest denoted by key, or an empty string.
func (mlh *ManifestListHandler) GetAnnotation(key string) string {
	if mlh.Descriptor.Annotations == nil {
		return ""
	}
	if s, ok := mlh.Descriptor.Annotations[key]; ok {
		return s
	}
	return ""
}

// Copy creates a fresh manifest handler based on the data contained in the current BottleManifest. This allows a
// duplicate handler to be modified (such as adding annotations) without updating the originating handler.
func (mlh *ManifestListHandler) Copy() ManifestHandler {
	return ManifestFromData(mlh.Descriptor.MediaType, mlh.rawManifest)
}

// GetSourceRefs returns a slice containing all child source references in a manifest list.  To get a single
// architecture ref, use GetPlatformSourceRef.
func (mlh *ManifestListHandler) GetSourceRefs(ctx context.Context, parentRef ref.SourceRef) []ref.SourceRef {
	outRefs := make([]ref.SourceRef, 0, len(mlh.ManifestList.Manifests))
	for _, d := range mlh.ManifestList.Manifests {
		newRef := parentRef.GetRef()
		newRef.Digest = d.Digest.String()
		newDecl := parentRef.GetDecl()
		if parentRef.GetRef().Digest != "" {
			newDecl = strings.Replace(newDecl, parentRef.GetRef().Digest, newRef.Digest, 1)
		}
		outRefs = append(outRefs, ref.NewDeclaredRef(newRef, newDecl))
	}
	return outRefs
}

// HasPlatformContext is a function to determine if a the context has any of the platform values set
// TODO I think this function is a bad idea.
func HasPlatformContext(ctx context.Context) bool {
	if val := ctx.Value(CtxPlatformOs); val != nil {
		return true
	}
	if val := ctx.Value(CtxPlatformArch); val != nil {
		return true
	}
	if val := ctx.Value(CtxPlatformVariant); val != nil {
		return true
	}
	if val := ctx.Value(CtxPlatform); val != nil {
		return true
	}
	return false
}

// GetPlatformSourceRef creates a new source ref from a parent source ref, redirecting the source from the original
// to a platform-appropriate manifest, based on platform values in the supplied context.
func (mlh *ManifestListHandler) GetPlatformSourceRef(ctx context.Context, parentRef ref.SourceRef) (ref.SourceRef, error) {
	var platform, os, arch, variant string
	iOS := ctx.Value(CtxPlatformOs)
	if iOS != nil {
		os = iOS.(string)
	}
	iArch := ctx.Value(CtxPlatformArch)
	if iArch != nil {
		arch = iArch.(string)
	}
	iVariant := ctx.Value(CtxPlatformVariant)
	if iVariant != nil {
		variant = iVariant.(string)
	}
	iPlatform := ctx.Value(CtxPlatform)
	if iPlatform != nil {
		platform = iPlatform.(string)
		platParts := strings.Split(platform, "/")
		if len(platParts) == 2 {
			os = platParts[0]
			arch = platParts[1]
		}
	}
	for _, d := range mlh.ManifestList.Manifests {
		if d.Platform == nil {
			continue
		}
		if os != "" && d.Platform.OS != os {
			continue
		}
		if arch != "" && d.Platform.Architecture != arch {
			continue
		}
		if variant != "" && d.Platform.Variant != variant {
			continue
		}
		newRef := parentRef.GetRef()
		newRef.Digest = d.Digest.String()
		newDecl := parentRef.GetDecl()
		if parentRef.GetRef().Digest != "" {
			newDecl = strings.Replace(newDecl, parentRef.GetRef().Digest, newRef.Digest, 1)
		}
		return ref.NewDeclaredRef(newRef, newDecl), nil
	}
	return ref.DeclRef{}, fmt.Errorf("no matching platform in manifest list")
}
