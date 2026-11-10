package oci

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestParseOCILayoutReference(t *testing.T) {
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

func TestTree_Run_Errors(t *testing.T) {
	// Test parsing error for OCILayout
	tree := &Tree{
		OCILayout: true,
	}
	var buf bytes.Buffer
	
	err := tree.Run(context.Background(), &buf, "invalid-reference")
	if err == nil {
		t.Error("expected error for invalid OCI layout reference, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "directory path and reference must be separated by") {
		t.Errorf("unexpected error message: %v", err)
	}

	// Test non-existent directory error for OCILayout
	err = tree.Run(context.Background(), &buf, "non-existent-dir:latest")
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}