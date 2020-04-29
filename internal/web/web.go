// Package web is responsible for creating and initializing endpoints for interacting with the database.
//
package web

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

// App structure containing the necessary server settings and responsible for starting and stopping it.
type App struct {
	srv *http.Server
}

// AppConfig structure containing the server settings necessary for its operation.
type AppConfig struct {
	NetInterface string
	Timeout      time.Duration
}

func (ac *AppConfig) checkConfig() {

	log.Debug().Msg("checking web application config")

	if ac.NetInterface == "" {
		ac.NetInterface = "localhost:9000"
	}
	if ac.Timeout <= 0 {
		ac.Timeout = 10 * time.Millisecond
	}
}

// NewApp returns a new ready-to-launch App object with adjusted settings.
func NewApp(ctx context.Context, appCfg AppConfig) (*App, func(), error) {
	log.Debug().Interface("web app config", appCfg).Msg("starting initialize web application")

	appCfg.checkConfig()

	r := chi.NewRouter()

	r.Use(middleware.Timeout(appCfg.Timeout))
	r.Use(logMiddleware)

	r.Route("/", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("{\"message\" : \"200 OK\"}")); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Err(err).Msg("can not write data to response")
			}
		})
	})

	log.Debug().Strs("endpoints", []string{"GET /healthcheck"}).Msg("endpoints registered")

	srv := &http.Server{
		Addr:    appCfg.NetInterface,
		Handler: r,
	}
	a := &App{srv: srv}
	return a, a.Close(ctx), nil
}

// Run start the server.
func (a *App) Run() error {
	log.Info().Msg("database server started")
	if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		log.Debug().
			Str("method", r.Method).
			Str("remote", r.RemoteAddr).
			Str("path", r.URL.Path).
			Int("duration", int(time.Since(start))).
			Msgf("called url %s", r.URL.Path)
	})
}

// Close smoothly stops the server with the completion of all network connections with a specified timeout.
func (a *App) Close(ctx context.Context) func() {
	return func() {
		log.Debug().Msg("start shutting down server")
		tctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()
		if err := a.srv.Shutdown(tctx); err != nil {
			log.Err(err).Msg("can not shutdown server correctly")
			return
		}
		log.Info().Msg("server stopped correctly")
	}
}
