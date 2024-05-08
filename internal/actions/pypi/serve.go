package pypi

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
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
	repo, err := action.Config.ConfigureRepository(ctx, repository)
	if err != nil {
		return err
	}

	myApp, err := NewApp(log, repo, repo.Reference.String(), action.Config.UserAgent(), action.Version(), action.AllowYanked)
	if err != nil {
		return err
	}

	router := chi.NewRouter()

	// add some middleware
	router.Use(
		// NOTE from a security perspective sharing your server version is considered a security issue by some, but not by me.
		// The troubleshooting value out weights the security concerns.
		httputil.ServerHeaderMiddleware(action.Config.UserAgent()),

		httputil.TracingMiddleware,
		httputil.LoggingMiddleware(logger.FromContext(ctx)),
		httputil.PrometheusMiddleware,
	)

	myApp.Initialize(router)

	// graceful shutdown adapted from https://github.com/gorilla/mux#graceful-shutdown

	srv := &http.Server{
		Addr: action.Listen,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	return httputil.Serve(ctx, srv, 10*time.Second)
}
