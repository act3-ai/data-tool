package bottle

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/act3-ai/data-tool/internal/bottle"
	"github.com/act3-ai/go-common/pkg/logger"
)

//go:embed templates
var baseTemplateFS embed.FS

// GUI represents the bottle gui action.
type GUI struct {
	*Action

	Listen         string
	DisableBrowser bool
}

// Run displays the bottle edit gui in a web browser and returns after the updated bottle data has been posted to the web server.
func (action *GUI) Run(ctx context.Context, out io.Writer) error {
	log := logger.FromContext(ctx)

	log.InfoContext(ctx, "Serve command activated")
	_, btl, err := action.prepare(ctx)
	if err != nil {
		return err
	}

	finishedEditingChannel := make(chan error)
	defer close(finishedEditingChannel)

	// Set up the http request router
	mux := http.NewServeMux()
	h := handlers{
		tmpl: template.Must(template.New("base").ParseFS(baseTemplateFS, "templates/*.html")),
		done: finishedEditingChannel,
		btl:  btl,
		log:  log.WithGroup("handler"),
	}
	mux.HandleFunc("GET /", h.bottleGet)
	mux.HandleFunc("POST /discard", h.bottleDiscard)
	mux.HandleFunc("POST /", h.bottlePost)

	// Also serving static assets
	templatesAssetsDir, err := fs.Sub(baseTemplateFS, "templates/assets")
	if err != nil {
		return fmt.Errorf("could not open templates/assets directory: %w", err)
	}
	mux.Handle("GET /assets/", http.StripPrefix("/assets", http.FileServerFS(templatesAssetsDir)))

	// we create our own Listener because we need a way to get th port that it chooses (when port is 0).
	// http server's ListenAndServe() does not provide access to that.
	listener, err := net.Listen("tcp", action.Listen)
	if err != nil {
		return fmt.Errorf("could not create tcp listener: %w", err)
	}

	// HACK to allow unit testing access to the dynamic port
	u := fmt.Sprintf("http://%s", listener.Addr())
	if err := os.WriteFile(GUIURLPath(btl.GetPath()), []byte(u), 0o666); err != nil {
		return fmt.Errorf("writing GUI URL to file: %w", err)
	}

	_, err = fmt.Fprintln(out, "Serving bottle edit UI at address", u)
	if err != nil {
		return err
	}

	httpServer := http.Server{
		Handler: mux,
	}

	// Using an errorgroup and channel to be able to gracefully shutdown in case of an error
	g := errgroup.Group{}
	g.Go(func() error {
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("could not serve http: %w", err)
		}
		return nil
	})

	if !action.DisableBrowser {
		if err = openBrowser(u); err != nil {
			return fmt.Errorf("error occurred while opening browser: %w", err)
		}
	}

	finishedEditingErr := <-finishedEditingChannel
	if finishedEditingErr != nil {
		_ = httpServer.Close()
		return finishedEditingErr
	}

	log.InfoContext(ctx, "Shutdown requested")

	// Create a deadline to wait for.
	wait := 10 * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	log.InfoContext(ctx, "Waiting for graceful shutdown", "timeout", wait)
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err := httpServer.Shutdown(timeoutCtx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	if err := g.Wait(); err != nil {
		return err
	}

	log.InfoContext(ctx, "Serve command completed")

	return nil
}

// GUIURLPath returns the path used to store the connection URL to the last started bottle edit GUI
// HACK for the unit test.
func GUIURLPath(pth string) string {
	return filepath.Join(pth, ".dt", "gui-url.txt")
}

type handlers struct {
	tmpl *template.Template
	done chan<- error
	btl  *bottle.Bottle
	log  *slog.Logger
}

func (h *handlers) bottleDiscard(http.ResponseWriter, *http.Request) {
	h.done <- nil
}

// bottleHandler returns an http.Handler function for bottle editing
// btl is the initial bottle object that will be displayed in the GUI
// done is an error channel to push to when editing is complete and the http handler is no longer needed.
func (h *handlers) bottleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.log.With("handler", "bottle GET")
	log.InfoContext(ctx, "GET called")
	var buf bytes.Buffer
	values := struct {
		Bottle *bottle.Bottle
	}{
		h.btl,
	}
	if err := h.tmpl.ExecuteTemplate(&buf, "bottle-edit.html", values); err != nil {
		h.done <- fmt.Errorf("could not execute bottle edit template: %w", err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	if _, err := buf.WriteTo(w); err != nil {
		h.done <- fmt.Errorf("error writing to template buffer: %w", err)
		return
	}
}

func (h *handlers) bottlePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.log.With("handler", "bottle POST")
	log.InfoContext(ctx, "POST called")
	// Were expecting a bottle definition in the JSON body of the request
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.done <- fmt.Errorf("could not read request body data: %w", err)
		return
	}
	log.InfoContext(ctx, "received bottle post data", "requestBody", string(data))

	if err = h.btl.Configure(data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.done <- fmt.Errorf("could not update bottle from request body: %w", err)
		return
	}

	if err = saveMetaChanges(r.Context(), h.btl); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.done <- err
		return
	}
	log.InfoContext(ctx, "Bottle data saved")

	h.done <- nil
}

func openBrowser(u string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", u).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	case "darwin":
		err = exec.Command("open", u).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		return err
	}
	return nil
}
