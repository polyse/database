package web

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/xlab/closer"
)

type App struct {
	srv *http.Server
}

type AppConfig struct {
	NetInterface string
	Timeout      time.Duration
}

func NewApp(appCfg AppConfig) (App, error) {
	log.Debug().Interface("web app config", appCfg).Msg("starting initialize web application")
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

	return App{srv: srv}, nil
}

func (a *App) Run(c context.Context) error {

	log.Info().Msg("database server started")
	closer.Bind(a.Close(c))
	if err := a.srv.ListenAndServe(); err != nil {
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

func (a *App) Close(c context.Context) func() {
	return func() {
		log.Debug().Msg("start shutting down server")
		ctx, cancel := context.WithTimeout(c, 10*time.Millisecond)
		defer cancel()
		if err := a.srv.Shutdown(ctx); err != nil {
			log.Err(err).Msg("can not shutdown server correctly")
			return
		}
		log.Info().Msg("server stopped correctly")
	}
}
