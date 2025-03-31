// Package ref provides functionality for interpreting and manipulating OCI references, including delineated data for
// scheme, registry, repository, name, tag, and digest that identifies an OCI object (or bottle).  Additionally,
// facilities for indexing a repository to retrieve a list of tags is included.
package ref

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/opencontainers/go-digest"
)

// Ref is a bottle location reference, including scheme, registry, repository Name, and tag.
// The Query and Fragment members allows additional uri query terms to be parsed and stored.
type Ref struct {
	Scheme    string
	Reg       string
	Repo      string
	Name      string
	Tag       string
	Digest    string
	Query     string
	Selectors []string
}

// IsEmpty will iterate over the fields in the given Ref struct r and return false if any field is a non-zero value.
func (r Ref) IsEmpty() bool {
	if len(r.Selectors) != 0 {
		return false
	}
	if r.Reg != "" {
		return false
	}
	if r.Repo != "" {
		return false
	}
	if r.Name != "" {
		return false
	}
	if r.Tag != "" {
		return false
	}
	if r.Digest != "" {
		return false
	}
	if r.Scheme != "" {
		return false
	}
	if r.Query != "" {
		return false
	}
	return true
}

// String for Ref returns a string representation of a reference
// Note, this leaves off the scheme if defined.
func (r Ref) String() string {
	retstr := r.RepoString()
	if r.Tag != "" {
		retstr += ":" + r.Tag
	}
	if r.Digest != "" {
		retstr += "@" + r.Digest
	}
	return retstr
}

// StringWithScheme for Ref returns a string representation of a reference including the scheme if defined.
func (r Ref) StringWithScheme() string {
	retstr := r.RepoStringWithScheme()
	if r.Tag != "" {
		retstr += ":" + r.Tag
	}
	if r.Digest != "" {
		retstr += "@" + r.Digest
	}
	return retstr
}

// APIString for Ref returns an API string representation of a reference, including the 'v2' path element
// Note, this leaves off the scheme if defined.
func (r Ref) APIString() string {
	retstr := ""
	switch r.Repo {
	case "":
		retstr = r.Reg + "/v2/" + r.Name
	case r.Name:
		retstr = r.Reg + "/v2/" + r.Repo
	default:
		retstr = r.Reg + "/v2/" + r.Repo + "/" + r.Name
	}
	return retstr
}

// RepoString returns the full repository value, including repo and name as appropriate.
func (r Ref) RepoString() string {
	retstr := ""
	if r.Repo != "" && r.Repo != r.Name {
		retstr = r.Reg + "/" + r.Repo + "/" + r.Name
	} else {
		retstr = r.Reg + "/" + r.Name
	}
	return retstr
}

// RepoStringWithScheme returns the full repository value, including repo and name as appropriate, and including scheme
// for registry name.
func (r Ref) RepoStringWithScheme() string {
	retstr := ""
	if r.Repo != "" && r.Repo != r.Name {
		retstr = r.Reg + "/" + r.Repo + "/" + r.Name
	} else {
		retstr = r.Reg + "/" + r.Name
	}
	if r.Scheme != "" {
		retstr = r.Scheme + "://" + retstr
	}
	return retstr
}

// MountRef returns a reference used by oras to discover cross-repo mount locations.
// https://github.com/opencontainers/distribution-spec/blob/main/spec.md#mounting-a-blob-from-another-repository.
func (r Ref) MountRef() string {
	retstr := ""
	if r.Repo != "" && r.Repo != r.Name {
		retstr = r.Repo + "/" + r.Name
	} else {
		retstr = r.Name
	}

	return retstr
}

// CombineRepoName combines the repo and name values into one, unless repo and name are already the same.
func (r Ref) CombineRepoName() string {
	retstr := ""
	if r.Repo != "" && r.Repo != r.Name {
		retstr = r.Repo + "/" + r.Name
	} else {
		retstr = r.Name
	}
	return retstr
}

