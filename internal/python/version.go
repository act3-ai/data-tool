package python

import (
	"fmt"
	"regexp"
	"strings"
)

// https://packaging.python.org/en/latest/specifications/version-specifiers/#version-specifiers

// Version is a Python distribution version.
type Version string

// VersionSpecifier specifies a range of versions used in python requirement.
// The nil or length zero VersionSpecifier matches everything.
type VersionSpecifier []VersionClause

// Matches returns true if the provided version is acceptable by this VersionSpecifier.
func (vs VersionSpecifier) Matches(version Version) bool {
	// to match it must match ALL clauses
	for _, c := range vs {
		if !c.Matches(version) {
			return false
		}
	}
	return true
}

// String formats the VersionSpecifier in it's requirements compatible form.
func (vs VersionSpecifier) String() string {
	switch len(vs) {
	case 0:
		return ""
	case 1:
		return vs[0].String()
	}

	var b strings.Builder
	b.Grow(len(vs) * 10) // should be sufficient
	b.WriteString(vs[0].String())
	for _, c := range vs[1:] {
		b.WriteRune(' ')
		b.WriteString(c.String())
	}
	return b.String()
}

// ParseVersionSpecifier parses the version specifier from a requirement spec.
func ParseVersionSpecifier(versionSpec string) (VersionSpecifier, error) {
	versionSpec = strings.TrimSpace(versionSpec)
	if len(versionSpec) == 0 {
		// No version specifier
		return nil, nil
	}

	parts := strings.Split(versionSpec, ",")
	vs := make(VersionSpecifier, len(parts))
	for i, c := range strings.Split(versionSpec, ",") {
		clause, err := ParseVersionClause(c)
		if err != nil {
			return nil, fmt.Errorf("parsing clause %s: %w", c, err)
		}
		vs[i] = *clause
	}

	return vs, nil
}

// VersionClause is a single part of a VersionSpecifier.
type VersionClause struct {
	// Operator is the comparison operator (==, !=, ~=, etc.)
	Operator string

	// Value is the partial version string (1.2.*, 1.2, 1.2.3a4 )
	Value string
}

// Matches returns true when the version matches the clause.
func (vc VersionClause) Matches(version Version) bool {
	switch vc.Operator {
	case "===":
		return string(version) == vc.Value
	case "~=", "==", "!=", "<=", ">=", "<", ">":
		// TODO Not implemented
		// lets just we match everything for now
		return true
	default:
		panic(fmt.Sprintf("Unknown version specifier: %q", vc.Operator))
	}
}

// String converts the clause back into the string representation.
func (vc VersionClause) String() string {
	return vc.Operator + vc.Value
}

var clauseRegex = regexp.MustCompile(`^\s*(~=|==|!=|<=|>=|<|>|===)\s*([\w-.\*]*)\s*$`)

// ParseVersionClause parses the version clause, splitting out the operator from the value.
func ParseVersionClause(clauseSpec string) (*VersionClause, error) {
	groups := clauseRegex.FindStringSubmatch(clauseSpec)
	if len(groups) != 3 {
		return nil, fmt.Errorf("clause %q must match %s", clauseSpec, clauseRegex)
	}

	return &VersionClause{
		Operator: groups[1],
		Value:    groups[2],
	}, nil
}
