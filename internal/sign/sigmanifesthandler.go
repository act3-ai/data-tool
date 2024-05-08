package sign

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"gitlab.com/act3-ai/asce/data/tool/internal/oci"
)

// AnnotationX509ChainThumbprint stores a certificate chain as a list of thumbprints. A manifest annotation key.
// Note: Notation keeps this internal at "github.com/notaryproject/notation-go/internal/envelope", which
// is odd as it's a required property of a notation signature.
const AnnotationX509ChainThumbprint = "io.cncf.notary.x509chain.thumbprint#S256"

// SigsManifestHandler wraps an act3oci.ManifestHandler with additional
// functions used for creating or updating a signature image manifest.
type SigsManifestHandler interface {
	oci.ManifestHandler
	// SetManifestData sets the OCI manifest data structure.
	SetManifestData(ocispec.Manifest)
	// UpdateManifestDescriptor updates the manifest descriptor to reflect current manifest data.
	UpdateManifestDescriptor() error
	// GetConfigData gets the OCI config data structure.
	GetConfigData() []byte
	// GetConfigRaw gets the json formatted slice of bytes representing the config data.
	GetConfigRaw() ([]byte, error)
	// SetConfigData sets the OCI config data structure.
	SetConfigData([]byte)
	// SetLayerDescriptors sets the layer descriptors for the manifest and DiffIDs for the config.
	SetLayerDescriptors([]ocispec.Descriptor) error
	// GetAnnotations retrieves the annotations collection from the manifest
	GetAnnotations() map[string]string
	// GetRawLayers returns the raw layers.
	GetRawLayers() map[digest.Digest][]byte
	// SetRawLayers sets the raw layers.
	SetRawLayers(map[digest.Digest][]byte)
	// WriteDisk writes the manifest, config, and layers to disk.
	WriteDisk(string, string, string) error
}

// SigsManifest is an alias for ocispec.Manifest which implements the manifest handler interface. It also contains
// information about the manifest descriptor itself (notably, size, Digest, etc), and a status structure used during
// transfer. This structure implements the ManifestHandler interface.
//
// Note: This struct is derived from the BottleManifest struct in act3oci.manifesthandler.go, but introduces setters
// and a config with methods for handling the config. Setters introduced are: SetManifestData, SetManifestRaw,
// SetConfigData, SetConfigRaw, GetConfigRaw, and AddLayerDescriptor.
type SigsManifest struct {
	Descriptor ocispec.Descriptor
	Manifest   ocispec.Manifest
	Config     []byte
	rawLayers  map[digest.Digest][]byte
	oci.ManifestStatusInfo
}

// WriteDisk writes the entire signature image data to disk. The signature
// directory is created if it does not already exist. The signature manifest
// is named locally and remotely by the manifest tag. The signature configuration
// is named similarly to the manifest with a "-Config.json" suffix rather than ".sig".
// WriteDisk implements the SigsManifestHandler interface.
func (s *SigsManifest) WriteDisk(dir, sigManifestTag, configFileName string) error {
	// ensure .signature dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// create the sig dir
		err = os.MkdirAll(dir, 0o777)
		if err != nil {
			return fmt.Errorf("creating signature directory: %w", err)
		}
	}

	// write layers if we don't already have them
	for layerDigest, data := range s.GetRawLayers() {
		sigPath := filepath.Join(dir, layerDigest.Hex())
		_, err := os.Stat(sigPath)
		if os.IsNotExist(err) {
			err = os.WriteFile(sigPath, data, 0o644)
			if err != nil {
				return fmt.Errorf("writing signature layer to disk, digest = %s: %w", layerDigest, err)
			}
		} else if err != nil {
			return fmt.Errorf("writing sig layer to disk: %w", err)
		}
	}

	// Notary style signatures use an empty config
	if configFileName != "" {
		// write config
		configData, err := s.GetConfigRaw()
		if err != nil {
			return fmt.Errorf("getting JSON encoded config data: %w", err)
		}
		err = os.WriteFile(filepath.Join(dir, configFileName), configData, 0o644)
		if err != nil {
			return fmt.Errorf("writing config data: %w", err)
		}
	}

	// write manifest
	manifestData, err := s.GetManifestRaw()
	if err != nil {
		return fmt.Errorf("getting JSON encoded manifest data: %w", err)
	}
	err = os.WriteFile(filepath.Join(dir, sigManifestTag), manifestData, 0o644)
	if err != nil {
		return fmt.Errorf("writing manifest data: %w", err)
	}

	return nil
}

// SetManifestData sets the manifest data structure.
// SetManifestDate implements the SigsManifestHandler interface.
func (s *SigsManifest) SetManifestData(manifest ocispec.Manifest) {
	s.Manifest = manifest
}

// GetConfigData returns the full manifest data object.
// GetConfigData implements the SigsManifestHandler interface.
func (s *SigsManifest) GetConfigData() []byte {
	return s.Config
}

// SetConfigData sets the config data structure.
// SetConfigData implements the SigsManifestHandler interface.
func (s *SigsManifest) SetConfigData(rawData []byte) {
	// Was config ocispec.Image
	s.Config = rawData
}

// GetConfigRaw returns the original raw data for the manifest.
// GetConfigRaw implements the SigsManifestHandler interface.
func (s *SigsManifest) GetConfigRaw() ([]byte, error) {
	return s.Config, nil
}

