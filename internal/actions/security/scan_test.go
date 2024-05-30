package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScan(t *testing.T) {

	t.Run("Name Uniqueness for Similar Images", func(t *testing.T) {
		image1 := "test1/outlier1/docker.io/image:v1"
		image2 := "test2/outlier2/docker.io/image:v1"

		newImage1, newImage2 := dilineateArtifactNames(image1, image2)
		assert.Equal(t, "outlier1/docker.io/image:v1", newImage1)
		assert.Equal(t, "outlier2/docker.io/image:v1", newImage2)
	})

}
