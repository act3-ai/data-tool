package python

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersionClause(t *testing.T) {

	tests := []struct {
		name    string
		arg     string
		want    *VersionClause
		wantErr error
	}{
		{"match", "==1.2", &VersionClause{"==", "1.2"}, nil},

		{"match spaces 1", "== 1.2", &VersionClause{"==", "1.2"}, nil},
		{"match spaces 2", " ==1.2", &VersionClause{"==", "1.2"}, nil},
		{"match spaces all", " ==   1.2 ", &VersionClause{"==", "1.2"}, nil},

		{"match !=", "!=1.2", &VersionClause{"!=", "1.2"}, nil},
		{"match ===", "===1.2", &VersionClause{"===", "1.2"}, nil},

		{"match long", "==1.2-foo.bar42", &VersionClause{"==", "1.2-foo.bar42"}, nil},
		{"match long", "~=1.2.3a4 ", &VersionClause{"~=", "1.2.3a4"}, nil},
		{"match star", ">1.2.*", &VersionClause{">", "1.2.*"}, nil},

		{"no match", "!==1.2", &VersionClause{"==", "1.2"}, fmt.Errorf(`clause "!==1.2" must match ^\s*(~=|==|!=|<=|>=|<|>|===)\s*([\w-.\*]*)\s*$`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersionClause(tt.arg)
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

func TestParseVersionSpecifier(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    VersionSpecifier
		wantErr error
	}{
		{
			name: "single", arg: "<= 1.2",
			want: VersionSpecifier{
				VersionClause{"<=", "1.2"},
			},
		},
		{
			name: "double", arg: "<= 1.2, > 1.0",
			want: VersionSpecifier{
				VersionClause{"<=", "1.2"},
				VersionClause{">", "1.0"},
			},
		},
		{
			name: "empty", arg: " ",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersionSpecifier(tt.arg)
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
