package oci

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func Test_ParseOCILayoutReference(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantPath string
		wantRef  string
		wantErr  bool
	}{
		{
			name:     "valid tag",
			raw:      "my/layout:latest",
			wantPath: "my/layout",
			wantRef:  "latest",
			wantErr:  false,
		},
		{
			name:     "valid digest",
			raw:      "my/layout@sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantPath: "my/layout",
			wantRef:  "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr:  false,
		},
		{
			name:     "invalid format missing separator",
			raw:      "my/layout-without-tag-or-digest",
			wantPath: "",
			wantRef:  "",
			wantErr:  true,
		},
		{
			name:     "path with colon and digest",
			raw:      "my:layout@sha256:abcdef",
			wantPath: "my:layout",
			wantRef:  "sha256:abcdef",
			wantErr:  false,
		},
		{
			name:     "multiple colons in tag",
			raw:      "my/layout:v1.0:latest",
			wantPath: "my/layout:v1.0",
			wantRef:  "latest",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotRef, err := parseOCILayoutReference(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseOCILayoutReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPath != tt.wantPath {
				t.Errorf("parseOCILayoutReference() gotPath = %v, want %v", gotPath, tt.wantPath)
			}
			if gotRef != tt.wantRef {
				t.Errorf("parseOCILayoutReference() gotRef = %v, want %v", gotRef, tt.wantRef)
			}
		})
	}
}

func Test_Tree_Run(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) string
		wantErr  string
		validate func(*testing.T, string)
	}{
		{
			name: "missing layout separator",
			setup: func(t *testing.T) string {
				t.Helper()

				return "invalid-reference"
			},
			wantErr: "directory path and reference must be separated by",
		},
		{
			name: "nonexistent layout directory",
			setup: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "missing") + ":latest"
			},
			wantErr: "opening OCI image layout directory",
		},
		{
			name: "valid layout digest",
			setup: func(t *testing.T) string {
				t.Helper()
				dir, childDigest := createTreeLayout(t)
				return dir + "@" + childDigest.String()
			},
			validate: func(t *testing.T, output string) {
				t.Helper()
				if output == "" {
					t.Error("Tree.Run printed no output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := tt.setup(t)
			var buf bytes.Buffer
			tree := &Tree{OCILayout: true, Depth: 1}
			err := tree.Run(context.Background(), &buf, ref)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Tree.Run error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Tree.Run expected success, got: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, buf.String())
			}
		})
	}
}
