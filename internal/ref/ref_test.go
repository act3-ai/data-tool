package ref

import (
	_ "crypto/sha256"
	"testing"
)

var (
	testNameTagTable = []struct {
		nametag string
		name    string
		tag     string
		digest  string
	}{
		{
			"reg.example.com/bottle/mnist:v1.3",
			"reg.example.com/bottle/mnist",
			"v1.3",
			"",
		},
		{
			"reg.git.act3-ace.com/bottle/mnist",
			"reg.git.act3-ace.com/bottle/mnist",
			"latest",
			"",
		},
		{
			"reg.example.com/bottle/mnist:normalTag",
			"reg.example.com/bottle/mnist",
			"normalTag",
			"",
		},
		{
			"reg.git.act3-ace.com/bottle/mnist@sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391",
			"reg.git.act3-ace.com/bottle/mnist",
			"",
			"sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391",
		},
		{
			"reg.example.com/bottle/mnist:v1.3@sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391",
			"reg.example.com/bottle/mnist",
			"v1.3",
			"sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391",
		},
		{
			"reg.example.com/bottle/mnist:v1.3@sha256:superbadsha",
			"reg.example.com/bottle/mnist",
			"v1.3",
			"",
		},
	}

	testRefTable = []struct {
		ref      Ref
		testref  string
		expected string
		wantErr  bool
	}{
		{
			ref:      Ref{Reg: "reg.example.com", Repo: "jack", Name: "distributed", Tag: "v1.3"},
			testref:  "reg.example.com/jack/distributed:v1.3",
			expected: "reg.example.com/jack/distributed:v1.3",
		},
		{
			ref:      Ref{Reg: "reg.git.act3-ace.ai", Repo: "bottle", Name: "mnist", Tag: "latest"},
			testref:  "reg.git.act3-ace.ai/bottle/mnist:latest",
			expected: "reg.git.act3-ace.ai/bottle/mnist:latest",
		},
		{
			ref:      Ref{Reg: "reg.git.act3-ace.ai", Name: "accothing", Tag: "v1.02-Dirty"},
			testref:  "reg.git.act3-ace.ai/accoThing:v1.02-Dirty",
			expected: "reg.git.act3-ace.ai/accothing:v1.02-Dirty",
			wantErr:  true,
		},
		{
			ref:      Ref{Reg: "reg.git.act3-ace.ai", Name: "accothing", Tag: "v1.02-Dirty"},
			testref:  "reg.git.act3-ace.ai/accothing:v1.02-Dirty",
			expected: "reg.git.act3-ace.ai/accothing:v1.02-Dirty",
		},
		{
			ref:      Ref{Reg: "reg.example.com", Repo: "jack", Name: "distributed", Tag: "v1.3", Digest: "sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391"},
			testref:  "reg.example.com/jack/distributed:v1.3@sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391",
			expected: "reg.example.com/jack/distributed:v1.3@sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391",
		},
		{
			ref:      Ref{Scheme: "http", Reg: "localhost:5000", Repo: "jack", Name: "distributed", Tag: "v1.3", Digest: "sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391"},
			testref:  "http://localhost:5000/jack/distributed:v1.3@sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391",
			expected: "localhost:5000/jack/distributed:v1.3@sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391",
		},
	}
)

func Test_GetNameTagAndDigest(t *testing.T) {
	for _, testNameTag := range testNameTagTable {
		name, tag, digest := getNameTagAndDigest(testNameTag.nametag)
		if name != testNameTag.name {
			t.Errorf("Expected %s but received %s for nametag: %s's name", testNameTag.name, name, testNameTag.nametag)
		}
		if tag != testNameTag.tag {
			t.Errorf("Expected %s but received %s for nametag: %s's tag", testNameTag.tag, tag, testNameTag.nametag)
		}
		if digest != testNameTag.digest {
			t.Errorf("Expected %s but received %s for nametag: %s's digest", testNameTag.digest, digest, testNameTag.nametag)
		}
	}
}

func Test_RefString(t *testing.T) {
	for _, testRef := range testRefTable {
		refString := testRef.ref.String()
		if refString != testRef.expected {
			t.Errorf("Expected %s but received %s", testRef.expected, refString)
		}
	}
}

func Test_RefFromString(t *testing.T) {
	_, err := FromString("")
	if err == nil {
		t.Error("Expected error for no registry specified")
	}
	_, err = FromString("bad hostname/repo:v1")
	if err == nil {
		t.Error("expected error for registry format invalid")
	}
	_, err = FromString("/")
	if err == nil {
		t.Error("expected error for no name specified")
	}
	_, err = FromString("/jack/distributed:v1.3@sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391")
	if err == nil {
		t.Error("expected error for no registry specified")
	}
	// This case is checked twice in RefFromString
	_, err = FromString("/bottle/test@sha256:bd055c27d897f327ddb4e846782fccfda3b1d3d2397d5c764d8038547acde391")
	if err == nil {
		t.Error("expected error for no repository specified")
	}
	for _, testRef := range testRefTable {
		ref, err := FromString(testRef.testref)
		if testRef.wantErr {
			if err == nil {
				t.Errorf("Expected parse failure, but parse succeeded for %s", ref)
			} else {
				continue
			}
		}
		if err != nil {
			t.Error(err)
		}
		if ref.String() != testRef.ref.String() {
			t.Errorf("Expected %s but received %s", testRef.ref, ref)
		}
		// Scheme is not included in the string() conversion for ref, so test it separately
		if ref.Scheme != testRef.ref.Scheme {
			t.Errorf("Expected %s but received %s", testRef.ref.Scheme, ref.Scheme)
		}
	}
}

func TestRef_IsEmpty(t *testing.T) {
	type fields struct {
		Scheme    string
		Reg       string
		Repo      string
		Name      string
		Tag       string
		Digest    string
		Query     string
		Selectors []string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "true test", fields: fields{}, want: true},
		{name: "false string", fields: fields{Scheme: "http"}, want: false},
		{name: "false string slice", fields: fields{Selectors: []string{"selector"}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Ref{
				Scheme:    tt.fields.Scheme,
				Reg:       tt.fields.Reg,
				Repo:      tt.fields.Repo,
				Name:      tt.fields.Name,
				Tag:       tt.fields.Tag,
				Digest:    tt.fields.Digest,
				Query:     tt.fields.Query,
				Selectors: tt.fields.Selectors,
			}
			if got := r.IsEmpty(); got != tt.want {
				t.Errorf("Ref.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