// URL returns a ref formatted as a URL (including the scheme).
func (r Ref) URL() string {
	tmpref := r
	tmpref.Tag = ""
	tmpref.Digest = ""
	if tmpref.Scheme == "" {
		tmpref.Scheme = "https"
	}
	return tmpref.Scheme + "://" + tmpref.String()
}

// AuthURL returns a ref formatted as the base url used to authenticate a user with a registry.
// Example: https://us-central1-docker.pkg.dev
func (r Ref) AuthURL() string {
	tmpref := r
	if tmpref.Scheme == "" {
		tmpref.Scheme = "https"
	}
	return tmpref.Scheme + "://" + tmpref.Reg
}

// API returns a ref formatted as an API appropriate URL (no tag info, but including api version `v2`).
func (r Ref) API() string {
	tmpref := r
	if tmpref.Scheme == "" {
		tmpref.Scheme = "https"
	}
	return tmpref.Scheme + "://" + tmpref.APIString()
}

// TagOrDigest returns the tag or digest of a ref.
func (r Ref) TagOrDigest() string {
	if r.Digest != "" {
		return r.Digest
	}
	return r.Tag
}

func getNameTagAndDigest(nametag string) (name string, tag string, dgst string) {
	nd := strings.Split(nametag, "@")
	if len(nd) > 1 {
		if d, err := digest.Parse(nd[1]); err == nil {
			dgst = d.String()
		}
	}
	nt := strings.Split(nd[0], ":")
	if len(nt) > 1 {
		name = nt[0]
		tag = nt[1]
	} else {
		name = nd[0]
		if dgst == "" {
			tag = "latest"
		}
	}
	return
}

const (
	// SchemeHTTP is the standard non-tls http url reference.
	SchemeHTTP = "http"
	// SchemeHTTPS is the standard tls http url reference.
	SchemeHTTPS = "https"
	// SchemeBottle defines a scheme for querying bottle location from a telemetry server.
	SchemeBottle = "bottle"
	// SchemeHash defines a scheme for querying bottle location from a telemetry server using hash uri format.
	SchemeHash = "hash"
)

// ExtractScheme looks for a transfer scheme prefix in the provided string, and returns the appropriate
// scheme string if found.  Additionally, the portion of the string not including any scheme value is
// returned.
func ExtractScheme(procstr string) (string, string) {
	var supportedSchemes = []string{
		SchemeHTTP,
		SchemeHTTPS,
		SchemeBottle,
		SchemeHash,
	}
	for _, scheme := range supportedSchemes {
		if strings.HasPrefix(procstr, scheme+"://") {
			return scheme, procstr[len(scheme)+3:]
		}
		if strings.HasPrefix(procstr, scheme+":") {
			return scheme, procstr[len(scheme)+1:]
		}
	}
	return "", procstr
}

func selectorsFromFragment(frag string) []string {
	if frag == "" {
		return nil
	}
	return strings.Split(frag, "|")
}

func parseScheme(scheme string, inref string) (Ref, bool) {
	switch scheme {
	case SchemeBottle:
		// bottle scheme has the format bottle:algorithm:digest#partselectors
		parts := strings.SplitN(inref, "#", 2)
		inref = parts[0]
		var fragment string
		if len(parts) > 1 {
			fragment = parts[1]
		}
		// verify the remaining portion of the ref is a valid digest spec
		if _, err := digest.Parse(inref); err != nil {
			return Ref{}, false
		}
		return Ref{Scheme: SchemeBottle, Digest: inref, Selectors: selectorsFromFragment(fragment)}, true
	case SchemeHash:
		parts := strings.SplitN(inref, "?", 2)
		inref = parts[0]
		// hash format uses a slash separator instead of a colon between algorithm and digest
		if strings.Contains(inref, "/") {
			inref = strings.Replace(inref, "/", ":", 1)
		} else {
			// if no algorithm is specified, invalid format
			return Ref{}, false
		}
		var query string
		if len(parts) > 1 {
			query = parts[1]
		}
		parts = strings.SplitN(query, "#", 2)
		query = parts[0]
		var fragment string
		if len(parts) > 1 {
			fragment = parts[1]
		}
		return Ref{Scheme: SchemeHash, Digest: inref, Query: query, Selectors: selectorsFromFragment(fragment)}, true
	default:
		return Ref{}, false
	}
}

