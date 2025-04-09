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

func exists(ctx context.Context, t *testing.T, repo, reference string) {
	t.Helper()
	cas, err := remote.NewRepository(repo)
	require.NoError(t, err)
	cas.PlainHTTP = true

	_, err = cas.Resolve(ctx, reference)
	require.NoError(t, err)
}

func notExists(ctx context.Context, t *testing.T, repo, reference string) {
	t.Helper()
	cas, err := remote.NewRepository(repo)
	require.NoError(t, err)
	cas.PlainHTTP = true
	_, err = cas.Resolve(ctx, reference)
	assert.Error(t, err)
}

func TestScatter_Run(t *testing.T) {
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
		strings.Join([]string{refImg1, "component=core", "module=test"}, ","),
		strings.Join([]string{refIdx1, "component=env", "module=test"}, ","),
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

	t.Run("go-template", func(t *testing.T) {
		scatter := Scatter{
			Action: mAction,
		}

		destTemplate := fmt.Sprintf(`{{- $name := index .Annotations "%[2]s" -}}
%[1]s/high/scatter/go-template/{{ trimPrefix "%[1]s/" $name -}}
{{ if hasPrefix "%[1]s/low/source1" $name }}
%[1]s/high/scatter/go-template/source1:tag
{{- end -}}`, u.Host, ref.AnnotationSrcRef)

		t.Log(destTemplate)

		templateFile := filepath.Join(dir, "dest.tmpl")
		require.NoError(t, os.WriteFile(templateFile, []byte(destTemplate), 0o666))

		assert.NoError(t, scatter.Run(ctx, gatherDest, "go-template="+templateFile))
		exists(ctx, t, u.Host+"/high/scatter/go-template/source1", "tag")
		exists(ctx, t, u.Host+"/high/scatter/go-template/low/source1", "v1")
		exists(ctx, t, u.Host+"/high/scatter/go-template/low/source2", idx1.Digest.String())
	})

	t.Run("first-prefix", func(t *testing.T) {
		scatter := Scatter{
			Action: mAction,
		}

		c := [][]string{
			{"# this is a test comment"},
			{u.Host + "/low/source1", u.Host + "/high/first-prefix/source1"},
			{u.Host + "/low/source2", u.Host + "/high/first-prefix/source2"},
		}
		f, err := os.Create(filepath.Join(dir, "dest.csv"))
		require.NoError(t, err)
		w := csv.NewWriter(f)
		err = w.WriteAll(c)
		require.NoError(t, err)

		err = scatter.Run(ctx, gatherDest, "first-prefix="+f.Name())
		assert.NoError(t, err)
	})

	t.Run("longest-prefix", func(t *testing.T) {
		scatter := Scatter{
			Action: mAction,
		}

		c := [][]string{
			{"# this is a test comment"},
			{u.Host + "/low/source", u.Host + "/high/longest-prefix/wrong"},
			{u.Host + "/low/source1", u.Host + "/high/longest-prefix/source1"},
			{u.Host + "/low/source2", u.Host + "/high/longest-prefix/source2"},
		}
		f, err := os.Create(filepath.Join(dir, "dest.csv"))
		require.NoError(t, err)
		w := csv.NewWriter(f)
		err = w.WriteAll(c)
		require.NoError(t, err)

		err = scatter.Run(ctx, gatherDest, "longest-prefix="+f.Name())
		assert.NoError(t, err)
		notExists(ctx, t, u.Host+"/high/longest-prefix/wrong", "v1")
		exists(ctx, t, u.Host+"/high/longest-prefix/source1", "v1")
		exists(ctx, t, u.Host+"/high/longest-prefix/source2", idx1.Digest.String())
	})

	t.Run("digests", func(t *testing.T) {
		rne := require.New(t).NoError

		scatter := Scatter{
			Action: mAction,
		}

		img1dest := u.Host + "/high/scatter/digests/image1"
		c := [][]string{
			{"# this is a test comment"},
			{img1.Digest.String(), img1dest},
		}
		f, err := os.Create(filepath.Join(dir, "dest.csv"))
		rne(err)
		w := csv.NewWriter(f)
		err = w.WriteAll(c)
		rne(err)

		err = scatter.Run(ctx, gatherDest, "digests="+f.Name())
		rne(err)

		exists(ctx, t, u.Host+"/low/source1", img1.Digest.String())
	})

	t.Run("all-prefix", func(t *testing.T) {
		rne := require.New(t).NoError

		scatter := Scatter{
			Action: mAction,
		}

		c := [][]string{
			{"# this is a test comment"},
			{u.Host + "/low/source1", u.Host + "/high/scatter/all-prefix/source1"},
			{u.Host + "/low/source1", u.Host + "/high/scatter/all-prefix/no1"},
			{u.Host + "/low/source2", u.Host + "/high/scatter/all-prefix/no2"},
			{u.Host + "/low/source2", u.Host + "/high/scatter/all-prefix/source2"},
			{u.Host + "/low/source2", u.Host + "/high/scatter/all-prefix/image2"},
		}
		f, err := os.Create(filepath.Join(dir, "dest.csv"))
		rne(err)
		w := csv.NewWriter(f)
		err = w.WriteAll(c)
		rne(err)

		err = scatter.Run(ctx, gatherDest, "all-prefix="+f.Name())
		rne(err)

		// verify that all of the image 2 destinations exist
		exists(ctx, t, u.Host+"/high/scatter/all-prefix/source1", "v1")
		exists(ctx, t, u.Host+"/high/scatter/all-prefix/no1", "v1")
		exists(ctx, t, u.Host+"/high/scatter/all-prefix/no2", idx1.Digest.String())
		exists(ctx, t, u.Host+"/high/scatter/all-prefix/source2", idx1.Digest.String())
		exists(ctx, t, u.Host+"/high/scatter/all-prefix/image2", idx1.Digest.String())
	})

	t.Run("check", func(t *testing.T) {
		rne := require.New(t).NoError

		scatter := Scatter{
			Action: mAction,
			Check:  true,
		}

		c := [][]string{
			{"# this is a test comment"},
			{u.Host + "/low/source1", u.Host + "/high/scatter/check/source1"},
			{u.Host + "/low/source2", u.Host + "/high/scatter/check/source2"},
		}
		f, err := os.Create(filepath.Join(dir, "dest.csv"))
		rne(err)
		w := csv.NewWriter(f)
		err = w.WriteAll(c)
		rne(err)

		err = scatter.Run(ctx, gatherDest, "all-prefix="+f.Name())
		rne(err)

		// verify that all of the image 2 destinations exist
		notExists(ctx, t, u.Host+"/high/scatter/check/source1", "v1")
		notExists(ctx, t, u.Host+"/high/scatter/check/source2", idx1.Digest.String())
	})

	t.Run("invalid mapper", func(t *testing.T) {
		rne := require.New(t).NoError

		scatter := Scatter{
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

		err = scatter.Run(ctx, gatherDest, "invalid-mapper="+f.Name())
		assert.ErrorContains(t, err, "unknown mapping type")
	})

	t.Run("subset", func(t *testing.T) {
		rne := require.New(t).NoError

		// we only want one of the images in the scatter repository
		subsetRef := []string{
			refImg1,
		}

		// create the sources.list file
		sources := filepath.Join(dir, "subset.list")
		err = os.WriteFile(sources, []byte(strings.Join(subsetRef, "\n")), 0o666)
		rne(err)

		scatter := Scatter{
			Action:     mAction,
			SourceFile: sources,
		}

		destTemplate := fmt.Sprintf(`{{- $name := index .Annotations "%[2]s" -}}
%[1]s/high/scatter/subset/{{ trimPrefix "%[1]s/" $name -}}
{{ if hasPrefix "%[1]s/low/source1" $name }}
%[1]s/high/scatter/subset/source1:tag
{{- end -}}`, u.Host, ref.AnnotationSrcRef)

		templateFile := filepath.Join(dir, "dest.tmpl")
		require.NoError(t, os.WriteFile(templateFile, []byte(destTemplate), 0o666))

		assert.NoError(t, scatter.Run(ctx, gatherDest, "go-template="+templateFile))
		exists(ctx, t, u.Host+"/high/scatter/subset/source1", "tag")
		exists(ctx, t, u.Host+"/high/scatter/subset/low/source1", "v1")
		notExists(ctx, t, u.Host+"/high/scatter/subset/low/source2", idx1.Digest.String())
	})

	// TODO: fix me or move to internal/mirror
	// t.Run("Copy Test", func(t *testing.T) {

	// 	cas3, err := remote.NewRepository(u.Host + "/low/source3")
	// 	rne(err)
	// 	cas3.PlainHTTP = true

	// 	// image 2
	// 	img2, err := pushRandomManifest(ctx, cas3, rng, nil, "v1", nil)
	// 	rne(err)
	// 	t.Log("img2", img2)

	// 	// construct the descriptor for image 1
	// 	image2Descriptor := &v1.Descriptor{
	// 		MediaType:   img2.MediaType,
	// 		Digest:      img2.Digest,
	// 		Size:        img2.Size,
	// 		Annotations: img2.Annotations,
	// 	}

	// 	// image 3 which refers to image 3
	// 	img3, err := pushRandomManifest(ctx, cas3, rng, image2Descriptor, "v1", nil)
	// 	rne(err)
	// 	t.Log("img3", img3)
	// 	refImg3 := u.Host + "/low/source3:v1"
	// 	destRef := u.Host + "/dest/copy/source1"

	// 	c, err := mirror.NewCopier(ctx, slog.Default(), refImg3, destRef, nil, reg.Reference{}, v1.Descriptor{}, nil, reg.Reference{}, true, nil, mAction.NewRepository)
	// 	rne(err)
	// 	err = mirror.Copy(ctx, c)
	// 	rne(err)
	// 	_, err = c.dest.Resolve(ctx, u.Host+"/dest/copy/source1@"+img3.Digest.String())
	// 	rne(err)
	// 	_, err = c.dest.Resolve(ctx, u.Host+"/dest/copy/source1@"+img2.Digest.String())
	// 	rne(err)
	// })

	t.Run("filter", func(t *testing.T) {
		scatter := Scatter{
			Action:    mAction,
			Selectors: []string{"component=core"},
		}

		destTemplate := fmt.Sprintf(`{{- $name := index .Annotations "%[2]s" -}}
%[1]s/high/scatter/filtered/{{ trimPrefix "%[1]s/low/" $name -}}`, u.Host, ref.AnnotationSrcRef)

		templateFile := filepath.Join(dir, "filter.tmpl")
		require.NoError(t, os.WriteFile(templateFile, []byte(destTemplate), 0o666))

		assert.NoError(t, scatter.Run(ctx, gatherDest, "go-template="+templateFile))
		exists(ctx, t, u.Host+"/high/scatter/filtered/source1", "v1")
		notExists(ctx, t, u.Host+"/high/scatter/filtered/source2", idx1.Digest.String())
	})
}
