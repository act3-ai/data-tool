package bottle

import (
	"context"
	"fmt"
	"io"

	"git.act3-ace.com/ace/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/data/tool/internal/actions/internal/format"
)

// SourceList represents the bottle source list action.
type SourceList struct {
	*Action
}

// Run runs the bottle source list action.
func (action *SourceList) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	sources := btl.Definition.Sources
	if len(sources) == 0 {
		log.InfoContext(ctx, "bottle has no sources to show", "path", btl.GetPath())
		return nil
	}

	// Print header and authors
	t := format.NewTable()
	t.AddRow("NAME", "URI")

	for _, src := range btl.Definition.Sources {
		t.AddRow(format.TitleCase(src.Name), src.URI)
	}

	_, err = fmt.Fprintln(out, t.String())
	if err != nil {
		return err
	}

	return nil
}
