package bottle

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getRootBottlePath(t *testing.T) {
	type args struct {
		path string
	}

	dir := t.TempDir()
	mkbottle := func(path ...string) {
		path = append(path, ".dt")
		dt := filepath.Join(path...)
		require.NoError(t, os.MkdirAll(dt, 0777))
	}
	mkbottle(dir, "foo")
	mkbottle(dir, "foo", "bar")
	mkbottle(dir, "mydir")

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{"current", args{filepath.Join(dir, "foo")}, filepath.Join(dir, "foo"), nil},
		{"not-found", args{dir}, "", errNoRootBottleFound},
		{"same", args{filepath.Join(dir, "foo", "bar")}, filepath.Join(dir, "foo", "bar"), nil},
		{"parent", args{filepath.Join(dir, "foo", ".dt")}, filepath.Join(dir, "foo"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRootBottlePath(tt.args.path)
			if err != nil && tt.wantErr == nil {
				assert.Fail(t, fmt.Sprintf(
					"Error not expected but got one:\n"+
						"error: %q", err),
				)
				return
			}
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
