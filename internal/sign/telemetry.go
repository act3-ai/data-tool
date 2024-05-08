package sign

import (
	"context"
	"fmt"

	"gitlab.com/act3-ai/asce/data/telemetry/pkg/types"
	"gitlab.com/act3-ai/asce/data/tool/internal/bottle"
)

// NewSummaryFromBottle generates a telemetry compatible signature summary, which includes the bottle and manifest
// digests, plus a list of signature detail records.  Each signature contains the certificate, certificate thumbprint,
// signature data, and custom annotations.
func NewSummaryFromBottle(ctx context.Context, btl *bottle.Bottle) (*types.SignaturesSummary, error) {
	sigPath := bottle.SigDir(btl.GetPath())
	bottleManifestDescriptor := btl.Manifest.GetManifestDescriptor()
	sigHandler, err := LoadLocalSignatures(ctx, bottleManifestDescriptor, sigPath)
	if err != nil {
		return nil, fmt.Errorf("loading local signatures %w", err)
	}
	sigs := sigHandler.Signatures()
	if len(sigs) == 0 {
		return nil, nil
	}

	summary := &types.SignaturesSummary{
		SubjectManifest: btl.Manifest.GetManifestDescriptor().Digest,
		SubjectBottleID: btl.GetBottleID(),
		Signatures:      make([]types.SignatureDetail, len(sigs)),
	}
	for i, sig := range sigHandler.Signatures() {
		summary.Signatures[i].SignatureType = "application/vnd.cncf.notary.payload.v1+json"
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
