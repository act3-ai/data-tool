package pypi

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"

	"github.com/act3-ai/data-tool/internal/python"
	"github.com/act3-ai/go-common/pkg/httputil"
	"github.com/act3-ai/go-common/pkg/logger"
)

func (a *App) handleAbout(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	return a.executeTemplate(ctx, w, "about.html", nil)
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	// list all projects by listing tags
	projects, err := registry.Tags(ctx, a.repository)
	if err != nil {
		return fmt.Errorf("error listing remote tags: %w", err)
	}

	log.InfoContext(ctx, "Found tags", "count", len(projects))

	slices.Sort(projects)

	// TODO support JSON output in addition to HTML based on the accept header

	values := struct {
		Projects []string
	}{projects}

	return a.executeTemplate(ctx, w, "index.html", values)
}

func (a *App) handleProject(w http.ResponseWriter, r *http.Request) error {
	project := r.PathValue("project")

	ctx := r.Context()
	log := logger.FromContext(ctx).With("project", project)

	pyDistIdx, err := newPythonDistributionIndex(ctx, a.repository, false, project)
	if err != nil {
		return err
	}

	distributions, err := pyDistIdx.GetDistributions(a.allowYanked)
	if err != nil {
		return err
	}

	log.InfoContext(ctx, "found distributions", "count", len(distributions))

	// TODO support JSON output in addition to HTML based on the accept header

	// build out the values
	values := struct {
		Project       string
		Distributions []python.DistributionEntry
	}{project, distributions}

	return a.executeTemplate(ctx, w, "project.html", values)
}

func (a *App) handleFile(w http.ResponseWriter, r *http.Request) error {
	project := r.PathValue("project")
	filename := r.PathValue("filename")

	ctx := r.Context()
	log := logger.FromContext(ctx).With("project", project, "filename", filename)
	log.InfoContext(ctx, "Retrieving file")

	pyDistIdx, err := newPythonDistributionIndex(ctx, a.repository, false, project)
	if err != nil {
		return err
	}

	var metadata bool
	if strings.HasSuffix(filename, python.MetadataSuffix) {
		filename = strings.TrimSuffix(filename, python.MetadataSuffix)
		metadata = true
		log.InfoContext(ctx, "Metadata requested")
	}

	man, ok := pyDistIdx.LookupDistribution(filename)
	if !ok {
		return httputil.NewHTTPError(nil, http.StatusNotFound, "File not found in python distribution index", "filename", filename)
	}

	dist, meta, err := pyDistIdx.FetchDistribution(ctx, man)
	if err != nil {
		return err
	}

	desc := dist
	if metadata {
		if meta == nil {
			return httputil.NewHTTPError(nil, http.StatusNotFound, "Metadata not available for this distribution", "filename", filename)
		}
		desc = *meta
	}

	log.InfoContext(ctx, "Found layer")

	// read a blob from OCI (stream it)
	rc, err := a.repository.Fetch(ctx, desc)
	if err != nil {
		return fmt.Errorf("error reading blob from OCI: %w", err)
	}
	defer rc.Close()

	httputil.AllowCaching(w.Header())

	// TODO layer.MediaType to set the content type
	// pypi.org for sdist  .tar.gz files uses Content-Type: application/x-tar
	// for .whl uses Content-Type: binary/octet-stream
	// for sdist .zip uses application/zip
	// Or we remove the below line and let content-type discovery happen on the data (for http.ServeContent())
	w.Header().Set("Content-Type", "binary/octet-stream")

	// parse the ocispec.AnnotationCreated annotations of layer to populate this
	modtime, err := time.Parse(time.RFC3339, desc.Annotations[ocispec.AnnotationCreated])
	if err != nil {
		// This is not fatal
		log.ErrorContext(ctx, "parsing descriptor created time", "error", err.Error(), "digest", desc.Digest)
	}

	if rcs, ok := rc.(io.ReadSeeker); ok {
		// ideally we would have a io.ReadSeeker and then we can use http.ServeContent() to get caching and range requests.
		// The ORAS does support io.ReadSeeker when the registry supports HTTP range requests.  See registry/remote/repository_test.go (func Test_BlobStore_Fetch_Seek()).  So we need to cast our rc to a io.Seeker (if possible) and serve the request with http.ServerContent()
		// ORAS only support io.ReadSeeker if the upstream registry support it
		log.InfoContext(ctx, "supporting range request")

		http.ServeContent(w, r, "", modtime, rcs)
		// NOTE there is an inefficiency in http.ServeContent.  It seeks to the end to find the size (we already know the size layer.Size but cannot provide it) which makes a new HTTP request to the upstream server. With HTTP/2 this should be insignificant.
		return nil
	}

	// fallback to the non-io.Reader way
	w.Header().Set("Content-Length", fmt.Sprint(desc.Size))

	// TODO still possible to support If-Modified-Since and relates headers by using modtime.
	_, err = io.Copy(w, rc)
	if err != nil {
		return fmt.Errorf("error streaming: %w", err)
	}

	return nil
}
