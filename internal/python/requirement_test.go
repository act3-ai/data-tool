package python

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRequirement(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name    string
		args    args
		want    *Requirement
		wantErr error
	}{
		{
			"name",
			args{`fsspec`},
			&Requirement{
				Name: "fsspec",
			},
			nil,
		},
		{
			"extra",
			args{`fsspec[myextra,foo]`},
			&Requirement{
				Name:   "fsspec",
				Extras: []string{"myextra", "foo"},
			},
			nil,
		},
		{
			"version",
			args{`fsspec==1.2.3`},
			&Requirement{
				Name:             "fsspec",
				VersionSpecifier: VersionSpecifier{VersionClause{"==", "1.2.3"}},
			},
			nil,
		},
		{
			"version2",
			args{`fsspec >= 2.8.1, == 2.8.*`},
			&Requirement{
				Name:             "fsspec",
				VersionSpecifier: VersionSpecifier{VersionClause{">=", "2.8.1"}, VersionClause{"==", "2.8.*"}},
			},
			nil,
		},
		{
			"extra and space",
			args{`fsspec [myextra]`},
			&Requirement{
				Name:   "fsspec",
				Extras: []string{"myextra"},
			},
			nil,
		},
		{
			"hash",
			args{`fsspec==2022.8.2 ; python_version >= "3.8" and python_version < "3.11" --hash=sha256:6374804a2c0d24f225a67d009ee1eabb4046ad00c793c3f6df97e426c890a1d9 --hash=sha256:7f12b90964a98a7e921d27fb36be536ea036b73bf3b724ac0b0bd7b8e39c7c18`},
			&Requirement{
				Name:             "fsspec",
				VersionSpecifier: VersionSpecifier{VersionClause{"==", "2022.8.2"}},
				Constraints:      `python_version >= "3.8" and python_version < "3.11"`,
				Digests: map[digest.Digest]struct{}{
					digest.Digest("sha256:6374804a2c0d24f225a67d009ee1eabb4046ad00c793c3f6df97e426c890a1d9"): {},
					digest.Digest("sha256:7f12b90964a98a7e921d27fb36be536ea036b73bf3b724ac0b0bd7b8e39c7c18"): {},
				},
			},
			nil,
		},
		{
			"hash with options",
			args{`brotli==1.0.9 ; platform_python_implementation == "CPython" and python_version >= "3.8" and python_version < "3.11" --hash=sha256:12effe280b8ebfd389022aa65114e30407540ccb89b177d3fbc9a4f177c4bd5d`},
			&Requirement{
				Name:             "brotli",
				VersionSpecifier: VersionSpecifier{VersionClause{"==", "1.0.9"}},
				Constraints:      `platform_python_implementation == "CPython" and python_version >= "3.8" and python_version < "3.11"`,
				Digests: map[digest.Digest]struct{}{
					digest.Digest("sha256:12effe280b8ebfd389022aa65114e30407540ccb89b177d3fbc9a4f177c4bd5d"): {},
				},
			},
			nil,
		},
		{
			"hash with extras",
			args{`brotli[x,y]==1.0.9 ; platform_python_implementation == "CPython" and python_version >= "3.8" and python_version < "3.11" --hash=sha256:12effe280b8ebfd389022aa65114e30407540ccb89b177d3fbc9a4f177c4bd5d`},
			&Requirement{
				Name:             "brotli",
				VersionSpecifier: VersionSpecifier{VersionClause{"==", "1.0.9"}},
				Extras:           []string{"x", "y"},
				Constraints:      `platform_python_implementation == "CPython" and python_version >= "3.8" and python_version < "3.11"`,
				Digests: map[digest.Digest]struct{}{
					digest.Digest("sha256:12effe280b8ebfd389022aa65114e30407540ccb89b177d3fbc9a4f177c4bd5d"): {},
				},
			},
			nil,
		},
		{
			"hash with extras and spaces",
			args{`brotli[x,y] == 1.0.9 ; platform_python_implementation == "CPython" and python_version >= "3.8" and python_version < "3.11" --hash=sha256:12effe280b8ebfd389022aa65114e30407540ccb89b177d3fbc9a4f177c4bd5d`},
			&Requirement{
				Name:             "brotli",
				VersionSpecifier: VersionSpecifier{VersionClause{"==", "1.0.9"}},
				Extras:           []string{"x", "y"},
				Constraints:      `platform_python_implementation == "CPython" and python_version >= "3.8" and python_version < "3.11"`,
				Digests: map[digest.Digest]struct{}{
					digest.Digest("sha256:12effe280b8ebfd389022aa65114e30407540ccb89b177d3fbc9a4f177c4bd5d"): {},
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRequirement(tt.args.line)
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

func TestParseRequirementsFile(t *testing.T) {
	rr := Requirements{}
	err := rr.ParseRequirementsFile(filepath.Join("testdata", "basic-requirements.txt"))
	require.NoError(t, err)

	require.NotEmpty(t, rr.reqs)
	require.Contains(t, rr.reqs, "absl-py")
	reqs := rr.reqs["absl-py"]
	require.Len(t, reqs, 1)
	assert.Equal(t, "absl-py", reqs[0].Name)
	assert.Equal(t, map[digest.Digest]struct{}{
		digest.Digest("sha256:5d15f85b8cc859c6245bc9886ba664460ed96a6fee895416caa37d669ee74a9a"): {},
		digest.Digest("sha256:f568809938c49abbda89826223c992b630afd23c638160ad7840cfe347710d97"): {},
	}, reqs[0].Digests)

	assert.Equal(t, "http://example.com", rr.IndexURL)
	assert.Equal(t, []string{"https://foo.com", "https://other.com"}, rr.ExtraIndexURLs)

	assert.Contains(t, rr.reqs, "pkg")
}

func TestRequirement_String(t *testing.T) {
	tests := []struct {
		name string
		req  Requirement
		want string
	}{
		{
			name: "name-only",
			req: Requirement{
				Name: "there",
			},
			want: `there`,
		},
		{
			name: "all",
			req: Requirement{
				Name:             "brotli",
				VersionSpecifier: VersionSpecifier{VersionClause{"==", "1.0.9"}},
				Extras:           []string{"x", "y"},
				Constraints:      `platform_python_implementation == "CPython" and python_version >= "3.8" and python_version < "3.11"`,
				Digests: map[digest.Digest]struct{}{
					digest.Digest("sha256:12effe280b8ebfd389022aa65114e30407540ccb89b177d3fbc9a4f177c4bd5d"): {},
					digest.Digest("sha256:42effe280b8ebfd389022aa65114e30407540ccb89b177d3fbc9a4f177c4bd5d"): {},
				},
			},
			want: `brotli[x,y] ==1.0.9 ; platform_python_implementation == "CPython" and python_version >= "3.8" and python_version < "3.11" \
	--hash=sha256:12effe280b8ebfd389022aa65114e30407540ccb89b177d3fbc9a4f177c4bd5d \
	--hash=sha256:42effe280b8ebfd389022aa65114e30407540ccb89b177d3fbc9a4f177c4bd5d`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.req.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
