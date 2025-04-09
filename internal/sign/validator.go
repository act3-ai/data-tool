package sign

import (
	"errors"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-core-go/signature/jws"
	notationreg "github.com/notaryproject/notation-go/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	act3valid "github.com/act3-ai/bottle-schema/pkg/validation"
)

// * * NOTICE * * //
// The following is nearly identical to github.com/act3-ai/bottle-schema/pkg/validation/manifest.go.
// It contains minor mondifications to suit the media types used in signature
// layers and the config.

// ValidateSigManifest validates a signature manifest for correctness.
func ValidateSigManifest(m ocispec.Manifest) error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.SchemaVersion, validation.Required, validation.In(2)),
		validation.Field(&m.MediaType, validation.Required, validation.In(ocispec.MediaTypeImageManifest)),
		validation.Field(&m.Config, configDescriptor),
		validation.Field(&m.Layers, validation.Each(sigLayerDescriptor)),
	)
	if err != nil {
		return fmt.Errorf("validating signature manifest: %w", err)
	}

	return nil
}

// config validation.
var configDescriptor = validation.By(func(value any) error {
	return validateSigConfigDescriptor(value.(ocispec.Descriptor))
})

func validateSigConfigDescriptor(config ocispec.Descriptor) error {
	err := validation.ValidateStruct(&config,
		validation.Field(&config.MediaType, sigConfigMediaType),
		validation.Field(&config.Digest, validation.Required, act3valid.IsDigest),
		validation.Field(&config.Size, validation.Min(0)),
	)
	if err != nil {
		return fmt.Errorf("validating signature configuration: %w", err)
	}

	return nil
}

var sigConfigMediaType = validation.By(func(value any) error {
	if !IsSigConfig(value.(string)) {
		return errors.New("invalid sig config media type")
	}
	return nil
})

// IsSigConfig verifies the media type of a notary signature config.
func IsSigConfig(mediaType string) bool {
	return mediaType == notationreg.ArtifactTypeNotation
}

// layer validation.
var sigLayerDescriptor = validation.By(func(value any) error {
	return validateSigLayerDescriptor(value.(ocispec.Descriptor))
})

func validateSigLayerDescriptor(d ocispec.Descriptor) error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.MediaType, validation.Required, sigLayerMediaType),
		validation.Field(&d.Digest, validation.Required, act3valid.IsDigest),
		validation.Field(&d.Size, validation.Min(0)),
		validation.Field(&d.Platform, validation.Empty), // necessary? sigs don't use platform fields.
	)
	if err != nil {
		return fmt.Errorf("validating signature layer descriptor: %w", err)
	}

	return nil
}

// SignatureMediaType validates accepted notary signature layer mediatypes.
var sigLayerMediaType = validation.By(func(value any) error {
	if value != jws.MediaTypeEnvelope {
		return fmt.Errorf("validating signature layer media type: %w", errors.New((&signature.UnsupportedSignatureFormatError{MediaType: value.(string)}).Error()))
	}
	return nil
})
