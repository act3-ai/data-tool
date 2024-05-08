package sign

import (
	"bytes"
	"context"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation-core-go/signature/cose"
	"github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go"
	notationreg "github.com/notaryproject/notation-go/registry"
	"github.com/notaryproject/notation-go/signer"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"gitlab.com/act3-ai/asce/data/tool/internal/oci"
	"gitlab.com/act3-ai/asce/data/tool/internal/ui"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// NotarySignatures represents a set of notary based manifest signatures, implementing SigsHandler.
type NotarySignatures struct {
	// Subject is the descriptor of the manifest being signed (becomes subject in signature manifest)
	Subject ocispec.Descriptor

	// HashFunc is the hash function used for calculating digests.
	HashFunc crypto.Hash

	// SigManifest is the signature manifest handler.
	SigManifests []SigsManifestHandler

	// Certificate chain for signing
	CertChain []*x509.Certificate

	// LocalPath is the local path to the signature image directory.
	LocalPath string

	// SigVerified is a list of signature layer digests that were successfully verified on the last call to Verify.
	SigVerified []digest.Digest

	// SigFailed is a list of signature layer digests that failed to verify on the last call to Verify.
	SigFailed []digest.Digest
}

// Sign signs a manifest digest along with annotations using the private key provided.
// Sign implements the SigsHandler interface.
// The LoadLocalSignatures function should be used prior to calling this method.
func (notarySigs *NotarySignatures) Sign(ctx context.Context, pkProvider PrivateKeyProvider, unsignedAnnos map[string]string, signedAnnos map[string]string) error {
	log := logger.FromContext(ctx)

	if unsignedAnnos == nil {
		unsignedAnnos = make(map[string]string)
	}
	if signedAnnos == nil {
		signedAnnos = make(map[string]string)
	}

	privKey, err := pkProvider.PrivateKey()
	if err != nil {
		return fmt.Errorf("failed to obtain private key for signing: %w", err)
	}
	if notarySigs.CertChain == nil {
		cert, err := pkProvider.Certificate()
		switch {
		case err != nil:
			return fmt.Errorf("retrieving certificate for notary signing: %w", err)
		case cert == nil:
			return fmt.Errorf("unable to locate certificate for notary signing")
		default:
			notarySigs.CertChain = []*x509.Certificate{cert}
		}
	}

	notarySigner, err := signer.New(privKey, notarySigs.CertChain)
	if err != nil {
		return fmt.Errorf("failed to create notary signer: %w", err)
	}

	sigOpts := notation.SignerSignOptions{
		SignatureMediaType: jws.MediaTypeEnvelope, // we always sign with jws, but support verification for cose as well
		ExpiryDuration:     0,
		PluginConfig:       nil,
		SigningAgent:       "ace-dt sign agent",
	}

	// build the signature payload, while adding all other "optional" annotations. These annotations are signed.
	log.InfoContext(ctx, "Constructing signature payload", "signedDigest", notarySigs.SignedSubject().Digest, "signedAnnotationns", signedAnnos)

	subj := notarySigs.SignedSubject()
	subj.Annotations = signedAnnos

	sigEnvelope, signerInfo, err := notarySigner.Sign(ctx, subj, sigOpts)
	if err != nil {
		return fmt.Errorf("failed to sign payload: %w", err)
	}

	sigManifest, err := notarySigs.MakeSignatureManifest(ctx, signerInfo, jws.MediaTypeEnvelope, sigEnvelope, unsignedAnnos)
	if err != nil {
		return fmt.Errorf("creating signature manifest: %w", err)
	}

	// add sigManifest to SigManifests
	notarySigs.SigManifests = append(notarySigs.SigManifests, sigManifest)

	// write the signature layer, and manifest.
	log.InfoContext(ctx, "Writing signature and updating sig manifest")
	err = notarySigs.WriteDisk(notarySigs.SignedSubject().Digest)
	if err != nil {
		return err
	}

	// signed successfully.
	return nil
}

// Verify verifies all loaded signatures, returning a boolean indicating
// if all signatures passed verification. More specific information regarding
// failed and verified signatures can be retrieved with the appropriate methods.
// TODO: Currently insecure as we verify the certificate chain against its
// own root certificate.
// Verify implements the SigsHandler interface.
// The LoadLocalSignatures function should be used prior to calling this method.
func (notarySigs *NotarySignatures) Verify(ctx context.Context) (bool, error) {
	log := logger.FromContext(ctx)
	rootUI := ui.FromContextOrNoop(ctx)

	// get the signatures.
	sigList := notarySigs.Signatures()

	if len(sigList) < 1 {
		return false, signature.SignatureNotFoundError{}
	}
	rootUI.Infof("Found %d signatures", len(sigList))

	// Reset signature verification statistics
	var failedSigs []error                     // collects failed verifications and the reason for failure.
	notarySigs.SigFailed = []digest.Digest{}   // collects the layer digests of failed signatures.
	notarySigs.SigVerified = []digest.Digest{} // collects the layer digests of verified signatures.

	// all signatures must pass verification.
	// TODO: Do we want to verify all? Only the ones we trust?
SigIterator:
	for _, sig := range sigList {

		kr, err := sig.GetKeyRetrieverForPayload()
		if err != nil {
			failedSigs = append(failedSigs, fmt.Errorf("getting cert retriever: %w", err))
			continue SigIterator
		}

		verifier, err := kr.RetrieveVerifier(notarySigs.HashFunc)
		if err != nil {
			notarySigs.SigFailed = append(notarySigs.SigFailed, sig.GetDescriptor().Digest)
			failedSigs = append(failedSigs, fmt.Errorf("retrieving verifier: %w", err))
			continue SigIterator
		}

		payload, err := sig.GetPayload()
		if err != nil {
			notarySigs.SigFailed = append(notarySigs.SigFailed, sig.GetDescriptor().Digest)
			failedSigs = append(failedSigs, fmt.Errorf("getting signature payload: %w", err))
			continue SigIterator
		}

		// verify this signature.
		log.InfoContext(ctx, "Verifying signature")
		err = verifier.VerifySignature(notarySigs.SignedSubject(), bytes.NewReader(payload))
		if err != nil {
			notarySigs.SigFailed = append(notarySigs.SigFailed, sig.GetDescriptor().Digest)
			failedSigs = append(failedSigs, fmt.Errorf("verifying signature, sig = %v: %w", sig, err))
			continue SigIterator
		}

		// verification passed
		notarySigs.SigVerified = append(notarySigs.SigVerified, sig.GetDescriptor().Digest)
	}

	// verification failed.
	// TODO: We we want all sigs to pass verification? What if it's signed by persons A, B, and C;
	// but we only trust A and B? Thus if C fails verification we don't really care as it is
	// approved by who we wanted it to be approved by.
	if len(failedSigs) > 0 {
		return false, errors.Join(failedSigs...)
	}

	// verification successful.
	return true, nil
}

// VerifiedSignatures returns a list of signature fingerprints that passed verification during the last Verify.
func (notarySigs *NotarySignatures) VerifiedSignatures() []digest.Digest {
	return notarySigs.SigVerified
}

// FailedSignatures returns a list of signature fingerprints that failed verification during the last Verify.
func (notarySigs *NotarySignatures) FailedSignatures() []digest.Digest {
	return notarySigs.SigFailed
}

// SignedSubject returns the manifest digest to be signed/verified.
// SignedSubject implements the SigsHandler interface.
func (notarySigs *NotarySignatures) SignedSubject() ocispec.Descriptor {
	return notarySigs.Subject
}

// Signatures returns a slice of all signature layers.
func (notarySigs *NotarySignatures) Signatures() []Signature {
	sigLayers := make([]Signature, 0)
	for _, sig := range notarySigs.SigManifests {
		for _, sigDescriptor := range sig.GetLayerDescriptors() {
			// Notary stores per-signature metadata at the signature manifest level, so we need to broadcast that
			// data back to the layer through the layer descriptor.
			if sigDescriptor.Annotations == nil {
				sigDescriptor.Annotations = sig.GetAnnotations()
			} else {
				for k, v := range sig.GetAnnotations() {
					sigDescriptor.Annotations[k] = v
				}
			}
			sigLayers = append(sigLayers, NewNotarySigLayer(sigDescriptor, notarySigs.LocalPath, sig.GetRawLayers()[sigDescriptor.Digest]))
		}
	}
	return sigLayers
}

// WriteDisk writes the config, manifest, and a signature layer to disk.
// Notary signature manifests are named based on their own digest.
func (notarySigs *NotarySignatures) WriteDisk(signedDigest digest.Digest) error {
	// Write the matching manifest to disk
	for _, sigManifest := range notarySigs.SigManifests {
		subject := sigManifest.GetManifestData().Subject
		sigManifestTag := signedDigest.Algorithm().String() + "-" + sigManifest.GetManifestDescriptor().Digest.Hex() + ".sig"
		if subject != nil && subject.Digest == signedDigest {
			// Notary style signature manifests do not need a config file
			err := sigManifest.WriteDisk(notarySigs.LocalPath, sigManifestTag, "")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// MakeSignatureManifest builds a signature manifest adhering to the notary signature spec.
// See https://github.com/notaryproject/specifications/blob/main/specs/signature-specification.md#oci-signatures.
func (notarySigs *NotarySignatures) MakeSignatureManifest(ctx context.Context,
	signerInfo *signature.SignerInfo,
	mediatype string, envelope []byte,
	unsignedAnnos map[string]string,
) (SigsManifestHandler, error) {
	log := logger.FromContext(ctx)

	// sanity check - only jws and cose are supported
	if mediatype != jws.MediaTypeEnvelope && mediatype != cose.MediaTypeEnvelope {
		return nil, &signature.UnsupportedSignatureFormatError{MediaType: mediatype}
	}

	// prepare signature layer
	layerDesc := ocispec.Descriptor{
		MediaType:   mediatype,
		Digest:      digest.FromBytes(envelope),
		Size:        int64(len(envelope)),
		Annotations: make(map[string]string),
	}
	log.InfoContext(ctx, "created signature layer", "mediatype", layerDesc.MediaType, "digest", layerDesc.Digest)

	// notation configuration is an empty config object, which is standard in oci spec
	configDesc := ocispec.Descriptor{
		MediaType: notationreg.ArtifactTypeNotation,
		Digest:    ocispec.DescriptorEmptyJSON.Digest,
		Size:      ocispec.DescriptorEmptyJSON.Size,
	}

	// prepare manifest annotations
	prints, err := generateThumbprints(signerInfo)
	if err != nil {
		return nil, fmt.Errorf("generating thumbprints: %w", err)
	}
	unsignedAnnos[AnnotationX509ChainThumbprint] = prints

	// build the manifest
	subj := notarySigs.SignedSubject()
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2, // historical value. does not pertain to OCI or docker version.
		},
		MediaType:   ocispec.MediaTypeImageManifest,
		Config:      configDesc,
		Subject:     &subj,
		Layers:      []ocispec.Descriptor{layerDesc}, // notation manifests always have one signature layer.
		Annotations: unsignedAnnos,
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal notary sig manifest to bytes: %w", err)
	}

	sigManifest := &SigsManifest{
		Descriptor: ocispec.Descriptor{
			MediaType: ocispec.MediaTypeImageManifest,
			Digest:    digest.FromBytes(manifestBytes),
			Size:      int64(len(manifestBytes)),
		},
		Manifest:           manifest,
		Config:             ocispec.DescriptorEmptyJSON.Data, // notations manifest configs are the empty config
		rawLayers:          make(map[digest.Digest][]byte),
		ManifestStatusInfo: oci.ManifestStatusInfo{Status: oci.ManifestOK},
	}

	// validate manifest data for sanity check
	if err := ValidateSigManifest(manifest); err != nil {
		return nil, fmt.Errorf("validating signature image manifest: %w", err)
	}
	log.InfoContext(ctx, "created signature manifest", "digest", sigManifest.GetManifestDescriptor().Digest, "signedSubjectDigest", notarySigs.SignedSubject().Digest)

	// prepare layers for writing
	layerMap := sigManifest.GetRawLayers()
	layerMap[layerDesc.Digest] = envelope
	sigManifest.SetRawLayers(layerMap)

	return sigManifest, nil
}

// generateThumbprints creates and returns the notation manifest annotation
// "io.cncf.notary.x509chain.thumbprint#S256". It is a list of SHA-256 fingerprints
// of signing certificate and certificate chain (including root) used for signature generation.
// See https://github.com/notaryproject/specifications/blob/main/specs/signature-specification.md#oci-signatures.
func generateThumbprints(signerInfo *signature.SignerInfo) (string, error) {
	// modified from source: https://github.com/notaryproject/notation-go/blob/main/notation.go#L492
	// sanity check
	if signerInfo == nil {
		return "", errors.New("failed to generate annotations: signerInfo cannot be nil")
	}
	thumbprints := make([]string, 0, len(signerInfo.CertificateChain))
	for _, cert := range signerInfo.CertificateChain {
		checkSum := sha256.Sum256(cert.Raw)
		thumbprints = append(thumbprints, hex.EncodeToString(checkSum[:]))
	}
	val, err := json.Marshal(thumbprints)
	if err != nil {
		return "", err
	}
	return string(val), nil
}
