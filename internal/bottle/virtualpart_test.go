package bottle

import (
	"os"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeVirtPartData(path string) *VirtualParts {
	vp := NewVirtualPartTracker(path)
	vp.VirtRecords = append(vp.VirtRecords, VirtRecord{
		LayerID:   "sha256:69c4b28d36c92e47004c44dcf9f2a3ea6d2c58322cc2dbd065de711e3c705fbf",
		ContentID: "sha256:69c4b28d36c92e47004c44dcf9f2a3ea6d2c58322cc2dbd065de711e3c705fbf",
	})
	vp.VirtRecords = append(vp.VirtRecords, VirtRecord{
		LayerID:   "sha256:2d3a84006d059d51b6cc0630cfae3a05368bc796d3d59ddd299ca5e512bcee7e",
		ContentID: "sha256:a3bc368571be769c8f49f79f58a7d28ea6ebf303aa1f2b78783fbf7afffe39ee",
	})
	vp.VirtRecords = append(vp.VirtRecords, VirtRecord{
		LayerID:   "sha256:b1edb61290815003b6f299696f6a2c5431a1d4d68fa7c39815ed2ff2f26c8e87",
		ContentID: "sha256:4606989f5dc480174908c8fad859045c20947d94f767233c1cdffc9ab0b51db6",
	})
	return vp
}

func makeVirtPartDataShort(path string) *VirtualParts {
	vp := NewVirtualPartTracker(path)
	vp.VirtRecords = append(vp.VirtRecords, VirtRecord{
		LayerID:   "sha256:69c4b28d36c92e47004c44dcf9f2a3ea6d2c58322cc2dbd065de711e3c705fbf",
		ContentID: "sha256:69c4b28d36c92e47004c44dcf9f2a3ea6d2c58322cc2dbd065de711e3c705fbf",
	})
	vp.VirtRecords = append(vp.VirtRecords, VirtRecord{
		LayerID:   "sha256:2d3a84006d059d51b6cc0630cfae3a05368bc796d3d59ddd299ca5e512bcee7e",
		ContentID: "sha256:a3bc368571be769c8f49f79f58a7d28ea6ebf303aa1f2b78783fbf7afffe39ee",
	})
	return vp
}

func makeJSONString() string {
	return `{
  "virt-records": [
    {
      "layer-id": "sha256:69c4b28d36c92e47004c44dcf9f2a3ea6d2c58322cc2dbd065de711e3c705fbf",
      "content-id": "sha256:69c4b28d36c92e47004c44dcf9f2a3ea6d2c58322cc2dbd065de711e3c705fbf"
    },
    {
      "layer-id": "sha256:2d3a84006d059d51b6cc0630cfae3a05368bc796d3d59ddd299ca5e512bcee7e",
      "content-id": "sha256:a3bc368571be769c8f49f79f58a7d28ea6ebf303aa1f2b78783fbf7afffe39ee"
    },
    {
      "layer-id": "sha256:b1edb61290815003b6f299696f6a2c5431a1d4d68fa7c39815ed2ff2f26c8e87",
      "content-id": "sha256:4606989f5dc480174908c8fad859045c20947d94f767233c1cdffc9ab0b51db6"
    }
  ]
}`
}

func Test_VirtualParts_Save(t *testing.T) {
	d := t.TempDir()
	tests := []struct {
		name       string
		fields     *VirtualParts
		wantWriter string
	}{
		{"write some records", makeVirtPartData(d), makeJSONString()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := &VirtualParts{
				filePath:    tt.fields.filePath,
				VirtRecords: tt.fields.VirtRecords,
			}
			t.TempDir()
			require.NoError(t, vp.Save())
			data, err := os.ReadFile(vp.filePath)
			require.NoError(t, err)
			assert.Equal(t, tt.wantWriter, string(data))
		})
	}
}

func Test_VirtualParts_Load(t *testing.T) {
	d := t.TempDir()
	tests := []struct {
		name    string
		compare *VirtualParts
		arg     string
	}{
		{"read some records", makeVirtPartData(d), makeJSONString()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cvp := &VirtualParts{
				filePath:    tt.compare.filePath,
				VirtRecords: tt.compare.VirtRecords,
			}
			require.NoError(t, os.WriteFile(tt.compare.filePath, []byte(tt.arg), 0o666))
			tvp := &VirtualParts{filePath: tt.compare.filePath}
			require.NoError(t, tvp.Load())
			assert.Equal(t, cvp, tvp)
		})
	}
}

func Test_VirtualParts_Add(t *testing.T) {
	d := t.TempDir()
	type args struct {
		layerID   string
		contentID string
		loc       string
	}
	tests := []struct {
		name string
		vp   *VirtualParts
		args args
		cmp  *VirtualParts
	}{
		{
			"add a record", makeVirtPartDataShort(d),
			args{
				"sha256:b1edb61290815003b6f299696f6a2c5431a1d4d68fa7c39815ed2ff2f26c8e87",
				"sha256:4606989f5dc480174908c8fad859045c20947d94f767233c1cdffc9ab0b51db6",
				"reg.example.com/bottle/different:v2.1",
			},
			makeVirtPartData(d),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ld, err := digest.Parse(tt.args.layerID)
			assert.NoError(t, err)

			cd, err := digest.Parse(tt.args.contentID)
			assert.NoError(t, err)

			tt.vp.Add(ld, cd)
			assert.Equal(t, tt.vp, tt.cmp)
		})
	}
}

func Test_VirtualParts_HasContent(t *testing.T) {
	d := t.TempDir()
	tests := []struct {
		name   string
		fields *VirtualParts
		arg    string
		want   bool
	}{
		{"content found", makeVirtPartData(d), "sha256:4606989f5dc480174908c8fad859045c20947d94f767233c1cdffc9ab0b51db6", true},
		{"content not found", makeVirtPartData(d), "sha256:b1edb61290815003b6f299696f6a2c5431a1d4d68fa7c39815ed2ff2f26c8e87", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := VirtualParts{
				filePath:    tt.fields.filePath,
				VirtRecords: tt.fields.VirtRecords,
			}
			dig, err := digest.Parse(tt.arg)
			assert.NoError(t, err)

			if got := vp.HasContent(dig); got != tt.want {
				t.Errorf("HasContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_VirtualParts_HasLayer(t *testing.T) {
	d := t.TempDir()
	tests := []struct {
		name   string
		fields *VirtualParts
		arg    string
		want   bool
	}{
		{"layer found", makeVirtPartData(d), "sha256:b1edb61290815003b6f299696f6a2c5431a1d4d68fa7c39815ed2ff2f26c8e87", true},
		{"layer not found", makeVirtPartData(d), "sha256:4606989f5dc480174908c8fad859045c20947d94f767233c1cdffc9ab0b51db6", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := VirtualParts{
				filePath:    tt.fields.filePath,
				VirtRecords: tt.fields.VirtRecords,
			}
			dig, err := digest.Parse(tt.arg)
			assert.NoError(t, err)

			if got := vp.HasLayer(dig); got != tt.want {
				t.Errorf("HasLayer() = %v, want %v", got, tt.want)
			}
		})
	}
}
