package pypi

import (
	"context"
	"net/http"
	"time"

	"github.com/act3-ai/go-common/pkg/httputil"
	"github.com/act3-ai/go-common/pkg/httputil/promhttputil"
	"github.com/act3-ai/go-common/pkg/logger"
)

// Serve is the action for starting the server.
type Serve struct {
	*Action

	Listen string
}

// Run is the action method.
func (action *Serve) Run(ctx context.Context, repository string) error {
	log := logger.FromContext(ctx)

	// setup crane options
	repo, err := action.Config.Repository(ctx, repository)
	if err != nil {
		return err
	}

	myApp, err := NewApp(log, repo, repo.Reference.String(), action.Config.UserAgent(), action.Version(), action.AllowYanked)
	if err != nil {
		return err
	}

	// router := chi.NewRouter()
	mux := http.NewServeMux()

	handler := httputil.WrapHandler(mux,
		httputil.ServerHeaderMiddleware(action.Config.UserAgent()),
		httputil.TracingMiddleware,
		httputil.LoggingMiddleware(logger.FromContext(ctx)),
		promhttputil.PrometheusMiddleware,
	)

	myApp.Initialize(handler)

	// graceful shutdown adapted from https://github.com/gorilla/mux#graceful-shutdown

	srv := &http.Server{
		Addr: action.Listen,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handler,
	}

	return httputil.Serve(ctx, srv, 10*time.Second)
}
