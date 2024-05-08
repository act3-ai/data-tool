/* Test this with
podman run -d -p 127.0.0.1:5000:5000 docker.io/library/registry:2
go run ./cmd/ace-dt pypi to-oci localhost:5000/pypi pyzmq
*/

package pypi

import (
	"os"
	"slices"

	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"

	"gitlab.com/act3-ai/asce/data/schema/pkg/selectors"
	"gitlab.com/act3-ai/asce/data/tool/internal/python"
)

// newLabelsCmd creates a new cobra.Command for the pypi labels subcommand.
func newLabelsCmd() *cobra.Command {
	var sel []string

	cmd := &cobra.Command{
		Use:   "labels DISTNAME [-l selector]...",
		Short: "Displays all the computed labels for a given python distribution filename DISTNAME",
		Long: `The labels are computed and displayed.  If a selector is present then the the output also shows if th selector matched the labels.

Labels selectors can be used refine the set of items processed by a task.  They work exactly like Kubernetes label selectors.  Each distribution file has the following labels:
- project - normalized project name (lower case with dash)
- version.major
- version.minor
- version.patch
- type (e.g., "sdist", "bdist_wheel")
- python
- abi
- platform

Limitations:
- Version specifiers of type === are supported but no others (e.g., ==, ~= are unsupported). Support can be added if needed.
- Only some of the directives in the requirements files are processed.
`,
		Example: `This will produce two sets of labels and the selectors match.
$ ace-dt pypi labels pyzmq-23.2.1-pp39-pypy39_pp73-manylinux_2_17_x86_64.manylinux2014_x86_64.whl -l 'version.major=23,version.minor<5,python=pp39' -l 'type=sdist'
`,

		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			distname := args[0]
			cmd.SetOut(os.Stdout)

			dist, err := python.NewDistribution(distname)
			if err != nil {
				return err
			}

			labelSets := dist.Labels()
			for _, lbls := range labelSets {
				cmd.Println("Label Set")
				keys := maps.Keys(lbls)
				slices.Sort(keys)
				for _, k := range keys {
					cmd.Printf("\t%s=%s\n", k, lbls[k])
				}
				cmd.Println()
			}

			if len(sel) != 0 {
				// get the selectors
				s, err := selectors.Parse(sel)
				if err != nil {
					return err
				}

				cmd.Print("Provided selectors do ")
				ll := selectors.LabelsFromSets(labelSets)
				if s.MatchAny(ll) {
					cmd.Println("MATCH")
				} else {
					cmd.Println("NOT MATCH")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&sel, "selector", "l", []string{}, "Selectors to test against the labels.")

	return cmd
}
