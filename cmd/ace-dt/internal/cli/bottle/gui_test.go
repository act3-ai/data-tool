package bottle

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	v1 "git.act3-ace.com/ace/data/schema/pkg/apis/data.act3-ace.io/v1"
	"gitlab.com/act3-ai/asce/data/tool/internal/actions/bottle"
)

func Test_Functional_GUI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.AddArbitraryFileParts(1)
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.BottleHelper.Load()
	helper.Bottle.AddAnnotation("foo", "bar")
	helper.Bottle.AddLabel("foo", "bar")
	assert.NoError(t, helper.Bottle.AddAuthorInfo(v1.Author{
		Name:  "Test Jones",
		Email: "test@test.org",
		URL:   "test.org/test",
	}))
	assert.NoError(t, helper.Bottle.AddMetricInfo(v1.Metric{
		Name:        "Test",
		Description: "This metric shows that this is a test",
		Value:       "100",
	}))
	assert.NoError(t, helper.Bottle.AddSourceInfo(v1.Source{
		Name: "Test Source",
		URI:  "sha256:e22540034211ad6021394902dfb42f49b36ea9d45c838955afddca825397c2e2",
	}))
	require.NoError(t, helper.Bottle.Save())

	updatedBottle := helper.Bottle.Definition
	updatedBottle.Annotations = map[string]string{
		"updated": "updated",
	}
	updatedBottle.Labels = map[string]string{
		"updated": "updated",
	}
	updatedBottle.Authors = []v1.Author{
		{
			Name:  "Updated Name",
			Email: "updated@test.org",
			URL:   "test.org/updated",
		},
	}
	updatedBottle.Metrics = []v1.Metric{
		{
			Name:        "Updated Metric",
			Description: "This metric shows that this is updated",
			Value:       "-100",
		},
	}
	updatedBottle.Sources = []v1.Source{
		{
			Name: "Updated Source",
			URI:  "https://updated.source.com/updated123",
		},
	}
	// TODO: test the client side submission of this data instead of marshalling it ourselves here
	updatedBottleJSON, err := json.Marshal(updatedBottle)
	require.NoError(t, err)

	ctx := context.TODO()
	g := errgroup.Group{}
	g.Go(func() error {
		helper.CommandHelper.RunCommand("gui", "--no-browser")
		return nil
	})

	// wait for the serve to come up
	// TODO it would be better call a function to start the server and then call http.Post()

	retries := 0
	maxRetries := 10
	for ; retries < maxRetries; retries++ {
		// Read the GUI url from file
		u, err := os.ReadFile(bottle.GUIURLPath(helper.RootDir))
		if err != nil {
			t.Logf("re-trying reading get-url.txt, attempt %d, %s", retries, err)
			time.Sleep(10 * time.Millisecond)
			continue
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, string(u), bytes.NewBuffer(updatedBottleJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, io.EOF) {
			t.Logf("re-trying http.Post(), attempt %d, %s", retries, err)
			time.Sleep(10 * time.Millisecond)
			continue
		}
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		require.Equal(t, http.StatusOK, resp.StatusCode)
		break
	}
	assert.Less(t, retries, maxRetries, "Retry limit exceeded")

	assert.NoError(t, g.Wait())

	helper.BottleHelper.Load()

	// assertions
	// require.Equal(t, updatedBottle.Authors, helper.Bottle.Definition.Authors)
	// require.Equal(t, updatedBottle.Sources, helper.Bottle.Definition.Sources)
	// require.Equal(t, updatedBottle.Annotations, helper.Bottle.Definition.Annotations)
	// require.Equal(t, updatedBottle.Metrics, helper.Bottle.Definition.Metrics)
	// require.Equal(t, updatedBottle.Labels, helper.Bottle.Definition.Labels)
}
