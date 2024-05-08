package util

import "slices"

// Set provides a set-like container for strings implemented using a map.
type Set struct {
	m map[string]struct{}
}

// NewSet creates an empty set container.
func NewSet() *Set {
	s := &Set{}
	s.m = make(map[string]struct{})
	return s
}

// NewSetFromList creates a set from a given slice of strings.
func NewSetFromList(list []string) *Set {
	s := NewSet()
	for _, l := range list {
		s.Add(l)
	}
	return s
}

// Add adds a new string to the set.
func (s *Set) Add(value string) {
	s.m[value] = struct{}{}
}

// Remove deletes a string from the set.
func (s *Set) Remove(value string) {
	delete(s.m, value)
}

// Has returns true if the given value exists in the set.
func (s *Set) Has(value string) bool {
	_, c := s.m[value]
	return c
}

// List returns an ordered list of items in the set (sorted).
func (s *Set) List() []string {
	l := make([]string, 0, len(s.m))
	for k := range s.m {
		l = append(l, k)
	}
	slices.Sort(l)
	return l
}
