// Package sign provides an interface abstraction and implementation for creating/verifying signatures of OCI or docker manifest digests.
package sign

import (
	"context"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// SigsHandler wraps the act3oci ManifestHandler with additional methods for handling signatures.
type SigsHandler interface {
	// Sign signs a manifest digest. unsignedAnnos allows unsigned annotations to be included with a signature layer
	// (such as userid/verify api), while signedAnnos are additional metadata to be included in the signed payload,
	// such as attestation data.
	Sign(ctx context.Context, pkProvider PrivateKeyProvider, unsignedAnnos, signedAnnos map[string]string) error

	// Verify verifies ALL existing signatures, using optional locally provided keys.  Returns true if all signatures
	// verify or pass integrity (when the only public key is untrusted).  Additional details can be returned via errors
	// or within the interface implementation
	Verify(context.Context) (bool, error)

	// Signatures returns all signature layers.
	Signatures() []Signature

	// VerifiedSignatures returns a list of signature layer digests that passed verification during the last Verify
	VerifiedSignatures() []digest.Digest

	// FailedSignatures returns a list of signature layer digests that failed verification during the last Verify
	FailedSignatures() []digest.Digest

	// SignedSubject returns the manifest digest to be signed/verified.
	SignedSubject() ocispec.Descriptor

	// WriteDisk writes the config, manifest, and a signature layer to disk.
	WriteDisk(digest.Digest) error
}
