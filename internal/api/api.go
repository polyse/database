// Package api is responsible for creating and initializing endpoints for interacting with the database.
//
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/polyse/database/internal/collection"
	"github.com/rs/zerolog/log"
)

// Context structure for handle context from main.
type Context struct {
	echo.Context
	Ctx context.Context
}

// API structure containing the necessary server settings and responsible for starting and stopping it.
type API struct {
	e    *echo.Echo
	addr string
	*collection.Manager
}

// AppConfig structure containing the server settings necessary for its operation.
type AppConfig struct {
	NetInterface string
	Timeout      time.Duration
}

func (ac *AppConfig) checkConfig() {
	log.Debug().Msg("checking api application config")

	if ac.NetInterface == "" {
		ac.NetInterface = "localhost:9000"
	}
	if ac.Timeout <= 0 {
		ac.Timeout = 10 * time.Millisecond
	}
}

// Documents is type to Bind for get []RawData
type Documents struct {
	Documents []collection.RawData `json:"documents" validate:"required,dive"`
}

// SearchRequest is strust for storage and validate query param.
type SearchRequest struct {
	Query  string `validate:"required" query:"q"`
	Limit  int    `validate:"gte=0" query:"limit"`
	Offset int    `validate:"gte=0" query:"offset"`
}

// Validator - to add custom validator in echo.
type Validator struct {
	validator *validator.Validate
}

// Validate add go-playground/validator in echo.
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

// NewApp returns a new ready-to-launch API object with adjusted settings.
func NewApp(ctx context.Context, appCfg AppConfig) (*API, error) {
	appCfg.checkConfig()

	log.Debug().Interface("api app config", appCfg).Msg("starting initialize api application")

	e := echo.New()

	a := &API{
		e:    e,
		addr: appCfg.NetInterface,
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &Context{
				Context: c,
				Ctx:     ctx,
			}
			return next(cc)
		}
	})
	e.Validator = &Validator{validator: validator.New()}
	e.Use(logMiddleware)

	e.GET("/healthcheck", a.handleHealthcheck)

	g := e.Group("/api")
	g.GET("/:collection/documents", a.handleSearch)
	g.POST("/:collection/documents", a.handleAddDocuments)

	log.Debug().Msg("endpoints registered")

	return a, nil
}

func (a *API) handleHealthcheck(c echo.Context) error {
	return ok(c)
}

func (a *API) handleSearch(c echo.Context) error {
	var err error

	collectionName := c.Param("collection")
	proc, err := a.Manager.GetProcessor(collectionName)
	if err != nil {
		log.Debug().Err(err).Msg("handleSearch GetProcessor err")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	request := &SearchRequest{}
	if err := c.Bind(request); err != nil {
		log.Debug().Err(err).Msg("handleSearch Bind err")
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	log.Debug().
		Str("collection", collectionName).
		Str("q", request.Query).
		Int("limit", request.Limit).
		Msg("handleSearch run")

	if err = c.Validate(request); err != nil {
		log.Debug().Err(err).Msg("handleSearch Validate err")
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	r, err := proc.ProcessAndGet(request.Query, request.Limit, request.Offset)

	if err != nil {
		log.Err(err).Msg("saving error")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, r)
}

func (a *API) handleAddDocuments(c echo.Context) error {
	collectionName := c.Param("collection")

	log.Debug().
		Str("collection", collectionName).
		Msg("adding documents")

	proc, err := a.Manager.GetProcessor(collectionName)
	if err != nil {
		log.Debug().Err(err).Msg("handleAddDocuments GetProcessor err")
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	docs := &Documents{}
	if err = c.Bind(docs); err != nil {
		log.Debug().Err(err).Msg("handleAddDocuments Bind err")
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	log.Debug().Interface("docs", docs).Msg("before validating")
	if len(docs.Documents) == 0 {
		return c.JSON(http.StatusOK, docs)
	}

	if err = c.Validate(docs); err != nil {
		log.Debug().Err(err).Msg("handleAddDocuments Validate err")
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if err = proc.ProcessAndInsertString(docs.Documents); err != nil {
		log.Debug().Err(err).Msg("handleAddDocuments ProcessAndInsertString err")
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	return c.JSON(http.StatusCreated, docs)
}

// Run start the server.
func (a *API) Run() error {
	return a.e.Start(a.addr)
}

// Close stop the server.
func (a *API) Close() error {
	log.Debug().Msg("shutting down server")
	return a.e.Close()
}

func ok(c echo.Context) error {
	return c.JSON(http.StatusOK, http.StatusText(http.StatusOK))
}

func logMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()
		start := time.Now()

		err := next(c)

		stop := time.Now()

		log.Debug().
			Str("remote", req.RemoteAddr).
			Str("user_agent", req.UserAgent()).
			Str("method", req.Method).
			Str("path", c.Path()).
			Int("status", res.Status).
			Dur("duration", stop.Sub(start)).
			Str("duration_human", stop.Sub(start).String()).
			Msgf("called url %s", req.URL)

		return err
	}
}
