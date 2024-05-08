package mirror

import (
	"context"
	"crypto/sha256"
	"encoding/json"
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
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"sigs.k8s.io/yaml"

	"git.act3-ace.com/ace/data/tool/internal/actions"
	"git.act3-ace.com/ace/data/tool/internal/mirror/encoding"
	"git.act3-ace.com/ace/data/tool/internal/ref"
	"git.act3-ace.com/ace/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
	"git.act3-ace.com/ace/go-common/pkg/logger"
	"git.act3-ace.com/ace/go-common/pkg/test"
)

func TestGatherRun(t *testing.T) {
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
	sources := filepath.Join(dir, "sources.list")
	err = os.WriteFile(sources, []byte(strings.Join(sourceRefs, "\n")), 0666)
	rne(err)

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

	t.Run("basic", func(t *testing.T) {
		rne := require.New(t).NoError

		gather := Gather{
			Action: mAction,
		}

		// Run the gather action
		gatherDest := u.Host + "/low/mirror1:sync-1"
		err = gather.Run(ctx, sources, gatherDest)
		rne(err)

		// verify that the images are now in the new location
		cas3, err := remote.NewRepository(gatherDest)
		rne(err)
		cas3.PlainHTTP = true

		// pull the destination descriptor
		_, data, err := oras.FetchBytes(ctx, cas3, "sync-1", oras.DefaultFetchBytesOptions)
		rne(err)

		// grab the manifest and check for the image and index
		var man ocispec.Index
		require.NoError(t, json.Unmarshal(data, &man))
		for _, mdesc := range man.Manifests {
			refname := mdesc.Annotations[ref.AnnotationSrcRef]
			if refname != refIdx1 && refname != refImg1 {
				rne(fmt.Errorf("unexpected descriptor: %+v", mdesc))
			}
		}
	})

	t.Run("index fallback", func(t *testing.T) {
		rne := require.New(t).NoError

		gather := Gather{
			Action:        mAction,
			IndexFallback: true,
		}

		// Run the gather action
		gatherDest := u.Host + "/low/mirror2:sync-1"
		err = gather.Run(ctx, sources, gatherDest)
		rne(err)

		// verify that the images are now in the new location
		cas3, err := remote.NewRepository(gatherDest)
		rne(err)
		cas3.PlainHTTP = true

		// pull the destination descriptor
		desc, err := cas3.Resolve(ctx, "sync-1")
		rne(err)

		// grab the manifest and check for the image and index
		successors, err := encoding.Successors(ctx, cas3, desc)
		rne(err)
		require.Len(t, successors, 2)
	})

	t.Run("basicWithPlatform", func(t *testing.T) {
		rne := require.New(t).NoError

		gather := Gather{
			Action:    mAction,
			Platforms: []string{"linux/arm64", "linux/arm64/v8"},
		}

		casPlatform, err := remote.NewRepository(u.Host + "/low/platform")
		rne(err)
		casPlatform.PlainHTTP = true
		// index 2 with platform
		idx2, err := pushRandomMultiArchIndex(ctx, casPlatform, rng, "v1")
		rne(err)
		t.Log("img2", idx2)
		refIdx2 := u.Host + "/low/platform:v1"

		sourceRefs := []string{
			refIdx2,
			refImg1,
		}
		t.Log(sourceRefs)

		dir := GetOrCreateTestDir(t)
		sources := filepath.Join(dir, "sources.list")
		err = os.WriteFile(sources, []byte(strings.Join(sourceRefs, "\n")), 0666)
		rne(err)

		// Run the gather action
		gatherDest := u.Host + "/low/mirror3:sync-1"
		err = gather.Run(ctx, sources, gatherDest)
		rne(err)

		// verify that the images are now in the new location
		cas3, err := remote.NewRepository(gatherDest)
		rne(err)
		cas3.PlainHTTP = true

		// pull the destination descriptor
		_, data, err := oras.FetchBytes(ctx, cas3, "sync-1", oras.DefaultFetchBytesOptions)
		rne(err)

		// grab the manifest and check for the image and index
		var man ocispec.Index
		require.NoError(t, json.Unmarshal(data, &man))
		for _, mdesc := range man.Manifests {
			refname := mdesc.Annotations[ref.AnnotationSrcRef]
			// make sure that the image with no platform defined is filtered out.
			assert.NotEqual(t, refname, refImg1)

			// grab the source index from the annotations and compute its digest. Make sure it is equal to the actual source index digest.
			srcIdx := mdesc.Annotations[encoding.AnnotationSrcIndex]
			h := sha256.New()
			h.Write([]byte(srcIdx))
			computed := fmt.Sprintf("%x", h.Sum(nil))
			assert.Equal(t, idx2.Digest.Encoded(), computed)
		}

	})

	t.Run("gather with labels", func(t *testing.T) {
		cas3, err := remote.NewRepository(u.Host + "/low/labels")
		rne(err)
		cas3.PlainHTTP = true

		rne := require.New(t).NoError

		// create new images
		// image 2
		img2, err := pushRandomManifest(ctx, cas3, rng, nil, "image2", nil)
		rne(err)
		t.Log("img2", img2)
		refImg2 := u.Host + "/low/labels:image2"

		// image 3
		img3, err := pushRandomManifest(ctx, cas3, rng, nil, "image3", nil)
		rne(err)
		t.Log("img3", img3)
		refImg3 := u.Host + "/low/labels:image3"

		// index 2
		idx2, err := pushRandomIndex(ctx, cas3, rng, "v2")
		rne(err)
		t.Log("idx2", idx2)
		refIdx2 := u.Host + "/low/labels:v2"

		labeledSourceRefs := []string{
			strings.Join([]string{refImg2, "component=core", "module=kuberay"}, ","),
			strings.Join([]string{refImg3, "component=env", "module=kuberay"}, ","),
			strings.Join([]string{refIdx2, "component=core"}, ","),
		}
		t.Log(labeledSourceRefs)

		// create source file with labels

		labeledSources := filepath.Join(dir, "labeled-sources.list")
		err = os.WriteFile(labeledSources, []byte(strings.Join(labeledSourceRefs, "\n")), 0666)
		rne(err)

		gather := Gather{
			Action:        mAction,
			IndexFallback: true,
		}

		// Run the gather action
		gatherDest := u.Host + "/low/mirror2:sync-2"
		err = gather.Run(ctx, labeledSources, gatherDest)
		rne(err)

		// TODO Pull labels and check them
	})

	// t.Run("parse source and labels", func(t *testing.T) {
	// 	rne := require.New(t).NoError
	// 	source, labels, err := processSourceLabels([]string{"localhost:5000/testing/image1:v1", "component = core", "module=kuberay"})
	// 	rne(err)
	// 	assert.Equal(t, "localhost:5000/testing/image1:v1", source)
	// 	assert.EqualValues(t, map[string]string{"component": "core", "module": "kuberay"}, labels)
	// })

	// 	t.Run("invalid labels", func(t *testing.T) {
	// 		_, _, err := mirror.processSourceLabels([]string{"localhost:5000/testing/image1:v1", "component=core!!!"})
	// 		assert.ErrorContains(t, err, "error validating labels")
	// 	})
}

func CreateConfigWithRegHTTP(tb testing.TB, fileName, hostName string) {
	tb.Helper()
	rne := require.New(tb).NoError
	regMap := map[string]v1alpha1.Registry{
		hostName: {
			Endpoints: []string{"http://" + hostName},
		},
	}
	cfg := v1alpha1.ConfigurationSpec{
		RegistryConfig: v1alpha1.RegistryConfig{
			Configs: regMap,
		},
	}
	b, err := yaml.Marshal(cfg)
	rne(err)
	err = os.WriteFile(fileName, b, 0666)
	rne(err)
}
