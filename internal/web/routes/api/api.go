package api

import (
	"net/http"

	"github.com/JoachimFlottorp/Linnea/internal/ctx"
	"github.com/JoachimFlottorp/Linnea/internal/web/router"
	"github.com/gorilla/mux"
)

type Route struct {
	Ctx ctx.Context
}

func API(gCtx ctx.Context) router.Route {
	return &Route{gCtx}
}

func (a *Route) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI:        "/api/v1",
		Method:     []string{http.MethodGet, http.MethodPost},
		Children:   []router.Route{},
		Middleware: []mux.MiddlewareFunc{},
	}
}

func (a *Route) Handler(w http.ResponseWriter, r *http.Request) {
	router.JSON(w, http.StatusOK, router.ApiResult{
		Success: true,
		Data:    "Hello World!",
	})
	// router.Redirect(w, r, "/")
}
