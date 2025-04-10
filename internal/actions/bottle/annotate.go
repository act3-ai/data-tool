package bottle

import (
	"context"
	"fmt"
	"io"
	"strings"

	apivalidation "k8s.io/apimachinery/pkg/api/validation"

	"github.com/act3-ai/go-common/pkg/logger"
)

// Annotate represents the bottle annotate action.
type Annotate struct {
	*Action
}

// Run runs the bottle annotate action.
func (action *Annotate) Run(ctx context.Context, annotations []string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "annotation command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	// check if user wants to remove annotation
	for _, annotation := range annotations {
		if isLabelDelete(annotation) {
			key := strings.TrimSuffix(annotation, "-")
			if err := btl.RemoveAnnotation(key); err != nil {
				return fmt.Errorf("could not remove annotation of specified key %s: %w", key, err)
			}
		} else {
			key, value, err := parseAnnotation(annotation)
			if err != nil {
				return fmt.Errorf("could not add annotation %s to bottle at %s: %w", annotation, btl.GetPath(), err)
			}
			btl.AddAnnotation(key, value)
		}
	}

	log.InfoContext(ctx, "annotate command completed")
	return saveMetaChanges(ctx, btl)
}

// parseAnnotation parses a single arguments from arguments for annotations for a single Key value pair.
// It also validates the values to check if the annotation (Key-Value) is a valid kubernetes annotation value.
func parseAnnotation(arg string) (string, string, error) {
	// Since annotation value can have an equal sign (i.e. base64 encoded
	// item), we'll look for the first equal sign to delineate the Key

	// find index of the first "="
	idx := strings.IndexAny(arg, "=")
	if idx == -1 {
		return "", "", fmt.Errorf("invalid format \"%s\", expected an\"=\"", arg)
	}

	// omit the "=" sign
	key, val := arg[:idx], arg[idx+1:]

	// verify if item is valid annotation
	if err := apivalidation.ValidateAnnotations(map[string]string{key: val}, nil).ToAggregate(); err != nil {
		return "", "", err
	}

	return key, val, nil
}
