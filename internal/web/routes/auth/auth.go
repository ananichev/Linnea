package auth

import (
	"net/http"
	"strings"

	"github.com/JoachimFlottorp/Linnea/internal/ctx"
	"github.com/JoachimFlottorp/Linnea/internal/web/router"
	"github.com/JoachimFlottorp/Linnea/pkg/url"

	"github.com/gorilla/mux"
)

type Route struct {
	Ctx ctx.Context
}

func NewAuth(gCtx ctx.Context) router.Route {
	return &Route{gCtx}
}

func (a *Route) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI:    "/auth",
		Method: []string{http.MethodGet},
		Children: []router.Route{
			newCallback(a.Ctx),
		},
		Middleware: []mux.MiddlewareFunc{},
	}
}

func (a *Route) Handler(w http.ResponseWriter, r *http.Request) {
	urlBuilder := url.NewURLBuilder()

	urlBuilder.SetURL(TWITCH_AUTH_URL)

	urlBuilder.AddParam("response_type", "code")
	urlBuilder.AddParam("client_id", a.Ctx.Config().Twitch.ClientID)

	urlBuilder.AddParam("redirect_uri", strings.Join([]string{a.Ctx.Config().Http.PublicAddr, "auth", "callback"}, "/"))
	urlBuilder.AddParam("scope", "user:read:email")

	url := urlBuilder.Build()

	router.Redirect(w, r, url)
}
