package python

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"github.com/opencontainers/go-digest"
	"golang.org/x/exp/maps"
	"golang.org/x/net/html"
)

// ErrProjectNotFound indicated that the project was not found in the python package index.
var ErrProjectNotFound = errors.New("python project not found")

// MediaTypes as defined in https://peps.python.org/pep-0691/
const (
	MediaTypeSimpleJSON   = "application/vnd.pypi.simple.v1+json"
	MediaTypeSimpleHTML   = "application/vnd.pypi.simple.v1+html"
	MediaTypeSimpleLegacy = "text/html"
)

// MetadataSuffix for metadata requests, see PEP-658.
const MetadataSuffix = ".metadata"

// This might be a useful resource https://pypi.org/project/dumb-pypi/

// Client behaves like http.Client.
// Abstracting at this level allows auth and retries.
// Same as oras/remote.Client.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// RetrieveAllDistributions get all the distributions from all Python package indexes.
// The distributions are deduplicated such that the first package index that provides a distribution file is used.
func RetrieveAllDistributions(ctx context.Context, client Client, pypis []string, project string) ([]DistributionEntry, error) {
	// TODO Consider the security implications of extra-index-urls https://mattkubilus.medium.com/pip-extra-index-url-considered-dangerous-43146e44f1c
	// Poetry provides --hash in the requirements files so we only pull those hashes which mitigates the issue for this tool used in that way
	var found bool
	allDists := make(map[string]DistributionEntry)
	for _, pypi := range pypis {
		dists, err := RetrieveDistributions(ctx, client, pypi, project)
		if err != nil {
			if errors.Is(err, ErrProjectNotFound) {
				continue
			}
			return nil, err
		}
		found = true
		for _, dist := range dists {
			if _, ok := allDists[dist.Filename]; !ok {
				allDists[dist.Filename] = dist
			}
		}
	}

	// NOTE if no package indexes have the project then it is not found
	// We still might have no distribution files even if the project is found
	if !found {
		return nil, fmt.Errorf("project %q: %w", project, ErrProjectNotFound)
	}

	return maps.Values(allDists), nil
}

// TODO as an alternative to the HTML based legacy API (PEP 503) we could implement the JSON API (see https://warehouse.pypa.io/api-reference/json.html) or the other JSON API (PEP-691 https://peps.python.org/pep-0691/)

// TODO we could cache to disk the results html or []DistributionEntry from PyPI so that we do not have to pull them every time.  We would need to expire them periodically.

// RetrieveDistributions reads and parses the Python Package Index standard HTML for a Python project and returns the slice of Distribution objects and possibly an error.
func RetrieveDistributions(ctx context.Context, client Client, pypi string, project string) ([]DistributionEntry, error) {
	log := logger.FromContext(ctx).With("project", project)

	// make a request to https://pypi.org/simple/<project> for example
	base := pypi + "/" + project + "/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating http get request: %w", err)
	}

	req.Header.Set("accept", MediaTypeSimpleHTML+";q=0.2, "+MediaTypeSimpleLegacy+";q=0.01")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing get request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("project %q: %w", project, ErrProjectNotFound)
	}

	// see https://warehouse.pypa.io/api-reference/legacy.html
	// serial := res.Header.Get("X-PyPI-Last-Serial")
	// This is PyPI specific and not standards compliant so it is low priority (also not sure exactly what we would use it for)

	mediaType, _, err := mime.ParseMediaType(res.Header.Get("content-type"))
	if err != nil {
		return nil, fmt.Errorf("error parsing media type: %w", err)
	}

	switch mediaType {
	case MediaTypeSimpleHTML, MediaTypeSimpleLegacy:
		return parseV1ProjectHTML(log, base, project, res.Body)
	case MediaTypeSimpleJSON:
		return nil, fmt.Errorf("not implemented yet")
	default:
		return nil, fmt.Errorf("unknown content type (%s) returned from python package index", mediaType)
	}
}

// DistributionEntry represents a raw distribution file reference entry for a project (minimally processed).
type DistributionEntry struct {
	// URL of the distribution file
	URL string

	// Digest is the digest (if present) of the distribution file
	Digest digest.Digest

	// Filename is the filename which encodes lots of information.
	Filename string

	// RequiresPython is an optional specifier that denotes a ranges of python version this distribution supports
	RequiresPython string

	// Yanked if non-nil then the package has been yanked (should no longer be used).  The value is the reason why it was yanked (may be empty).  See https://peps.python.org/pep-0592/
	Yanked *string

	// MetadataDigest is the digest (if present) of the metadata
	MetadataDigest digest.Digest
}

// MetadataURL is the URL of the metadata (if present).
func (entry *DistributionEntry) MetadataURL() string {
	if entry.MetadataDigest != "" {
		// append ".metadata" to the path part of the URL
		// TODO use URLs instead to preserve query string parameters
		return entry.URL + MetadataSuffix
	}

	return ""
}

// Digest extracts the optional digest from the location.  If not present it returns an empty URL and  digest.
func processURL(base *url.URL, locationURL string) (string, digest.Digest) {
	ref, err := url.Parse(locationURL)
	if err != nil {
		return "", ""
	}

	d := parseDigest(ref.Fragment)
	ref.Fragment = "" // remove the fragment

	// handle relative URLs properly
	ref = base.ResolveReference(ref)
	return ref.String(), d
}

