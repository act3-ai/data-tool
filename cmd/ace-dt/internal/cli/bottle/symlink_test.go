package bottle

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.act3-ace.com/ace/data/tool/pkg/apis"
	"git.act3-ace.com/ace/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
	"git.act3-ace.com/ace/go-common/pkg/config"
)

// Yes they all test different logic in ace-dt

func getConfig(configLoc string) (*v1alpha1.Configuration, error) {
	c := &v1alpha1.Configuration{}
	discard := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := config.Load(discard, apis.NewScheme(), c, []string{configLoc})
	if err != nil {
		return nil, err
	}
	return c, nil
}

func Test_Functional_SymLinkFileToSymLinkFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support symlinks")
	}

	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	// add sym link file
	symlinkFile := NewBottleHelper(t)
	filePath := "originalPartLocation"
	symlinkFile.AddBottlePart(filePath)

	// add another level of indirection (to make sure we traverse through symlinks to symlinks)
	intermediateFile := filepath.Join(t.TempDir(), "intermediate")
	assert.NoError(t, os.Symlink(filepath.Join(symlinkFile.RootDir, filePath), intermediateFile))

	// the final destination
	newFileName := "symlinkedFile"
	assert.NoError(t, os.Symlink(intermediateFile, filepath.Join(helper.BottleHelper.RootDir, newFileName)))

	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.CommandHelper.RunCommand("commit")

	// check for 1 symlink file
	helper.BottleHelper.Load()
	bottleParts := helper.BottleHelper.Bottle.GetParts()
	assert.Len(t, bottleParts, 1)
	assert.Equal(t, newFileName, bottleParts[0].GetName())

	cfg, err := getConfig(helper.CommandHelper.GetConfigFile())
	assert.NoError(t, err)

	// check cachepath, looking for part that is larger than a symlink (check if ace-dt cached the symlink and not the compressed+digested part)
	dgst := bottleParts[0].GetLayerDigest()
	cachePart := filepath.Join(cfg.CachePath, dgst.Algorithm().String(), dgst.Encoded())
	finfo, err := os.Stat(cachePart)
	assert.NoError(t, err)
	assert.Less(t, int64(1000), finfo.Size())
}

func Test_Functional_SymLinkDirToSymLinkDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support symlinks")
	}

	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	// add sym link directory
	symlinkDir := NewBottleHelper(t)
	symlinkDir.AddArbitraryFileParts(10)
	newDirName := "symlinkedDir/"
	// add another level of indirection (to make sure we traverse through symlinks to symlinks)
	intermediateDir := filepath.Join(t.TempDir(), "intermediate")
	assert.NoError(t, os.Symlink(symlinkDir.RootDir, intermediateDir))
	// the final destination
	assert.NoError(t, os.Symlink(intermediateDir, filepath.Join(helper.BottleHelper.RootDir, newDirName)))

	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.CommandHelper.RunCommand("commit")

	// check for 1 symlink file
	helper.BottleHelper.Load()
	bottleParts := helper.BottleHelper.Bottle.GetParts()
	assert.Len(t, bottleParts, 1)
	assert.Equal(t, newDirName, bottleParts[0].GetName())

	cfg, err := getConfig(helper.CommandHelper.GetConfigFile())
	assert.NoError(t, err)

	// check cachepath, looking for part that is larger than a symlink (check if ace-dt cached the symlink and not the compressed+digested part)
	dgst := bottleParts[0].GetLayerDigest()
	cachePart := filepath.Join(cfg.CachePath, dgst.Algorithm().String(), dgst.Encoded())
	finfo, err := os.Stat(cachePart)
	assert.NoError(t, err)
	assert.Less(t, int64(1000), finfo.Size())
}