// Validator defines a function that can examine a parsed reference and return an error or transformed Ref.
// Note, if no transform is done or there's no error, the supplied ref should be returned.
type Validator func(Ref) (Ref, error)

// DefaultRefValidator is the default validation function for parsed references, validating character sets appropriate
// for use with OCI spec references.  Characters outside of the spec cause an error and empty Ref to be returned.
func DefaultRefValidator(r Ref) (Ref, error) {
	isValid := regexp.MustCompile(`^[a-z0-9:.\-_]+$`).MatchString(r.Reg)
	if !isValid {
		return Ref{}, fmt.Errorf("invalid registry host format, must contain only a-z, 0-9, ., -, _ : %s", r.Reg)
	}
	if r.Repo != "" {
		isValid := regexp.MustCompile(`^[a-z0-9./\-_]+$`).MatchString(r.Repo)
		if !isValid {
			return Ref{}, fmt.Errorf("invalid repository format, must contain only a-z, 0-9, /, ., -, _ : %s", r.Repo)
		}
	}
	isValid = regexp.MustCompile(`^[a-z0-9.\-_]+$`).MatchString(r.Name)
	if !isValid {
		return Ref{}, fmt.Errorf("invalid name format, must contain only a-z, 0-9, ., -, _ : %s", r.Name)
	}
	// Tags can contain specific capitalization
	isValid = regexp.MustCompile(`^[A-Za-z0-9.\-_]*$`).MatchString(r.Tag)
	if !isValid {
		return Ref{}, fmt.Errorf("invalid tag format, must contain only a-Z, 0-9, ., -, _ : %s", r.Tag)
	}
	return r, nil
}

// SkipRefValidation is a helper validator that simply returns the provided ref, allowing validation to be skipped.
func SkipRefValidation(r Ref) (Ref, error) {
	return r, nil
}

// FromString parses a string representing a data bottle and returns a Ref object that identifies registry,
// repository, name, and tag.  By default, each element is validated using a regular expression using
// DefaultRefValidator, but this behavior can be optionally overridden by supplying a RefValidator function.  Note,
// the validator argument is specified as a variadic list to allow the default argument, only the first element of the
// list will be called.
func FromString(inref string, validator ...Validator) (Ref, error) {
	if strings.HasSuffix(inref, "/") {
		return Ref{}, fmt.Errorf("oci reference input does not specify a name")
	}

	scheme, procstr := ExtractScheme(inref)

	if ref, parsed := parseScheme(scheme, procstr); parsed {
		return ref, nil
	}

	if strings.Count(procstr, "/") == 0 {
		return Ref{}, fmt.Errorf("no registry or repository specified in reference %s", inref)
	}

	l := strings.LastIndex(procstr, "/")
	f := strings.Index(procstr, "/")

	checkStr := strings.SplitN(procstr[f+1:], "@", 2)

	if strings.Count(checkStr[0], ":") > 1 {
		return Ref{}, fmt.Errorf("invalid reference format, too many `:` in %s", procstr[f+1:])
	}

	reg := procstr[:f]
	if reg == "" {
		return Ref{}, fmt.Errorf("no registry specified in reference %s", inref)
	}
	repo := ""
	name, tag, dgst := getNameTagAndDigest(procstr[l+1:])
	var combineRepoName bool
	if l != f {
		repo = procstr[f+1 : l]
		if repo == name {
			// hack: combining repo and name if the two values are the same (eg /bats/bats:v1) to avoid confusion
			// elsewhere
			combineRepoName = true
		}
	}

	if name == "" {
		return Ref{}, fmt.Errorf("unable to determine name from reference %s", inref)
	}

	ref := Ref{
		Scheme: scheme,
		Reg:    reg,
		Repo:   repo,
		Name:   name,
		Tag:    tag,
		Digest: dgst,
	}

	if validator != nil {
		return validator[0](ref)
	}

	outRef, err := DefaultRefValidator(ref)
	if err != nil {
		return outRef, err
	}
	if combineRepoName {
		outRef.Repo = outRef.Repo + "/" + outRef.Name
		outRef.Name = outRef.Repo
	}
	return outRef, nil
}

