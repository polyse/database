// Package web is responsible for creating and initializing endpoints for interacting with the database.
//
package web

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

type Document struct {
	Title   string `json:"title" validate:"required"`
	URL     string `json:"url" validate:"required"`
	Content string `json:"content" validate:"required"`
}

func newDocuments() *[]Document {
	return &[]Document{}
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
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

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Use(middleware.Logger())

	e.GET("/healthcheck", handleHealthcheck)
	e.GET("/api/:collection/documents", handleSearch)
	e.POST("/api/:collection/documents", handleAddDocuments)

	// log.Debug().Strs("endpoints", []string{"GET /healthcheck"}).Msg("endpoints registered")

	srv := &http.Server{
		Addr:    appCfg.NetInterface,
		Handler: e,
	}
	a := &App{srv: srv}
	return a, a.Close(ctx), nil
}

func handleHealthcheck(c echo.Context) error {
	encodedJSON := []byte(`{"message": "200 OK"}`) // Encoded JSON from external source
	return c.JSONBlob(http.StatusOK, encodedJSON)
}

func handleSearch(c echo.Context) error {
	collection := c.Param("collection")
	query := c.QueryParam("q")
	if len(query) == 0 {
		encodedJSON := []byte(`{"message": "400 Bad request"}`)
		return c.JSONBlob(http.StatusOK, encodedJSON)
	}
	// limit := c.QueryParam("limit")

	encodedJSON := []byte(`{"message": "200 OK"}`) // Encoded JSON from external source
	return c.JSONBlob(http.StatusOK, encodedJSON)
}

func handleAddDocuments(c echo.Context) error {
	// collection := c.Param("collection")
	documents := newDocuments()
	if err := c.Bind(documents); err != nil {
		return err
	}
	for _, document := range *documents {
		if err := c.Validate(document); err != nil {
			encodedJSON := []byte(`{"message": "400 Bad request"}`)
			return c.JSONBlob(http.StatusOK, encodedJSON)
		}
	}

	// here will be sending documents to bd

	encodedJSON := []byte(`{"message": "200 OK"}`)
	return c.JSONBlob(http.StatusOK, encodedJSON)
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