func Test_Functional_SymLinkFileAndDirPreInitThenCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support symlinks")
	}

	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)
	rne := require.New(t).NoError

	// add sym link file
	symlinkFile := NewBottleHelper(t)
	filePath := "originalPartLocation"
	symlinkFile.AddBottlePart(filePath)
	// add another level of indirection (to make sure we traverse through symlinks to symlinks)
	intermediateFile := filepath.Join(t.TempDir(), "intermediate")
	rne(os.Symlink(filepath.Join(symlinkFile.RootDir, filePath), intermediateFile))
	// the final destination
	newFileName := "symlinkedFile"
	rne(os.Symlink(intermediateFile, filepath.Join(helper.BottleHelper.RootDir, newFileName)))

	// add sym link directory
	symlinkDir := NewBottleHelper(t)
	symlinkDir.AddArbitraryFileParts(10)
	newDirName := "symlinkedDir/"
	// add another level of indirection (to make sure we traverse through symlinks to symlinks)
	intermediateDir := filepath.Join(t.TempDir(), "intermediate")
	rne(os.Symlink(symlinkFile.RootDir, intermediateDir))
	// the final destination
	rne(os.Symlink(intermediateDir, filepath.Join(helper.BottleHelper.RootDir, newDirName)))

	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")
	helper.CommandHelper.RunCommand("commit")

	// check for 2 parts (1 symlink file, 1 symlink dir)
	helper.BottleHelper.Load()
	bottleParts := helper.BottleHelper.Bottle.GetParts()
	assert.Len(t, bottleParts, 2)
	for _, part := range bottleParts {
		n := part.GetName()
		if !(n == newFileName || n == newDirName) {
			t.Errorf("unexpected bottle part name %s", n)
		}
	}
}

func Test_Functional_SymLinkFileAndDirPostInitThenCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support symlinks")
	}

	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)
	rne := require.New(t).NoError

	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	// add sym link file
	symlinkFile := NewBottleHelper(t)
	filePath := "originalPartLocation"
	symlinkFile.AddBottlePart(filePath)
	newFileName := "symlinkedFile"
	rne(os.Symlink(filepath.Join(symlinkFile.RootDir, filePath), filepath.Join(helper.BottleHelper.RootDir, newFileName)))

	// add sym link directry
	symlinkDir := NewBottleHelper(t)
	symlinkDir.AddArbitraryFileParts(10)
	newDirName := "symlinkedDir/"
	rne(os.Symlink(symlinkDir.RootDir, filepath.Join(helper.BottleHelper.RootDir, newDirName)))

	helper.CommandHelper.RunCommand("commit")

	// check for 2 parts (1 symlink file, 1 symlink dir)
	helper.BottleHelper.Load()
	bottleParts := helper.BottleHelper.Bottle.GetParts()

	assert.Len(t, bottleParts, 2)
	for _, part := range bottleParts {
		n := part.GetName()
		if !(n == newFileName || n == newDirName) {
			t.Errorf("unexpected bottle part name %s", n)
		}
	}
}

func Test_Functional_SymLinkMatchFileName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support symlinks")
	}

	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)
	rne := require.New(t).NoError

	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	// add sym link file
	symlinkFile := NewBottleHelper(t)
	filePath := "originalPartLocation"
	symlinkFile.AddBottlePart(filePath)
	newFileName := "symlinkedFile"
	rne(os.Symlink(filepath.Join(symlinkFile.RootDir, filePath), filepath.Join(helper.BottleHelper.RootDir, newFileName)))

	helper.CommandHelper.RunCommand("commit")

	helper.BottleHelper.Load()
	bottleParts := helper.BottleHelper.Bottle.GetParts()
	assert.Len(t, bottleParts, 1)
	for _, part := range bottleParts {
		n := part.GetName()
		if !(n == newFileName) {
			t.Errorf("unexpected bottle part name %s", n)
		}
	}
}

func Test_Functional_SymLinkOneDirTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support symlinks")
	}

	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)

	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	// add sym link directory
	symlinkDir := NewBottleHelper(t)
	symlinkDir.AddArbitraryFileParts(10)
	newDirName := "symlinkedDir/"
	newDirPath := filepath.Join(helper.BottleHelper.RootDir, newDirName)
	require.NoError(t, os.Symlink(symlinkDir.RootDir, newDirPath))

	helper.CommandHelper.RunCommand("commit")

	helper.BottleHelper.Load()
	bottleParts := helper.BottleHelper.Bottle.GetParts()
	require.Len(t, bottleParts, 1)
	assert.Equal(t, newDirName, bottleParts[0].GetName())

	cfg, err := getConfig(helper.CommandHelper.GetConfigFile())
	require.NoError(t, err)

	// check cachepath, looking for part that is larger than a symlink (check if ace-dt cached the symlink and not the compressed+digested part)
	dgst := bottleParts[0].GetLayerDigest()
	cachePart := filepath.Join(cfg.CachePath, dgst.Algorithm().String(), dgst.Encoded())
	finfo, err := os.Stat(cachePart)
	assert.NoError(t, err)
	assert.Less(t, int64(1000), finfo.Size())
}