// ResolveRegistry resolves a registry hostname to a fully specified one, for instance docker.io to index.docker.io.
// this utilizes a configured map of registry hostnames accessed through ace-dt config.  If no matching hostnames are discovered
// the input string is returned.
func ResolveRegistry(inReg string) string {
	// This used to be configurable but that used a singleton.
	// TODO make this configurable again
	m := map[string]string{
		"docker.io": "index.docker.io",
	}
	if r, ok := m[inReg]; ok {
		return r
	}
	return inReg
}

// RepoFromString parses a string representing a data bottle and returns a Ref object that identifies
// registry and repository.  This is similar to RefFromString, but is more tolerant of missing components,
// and doesn't treat name different from repository.
func RepoFromString(inref string) (outRef Ref) {
	procstr := inref
	outRef.Scheme, procstr = ExtractScheme(procstr)

	if ref, parsed := parseScheme(outRef.Scheme, procstr); parsed {
		return ref
	}

	f := strings.Index(procstr, "/")
	if f == -1 {
		outRef.Reg = procstr
		return
	}

	outRef.Reg = procstr[:f]
	outRef.Repo, outRef.Tag, outRef.Digest = getNameTagAndDigest(procstr[f+1:])
	outRef.Name = outRef.Repo
	return
}

// MatchType is a bitfield type that provides flags for what parts of a Ref object to match during a Match call.
type MatchType uint8

const (
	// RefMatchRepo indicates that a match should consider the repo portion of the reference.
	RefMatchRepo MatchType = 1 << iota
	// RefMatchReg indicates that a match should consider the registry portion of the reference.
	RefMatchReg
	// RefMatchTag Indicates that a match should consider the tag portion of the reference.
	RefMatchTag
	// RefMatchDigest indicates that a match should consider the digest portion of the reference.
	RefMatchDigest
	// RefMatchScheme indicates that a match should consider the scheme portion of the reference.
	RefMatchScheme
	// RefMatchRegRepo combines RefMatchRepo and RefMatchReg options.
	RefMatchRegRepo = RefMatchReg | RefMatchRepo
	// RefMatchAll combines RefMatchReg and RefMatchRepo and RefMatchTag options.
	RefMatchAll = RefMatchReg | RefMatchRepo | RefMatchTag
)

// Match compares the selector Ref to the provided other Ref and returns true if they match.  matchType provides
// options that determine what portions of the Ref are considered.
func (r Ref) Match(other Ref, matchType MatchType) bool {
	if matchType&RefMatchRepo != 0 {
		if !strings.EqualFold(r.RepoString(), other.RepoString()) {
			return false
		}
	}
	if matchType&RefMatchReg != 0 {
		if !strings.EqualFold(r.Reg, other.Reg) {
			return false
		}
	}
	if matchType&RefMatchTag != 0 {
		if !strings.EqualFold(r.Tag, other.Tag) {
			return false
		}
	}
	if matchType&RefMatchDigest != 0 {
		if !strings.EqualFold(r.Digest, other.Digest) {
			return false
		}
	}
	if matchType&RefMatchScheme != 0 {
		if !strings.EqualFold(r.Scheme, other.Scheme) {
			return false
		}
	}
	return true
}

// IsInsecure returns true iff the reference scheme is set to "http".
func (r Ref) IsInsecure() bool {
	return r.Scheme == SchemeHTTP
}
