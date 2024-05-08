// Package docs embeds relevant documentation to be surfaced in the ace-dt CLI.
package docs

import (
	"embed"
	"io/fs"

	"github.com/spf13/cobra"

	"git.act3-ace.com/ace/go-common/pkg/cmd"
	"git.act3-ace.com/ace/go-common/pkg/config"
	"git.act3-ace.com/ace/go-common/pkg/embedutil"
)

// GeneralDocs contains general documentation
//
//go:embed get-started/quick-start-guide.md
//go:embed usage/user-guide.md
//go:embed usage/concepts/bottle-creator.md
//go:embed support/faq.md
var GeneralDocs embed.FS

// BottleDocs contains bottle-related docs
//
//go:embed usage/concepts/bottle-anatomy.md
//go:embed usage/concepts/labels-selectors.md
var BottleDocs embed.FS

// MirrorDocs contains mirror-related docs
//
//go:embed usage/tutorials/mirror.md
//go:embed usage/concepts/mirror.md
var MirrorDocs embed.FS

// ConfigDocs contains config-related docs
//
//go:embed apis/config.dt.act3-ace.io/v1alpha1.md
//go:embed get-started/configuration-guide.md
var ConfigDocs embed.FS

// schemas contains JSON Schema definitions of the API schemas
//
//go:embed apis/schemas/*.schema.json
var schemas embed.FS

// Schemas returns the embedded schemas for ace-dt.
func Schemas() fs.FS {
	filesys, err := fs.Sub(schemas, "apis/schemas")
	if err != nil {
		panic(err)
	}

	return filesys
}

// Embedded loads and categorizes the embedded documentation for use in the ace-dt CLI.
func Embedded(root *cobra.Command) *embedutil.Documentation {
	return &embedutil.Documentation{
		Title:   "ACE Data Tool",
		Command: root,
		Categories: []*embedutil.Category{
			embedutil.NewCategory(
				"docs", "General Documentation", root.Name(), 1,
				embedutil.LoadMarkdown("quick-start-guide", "Quick Start Guide", "get-started/quick-start-guide.md", GeneralDocs),
				embedutil.LoadMarkdown("user-guide", "User Guide", "usage/user-guide.md", GeneralDocs),
				embedutil.LoadMarkdown("bottle-creator-guide", "Bottle Creator Guide", "usage/concepts/bottle-creator.md", GeneralDocs),
				embedutil.LoadMarkdown("faq", "FAQ", "support/faq.md", GeneralDocs),
			),
			embedutil.NewCategory(
				"bottle", "Bottle Documentation", root.Name()+"-bottle", 1,
				embedutil.LoadMarkdown("anatomy", "Anatomy of a Bottle", "usage/concepts/bottle-anatomy.md", BottleDocs),
				embedutil.LoadMarkdown("labels", "Bottle Labels and Selectors", "usage/concepts/labels-selectors.md", BottleDocs),
			),
			embedutil.NewCategory(
				"mirror", "Mirror Documentation", root.Name()+"-mirror", 1,
				embedutil.LoadMarkdown("mirror-tutorial", "Mirror Tutorial", "usage/tutorials/mirror.md", MirrorDocs),
				embedutil.LoadMarkdown("mirror-usage", "Mirror Usage", "usage/concepts/mirror.md", MirrorDocs),
			),
			embedutil.NewCategory(
				"config", "Configuration Documentation", root.Name(), 5,
				embedutil.LoadMarkdown("config-v1alpha1", "Configuration API Documentation", "apis/config.dt.act3-ace.io/v1alpha1.md", ConfigDocs),
				embedutil.LoadMarkdown("registry-config", "Registry Configuration Options", "get-started/configuration-guide.md", ConfigDocs),
			),
		},
	}
}

// SchemaAssociations associates JSON Schema definitions with the config files they validate.
var SchemaAssociations = []cmd.SchemaAssociation{
	{
		Definition: "data.act3-ace.io.schema.json",
		FileMatch:  []string{"entry.json", "entry.yaml"},
	},
	{
		Definition: "config.dt.act3-ace.io.schema.json",
		FileMatch:  config.DefaultConfigValidatePath("ace", "dt", "config.yaml"),
	},
}