func Test_Functional_SymLinkTwoDirTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support symlinks")
	}

	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)
	rne := require.New(t).NoError

	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	helper.BottleHelper.AddArbitraryFileParts(1)

	// add sym link directory
	symlinkDir := NewBottleHelper(t)
	symlinkDir.AddArbitraryFileParts(10)
	newDirName := "symlinkedDir"
	rne(os.Symlink(symlinkDir.RootDir, filepath.Join(helper.BottleHelper.RootDir, newDirName)))

	// add another sym link directory
	symlinkDir2 := NewBottleHelper(t)
	symlinkDir2.AddArbitraryFileParts(10)
	newDirName2 := "symlinkedDir2"
	rne(os.Symlink(symlinkDir2.RootDir, filepath.Join(helper.BottleHelper.RootDir, newDirName2)))

	helper.CommandHelper.RunCommand("commit")

	helper.RequirePartNum(3)
}

func Test_Functional_SymLinkTwoDirTestPreAndPostInit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support symlinks")
	}

	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)
	rne := require.New(t).NoError

	// add sym link directory
	symlinkDir := NewBottleHelper(t)
	symlinkDir.AddArbitraryFileParts(10)
	newDirName := "symlinkedDir"
	rne(os.Symlink(symlinkDir.RootDir, filepath.Join(helper.BottleHelper.RootDir, newDirName)))

	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	// add another sym link directory
	symlinkDir2 := NewBottleHelper(t)
	symlinkDir2.AddArbitraryFileParts(10)
	newDirName2 := "symlinkedDir2"
	rne(os.Symlink(symlinkDir2.RootDir, filepath.Join(helper.BottleHelper.RootDir, newDirName2)))

	helper.CommandHelper.RunCommand("commit")

	helper.RequirePartNum(2)
}

func Test_Functional_SymLinkPostInitThenPush(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support symlinks")
	}

	rootReg := os.Getenv("TEST_REGISTRY")
	if rootReg == "" {
		t.Skip("Skipping because TEST_REGISTRY is not set")
	}

	rootCmd := rootTestCmd()

	helper := NewTestHelper(t, rootCmd)
	rne := require.New(t).NoError

	// add sym link directory
	symlinkDir := NewBottleHelper(t)
	symlinkDir.AddArbitraryFileParts(10)
	newDirName := "symlinkedDir"
	rne(os.Symlink(symlinkDir.RootDir, filepath.Join(helper.BottleHelper.RootDir, newDirName)))

	// set the bottle dir for all other commands to use when running
	helper.CommandHelper.SetBottleDir(helper.BottleHelper.RootDir)
	helper.CommandHelper.RunCommand("init")

	// add another sym link directory
	symlinkDir2 := NewBottleHelper(t)
	symlinkDir2.AddArbitraryFileParts(10)
	newDirName2 := "symlinkedDir2"
	rne(os.Symlink(symlinkDir2.RootDir, filepath.Join(helper.BottleHelper.RootDir, newDirName2)))

	helper.BottleHelper.SetTempBottleRef(rootReg)
	helper.CommandHelper.RunCommand("push", helper.BottleHelper.RegRef)

	helper.BottleHelper.Load()
	bottleParts := helper.BottleHelper.Bottle.GetParts()
	assert.Len(t, bottleParts, 2)

	cfg, err := getConfig(helper.CommandHelper.GetConfigFile())
	assert.NoError(t, err)

	// check cachepath, looking for part that is larger than a symlink (check if ace-dt cached the symlink and not the compressed+digested part)
	dgst0 := bottleParts[0].GetLayerDigest()
	cachePart := filepath.Join(cfg.CachePath, dgst0.Algorithm().String(), dgst0.Encoded())
	finfo, err := os.Stat(cachePart)
	assert.NoError(t, err)
	assert.Less(t, int64(1000), finfo.Size())

	// check cachepath, looking for part that is larger than a symlink (check if ace-dt cached the symlink and not the compressed+digested part)
	dgst := bottleParts[1].GetLayerDigest()
	cachePart = filepath.Join(cfg.CachePath, dgst.Algorithm().String(), dgst.Encoded())
	finfo, err = os.Stat(cachePart)
	assert.NoError(t, err)
	assert.Less(t, int64(1000), finfo.Size())

	assert.NoError(t, helper.CheckRegForBottle(helper.BottleHelper.RegRef, ""))
}