func parseDigest(pydigest string) digest.Digest {
	// Handle the digest (optional)
	dgst := strings.Replace(pydigest, "=", ":", 1)
	d, err := digest.Parse(dgst)
	if err == nil {
		return d
	}

	return ""
}

// parseV1ProjectHTML is https://peps.python.org/pep-0503/ compliant parser for the HTML.
func parseV1ProjectHTML(log *slog.Logger, pypi, project string, in io.Reader) ([]DistributionEntry, error) {
	// parse the HTML <a> tags to find the version

	// href can be complete, absolute, or relative
	/*
		<a href="https://files.pythonhosted.org/packages/07/71/9f0eaddad031f593018623d80e7169ba95fbf47c356539494dd70bbbffc7/numpy-1.26.0b1-cp39-cp39-musllinux_1_1_x86_64.whl#sha256=eea337d6d5ab2b6eb657b3f18e8b57a280f16fb5f94df484d9c1a8d3450d9ae9"
			data-requires-python="<3.13,>=3.9"
			data-dist-info-metadata="sha256=460946f848b89ed152a9ff396392a2d82f0513d94f74308ce748c4ac65eea77b"
			data-core-metadata="sha256=460946f848b89ed152a9ff396392a2d82f0513d94f74308ce748c4ac65eea77b">
			numpy-1.26.0b1-cp39-cp39-musllinux_1_1_x86_64.whl
		</a>

		<a href="/packages/07/71/9f0eaddad031f593018623d80e7169ba95fbf47c356539494dd70bbbffc7/numpy-1.26.0b1-cp39-cp39-musllinux_1_1_x86_64.whl#sha256=eea337d6d5ab2b6eb657b3f18e8b57a280f16fb5f94df484d9c1a8d3450d9ae9"
			data-requires-python="<3.13,>=3.9"
			data-dist-info-metadata="sha256=460946f848b89ed152a9ff396392a2d82f0513d94f74308ce748c4ac65eea77b"
			data-core-metadata="sha256=460946f848b89ed152a9ff396392a2d82f0513d94f74308ce748c4ac65eea77b">
			numpy-1.26.0b1-cp39-cp39-musllinux_1_1_x86_64.whl
		</a>

		<a href="../../numpy-1.26.0b1-cp39-cp39-musllinux_1_1_x86_64.whl#sha256=eea337d6d5ab2b6eb657b3f18e8b57a280f16fb5f94df484d9c1a8d3450d9ae9"
			data-requires-python="<3.13,>=3.9"
			data-dist-info-metadata="sha256=460946f848b89ed152a9ff396392a2d82f0513d94f74308ce748c4ac65eea77b"
			data-core-metadata="sha256=460946f848b89ed152a9ff396392a2d82f0513d94f74308ce748c4ac65eea77b">
			numpy-1.26.0b1-cp39-cp39-musllinux_1_1_x86_64.whl
		</a>
	*/

	base, err := url.Parse(pypi)
	if err != nil {
		return nil, fmt.Errorf("parsing pypi URL: %w", err)
	}

	doc, err := html.Parse(in)
	if err != nil {
		return nil, fmt.Errorf("error parsing html: %w", err)
	}

	var dists []DistributionEntry
	var f func(*html.Node) error
	f = func(n *html.Node) error {
		if n.Type == html.ElementNode && n.Data == "a" {
			dist := parseDistNode(base, n)
			if dist.URL != "" && dist.Filename != "" {
				logger.V(log, 1).Info("Found distribution", "filename", dist.Filename) //nolint:sloglint
				dists = append(dists, dist)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := f(c); err != nil {
				return err
			}
		}
		return nil
	}
	if err := f(doc); err != nil {
		return nil, err
	}

	logger.V(log, 1).Info("Found packages", "number", len(dists)) //nolint:sloglint

	return dists, nil
}

func parseDistNode(base *url.URL, n *html.Node) DistributionEntry {
	entry := DistributionEntry{}
	for _, a := range n.Attr {
		switch a.Key {
		case "href":
			entry.URL, entry.Digest = processURL(base, a.Val)
		case "data-requires-python":
			entry.RequiresPython = a.Val
		case "data-yanked":
			val := a.Val // make a copy
			entry.Yanked = &val
		case "data-gpg-sig":
			// found a GPG signature
			// I have not found a case where this is used in the wild yet.  This is low priority.
		case "data-core-metadata":
			/* numpy is a package that does have this HTML attribute
			https://peps.python.org/pep-0658/ original but had PyPI and PIP both had bugs and implemented it incorrectly thus burning data-dist-info-metadata.
			https://peps.python.org/pep-0714/ TLDR, use data-core-metadata now.
			https://PyPI.org finally does implement this (thought the issue is not closed yet)https://github.com/pypi/warehouse/issues/8254
			*/
			entry.MetadataDigest = parseDigest(a.Val)
		case "data-dist-info-metadata":
			// fallback to this attribute if available, as per PEP-714
			if entry.MetadataDigest != "" {
				// we already came across "data-core-metadata" so we do not want to overwrite it
				break
			}
			entry.MetadataDigest = parseDigest(a.Val)
		}
	}

	if n.FirstChild != nil {
		// Filename is taken from the body of the <a> tag and not the filename in the Location since they are allowed to be different.
		entry.Filename = n.FirstChild.Data
	}
	return entry
}
