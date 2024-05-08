package bottle

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"

	"git.act3-ace.com/ace/data/schema/pkg/selectors"
	"git.act3-ace.com/ace/data/schema/pkg/util"
	"git.act3-ace.com/ace/data/tool/internal/storage"
	"git.act3-ace.com/ace/go-common/pkg/logger"
)

// PartSelectorOptions are options for creating a PartSelector.
type PartSelectorOptions struct {
	Empty     bool     // Only pull metadata from the bottle
	Selectors []string // Label selectors to apply
	Parts     []string // Names of parts to retrieve
	Artifacts []string // Filter by public artifact type
}

// New creates a new PartSelectorFunc from the options.
//
//nolint:sloglint
func (opts PartSelectorOptions) New(ctx context.Context) (PartSelectorFunc, error) {
	switch {
	case opts.Empty:
		return func(storage.PartInfo) bool { return false }, nil
	case len(opts.Selectors) == 0 && len(opts.Parts) == 0 && len(opts.Artifacts) == 0:
		// return nil should also work
		return func(storage.PartInfo) bool { return true }, nil
	default:
		sels, err := selectors.Parse(opts.Selectors)
		if err != nil {
			return nil, err
		}
		return func(part storage.PartInfo) bool {
			log := logger.V(logger.FromContext(ctx).With("part", part.GetName(), "labels", part.GetLabels()), 2)
			for _, partName := range opts.Parts {
				if part.GetName() == partName {
					log.Info("Selecting part due to the part name")
					return true
				}
			}

			for _, artifactPath := range opts.Artifacts {
				if util.IsPathPrefix(artifactPath, part.GetName()) {
					log.Info("Selecting part because it contains the artifact", "artifact", artifactPath)
					return true
				}
			}

			if sels.Matches(labels.Set(part.GetLabels())) {
				log.Info("Selecting part because it matches a selector")
				return true
			}

			log.Info("Did not select part")
			return false
		}, nil
	}
}
