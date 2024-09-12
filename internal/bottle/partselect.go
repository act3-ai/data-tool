package bottle

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"

	"git.act3-ace.com/ace/data/schema/pkg/selectors"
	"git.act3-ace.com/ace/data/schema/pkg/util"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/storage"
)

// PartSelectorOptions are options for creating a PartSelector.
type PartSelectorOptions struct {
	Empty     bool     // Only pull metadata from the bottle
	Labels    []string // Label selectors to apply
	Names     []string // Names of parts to retrieve
	Artifacts []string // Filter by public artifact type
}

// New creates a new PartSelectorFunc from the options.
//
//nolint:sloglint
func (opts *PartSelectorOptions) New(ctx context.Context) (PartSelectorFunc, error) {
	switch {
	case opts.Empty:
		return func(PartInfo) bool { return false }, nil
	case len(opts.Labels) == 0 && len(opts.Names) == 0 && len(opts.Artifacts) == 0:
		// return nil should also work
		return func(PartInfo) bool { return true }, nil
	default:
		// we only select all parts if all selectors are nil or empty, since
		// selectors.Parse will select all if nil we alter this behavior.
		if opts.Labels == nil {
			opts.Labels = make([]string, 0)
		}
		sels, err := selectors.Parse(opts.Labels)
		if err != nil {
			return nil, err
		}
		return func(part PartInfo) bool {
			log := logger.V(logger.FromContext(ctx).With("part", part.GetName(), "labels", part.GetLabels()), 2)
			if opts.Names != nil {
				for _, partName := range opts.Names {
					if part.GetName() == partName {
						log.Info("Selecting part due to the part name")
						return true
					}
				}
			}

			if opts.Artifacts != nil {
				for _, artifactPath := range opts.Artifacts {
					if util.IsPathPrefix(artifactPath, part.GetName()) {
						log.Info("Selecting part because it contains the artifact", "artifact", artifactPath)
						return true
					}
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
