package telemetry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"

	"gitlab.com/act3-ai/asce/data/telemetry/pkg/types"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
	sigcustom "gitlab.com/act3-ai/asce/data/tool/internal/sign"
)

// sendSignatures sends any existing signatures to telemetry for validation and storage.
func (a *Adapter) sendSignatures(ctx context.Context, summary *types.SignaturesSummary) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	sigSummaryJSON, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("error marshalling signature summary: %w", err)
	}

	return a.client.SendSignature(ctxTimeout, digest.Canonical, sigSummaryJSON, nil)
}

// newSummaryFromLocal generates a telemetry compatible signature summary, which includes the bottle and manifest
// digests, plus a list of signature detail records.  Each signature contains the certificate, certificate thumbprint,
// signature data, and custom annotations.
func newSummaryFromLocal(ctx context.Context, bottleDir string, manDesc ocispec.Descriptor) (*types.SignaturesSummary, error) {
	sigPath := bottle.SigDir(bottleDir)
	sigHandler, err := sigcustom.LoadLocalSignatures(ctx, manDesc, sigPath)
	if err != nil {
		return nil, fmt.Errorf("loading local signatures %w", err)
	}
	sigs := sigHandler.Signatures()
	if len(sigs) == 0 {
		return nil, nil
	}

	bottleID, err := bottle.ReadBottleIDFile(bottleDir)
	if err != nil {
		return nil, fmt.Errorf("reading bottleID file: %w", err)
	}

	summary := &types.SignaturesSummary{
		SubjectManifest: manDesc.Digest,
		SubjectBottleID: bottleID,
		Signatures:      make([]types.SignatureDetail, len(sigs)),
	}
	for i, sig := range sigHandler.Signatures() {
		summary.Signatures[i].SignatureType = types.NotarySignatureType // "application/vnd.cncf.notary.payload.v1+json"
		summary.Signatures[i].Descriptor = sig.GetDescriptor()
		annos, _ := sig.Annotations()
		// In notary signatures, the annotations are copied from the manifest to the layer descriptor, which would
		// cause duplicate information here, so we clear the descriptor annotations before creating the summary
		clear(summary.Signatures[i].Descriptor.Annotations)
		summary.Signatures[i].Annotations = annos
		sigdata, err := sig.GetPayloadBase64()
		if err != nil {
			return nil, fmt.Errorf("loading signature payload for telemetry %w", err)
		}
		summary.Signatures[i].Signature = sigdata
	}

	return summary, nil
}

// newSummaryFromRemote generates a telemetry compatible signature summary from remotely stored signatures, which
// includes the bottle and manifest digests, plus a list of signature detail records.  Each signature contains the
// certificate, certificate thumbprint, signature data, and custom annotations.
func newSummaryFromRemote(ctx context.Context, src content.ReadOnlyGraphStorage, subjectDesc ocispec.Descriptor, bottleID digest.Digest) (*types.SignaturesSummary, error) {
	sigManDescs, err := registry.Referrers(ctx, src, subjectDesc, types.NotarySignatureType) // currently only "application/vnd.cncf.notary.signature"" sigs are supported
	if err != nil {
		return nil, fmt.Errorf("resolving bottle signature referrers: %w", err)
	}

	if len(sigManDescs) == 0 {
		return nil, nil
	}

	summary := &types.SignaturesSummary{
		SubjectManifest: subjectDesc.Digest,
		SubjectBottleID: bottleID,
		Signatures:      make([]types.SignatureDetail, len(sigManDescs)),
	}
	for i, desc := range sigManDescs {
		summary.Signatures[i].SignatureType = desc.MediaType // currently only "application/vnd.cncf.notary.signature" sigs are supported
		summary.Signatures[i].Descriptor = desc
		annos := desc.Annotations
		// In notary signatures, the annotations are copied from the manifest to the layer descriptor, which would
		// cause duplicate information here, so we clear the descriptor annotations before creating the summary
		clear(summary.Signatures[i].Descriptor.Annotations)
		summary.Signatures[i].Annotations = annos

		sigPayload, err := content.FetchAll(ctx, src, desc)
		if err != nil {
			return nil, fmt.Errorf("fetching signature: %w", err)
		}

		summary.Signatures[i].Signature = base64.StdEncoding.EncodeToString(sigPayload)
	}

	return summary, nil
}
