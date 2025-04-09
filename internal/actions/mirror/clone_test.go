package mirror

import (
	"context"
	"encoding/csv"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/act3-ai/data-tool/internal/actions"
	"github.com/act3-ai/data-tool/internal/ref"
	"github.com/act3-ai/go-common/pkg/logger"
	"github.com/act3-ai/go-common/pkg/test"
)

func TestClone_Run(t *testing.T) {
	defer leaktest.Check(t)() //nolint

	log := test.Logger(t, 0)
	ctx := logger.NewContext(context.Background(), log)

	rne := require.New(t).NoError

	// Set up a fake registry
	s := httptest.NewServer(registry.New())
	defer s.Close()
	u, err := url.Parse(s.URL)
	rne(err)

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

	// create a test dir to store the config
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

	clone := Clone{
		Action: mAction,
	}
	sources := filepath.Join(dir, "sources.list")
	err = os.WriteFile(sources, []byte(strings.Join(sourceRefs, "\n")), 0o666)
	rne(err)

	t.Run("go-template", func(t *testing.T) {
		destTemplate := fmt.Sprintf(`{{- $name := index .Annotations "%[2]s" -}}
%[1]s/high/clone/go-template/{{ trimPrefix "%[1]s/" $name -}}
{{ if hasPrefix "%[1]s/low/source1" $name }}
%[1]s/high/clone/go-template/source1:tag
{{- end -}}`, u.Host, ref.AnnotationSrcRef)

		t.Log(destTemplate)

		templateFile := filepath.Join(dir, "dest.tmpl")
		require.NoError(t, os.WriteFile(templateFile, []byte(destTemplate), 0o666))

		// Run the action
		assert.NoError(t, clone.Run(ctx, sources, "go-template="+templateFile))
	})

	t.Run("first-prefix", func(t *testing.T) {
		c := [][]string{
			{"# this is a test comment"},
			{u.Host + "/low/source1", u.Host + "/high/clone/first-prefix/source1"},
			{u.Host + "/low/source1", u.Host + "/high/clone/second-prefix/source1"},
			{u.Host + "/low/source2", u.Host + "/high/clone/first-prefix/source2"},
		}
		f, err := os.Create(filepath.Join(dir, "dest.csv"))
		require.NoError(t, err)
		w := csv.NewWriter(f)
		err = w.WriteAll(c)
		require.NoError(t, err)

		err = clone.Run(ctx, sources, "first-prefix="+f.Name())
		assert.NoError(t, err)
		exists(ctx, t, u.Host+"/high/clone/first-prefix/source1", "v1")
		notExists(ctx, t, u.Host+"/high/clone/second-prefix/source1", "v1")
	})

	t.Run("digests", func(t *testing.T) {
		img1dest := u.Host + "/high/clone/digests/image1"
		c := [][]string{
			{"# this is a test comment"},
			{img1.Digest.String(), img1dest},
		}
		f, err := os.Create(filepath.Join(dir, "dest.csv"))
		rne(err)
		w := csv.NewWriter(f)
		err = w.WriteAll(c)
		rne(err)

		err = clone.Run(ctx, sources, "digests="+f.Name())
		rne(err)

		exists(ctx, t, u.Host+"/low/source1", img1.Digest.String())
	})

	t.Run("all-prefix", func(t *testing.T) {
		c := [][]string{
			{"# this is a test comment"},
			{u.Host + "/low/source1", u.Host + "/high/clone/all-prefix/source1"},
			{u.Host + "/low/source1", u.Host + "/high/clone/all-prefix/no1"},
			{u.Host + "/low/source2", u.Host + "/high/clone/all-prefix/no2"},
			{u.Host + "/low/source2", u.Host + "/high/clone/all-prefix/source2"},
			{u.Host + "/low/source2", u.Host + "/high/clone/all-prefix/image2"},
		}
		f, err := os.Create(filepath.Join(dir, "dest.csv"))
		rne(err)
		w := csv.NewWriter(f)
		err = w.WriteAll(c)
		rne(err)

		err = clone.Run(ctx, sources, "all-prefix="+f.Name())
		rne(err)

		// verify that all of the image 2 destinations exist
		exists(ctx, t, u.Host+"/high/clone/all-prefix/source1", "v1")
		exists(ctx, t, u.Host+"/high/clone/all-prefix/no1", "v1")
		exists(ctx, t, u.Host+"/high/clone/all-prefix/no2", idx1.Digest.String())
		exists(ctx, t, u.Host+"/high/clone/all-prefix/source2", idx1.Digest.String())
		exists(ctx, t, u.Host+"/high/clone/all-prefix/image2", idx1.Digest.String())
	})

	t.Run("invalid mapper", func(t *testing.T) {
		rne := require.New(t).NoError

		clone := Clone{
			Action: mAction,
		}

		c := [][]string{
			{u.Host + "/low/source1", u.Host + "/high/scatter/check/source1"},
			{u.Host + "/low/source2", u.Host + "/high/scatter/check/source2"},
		}
		f, err := os.Create(filepath.Join(dir, "dest.csv"))
		rne(err)
		w := csv.NewWriter(f)
		err = w.WriteAll(c)
		rne(err)

		err = clone.Run(ctx, sources, "invalid-mapper="+f.Name())
		assert.ErrorContains(t, err, "unknown mapping type")
	})

	t.Run("clone with labels", func(t *testing.T) {
		clone := Clone{
			Action:    mAction,
			Selectors: []string{"component=core"},
		}

		sourceRefsWithLabels := []string{
			strings.Join([]string{refImg1, "component=core", "module=test"}, ","),
			strings.Join([]string{refIdx1, "component=env", "module=test"}, ","),
		}
		labeledSources := filepath.Join(dir, "labeledSources.list")
		err = os.WriteFile(labeledSources, []byte(strings.Join(sourceRefsWithLabels, "\n")), 0o666)
		rne(err)

		labelTemplate := fmt.Sprintf(`{{- $name := index .Annotations "%[2]s" -}}
%[1]s/high/clone/core/{{ trimPrefix "%[1]s/low/" $name -}}`, u.Host, ref.AnnotationSrcRef)

		t.Log(labelTemplate)

		templateFile := filepath.Join(dir, "labels.tmpl")
		require.NoError(t, os.WriteFile(templateFile, []byte(labelTemplate), 0o666))

		// Run the action
		assert.NoError(t, clone.Run(ctx, labeledSources, "go-template="+templateFile))
		exists(ctx, t, u.Host+"/high/clone/core/source1", "v1")
		notExists(ctx, t, u.Host+"/high/clone/core/source2", idx1.Digest.String())
	})

	t.Run("clone with filter on multiarch index", func(t *testing.T) {
		clone := Clone{
			Action:    mAction,
			Platforms: []string{"linux/arm64", "linux/arm64/v8"},
		}

		casPlatform, err := remote.NewRepository(u.Host + "/low/multiarch")
		rne(err)
		casPlatform.PlainHTTP = true
		// index 2 with platform
		idx2, err := pushRandomMultiArchIndex(ctx, casPlatform, rng, "v1")
		rne(err)
		t.Log("img2", idx2)
		refIdx2 := u.Host + "/low/multiarch:v1"

		sourceRefsWithLabels := []string{
			strings.Join([]string{refImg1, "component=core", "module=test"}, ","),
			strings.Join([]string{refIdx2, "component=env", "module=test"}, ","),
		}
		labeledSources := filepath.Join(dir, "labeledSources.list")
		err = os.WriteFile(labeledSources, []byte(strings.Join(sourceRefsWithLabels, "\n")), 0o666)
		rne(err)

		labelTemplate := fmt.Sprintf(`{{- $name := index .Annotations "%[2]s" -}}
%[1]s/high/clone/platform/{{ trimPrefix "%[1]s/low/" $name -}}`, u.Host, ref.AnnotationSrcRef)

		t.Log(labelTemplate)

		templateFile := filepath.Join(dir, "labels.tmpl")
		require.NoError(t, os.WriteFile(templateFile, []byte(labelTemplate), 0o666))

		// Run the action
		assert.NoError(t, clone.Run(ctx, labeledSources, "go-template="+templateFile))
		notExists(ctx, t, u.Host+"/high/clone/platform/source1", "v1")
		exists(ctx, t, u.Host+"/high/clone/platform/multiarch", "v1")
	})
}
