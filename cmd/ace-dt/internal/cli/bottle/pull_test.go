package bottle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fortytw2/leaktest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	"git.act3-ace.com/ace/go-common/pkg/fsutil"
)

func Test_Functional_PullBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	remoteBottle := NewBottleHelper(t)
	remoteBottle.AddArbitraryFileParts(3)
	remoteBottle.AddArbitraryDirParts(2)
	remoteBottle.SetTempBottleRef(rootReg)
	helper.SendBottleToReg(remoteBottle.RootDir, remoteBottle.RegRef)
	helper.PruneCache()

	helper.CommandHelper.RunCommand("pull", remoteBottle.RegRef, "--bottle-dir", helper.BottleHelper.RootDir)

	helper.EqualBottles(remoteBottle.RootDir, helper.BottleHelper.RootDir)
}

func Test_Functional_PullWriteBottleID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	remoteBottle := NewBottleHelper(t)
	remoteBottle.AddArbitraryFileParts(15)
	remoteBottle.SetTempBottleRef(rootReg)
	helper.SendBottleToReg(remoteBottle.RootDir, remoteBottle.RegRef)
	helper.PruneCache()

	bottleIDFile := filepath.Join(t.TempDir(), "bottleID")
	helper.CommandHelper.RunCommand("pull", remoteBottle.RegRef, "--bottle-dir", helper.BottleHelper.RootDir, "--write-bottle-id", bottleIDFile)

	helper.BottleHelper.Load()
	helper.VerifyBottleIDFile(bottleIDFile)

	helper.EqualBottles(remoteBottle.RootDir, helper.BottleHelper.RootDir)
}

func Test_Functional_PullVerifyID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	remoteBottle := NewBottleHelper(t)
	remoteBottle.AddArbitraryFileParts(15)
	remoteBottle.SetTempBottleRef(rootReg)
	helper.SendBottleToReg(remoteBottle.RootDir, remoteBottle.RegRef)
	helper.PruneCache()

	remoteBottle.Load()
	bottleID := remoteBottle.Bottle.GetBottleID()

	helper.CommandHelper.RunCommand("pull", remoteBottle.RegRef, "--bottle-dir", helper.BottleHelper.RootDir, "--check-bottle-id", bottleID.String())

	helper.EqualBottles(remoteBottle.RootDir, helper.BottleHelper.RootDir)
}

func Test_Functional_PullName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	remoteBottle := NewBottleHelper(t)
	remoteBottle.AddArbitraryFileParts(5)

	labeledPart := "labeledPart.txt"
	remoteBottle.AddBottlePart(labeledPart)
	remoteBottle.SetTempBottleRef(rootReg)
	helper.SendBottleToReg(remoteBottle.RootDir, remoteBottle.RegRef)
	helper.PruneCache()

	helper.CommandHelper.RunCommand("pull", remoteBottle.RegRef, "--bottle-dir", helper.BottleHelper.RootDir, "--part", labeledPart)

	numLocalParts, err := helper.GetNumLocalParts(helper.BottleHelper.RootDir)
	assert.NoError(t, err)
	assert.Equal(t, 1, numLocalParts)
}

func Test_Functional_PullSelector(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	remoteBottle := NewBottleHelper(t)
	remoteBottle.AddArbitraryFileParts(5)
	labeledPart := "labeledPart.txt"
	remoteBottle.AddBottlePart(labeledPart)
	assert.NoError(t, helper.SaveBottle(remoteBottle.RootDir))
	remoteBottle.Load()
	remoteBottle.AddPartLabel(labeledPart, "test", "true")
	remoteBottle.SetTempBottleRef(rootReg)
	helper.SendBottleToReg(remoteBottle.RootDir, remoteBottle.RegRef)
	helper.PruneCache()

	helper.CommandHelper.RunCommand("pull", remoteBottle.RegRef, "--bottle-dir", helper.BottleHelper.RootDir, "-l", "test=true")

	numLocalParts, err := helper.GetNumLocalParts(helper.BottleHelper.RootDir)
	assert.NoError(t, err)
	assert.Equal(t, 1, numLocalParts)
}

func Test_Functional_PullFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.BottleHelper.SetTempBottleRef(rootReg)

	err := helper.CommandHelper.RunCommandWithError("pull", helper.BottleHelper.RegRef, "--bottle-dir", helper.BottleHelper.RootDir)
	assert.ErrorIs(t, err, errdef.ErrNotFound)
}

func Test_Functional_Pull_Versions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	defer leaktest.Check(t)() //nolint
	rootCmd := rootTestCmd()

	/* NOTE to Windows users
	You need the following in your GIT config to run this test otherwise the files are modified by GIT and that breaks the test
	[core]
	   autocrlf = false
	   eol = lf
	*/

	// to add more versions and test bottles run
	// cd cmd/ace-dt/cli/testdata; ./generate.sh localhost-registry

	versions := []string{
		"v0.25.4",
		"master",
	}

	testData := []string{
		"basic",
		"subparts",
	}

	for _, d := range testData {
		base := filepath.Join("testdata", d)
		// the truth
		dataDir := os.DirFS(filepath.Join(base, "original"))
		for _, ver := range versions {
			t.Run(d+"_"+ver, func(t *testing.T) {
				helper := NewTestHelper(t, rootCmd)
				helper.BottleHelper.SetTempBottleRef(rootReg)

				ref := helper.BottleHelper.RegRef + "-" + ver

				desc, err := pushImage(context.TODO(), filepath.Join(base, "oci-"+ver), ref)
				require.NoError(t, err)
				t.Logf("pushed image with manifest ID %s", desc.Digest)

				dir := filepath.Join(helper.BottleHelper.RootDir, ver)
				helper.CommandHelper.RunCommand("pull", ref, "--bottle-dir", dir)
				options := fsutil.DefaultComparisonOpts
				options.Mode = false // TODO need a bitmask for mode comparison not just a boolean
				assert.NoError(t, fsutil.EqualFilesystem(dataDir, os.DirFS(dir), options))
			})
		}
	}
}

// pushImage pushing a single image from dir (OCI image layout directory) to the insecure registry at reg.
func pushImage(ctx context.Context, dir string, ref string) (ocispec.Descriptor, error) {
	// Create a OCI layout store
	store, err := oci.NewFromFS(ctx, os.DirFS(dir))
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	repo, err := remote.NewRepository(ref)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	repo.PlainHTTP = true

	// Collect all the tags in the oci/index.json file
	allTags, err := registry.Tags(ctx, store)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	if len(allTags) != 1 {
		return ocispec.Descriptor{}, fmt.Errorf("expected one tag but found these tags %q", allTags)
	}

	return oras.Copy(ctx, store, allTags[0], repo, repo.Reference.ReferenceOrDefault(), oras.DefaultCopyOptions)
}
