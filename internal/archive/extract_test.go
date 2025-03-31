package archive

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO these should be files in the testdata directory with a script to reproduce them

var (
	dataUncompressed = []byte{116, 104, 105, 115, 32, 105, 115, 32, 97, 32, 116, 101, 115, 116, 32, 102, 111, 114, 32, 122, 115, 116, 100, 32, 99, 111, 109, 112, 114, 101, 115, 115, 105, 111, 110, 10}
	dataCompressed   = []byte{40, 181, 47, 253, 4, 88, 33, 1, 0, 116, 104, 105, 115, 32, 105, 115, 32, 97, 32, 116, 101, 115, 116, 32, 102, 111, 114, 32, 122, 115, 116, 100, 32, 99, 111, 109, 112, 114, 101, 115, 115, 105, 111, 110, 10, 234, 95, 190, 146}
	// dataTarGzip        = []byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 3, 237, 212, 223, 10, 130, 48, 20, 199, 241, 93, 247, 20, 123, 130, 218, 230, 254, 60, 143, 136, 161, 213, 85, 174, 232, 241, 115, 129, 177, 32, 235, 70, 13, 233, 251, 97, 176, 193, 4, 143, 252, 60, 39, 214, 93, 44, 207, 85, 211, 94, 235, 157, 152, 137, 234, 133, 224, 210, 174, 131, 51, 249, 62, 16, 186, 8, 214, 171, 66, 187, 254, 172, 180, 81, 202, 11, 233, 230, 42, 40, 119, 73, 223, 47, 165, 56, 148, 213, 241, 211, 115, 223, 238, 87, 42, 102, 249, 167, 243, 190, 61, 213, 122, 27, 111, 113, 194, 119, 164, 128, 189, 183, 227, 249, 235, 48, 228, 175, 76, 250, 79, 180, 181, 206, 9, 169, 38, 172, 97, 212, 191, 231, 223, 180, 157, 236, 87, 41, 83, 250, 155, 95, 151, 131, 133, 189, 235, 127, 179, 116, 255, 155, 231, 252, 79, 19, 224, 209, 255, 69, 160, 255, 151, 240, 218, 255, 134, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 192, 10, 221, 1, 181, 23, 240, 105, 0, 40, 0, 0}
	dataTarZstd        = []byte{40, 181, 47, 253, 4, 88, 189, 4, 0, 130, 135, 22, 23, 144, 135, 13, 208, 157, 25, 119, 183, 127, 228, 59, 198, 28, 114, 115, 73, 166, 211, 209, 56, 134, 29, 31, 197, 37, 185, 147, 223, 21, 171, 104, 235, 148, 166, 210, 116, 188, 140, 88, 138, 221, 152, 247, 180, 98, 204, 38, 22, 96, 168, 144, 4, 171, 104, 77, 48, 206, 192, 85, 134, 179, 122, 48, 150, 205, 127, 118, 255, 104, 228, 166, 36, 49, 31, 166, 253, 55, 102, 254, 191, 3, 8, 145, 14, 195, 112, 86, 209, 6, 21, 0, 253, 13, 122, 208, 14, 84, 28, 7, 20, 6, 98, 128, 2, 27, 72, 121, 128, 76, 219, 98, 174, 6, 128, 110, 165, 74, 24, 80, 10, 188, 24, 216, 3, 3, 244, 7, 8, 46, 104, 128, 130, 161, 112, 0, 132, 63, 0, 169, 193, 11, 172, 112, 53, 77, 224, 11, 136, 245, 114, 242}
	dataComponent1     = []byte{116, 104, 105, 115, 32, 105, 115, 32, 97, 32, 116, 101, 115, 116, 10}
	dataComponent2     = []byte{116, 104, 105, 115, 32, 105, 115, 32, 97, 32, 116, 101, 115, 116, 50, 10}
	fileName           = "test1.txt"
	fileNameCompressed = "test1.zst"
	// fileNameTarGzip    = "testarchive.tar.gz"
	fileNameTarZstd    = "testarchive.tar.zst"
	fileNameComponent1 = "testarchive/testfile1.txt"
	fileNameComponent2 = "testarchive/testfile2.txt"
)

type fsmap map[string][]byte

// Create as many temp files as needed from fileMap, where fileMap is
// desired filePath: fileData
func setup(t *testing.T, fileMap fsmap) (string, []string) {
	t.Helper()
	files := []string{}
	dir := t.TempDir()
	for k, v := range fileMap {
		newFile := filepath.Join(dir, k)
		assert.NoError(t, os.WriteFile(newFile, v, 0644))
		files = append(files, newFile)
	}
	return dir, files
}

func Test_ExtractZstdBasic(t *testing.T) {
	fileMap := fsmap{fileNameCompressed: dataCompressed}
	dir, files := setup(t, fileMap)

	for _, file := range files {
		f, err := os.Open(file)
		assert.NoError(t, err)
		assert.NoError(t, ExtractZstd(context.Background(), f, filepath.Join(dir, fileName)))
	}
}

func Test_ExtractZstComparison(t *testing.T) {
	fileMap := fsmap{fileNameCompressed: dataCompressed}
	dir, files := setup(t, fileMap)

	for _, file := range files {
		f, err := os.Open(file)
		assert.NoError(t, err)
		assert.NoError(t, ExtractZstd(context.Background(), f, filepath.Join(dir, fileName)))
		// Check file to see if its data matches dataUncompressed
		data, readErr := os.ReadFile(filepath.Join(dir, fileName))
		assert.NoError(t, readErr)
		assert.Equal(t, dataUncompressed, data)
	}
}

func Test_ExtractTarZstdBasic(t *testing.T) {
	fileMap := fsmap{fileNameTarZstd: dataTarZstd}
	dir, files := setup(t, fileMap)

	for _, file := range files {
		f, err := os.Open(file)
		assert.NoError(t, err)
		assert.NoError(t, ExtractTarZstd(context.Background(), f, dir))
	}

	// Check file to see if its data matches dataUncompressed
	data1, readErr1 := os.ReadFile(filepath.Join(dir, fileNameComponent1))
	assert.NoError(t, readErr1)
	assert.Equal(t, dataComponent1, data1)
	// Check file to see if its data matches dataUncompressed
	data2, readErr2 := os.ReadFile(filepath.Join(dir, fileNameComponent2))
	assert.NoError(t, readErr2)
	assert.Equal(t, dataComponent2, data2)
}
