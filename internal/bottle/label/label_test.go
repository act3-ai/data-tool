package label

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeLabelsTest() *partLabels {
	return &partLabels{
		Labels: map[string]map[string]string{
			"fileA": {
				"key1": "val1",
			},
			"fileB": {
				"key2": "val2",
			},
		},
	}
}
func makeLegacyTest() *partLabels {
	return &partLabels{
		Labels: map[string]map[string]string{
			"bigfile": {
				"type": "main",
			},
		},
	}
}

func matchFileLabels(t *testing.T, got *partLabels, want *partLabels) {
	t.Helper()
	require.NotNil(t, got)
	require.NotNil(t, want)
	if len(want.Labels) != len(got.Labels) {
		t.Errorf("mismatched label map size, want %d, got %d", len(want.Labels), len(got.Labels))
	}
	for k, wlbl := range want.Labels {
		glbl, ok := got.Labels[k]
		if !ok {
			t.Errorf("missing label file ref, want %s", k)
		}
		for lk, wv := range wlbl {
			v, ok := glbl[lk]
			if !ok {
				t.Errorf("missing label key in fileref %s, want %s", k, lk)
			}
			if v != wv {
				t.Errorf("mismatched label value for fileref %s, want %s got %s", k, wv, v)
			}
		}
	}
}

func Test_Provider_LoadFromPath(t *testing.T) {

	fsys := fstest.MapFS{
		".labels.yaml": &fstest.MapFile{
			Data: []byte("{}\n"),
		},
		"part/.labels.yaml": &fstest.MapFile{
			Data: []byte("{}\n"),
		},
		"part-with-labels/.labels.yaml": &fstest.MapFile{
			Data: []byte("labels:\n    fileA:\n        key1: val1\n    fileB:\n        key2: val2\n"),
		},
		"part-legacy/.labels.yml": &fstest.MapFile{
			Data: []byte("bigfile:\n    type: main\n"),
		},
		"part-other/foo": &fstest.MapFile{
			Data: []byte("why are you looking here?"),
		},
	}

	subparts, err := HasSubparts(fsys, "part-other")
	assert.NoError(t, err)
	assert.False(t, subparts)

	subparts, err = HasSubparts(fsys, ".")
	assert.NoError(t, err)
	assert.True(t, subparts)

	subparts, err = HasSubparts(fsys, "part")
	assert.NoError(t, err)
	assert.True(t, subparts)

	subparts, err = HasSubparts(fsys, "part-legacy")
	assert.NoError(t, err)
	assert.True(t, subparts)

	subparts, err = HasSubparts(fsys, "part-not-here")
	assert.Error(t, err)
	assert.False(t, subparts)

	p, err := NewProviderFromFS(fsys)
	require.NoError(t, err)

	tests := []struct {
		name    string
		match   *partLabels
		pth     string
		wantErr bool
	}{
		{"load empty", &partLabels{}, "part", false},
		{"load labels", makeLabelsTest(), "part-with-labels", false},
		{"load legacy", makeLegacyTest(), "part-legacy", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchFileLabels(t, p.pathLabels[tt.pth], tt.match)
		})
	}
}

func Test_Provider_LabelsForPart(t *testing.T) {
	tests := []struct {
		name    string
		fname   string
		wantKey string
		wantVal string
		search  *partLabels
		wantErr bool
	}{
		{name: "proc label empty", fname: "nofile", wantKey: "", wantVal: "", search: &partLabels{}},
		{name: "proc label notfound", fname: "nofile", wantKey: "", wantVal: "", search: makeLabelsTest()},
		{name: "proc label simple1", fname: "fileA", wantKey: "key1", wantVal: "val1", search: makeLabelsTest()},
		{name: "proc label simple2", fname: "fileB", wantKey: "key2", wantVal: "val2", search: makeLabelsTest()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				pathLabels: make(map[string]*partLabels),
			}
			p.pathLabels["path/to/data"] = tt.search

			lbl := p.LabelsForPart("path/to/data/" + tt.fname)
			assert.Equal(t, lbl.Get(tt.wantKey), tt.wantVal)
		})
	}
}

func Test_splitPartName(t *testing.T) {
	tests := []struct {
		name     string
		partName string
		wantP    string
		wantN    string
	}{
		{"top file", "file.txt", ".", "file.txt"},
		{"top dir", "my-dir/", ".", "my-dir/"},
		{"nested file", "path/to/file.txt", "path/to", "file.txt"},
		{"nested dir", "path/to/dir/", "path/to", "dir/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP, gotN := splitPartName(tt.partName)
			assert.Equal(t, tt.wantP, gotP)
			assert.Equal(t, tt.wantN, gotN)
		})
	}
}
