package bottle

import (
	"context"
	"fmt"
	"io"

	"gitlab.com/act3-ai/asce/data/tool/internal/actions/internal/format"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
)

// AuthorList represents the bottle author list action.
type AuthorList struct {
	*Action
}

// Run runs the bottle author list action.
func (action *AuthorList) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	authors := btl.Definition.Authors
	if len(authors) == 0 {
		log.InfoContext(ctx, "bottle has no authors to show", "path", action.Dir)
		return nil
	}

	t := format.NewTable()
	t.AddRow("NAME", "EMAIL", "URL")

	for _, author := range btl.Definition.Authors {
		t.AddRow(author.Name, author.Email, author.URL)
	}

	_, err = fmt.Fprintln(out, t.String())
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "author list command completed")
	return nil
}
