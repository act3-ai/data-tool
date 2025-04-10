package mirror

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"

	"github.com/act3-ai/data-tool/internal/ref"
)

type mapperFunc func(ocispec.Descriptor) ([]string, error)

var mappers = map[string]func(mapFile string) (mapperFunc, error){
	"first-prefix":   firstPrefixMapper,
	"go-template":    templateMapper,
	"digests":        digestMapper,
	"all-prefix":     allPrefixMapper,
	"longest-prefix": longestPrefixMapper,
	"nest":           nestMapper,
}

// newMapper returns a new mapping function for the given mapping directive.
func newMapper(mappingSpec string) (mapperFunc, error) {
	parts := strings.SplitN(mappingSpec, "=", 2)
	if len(parts) != 2 || parts[1] == "" {
		return nil, fmt.Errorf("invalid mapper, please use the format 'MAP-TYPE=MAP-ARG': got '%s'", mappingSpec)
	}
	mapType, mapArg := parts[0], parts[1]

	if m, ok := mappers[mapType]; ok {
		return m(mapArg)
	}

	return nil, fmt.Errorf("unknown mapping type %q", mapType)
}

// csvParser parses a given CSV mapfile and returns a matrix of [source][]destination.
func csvParser(mapfile string) ([][]string, error) {
	f, err := os.Open(mapfile)
	if err != nil {
		return nil, fmt.Errorf("error opening destination map file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comment = '#'
	r.FieldsPerRecord = 2
	return r.ReadAll()
}

// nestMapper returns a mapper function that will nest all the images under the given repository.
func nestMapper(repository string) (mapperFunc, error) {
	return func(d ocispec.Descriptor) ([]string, error) {
		p := path.Join(repository, d.Annotations[ref.AnnotationSrcRef])
		return []string{p}, nil
	}, nil
}

// firstPrefixMapper parses a CSV map file and returns a function that takes a descriptor and returns the first prefix-matched destination defined in the map.
func firstPrefixMapper(mapfile string) (mapperFunc, error) {
	locations, err := csvParser(mapfile)
	if err != nil {
		return nil, err
	}

	return func(d ocispec.Descriptor) ([]string, error) {
		sref := d.Annotations[ref.AnnotationSrcRef]
		return mapPrefixes(sref, locations, true), nil
	}, nil
}

// allPrefixMapper parses a CSV map file and returns a function that takes a descriptor and returns all of the destinations mapped to the reference.
func allPrefixMapper(mapfile string) (mapperFunc, error) {
	locations, err := csvParser(mapfile)
	if err != nil {
		return nil, err
	}

	return func(d ocispec.Descriptor) ([]string, error) {
		sref := d.Annotations[ref.AnnotationSrcRef]
		return mapPrefixes(sref, locations, false), nil
	}, nil
}

// longestPrefixMapper parses a CSV map file and returns a function that takes a descriptor and returns the destination that has the longest prefix matching the fully-qualified source in the descriptor's annotations.
func longestPrefixMapper(mapfile string) (mapperFunc, error) {
	locations, err := csvParser(mapfile)
	if err != nil {
		return nil, err
	}

	return func(d ocispec.Descriptor) ([]string, error) {
		sref := d.Annotations[ref.AnnotationSrcRef]
		return []string{getLongestPrefix(sref, locations)}, nil
	}, nil
}

// mapPrefixes is a prefix-matching helper function that returns a list of destinations by prefix.
func mapPrefixes(sref string, locations [][]string, first bool) []string {
	var destinations []string

	for _, location := range locations {
		from, to := location[0], location[1]
		if from == "" {
			// blank not supported
			continue
		}
		if strings.HasPrefix(sref, from) {
			// create the new completed destination location
			// trim the prefix of the original ref so it doesn't get copied over
			newLoc := to + strings.TrimPrefix(sref, from)
			destinations = append(destinations, newLoc)
			if first {
				break
			}
		}
	}

	return destinations
}

// getLongestPrefix will return the destination with the longest matching prefix.
// Ties are broken by selecting the last match.
func getLongestPrefix(sref string, locations [][]string) string {
	var destination string
	var matchLength int
	for _, location := range locations {
		from, to := location[0], location[1]
		if len(from) < matchLength {
			// this (if matched) will not be long enough to overturn the exiting match
			continue
		}

		// check if this is a prefix match
		if strings.HasPrefix(sref, from) {
			destination = to + strings.TrimPrefix(sref, from)
			matchLength = len(from)
		}
	}
	return destination
}

// templateMapper initializes a go-template and returns a mapperFunc that accepts a descriptor to scatter and returns a list of string destinations.
func templateMapper(mapFile string) (mapperFunc, error) {
	funcmap := template.FuncMap{
		"Tag":        extractTag,
		"Repository": extractRepo,
		"Registry":   extractReg,
		"Package":    extractPackage,
	}

	// grab the file name for the template creation
	filename := filepath.Base(mapFile)
	t, err := template.New(filename).Funcs(sprig.HermeticTxtFuncMap()).Funcs(funcmap).ParseFiles(mapFile)
	if err != nil {
		return nil, fmt.Errorf("error generating template: %w", err)
	}

	return func(d ocispec.Descriptor) ([]string, error) {
		buf := &bytes.Buffer{}
		if err := t.Execute(buf, d); err != nil {
			return nil, fmt.Errorf("error executing template: %w", err)
		}
		if buf.Len() == 0 {
			return []string{}, nil
		}

		s := strings.TrimSpace(buf.String())
		return strings.Split(s, "\n"), nil
	}, nil
}

func extractTag(sref string) (string, error) {
	// don't use a registry.EndpointReferenceParser, preserving the original reference
	r, err := registry.ParseReference(sref)
	if err != nil {
		return "", fmt.Errorf("error parsing the reference: %w", err)
	}
	return r.Reference, nil
}

func extractRepo(sref string) (string, error) {
	// don't use a registry.EndpointReferenceParser, preserving the original reference
	r, err := registry.ParseReference(sref)
	if err != nil {
		return "", fmt.Errorf("error parsing the reference: %w", err)
	}
	return r.Repository, nil
}

func extractReg(sref string) (string, error) {
	// don't use a registry.EndpointReferenceParser, preserving the original reference
	r, err := registry.ParseReference(sref)
	if err != nil {
		return "", fmt.Errorf("error parsing the reference: %w", err)
	}
	return r.Registry, nil
}

// Maybe a better name out there? A package is the full repository path and reference together (no registry).
func extractPackage(sref string) (string, error) {
	// don't use a registry.EndpointReferenceParser, preserving the original reference
	r, err := registry.ParseReference(sref)
	if err != nil {
		return "", fmt.Errorf("error parsing the reference: %w", err)
	}
	// i want the full path with the reference (digest or tag) at the end so using trim prefix
	p := strings.TrimPrefix(sref, r.Registry+"/")
	return p, nil
}

// digestMapper takes a csv file (digest, destination) and maps them.
func digestMapper(mapFile string) (mapperFunc, error) {
	locations, err := csvParser(mapFile)
	if err != nil {
		return nil, err
	}
	return func(d ocispec.Descriptor) ([]string, error) {
		var destinations []string
		for _, location := range locations {
			if d.Digest.String() == location[0] {
				destinations = append(destinations, location[1])
			}
		}
		return destinations, nil
	}, nil
}
