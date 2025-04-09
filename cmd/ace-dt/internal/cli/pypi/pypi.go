// Package pypi declares pypi related commands, such as serving an OCI backed pypi registry, and performing a sync.
package pypi

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-tool/internal/actions"
	"github.com/act3-ai/data-tool/internal/actions/pypi"
)

// NewPypiCmd represents the base command for all Python package index (PyPI commands).
func NewPypiCmd(tool *actions.DataTool) *cobra.Command {
	action := &pypi.Action{DataTool: tool}
	cmd := &cobra.Command{
		GroupID: "core",
		Use:     "pypi",
		Short:   "Python package syncing operations",
		Long: `Python package index and OCI credentials are both retrieved from the docker credential store.
Use "ace-dt login --no-auth-check" to add your python package index credentials to the store.
Only the hostname (and port) is used for the credential lookup.  The full URL path to the index is not used.`,
		Example: `The first step is to fetch distribution files from remote sources.
$ ace-dt pypi to-oci reg.example.com/my/pypi numpy -l 'version.major=1,version.minor>5'

or with a requirements file
$ ace-dt pypi to-oci reg.example.com/my/pypi -r requirements.txt
		
After you have fetched you can serve up the PyPI compliant (PEP-691) package index with
$ ace-dt pypi serve reg.example.com/my/pypi
`,
	}

	cmd.PersistentFlags().BoolVar(&action.AllowYanked, "allow-yanked", false, "Do not ignore yanked distribution files")

	cmd.AddCommand(
		newToOCICmd(action),
		newToPyPICmd(action),
		newServeCmd(action),
		newLabelsCmd(),
		// TODO we probably need a prune command to prune away distribution files in OCI
	)
	return cmd
}
