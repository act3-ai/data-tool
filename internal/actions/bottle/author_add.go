package bottle

import (
	"context"
	"io"

	"github.com/act3-ai/go-common/pkg/logger"

	latest "github.com/act3-ai/bottle-schema/pkg/apis/data.act3-ace.io/v1"
)

// AuthorAdd represents the bottle author add action.
type AuthorAdd struct {
	*Action
}

// Run runs the bottle author add action.
func (action *AuthorAdd) Run(ctx context.Context, authorName, authorEmail, authorURL string, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "author add command activated")

	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	// Create author info
	author := latest.Author{
		Name:  authorName,
		Email: authorEmail,
		URL:   authorURL,
	}

	// add author info
	err = btl.AddAuthorInfo(author)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "Saving bottle with added author", "name", author.Name, "email",
		author.Name, "url", author.URL)

	return saveMetaChanges(ctx, btl)
}
