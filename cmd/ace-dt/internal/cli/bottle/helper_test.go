package bottle

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.act3-ace.com/ace/data/tool/cmd/ace-dt/internal/cli/internal/functesting"
	"git.act3-ace.com/ace/data/tool/internal/bottle"
)

const (
	// KB = 1000 bytes.
	KB = int64(1000)
	// MB = 1,000,000 bytes.
	MB = int64(1000000)
)

// NewBottleHelper returns a BottleHelper with the given bottleDir.
func NewBottleHelper(tb testing.TB) *BottleHelper {
	tb.Helper()
	rsource := rand.NewSource(rand.Int63())
	fsys, err := NewFSUtilWithSource("", rsource)
	if err != nil {
		panic(err)
	}

	// register cleanup function for fsutil close
	tb.Cleanup(func() {
		err := fsys.Close()
		if err != nil {
			tb.Errorf("error closing fsutil: %v", err)
		}
	})

	return &BottleHelper{
		FSUtil: *fsys,
		t:      tb,
		r:      rsource, // We use a new time as the seed to prevent duplicated data.
	}
}

// BottleHelper contains methods to easily create, manipulate, and destroy a bottle.
type BottleHelper struct {
	FSUtil
	RegRef string
	Bottle *bottle.Bottle
	t      testing.TB
	r      rand.Source
}

// Load will load the bottle from b.Dir into b.Bottle (using bottle pkg funcs)
// Expects that bottle location is set, and is indeed a (post init) bottle.
func (b *BottleHelper) Load() {
	b.t.Helper()
	require.NoError(b.t, b.load())
}

func (b *BottleHelper) load() error {
	tmpBottle, err := bottle.LoadBottle(b.RootDir, bottle.WithLocalLabels())
	if err != nil {
		return err
	}
	b.Bottle = tmpBottle
	return nil
}

// SetTempBottleRef will create a temp bottle location based on the given registry. Will validate for you.
func (b *BottleHelper) SetTempBottleRef(registry string) {
	b.t.Helper()
	b.RegRef = functesting.TempOCIRef(registry)
}

// AddPartLabel will add the label to the given part.
// This command must follow both bottle init (OR) commit AND a BottleHelper.load().
func (b *BottleHelper) AddPartLabel(partPath string, key string, value string) {
	b.t.Helper()
	assert.NoError(b.t, b.Bottle.AddPartLabel(context.Background(), key, value, b.Bottle.NativePath(partPath)))
	assert.NoError(b.t, b.Bottle.Save())
}

// AddBottlePart adds a part to bottle by writing the given data to the part path
// Important. Must have a bottle dir set before and then commit changes after.
func (b *BottleHelper) AddBottlePart(partPath string) {
	b.t.Helper()
	require.NoError(b.t, b.AddFileOfSize(partPath, MB))
}

// tempFilePath generates a temporary file path with a given pattern in the specified directory.
// If the pattern includes a "*", the random string replaces the last "*".
// If dir is the empty string, tempFilePath uses the default directory for temporary files.
func tempFilePath(dir, pattern string) (string, error) {
	if dir == "" {
		dir = os.TempDir()
	}

	// add a * to the end of the pattern if it doesn't already have one
	if !strings.Contains(pattern, "*") {
		pattern += "*"
	}

	randomBytes := make([]byte, 8)
	_, err := crand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	randomString := hex.EncodeToString(randomBytes)
	fileName := strings.Replace(pattern, "*", randomString, 1)
	return filepath.Join(dir, fileName), nil
}

// AddArbitraryDirParts writes partNum parts to a bottle, all are directories.
func (b *BottleHelper) AddArbitraryDirParts(partNum int) {
	b.t.Helper()

	for i := 0; i < partNum; i++ {
		// partDir, err := os.MkdirTemp(b.RootDir, "dirPart*") // This does not work, maybe invalid directory characters
		partDir, err := tempFilePath(b.RootDir, "dirPart*")
		require.NoError(b.t, err)

		// let's put some files in the directory
		for i := 0; i < 3; i++ {
			pth := filepath.Join(filepath.Base(partDir), fmt.Sprintf("subfile%d-%d", i, b.r.Int63()))
			require.NoError(b.t, b.AddFileOfSize(pth, MB))
		}
	}
}

// AddArbitraryFileParts writes partNum parts to a bottle, all are directories.
func (b *BottleHelper) AddArbitraryFileParts(count int) {
	b.t.Helper()

	for i := 0; i < count; i++ {
		// make file part
		pth := fmt.Sprintf("file%d-%d", i, b.r.Int63())
		require.NoError(b.t, b.AddFileOfSize(pth, MB))
	}
}

// RemoveBottlePart removes a part from a bottle by deleting that part from the filesystem
// partPath is a part location relative to the bottle root dir.
func (b *BottleHelper) RemoveBottlePart(partPath string) {
	assert.NoError(b.t, b.removeBottlePart(partPath))
}

func (b *BottleHelper) removeBottlePart(partPath string) error {
	trueFilePath := filepath.Clean(filepath.Join(b.RootDir, partPath))
	if !filepath.IsAbs(trueFilePath) {
		return fmt.Errorf("expected correct file path, got root: %v, and part %v, to form: %v", b.RootDir, partPath, trueFilePath)
	}

	if err := os.RemoveAll(trueFilePath); err != nil {
		return fmt.Errorf("error removing bottle part: %w", err)
	}

	return nil
}
