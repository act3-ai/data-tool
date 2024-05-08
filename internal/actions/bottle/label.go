package bottle

import (
	"context"
	"fmt"
	"io"
	"strings"

	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// Label represents the bottle label action.
type Label struct {
	*Action
}

// Run runs the bottle label action.
func (action *Label) Run(ctx context.Context, labels []string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "label command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	// check if user wants to remove a label
	for _, lbl := range labels {
		if isLabelDelete(lbl) {
			key := strings.TrimSuffix(lbl, "-")
			if err := btl.RemoveLabel(key); err != nil {
				return fmt.Errorf("could not remove label %s in bottle at %s: %w", key, btl.GetPath(), err)
			}
		} else {
			key, value, err := parseLabel(lbl)
			if err != nil {
				return fmt.Errorf("could not add label %s to bottle at %s: %w", lbl, btl.GetPath(), err)
			}
			btl.AddLabel(key, value)
		}
	}

	log.InfoContext(ctx, "label command completed")
	return saveMetaChanges(ctx, btl)
}

// parseLabel parses a key=value for a label and validates that the result is a valid label and label value.
func parseLabel(arg string) (string, string, error) {
	// find index of the first "="
	idx := strings.IndexAny(arg, "=")
	if idx == -1 {
		return "", "", fmt.Errorf("invalid format \"%s\", expected an\"=\"", arg)
	}

	// omit the "=" sign
	key, val := arg[:idx], arg[idx+1:]

	if err := v1validation.ValidateLabels(map[string]string{key: val}, nil).ToAggregate(); err != nil {
		return "", "", err
	}

	return key, val, nil
}
