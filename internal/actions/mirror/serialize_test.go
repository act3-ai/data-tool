package mirror

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/google/go-containerregistry/pkg/registry"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/act3-ai/data-tool/internal/actions"
	"github.com/act3-ai/data-tool/internal/mirror"
	"github.com/act3-ai/go-common/pkg/logger"
	"github.com/act3-ai/go-common/pkg/test"
)

func TestSerializeRun(t *testing.T) {
	defer leaktest.Check(t)() //nolint

	log := test.Logger(t, 0)
	ctx := logger.NewContext(context.Background(), log)

	rne := require.New(t).NoError

	// Set up a fake registry
	s := httptest.NewServer(registry.New())
	defer s.Close()
	u, err := url.Parse(s.URL)
	rne(err)

	// cas := memory.New()
	cas, err := remote.NewRepository(u.Host + "/index")
	rne(err)
	cas.PlainHTTP = true

	// Populating the registry with a few images
	rng := rand.New(rand.NewSource(1))

	// index 1
	idx1, err := pushRandomIndex(ctx, cas, rng, "sync-1")
	rne(err)
	t.Log("idx1", idx1)
	ref := u.Host + "/index:sync-1"

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

	// oci.TreeRemote(context.Background(), gatherDest, os.Stdout)

	t.Run("basic", func(t *testing.T) {
		rne := require.New(t).NoError

		// create the temp directory
		tmpdir := t.TempDir()

		// create the local tar destination
		tf := filepath.Join(tmpdir, "test.tar")

		// build up the serialize action
		serialize := Serialize{
			Action: mAction,
		}

		bs := 1024 * 1024

		err = serialize.Run(ctx, ref, tf, nil, 0, bs, 90)
		rne(err)
	})

	t.Run("checkpoint", func(t *testing.T) {
		rne := require.New(t).NoError

		// create the temp directory
		tmpdir := t.TempDir()

		// create the local tar destination
		tf := filepath.Join(tmpdir, "test.tar")

		cp := filepath.Join(tmpdir, "checkpoint")

		// build up the serialize action
		serialize := Serialize{
			Action:     mAction,
			Checkpoint: cp,
		}

		bs := 1024 * 1024

		err = serialize.Run(ctx, ref, tf, nil, 0, bs, 90)
		rne(err)

		// make sure that the checkpoint file exists
		assert.FileExists(t, cp)
		// TODO check more of the checkpoint file
	})

	t.Run("existing images", func(t *testing.T) {
		rne := require.New(t).NoError

		// create the temp directory
		tmpdir := t.TempDir()

		// create the local tar destination
		tf := filepath.Join(tmpdir, "test.tar")

		cp := filepath.Join(tmpdir, "checkpoint")

		// build a destExist
		successors, err := content.Successors(ctx, cas, idx1)
		rne(err)
		existingDigest := successors[0].Digest
		successors, err = content.Successors(ctx, cas, successors[0])
		rne(err)
		existingConfigDigest := successors[0].Digest
		rne(err)
		existingImages := []string{
			u.Host + "/index@" + existingDigest.String(),
		}

		// build up the serialize action
		serialize := Serialize{
			Action:     mAction,
			Checkpoint: cp,
		}

		bs := 1024 * 1024
		err = serialize.Run(ctx, ref, tf, existingImages, 0, bs, 90)
		rne(err)

		// make sure that the checkpoint file exists
		assert.FileExists(t, cp)
		// TODO check more of the checkpoint file

		// make sure that the checkpoint file is not empty
		fi, err := os.Stat(cp)
		rne(err)
		if fi.Size() == 0 {
			rne(fmt.Errorf("error empty checkpoint file"))
		}

		// make sure that the "existing image" is not in the checkpoint file
		f, err := os.Open(cp)
		rne(err)
		defer f.Close()

		decoder := json.NewDecoder(f)
		for {
			desc := ocispec.Descriptor{}
			err := decoder.Decode(&desc)
			if errors.Is(err, io.EOF) {
				break
			}
			rne(err)
			if existingConfigDigest == desc.Digest {
				t.Errorf("existingImage manifest was serialized.  Digest is %s", existingConfigDigest)
			}
		}
	})

	t.Run("resume", func(t *testing.T) {
		rne := require.New(t).NoError

		// create the temp directory
		tmpdir := t.TempDir()

		// create the local tar destination
		tf := filepath.Join(tmpdir, "test.tar")

		cp := filepath.Join(tmpdir, "checkpoint")
		rne(os.WriteFile(cp, []byte(``), 0o600))

		// build up the serialize action
		serialize := Serialize{
			Action: mAction,
			ExistingCheckpoints: []mirror.ResumeFromLedger{
				{Path: cp, Offset: 123456},
			},
		}

		bs := 1024 * 1024

		err = serialize.Run(ctx, ref, tf, nil, 0, bs, 90)
		rne(err)
	})

	t.Run("buffer", func(t *testing.T) {
		rne := require.New(t).NoError

		// create the temp directory
		tmpdir := t.TempDir()

		// create the local tar destination
		tf := filepath.Join(tmpdir, "test.tar")

		// build up the serialize action
		serialize := Serialize{
			Action: mAction,
		}
		n := 4096
		bs := 1024 * 1024

		err = serialize.Run(ctx, ref, tf, nil, n, bs, 90)
		rne(err)
	})

	t.Run("referrers", func(t *testing.T) {
		rne := require.New(t).NoError

		// create the temp directory
		tmpdir := t.TempDir()

		// create the local tar destination
		tf := filepath.Join(tmpdir, "test.tar")

		// build up the serialize action
		serialize := Serialize{
			Action: &Action{
				DataTool:  tAction,
				Recursive: true,
			},
		}
		// TODO: how to link predecessors in cas?
		err = serialize.Run(ctx, ref, tf, nil, 0, 0, 90)
		rne(err)
	})
}
