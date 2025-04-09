// Package pypi facilitates executing the program as a proxy server for translating pypi requests into OCI registry
// operations. This enables an OCI registry to act as a pypi package registry.
package pypi

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"oras.land/oras-go/v2/registry"

	"github.com/act3-ai/data-tool/internal/python"
	"github.com/act3-ai/go-common/pkg/httputil"
	"github.com/act3-ai/go-common/pkg/logger"
)

// let's try go embedding to get the templates since they are cool

//go:embed templates/*.html
var templateFS embed.FS

// App implements the HTML web application.
type App struct {
	templates    *template.Template
	log          *slog.Logger
	globalValues globalValues
	repository   registry.Repository
	allowYanked  bool
	userAgent    string
}

// NewApp creates the web application used for serving the PyPI site (backed by OCI).
func NewApp(log *slog.Logger, repo registry.Repository, repoRef, userAgent, version string, allowYanked bool) (*App, error) {
	a := &App{
		log:         log.WithGroup("app"),
		repository:  repo,
		allowYanked: allowYanked,
		userAgent:   userAgent,
	}

	t, err := parseTemplates()
	if err != nil {
		return nil, err
	}
	a.templates = t

	a.globalValues = globalValues{
		Version:    version,
		Repository: repoRef,
	}

	return a, nil
}

// Initialize the routes.
func (a *App) Initialize(handler httputil.Router) {
	handler.Handle("GET /", httputil.RootHandler(a.handleAbout))
	handler.Handle("GET /simple/", httputil.RootHandler(a.handleIndex))

	handler.Handle("GET /simple/{project}/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := r.PathValue("project")
		norm := python.Normalize(project)
		if norm == project {
			httputil.RootHandler(a.handleProject).ServeHTTP(w, r)
		} else {
			// be nice and redirect to the normalized project name
			// http.Redirect(w, r, "/simple/"+norm+"/", http.StatusMovedPermanently)
			w.Header().Set("Location", "../"+norm+"/")
			w.WriteHeader(http.StatusMovedPermanently)
		}
	}))
	handler.Handle("GET /simple/{project}/{filename}", httputil.RootHandler(a.handleFile))

	handler.Handle("GET /simple", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// http.Redirect(w, r, "/simple/", http.StatusMovedPermanently)
		w.Header().Set("Location", "simple/")
		w.WriteHeader(http.StatusMovedPermanently)
	}))

	handler.Handle("GET /simple/{project}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := r.PathValue("project")
		// http.Redirect(w, r, "/simple/"+project+"/", http.StatusMovedPermanently)
		w.Header().Set("Location", project+"/")
		w.WriteHeader(http.StatusMovedPermanently)
	}))
	// middleware.GetHead might be useful on the files handler
}

func parseTemplates() (*template.Template, error) {
	// // TemplateFuncs are functions usable in the HTML templates
	// var templateFuncs = template.FuncMap{
	// 	"ByteSize": bytefmt.ByteSize,
	// 	"ToAge": func(t time.Time) string {
	// 		return duration.ShortHumanDuration(time.Since(t))
	// 	},
	// }

	var t *template.Template
	var err error
	if t, err = template.New("base").Funcs(sprig.FuncMap()).ParseFS(templateFS, "templates/*.html"); err != nil {
		return t, fmt.Errorf("error parsing template: %w", err)
	}

	return t, nil
}

type globalValues struct {
	Version    string
	Repository string
}

func (a *App) executeTemplate(ctx context.Context, w http.ResponseWriter, templateName string, values any) error {
	allValues := struct {
		Values  any
		Globals globalValues
	}{values, a.globalValues}

	log := logger.FromContext(ctx).With("values", allValues)
	log.InfoContext(ctx, "Rendering template", "name", templateName)

	// See https://medium.com/@leeprovoost/dealing-with-go-template-errors-at-runtime-1b429e8b854a
	// We do not write directly to the HTML response because if a template error occurs we send a partial response
	var buf bytes.Buffer
	if err := a.templates.ExecuteTemplate(&buf, templateName, allValues); err != nil {
		return fmt.Errorf("failed to render template %s: %w", templateName, err)
	}
	// if all is good then write buffer to the response writer
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("error writing to template buffer: %w", err)
	}
	return nil
}
