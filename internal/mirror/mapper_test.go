package mirror

import (
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/act3-ai/data-tool/internal/ref"
)

func Test_getLongestPrefix(t *testing.T) {
	locations := [][]string{
		{"", "high.example.com/default/"},
		{"reg.example.com/low/source1", "high.example.com/one"},
		{"reg.example.com/low/source", "high.example.com/two/"},
		{"reg.example.com/low/sources2", "high.example.com/three/"},
		{"reg.example.com/low/sources2:v2", "high.example.com/four:v3"},
	}

	tests := []struct {
		name string
		sref string
		want string
	}{
		{name: "one match", sref: "reg.example.com/low/source1:v1", want: "high.example.com/one:v1"},
		{name: "full", sref: "reg.example.com/low/sources2:v2", want: "high.example.com/four:v3"},
		{name: "other", sref: "reg2.example.com/other", want: "high.example.com/default/reg2.example.com/other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getLongestPrefix(tt.sref, locations)
			if got != tt.want {
				t.Errorf("Expected %q but got %q", tt.want, got)
			}
		})
	}
}

func TestExtractFromReference(t *testing.T) {
	rne := require.New(t).NoError
	sref := "quay.io/ceph/ceph:v16.2.7"

	t.Run("package", func(t *testing.T) {
		pkg, err := extractPackage(sref)
		rne(err)
		assert.Equal(t, pkg, "ceph/ceph:v16.2.7")
	})

	t.Run("tag", func(t *testing.T) {
		tag, err := extractTag(sref)
		rne(err)
		assert.Equal(t, "v16.2.7", tag)
	})

	t.Run("repository", func(t *testing.T) {
		repo, err := extractRepo(sref)
		rne(err)
		assert.Equal(t, "ceph/ceph", repo)
	})

	t.Run("registry", func(t *testing.T) {
		reg, err := extractReg(sref)
		rne(err)
		assert.Equal(t, "quay.io", reg)
	})

	t.Run("nest mapper", func(t *testing.T) {
		mapperFunc, err := nestMapper("localhost:5000/nest")
		rne(err)

		desc := ocispec.Descriptor{
			Annotations: map[string]string{
				ref.AnnotationSrcRef: "docker.io/testImage",
			},
		}
		dest, err := mapperFunc(desc)
		rne(err)
		assert.Equal(t, 1, len(dest))
		assert.Equal(t, dest[0], "localhost:5000/nest/docker.io/testImage")
	})
}
