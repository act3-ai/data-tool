package mirror

import (
	"context"
	"fmt"
	"math/rand"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/registry/remote"

	"gitlab.com/act3-ai/asce/data/tool/internal/actions"
	"gitlab.com/act3-ai/asce/data/tool/internal/ref"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/test"
)

func TestE2E_Smoke(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defer leaktest.Check(t)() //nolint

	log := test.Logger(t, 0)
	ctx := logger.NewContext(context.Background(), log)

	rne := require.New(t).NoError

	// Set up a fake registry
	s := httptest.NewServer(registry.New())
	defer s.Close()
	u, err := url.Parse(s.URL)
	rne(err)

	// cas1 := memory.New()
	// cas2 := memory.New()
	cas1, err := remote.NewRepository(u.Host + "/low/source1")
	rne(err)
	cas1.PlainHTTP = true

	cas2, err := remote.NewRepository(u.Host + "/low/source2")
	rne(err)
	cas2.PlainHTTP = true

	// Populating the registry with a few images
	rng := rand.New(rand.NewSource(1))

	// image 1
	img1, err := pushRandomManifest(ctx, cas1, rng, nil, "v1", nil)
	rne(err)
	t.Log("img1", img1)
	refImg1 := u.Host + "/low/source1:v1"

	// index 1
	idx1, err := pushRandomIndex(ctx, cas2, rng, "v2")
	rne(err)
	t.Log("idx1", idx1)
	refIdx1 := u.Host + "/low/source2@" + idx1.Digest.String()

	sourceRefs := []string{
		refImg1,
		refIdx1,
	}
	t.Log(sourceRefs)

	dir := GetOrCreateTestDir(t)
	// build up the action
	tAction := actions.NewTool("0.0.0")
	// create a config file that defines the registry as HTTP
	config := filepath.Join(dir, "config.yaml")
	CreateConfigWithRegHTTP(t, config, u.Host)
	// add it to the config files
	tAction.Config.ConfigFiles = []string{config}
	mAction := &Action{
		DataTool: tAction,
	}

	gather := Gather{
		Action: mAction,
	}

	sources := filepath.Join(dir, "sources.list")
	err = os.WriteFile(sources, []byte(strings.Join(sourceRefs, "\n")), 0o666)
	rne(err)

	// Run the actions
	gatherDest := u.Host + "/low/mirror:sync-1"
	err = gather.Run(ctx, sources, gatherDest)
	rne(err)

	serialize := Serialize{
		Action: mAction,
	}

	tape := filepath.Join(dir, "tape.tar")
	err = serialize.Run(ctx, gatherDest, tape, nil, 0, 0, 0)
	rne(err)

	deserialize := Deserialize{
		Action: mAction,
		Strict: true,
	}

	// commands to help debug deserialize
	// To see the output streaming run "cd pkg/actions/mirror; go test -v ."
	// To see the output during a debugging session add the following to your launch.json
	/*
			{
		            "name": "Launch test function",
		            "type": "go",
		            "request": "launch",
		            "mode": "test",
		            "program": "${workspaceFolder}/pkg/actions/mirror",
		            "args": [
		                "-test.run",
		                "TestE2E_Smoke",
		                "-test.v" // this is the important addition
		            ]
		        },
	*/
	t.Logf(`Commands to help with debugging:\nmkdir "%[1]s/oci"; tar xvf "%[1]s/tape.tar" -C "%[1]s/oci"; ace-dt oci tree -d "%[1]s/oci"`, dir)

	scatterSrc := u.Host + "/high/mirror:sync-1"
	err = deserialize.Run(ctx, tape, scatterSrc)
	rne(err)

	scatter := Scatter{
		Action: mAction,
	}

	destTemplate := fmt.Sprintf(`{{- $name := index .Annotations "%[2]s" -}}
%[1]s/high/{{ trimPrefix "%[1]s/" $name -}}
{{ if hasPrefix "%[1]s/low/source1" $name }}
%[1]s/high/scatter/source1:tag
{{- end -}}`, u.Host, ref.AnnotationSrcRef)

	templateFile := filepath.Join(dir, "dest.tmpl")
	err = os.WriteFile(templateFile, []byte(destTemplate), 0o666)
	rne(err)

	err = scatter.Run(ctx, scatterSrc, "go-template="+templateFile)
	rne(err)
}

// GetOrCreateTestDir will return a temporary test directory that is deleted upon cleanup unless
// the env TEST_DIR is set, then it will use that directory for this test.
func GetOrCreateTestDir(tb testing.TB) string {
	tb.Helper()
	knownDir := os.Getenv("TEST_DIR")
	if knownDir != "" {
		return knownDir
	}
	return tb.TempDir()
}
