package python

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/opencontainers/go-digest"
)

// Requirements represents a set of requirements (including index urls, constraints, etc).
type Requirements struct {
	// Reqs is a map from project name to requirements (of the same project name).
	// Any requirement may be satisfied.
	reqs map[string][]Requirement

	// IndexURL and ExtraIndexURLs are places to look for packages
	IndexURL       string
	ExtraIndexURLs []string
}

// RequirementsForProject returns a slice of requirements for a project.
// The caller should not modify the data pointed to by the slice.
func (rr *Requirements) RequirementsForProject(project string) []Requirement {
	return rr.reqs[project]
}

// ParseRequirementsFile extracts all requirements from the given file.
func (rr *Requirements) ParseRequirementsFile(requirementsFile string) error {
	reqFile, err := os.Open(requirementsFile)
	if err != nil {
		return fmt.Errorf("error opening requirements file: %w", err)
	}
	defer reqFile.Close()

	if err := rr.parseRequirements(requirementsFile, reqFile); err != nil {
		return err
	}

	return reqFile.Close()
}

// ParseRequirements parses all the requirements from the reader.
func (rr *Requirements) parseRequirements(requirementsFile string, r io.Reader) error {
	// read the file line by line using scanner
	scanner := bufio.NewScanner(r)

	prior := ""
	for scanner.Scan() {
		// do something with a line
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasSuffix(line, `\`) {
			// multi-line string
			prior += strings.TrimSuffix(line, `\`)
			continue
		}
		line = prior + line

		// everything after a # is a comment
		line, _, _ = strings.Cut(line, "#")

		if line == "" {
			// empty line
			continue
		}

		// This is a possible way we could implement the "file include feature"
		if strings.HasPrefix(line, "-r") /* || strings.HasPrefix(line, "-c") */ {
			// This is including another requirements file
			fields := strings.Fields(line)
			n := len(fields)
			if n != 2 {
				return fmt.Errorf("requirement file includes should be of the form \"-r filename\", but got %d instead of 2", n)
			}
			filename := filepath.Join(filepath.Dir(requirementsFile), fields[1])

			err := rr.ParseRequirementsFile(filename)
			if err != nil {
				return err
			}
			continue
		}

		if strings.HasPrefix(line, "--index-url") {
			fields := strings.Fields(line)
			n := len(fields)
			if n != 2 {
				return fmt.Errorf("index-url should be of the form \"--index-url URL\"")
			}
			rr.IndexURL = fields[1]
			continue
		}

		if strings.HasPrefix(line, "--extra-index-url") {
			fields := strings.Fields(line)
			n := len(fields)
			if n != 2 {
				return fmt.Errorf("extra-index-url should be of the form \"--extra-index-url URL\"")
			}
			rr.ExtraIndexURLs = append(rr.ExtraIndexURLs, fields[1])
			continue
		}

		err := rr.AddRequirementString(line)
		if err != nil {
			return err
		}
		prior = ""
	}

	return scanner.Err()
}

// AddRequirementString parses then adds a single requirement to the set.
func (rr *Requirements) AddRequirementString(requirement string) error {
	req, err := ParseRequirement(requirement)
	if err != nil {
		return fmt.Errorf("adding %q: %w", requirement, err)
	}

	rr.AddRequirement(*req)
	return nil
}

// AddRequirement adds the requirement to the set.
func (rr *Requirements) AddRequirement(req Requirement) {
	if rr.reqs == nil {
		rr.reqs = make(map[string][]Requirement)
	}

	rr.reqs[req.Name] = append(rr.reqs[req.Name], req)
}

// Indexes is the list of ordered python package indexes to use.
func (rr *Requirements) Indexes() []string {
	pypis := []string{rr.IndexURL}
	pypis = append(pypis, rr.ExtraIndexURLs...)
	return pypis
}

// Projects is a sorted list of project names referenced by these requirements.
func (rr *Requirements) Projects() []string {
	keys := make([]string, 0, len(rr.reqs))
	for k := range rr.reqs {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// IncludeExtras pulls out all the extra packages specified already and adds them to this set of requirements.
func (rr *Requirements) IncludeExtras() error {
	// TODO this deduplication should be moved to AddRequirement()
	seen := map[string]struct{}{}

	for _, reqs := range rr.reqs {
		for _, r := range reqs {
			for _, extra := range r.Extras {
				if _, ok := seen[extra]; ok {
					// deduplicate
					continue
				}
				// TODO this currently adds the blanket requirement and not one specific to this dependency (r) so arch/platform are not considered.
				if err := rr.AddRequirementString(extra); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Requirement represents a Python package requirement.
type Requirement struct {
	// Name is the normalized name of the project
	Name string

	// VersionSpecifier is the version specifier string (e.g., "==4.5.3", ">=4.5,<5")
	VersionSpecifier VersionSpecifier

	// Extras are the extra packages to be installed with this requirement
	Extras []string

	// Constrains are the non-hash constraints (e.g., "python_version > 3.6 and sys_platform == win32")
	Constraints string

	// Digests are the digests of the distribution files that this requirement can match
	Digests map[digest.Digest]struct{}
}

// ParseRequirement parses a Python requirement line
// currently only supports the pinned format produced by poetry.
func ParseRequirement(line string) (*Requirement, error) {
	var req Requirement

	requirement, constraint, _ := strings.Cut(line, ";")

	req.Name = requirement
	idx := strings.IndexAny(requirement, "[=<> \t")
	var rest string
	if idx != -1 {
		req.Name = requirement[:idx]
		rest = strings.TrimSpace(requirement[idx:])
	}
	req.Name = Normalize(strings.TrimSpace(req.Name))

	if strings.HasPrefix(rest, "[") {
		// has extras
		extras, r, _ := strings.Cut(rest[1:], "]")
		req.Extras = strings.Split(extras, ",")
		rest = r
	}
	vs, err := ParseVersionSpecifier(rest)
	if err != nil {
		return nil, fmt.Errorf("parsing version specifier in requirement %q: %w", line, err)
	}
	req.VersionSpecifier = vs

	req.Constraints = constraint
	const prefix = "--hash="
	if idx := strings.Index(constraint, prefix); idx != -1 {
		req.Constraints = constraint[:idx]
		req.Digests = make(map[digest.Digest]struct{})
		for _, field := range strings.Fields(constraint[idx:]) {
			if !strings.HasPrefix(field, prefix) {
				continue
			}

			hash := strings.TrimPrefix(field, prefix)
			hash = strings.TrimSpace(hash)
			dgst, err := digest.Parse(hash)
			if err != nil {
				return nil, fmt.Errorf("error parsing req. hash: %w", err)
			}
			req.Digests[dgst] = struct{}{}
		}
		if len(req.Digests) == 0 {
			panic("We should have found at least one --hash=")
		}
	}

	req.Constraints = strings.TrimSpace(req.Constraints)

	return &req, nil
}

// String formats this requirement as a pip compatible requirements specifier.
func (r *Requirement) String() string {
	var b strings.Builder
	b.Grow(200) // TODO compute this as an optimization
	b.WriteString(r.Name)

	if len(r.Extras) > 0 {
		b.WriteRune('[')
		b.WriteString(r.Extras[0])
		for _, s := range r.Extras[1:] {
			b.WriteRune(',')
			b.WriteString(s)
		}
		b.WriteRune(']')
	}

	if version := r.VersionSpecifier.String(); len(version) != 0 {
		b.WriteRune(' ')
		b.WriteString(r.VersionSpecifier.String())
	}

	if len(r.Constraints) != 0 {
		b.WriteString(" ; ")
		b.WriteString(r.Constraints)
	}

	hashes := make([]digest.Digest, 0, len(r.Digests))
	for k := range r.Digests {
		hashes = append(hashes, k)
	}
	slices.Sort(hashes)
	// hashes, put them on their own lines
	for _, d := range hashes {
		b.WriteString(" \\\n\t--hash=")
		b.WriteString(d.String())
	}

	return b.String()
}

/*
# This is a comment, to show how #-prefixed lines are ignored.
# It is possible to specify requirements as plain names.
pytest
pytest-cov
beautifulsoup4

# The syntax supported here is the same as that of requirement specifiers.
docopt == 0.6.1
requests [security] >= 2.8.1, == 2.8.* ; python_version < "2.7"
urllib3 @ https://github.com/urllib3/urllib3/archive/refs/tags/1.26.8.zip

# It is possible to refer to other requirement files or constraints files.
-r other-requirements.txt
-c constraints.txt

# These are also possible
--pre
--no-index
--find-links /my/local/archives
--find-links http://some.archives.com/archives
# TODO add support for these but handling the auth is hard.  Poetry does not include the credentials in the URL like PIP expects.  It stores credentials in a poetry specific place.
--index-url http://example.com/simple
--extra-index-url http://example2.com/simple
--extra-index-url http://example3.com/simple

# It is possible to refer to specific local distribution paths.
./downloads/numpy-1.9.2-cp34-none-win32.whl

# It is possible to refer to URLs.
http://wxpython.org/Phoenix/snapshot-builds/wxPython_Phoenix-3.0.3.dev1820+49a8884-cp34-none-win_amd64.whl

fsspec==2022.8.2 ; python_version >= "3.8" and python_version < "3.11" \
    --hash=sha256:6374804a2c0d24f225a67d009ee1eabb4046ad00c793c3f6df97e426c890a1d9 \
    --hash=sha256:7f12b90964a98a7e921d27fb36be536ea036b73bf3b724ac0b0bd7b8e39c7c18
*/

// Requirement specification is here https://peps.python.org/pep-0508/
// and PIP has some here https://pip.pypa.io/en/stable/reference/requirements-file-format/

// DistributionSatisfiesRequirement returns true iff the given distribution meets the requirement provided.
func DistributionSatisfiesRequirement(entry DistributionEntry, dist Distribution, req Requirement) bool {
	if dist.Project() != req.Name {
		return false
	}

	// Check that the version matches
	// TODO make dist.Version() return a Version
	if !req.VersionSpecifier.Matches(Version(dist.Version())) {
		return false
	}

	if req.Digests != nil {
		// Also check that the digest is included in the requirement
		_, exists := req.Digests[entry.Digest]
		// NOTE there seems to be a suitle difference between contraint files and requirement files that we need to iron out.
		// Why do we return exists here instead of just returning false if it is not in the digests?  That seems to a loose end we need to nail down.
		return exists
	}

	return true
}