// SetLayerDescriptors sets layer information in both the manifest and config.
// SetLayerDescriptors implements the SigsManifestHandler interface.
func (s *SigsManifest) SetLayerDescriptors(layers []ocispec.Descriptor) error {
	cfgImage := ocispec.Image{}
	err := json.Unmarshal(s.Config, &cfgImage)
	if err == nil {
		// Image type config (cosign style) so update config with layer digests
		diffIDs := make([]digest.Digest, 0, len(layers))
		for _, layer := range layers {
			diffIDs = append(diffIDs, layer.Digest)
		}
		cfgImage.RootFS.DiffIDs = diffIDs
		configBytes, err := json.Marshal(cfgImage)
		if err != nil {
			return err
		}

		s.Manifest.Config.Size = int64(len(configBytes))
		s.Manifest.Config.Digest = digest.FromBytes(configBytes)
	}

	// update manifest
	s.Manifest.Layers = layers
	err = s.UpdateManifestDescriptor()
	if err != nil {
		return err
	}

	return nil
}

// UpdateManifestDescriptor updates the manifest descriptor to reflect the current manifest data.
// UpdateManifestDescriptor implements the SigsManifestHandler interface.
func (s *SigsManifest) UpdateManifestDescriptor() error {
	manifestBytes, err := s.GetManifestRaw()
	if err != nil {
		return fmt.Errorf("getting manifest bytes for descriptor update: %w", err)
	}
	s.Descriptor = ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}

	return nil
}

// GetAnnotations retrieves the annotations collection from the manifest.
func (s *SigsManifest) GetAnnotations() map[string]string {
	return s.Manifest.Annotations
}

// GetRawLayers returns a map of digests to json formatted slice of bytes of the signature layers.
// GetRawLayers implements the SigsManifestHandler interface.
func (s *SigsManifest) GetRawLayers() map[digest.Digest][]byte {
	if s.rawLayers == nil {
		return make(map[digest.Digest][]byte)
	}
	return s.rawLayers
}

// SetRawLayers sets the map of digests to json formatted slice of bytes of the signature layers.
// SetRawLayers implements the SigsManifestHandler interface.
func (s *SigsManifest) SetRawLayers(layers map[digest.Digest][]byte) {
	s.rawLayers = layers
}

// GetManifestDescriptor returns the OCI descriptor for the manifest data.
// GetManifestDescriptor implements the ManifestHandler interface.
func (s *SigsManifest) GetManifestDescriptor() ocispec.Descriptor {
	return s.Descriptor
}

// GetManifestData returns an OCI manifest structure filled with manifest data.
// GetManifestData implements the ManifestHandler interface.
func (s *SigsManifest) GetManifestData() ocispec.Manifest {
	return s.Manifest
}

// GetManifestRaw returns a json formatted slice of bytes representing the manifest data.
// GetManifestRaw implements the ManifestHandler interface.
func (s *SigsManifest) GetManifestRaw() ([]byte, error) {
	manifestBytes, err := json.Marshal(s.Manifest)
	if err != nil {
		return nil, err
	}
	return manifestBytes, nil
}

// GetConfigDescriptor returns the OCI descriptor for the config record within the manifest data.
// GetConfigDescriptor implements the ManifestHandler interface.
func (s *SigsManifest) GetConfigDescriptor() ocispec.Descriptor {
	return s.Manifest.Config
}

// GetLayerDescriptors returns a list of OCI descriptors for each layer specified in a manifest.
// GetLayerDescriptors implements the ManifestHandler interface.
func (s *SigsManifest) GetLayerDescriptors() []ocispec.Descriptor {
	return s.Manifest.Layers
}

// GetContentDescriptors returns a collection of content descriptors, with the config blob first.
// GetContentDescriptors implements the ManifestHandler interface.
func (s *SigsManifest) GetContentDescriptors() []ocispec.Descriptor {
	descs := make([]ocispec.Descriptor, 0, len(s.Manifest.Layers)+1)
	descs = append(descs, s.Manifest.Config)
	descs = append(descs, s.Manifest.Layers...)

	return descs
}

// GetStatus returns information about the transfer of a manifest, including http and internal error information.
// GetStatus implements the ManifestHandler interface.
func (s *SigsManifest) GetStatus() oci.ManifestStatusInfo {
	return s.ManifestStatusInfo
}

// AddAnnotation adds a key value annotation to the manifest.
// AddAnnotations implements the ManifestHandler interface.
func (s *SigsManifest) AddAnnotation(key string, value string) {
	if s.Manifest.Annotations == nil {
		s.Manifest.Annotations = make(map[string]string)
	}
	s.Manifest.Annotations[key] = value

	_ = s.UpdateManifestDescriptor()
}

// GetAnnotation returns a annotation value based on a given key, or empty if the key is not found
// GetAnnotation implements the ManifestHandler interface.
func (s *SigsManifest) GetAnnotation(key string) string {
	if s.Manifest.Annotations == nil {
		return ""
	}
	if s, ok := s.Manifest.Annotations[key]; ok {
		return s
	}
	return ""
}

// Copy creates a fresh manifest handler based on the data contained in the current SigManifest. This allows a
// duplicate handler to be modified (such as adding annotations) without updating the originating handler.
//
// NOTE: This copy function returns a read-only handler. If there is an error marshaling the manifest
// a ManifestHandler will have an empty manifest.
func (s *SigsManifest) Copy() oci.ManifestHandler {
	manifestBytes, err := s.GetManifestRaw()
	if err != nil {
		manifestBytes = nil
	}
	return oci.ManifestFromData(s.Descriptor.MediaType, manifestBytes)
}
