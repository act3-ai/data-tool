package python

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func whlFilename(whl *wheelDistribution) string {
	parts := make([]string, 0, 6)
	parts = append(parts,
		whl.project,
		whl.version,
	)
	if whl.build != nil {
		parts = append(parts, *whl.build)
	}
	parts = append(parts,
		strings.Join(whl.python, "."),
		strings.Join(whl.abi, "."),
		strings.Join(whl.platform, "."),
	)
	return strings.Join(parts, "-") + ".whl"
}

func Test_newWheel(t *testing.T) {
	build := "43"
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    Distribution
		wantErr error
	}{
		{
			"long",
			args{"pyzmq-23.2.1-pp39-pypy39_pp73-manylinux_2_17_x86_64.manylinux2014_x86_64.whl"},
			&wheelDistribution{
				project:  "pyzmq",
				version:  "23.2.1",
				python:   []string{"pp39"},
				abi:      []string{"pypy39_pp73"},
				platform: []string{"manylinux_2_17_x86_64", "manylinux2014_x86_64"},
			},
			nil,
		},
		{
			"source",
			args{"pyzmq-23.2.1.tar.gz"},
			nil,
			fmt.Errorf("pyzmq-23.2.1.tar.gz is not a wheel: %w", ErrInvalidPythonDistributionFilename),
		},
		{
			"build",
			args{"pyzmq-23.2.1-43-pp39.pp310-pypy39_pp73-manylinux_2_17_x86_64.manylinux2014_x86_64.whl"},
			&wheelDistribution{
				project:  "pyzmq",
				version:  "23.2.1",
				build:    &build,
				python:   []string{"pp39", "pp310"},
				abi:      []string{"pypy39_pp73"},
				platform: []string{"manylinux_2_17_x86_64", "manylinux2014_x86_64"},
			},
			nil,
		},
		{
			"universal",
			args{"pip-22.1.2-py3-none-any.whl"},
			&wheelDistribution{
				project:  "pip",
				version:  "22.1.2",
				python:   []string{"py3"},
				abi:      []string{"none"},
				platform: []string{"any"},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newWheel(tt.args.filename)
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

			assert.Equal(t, tt.args.filename, whlFilename(got), "round trip")
		})
	}
}

// Run this fuzz test with
// go test -fuzz=Fuzz -fuzztime 30s ./pkg/python
func Fuzz_parseWheelFilename(f *testing.F) {
	f.Add("pyzmq-23.2.1-43-pp39.pp310-pypy39_pp73-manylinux_2_17_x86_64.manylinux2014_x86_64.whl")
	f.Add("pyzmq-23.2.1-43-pp39.pp310-_64.whl")
	f.Add("pyzmq-23.2.1-43-pp39.pp310-_64.not-wheel")
	f.Add("")
	f.Fuzz(func(t *testing.T, orig string) {
		whl, err := newWheel(orig)
		if err != nil {
			return
		}
		origAgain := whlFilename(whl)
		assert.Equal(t, orig, origAgain)
	})
}

/*
func TestGoOSArch(t *testing.T) {
	type args struct {
		platform string
	}
	tests := []struct {
		name     string
		args     args
		wantOs   string
		wantArch string
	}{
		{"linux", args{"manylinux2010_x86_64"}, "linux", "amd64"},
		{"musl", args{"musllinux_2_17_ppc64"}, "linux", "ppc64"},
		// {"win", args{"win_amd64"}, "win", "i686"},
		{"darwin", args{"macosx_10_15_x86_64"}, "darwin", "amd64"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOs, gotArch := GoOSArch(tt.args.platform)
			assert.Equal(t, tt.wantOs, gotOs)
			assert.Equal(t, tt.wantArch, gotArch)
		})
	}
}*/

func Test_newSourceDistribution(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    *sourceDistribution
		wantErr error
	}{
		{"tar.gz", args{"pyzmq-14.0.0.tar.gz"}, &sourceDistribution{project: "pyzmq", version: "14.0.0"}, nil},
		{"zip", args{"pyzmq-14.2.0.zip"}, &sourceDistribution{project: "pyzmq", version: "14.2.0"}, nil},
		{"dash", args{"aiohttp-cors-0.7.0.tar.gz"}, &sourceDistribution{project: "aiohttp-cors", version: "0.7.0"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newSourceDistribution(tt.args.filename)
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

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"basic", "aiohttp", "aiohttp"},
		{"dash", "aiohttp-cors", "aiohttp-cors"},
		{"dash", "aiohttp_cors", "aiohttp-cors"},
		{"upper", "aiohttp_Cors", "aiohttp-cors"},
		{"duplicate", "a.--.b_C", "a-b-c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.arg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_decodePlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		wantOs   string
		wantArch string
	}{
		{"windows", "win32", "windows", ""},
		{"windows 64", "win_amd64", "windows", "amd64"},
		{"linux", "manylinux_2_17_ppc64", "manylinux", "ppc64"},
		{"linux", "musllinux_1_1_x86_64", "musllinux", "x86_64"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOs, gotArch := decodePlatform(tt.platform)
			assert.Equal(t, tt.wantOs, gotOs)
			assert.Equal(t, tt.wantArch, gotArch)
		})
	}
}
