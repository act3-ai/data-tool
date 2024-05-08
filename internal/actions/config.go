package actions

import (
	"context"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"

	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
)

// Config is the action for getting the server configuration.
type Config struct {
	*DataTool

	Sample bool
}

// Run is the action method.
func (action *Config) Run(ctx context.Context, out io.Writer) error {
	if action.Sample {
		_, err := fmt.Fprint(out, v1alpha1.SampleConfig)
		return err
	}

	serverConfig := action.Config.Get(ctx)

	confYAML, err := yaml.Marshal(serverConfig)
	if err != nil {
		return fmt.Errorf("error marshalling server config: %w", err)
	}
	_, err = fmt.Fprintln(out, string(confYAML))
	return err
}
