package web

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/JoachimFlottorp/Linnea/internal/ctx"
	"github.com/JoachimFlottorp/Linnea/internal/web/router"
	"github.com/JoachimFlottorp/Linnea/internal/web/routes/api"
	"github.com/JoachimFlottorp/Linnea/internal/web/routes/auth"
	"github.com/JoachimFlottorp/Linnea/internal/web/routes/image"
	"github.com/JoachimFlottorp/Linnea/internal/web/routes/upload"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type l struct {
	logger *zap.SugaredLogger
}

func (log *l) Write(p []byte) (int, error) {
	log.logger.Errorw("HTTPError", "error", string(p))
	return len(p), nil
}

type Server struct {
	listener net.Listener
	router *mux.Router
}

func New(gCtx ctx.Context) error {	
	port := gCtx.Config().Http.Port
	addr := fmt.Sprintf("%s:%d", "localhost", port)

	s := Server{}

	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.router = mux.NewRouter().StrictSlash(false)	

	logger := log.New(&l{zap.S()}, "", 0)
	
	server := http.Server{
		Handler: s.router,
		ErrorLog: logger,
	}

	s.setupRoutes(api.API(gCtx), s.router)
	s.setupRoutes(upload.NewUpload(gCtx), s.router)
	s.setupRoutes(auth.NewAuth(gCtx), s.router)
	s.setupRoutes(image.NewImage(gCtx), s.router)

	go func() {
		<-gCtx.Done()

		_ = server.Shutdown(gCtx)
	}()

	return server.Serve(s.listener)
}

func (s *Server) setupRoutes(r router.Route, parent *mux.Router) {
	routeConfig := r.Configure()

	route := parent.
		PathPrefix(routeConfig.URI).
		Methods(routeConfig.Method...).
		Subrouter().
		StrictSlash(false)

	// Allow endpoint without trailing slash
	route.HandleFunc("", r.Handler)
	route.HandleFunc("/", r.Handler)

	zap.S().
		With("route", routeConfig.URI,).
		Debug("Setup route")

	for _, child := range routeConfig.Children {
		s.setupRoutes(child, route)
	}

	for _, middleware := range routeConfig.Middleware {
		route.Use(middleware)
	}
}