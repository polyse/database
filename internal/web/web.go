// Package web is responsible for creating and initializing endpoints for interacting with the database.
//
package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	// "github.com/polyse/databaseinternal/collection"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
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

// Source structure for domain\article\site\source description
type Source struct {
	Date  time.Time `json:"date" validate:"required"` // format: 2006-01-02T15:04:05+07:00
	Title string    `json:"title" validate:"required"`
}

// RawData structure for json data description
type RawData struct {
	Source `json:"source" validate:"required,dive"`
	Url    string `json:"url" validate:"required,url"`
	Data   string `json:"data" validate:"required"`
}

// Documents is type to Bind for get []RawData
type Documents struct {
	Documents []RawData `json:"documents" validate:"required,dive"`
}

// SearchRequest is strust for storage and validate query param
type SearchRequest struct {
	Query  string `validate:"required"`
	Limit  int    `validate:"gte=0"`
	Offset int    `validate:"gte=0"`
}

// Validator - to add custom validator in echo.
type Validator struct {
	validator *validator.Validate
}

// Validate add go-playground/validator in echo.
func (cv *Validator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// WebError - to add custom msg and code to err.
type WebError struct {
	err  error
	code int
	msg  string
}

func (w WebError) Error() string {
	return w.err.Error()
}

// WrapWebError is wrap code, message and err in WebError.
func WrapWebError(code int, err error) WebError {
	return WebError{code: code, err: err, msg: http.StatusText(code)}
}

func httpErrorHandler(err error, c echo.Context) {
	log.Err(err).Msg("web exception")

	var errJSON error
	if we, ok := err.(WebError); ok {
		errJSON = c.JSONBlob(we.code, []byte(fmt.Sprintf(`{"message": "%d %s"}`, we.code, we.msg)))
	} else {
		errJSON = c.JSONBlob(http.StatusInternalServerError, []byte(fmt.Sprintf(`{"message": "%d %s"}`, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))))
	}

	if errJSON != nil {
		log.Err(errJSON).Msg("can not write error to response")
	}
}

func ok(c echo.Context) error {
	encodedJSON := []byte(fmt.Sprintf(`{"message": "%d %s"}`, http.StatusOK, http.StatusText(http.StatusOK)))
	return c.JSONBlob(http.StatusOK, encodedJSON)
}

// NewApp returns a new ready-to-launch App object with adjusted settings.
func NewApp(ctx context.Context, appCfg AppConfig) (*App, func(), error) {
	log.Debug().Interface("web app config", appCfg).Msg("starting initialize web application")

	appCfg.checkConfig()

	e := echo.New()
	e.Validator = &Validator{validator: validator.New()}
	e.Use(logMiddleware())
	e.HTTPErrorHandler = httpErrorHandler

	e.GET("/healthcheck", handleHealthcheck)

	g := e.Group("/api")
	g.GET("/:collection/documents", handleSearch)
	g.POST("/:collection/documents", handleAddDocuments)

	log.Debug().Msg("endpoints registered")

	srv := &http.Server{
		Addr:    appCfg.NetInterface,
		Handler: e,
	}
	a := &App{srv: srv}
	return a, a.Close(ctx), nil
}

func handleHealthcheck(c echo.Context) error {
	log.Debug().Msg("handleHealthcheck run")
	return ok(c)
}

func handleSearch(c echo.Context) error {
	var request SearchRequest
	var err error

	request.Query = c.QueryParam("q")
	request.Limit, _ = strconv.Atoi(c.QueryParam("limit"))
	request.Offset, _ = strconv.Atoi(c.QueryParam("offset"))

	collection := c.Param("collection")

	log.Debug().
		Str("collection", collection).
		Str("q", request.Query).
		Int("limit", request.Limit).
		Msg("handleSearch run")

	if err = c.Validate(request); err != nil {
		return WrapWebError(http.StatusBadRequest, err)
	}

	// here will be searching

	return ok(c)
}

func handleAddDocuments(c echo.Context) error {
	collection := c.Param("collection")

	log.Debug().
		Str("collection", collection).
		Msg("handleSearch run")

	var docs Documents
	if err := c.Bind(&docs); err != nil {
		return WrapWebError(http.StatusBadRequest, err)
	}
	if err := c.Validate(docs); err != nil {
		return WrapWebError(http.StatusBadRequest, err)
	}

	// here will be sending document to bd

	return ok(c)
}

// Run start the server.
func (a *App) Run() error {
	log.Info().Msg("database server started")
	if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func logMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			start := time.Now()

			log.Debug().
				Str("method", c.Request().Method).
				Str("remote", c.Request().RemoteAddr).
				Str("path", c.Path()).
				Int("duration", int(time.Since(start))).
				Msgf("called url %s", c.Request().URL)

			return next(c)
		}
	}
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
