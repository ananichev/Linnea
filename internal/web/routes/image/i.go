package image

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JoachimFlottorp/Linnea/internal/ctx"
	"github.com/JoachimFlottorp/Linnea/internal/models"
	"github.com/JoachimFlottorp/Linnea/internal/web/router"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Route struct {
	Ctx ctx.Context
}

func NewImage(gCtx ctx.Context) router.Route {
	return &Route{gCtx}
}

func (a *Route) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI: "/i/{id}",
		Method: []string{http.MethodGet},
		Children: []router.Route{},
		Middleware: []mux.MiddlewareFunc{},
	}
}

func (a *Route) Handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if _, ok := vars["id"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
	}

	id := vars["id"]
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	img := aws.NewWriteAtBuffer([]byte{})
	
	if err := a.Ctx.Inst().Storage.DownloadFile(r.Context(), img, &s3.GetObjectInput{
		Key:   aws.String(id),
	}); err != nil {
		switch err.Error() {
		case s3.ErrCodeNoSuchKey: {
			w.WriteHeader(http.StatusNotFound)
			zap.S().Error(err)
			return
		}
		default: {
			w.WriteHeader(http.StatusInternalServerError)
			zap.S().Errorw("Error downloading image", "error", err)
			return
		}
		}
	}

	file 	:= models.File{}
	key 	:= fmt.Sprintf("file:%s", id)
	
	fileSrc, err := a.Ctx.Inst().Redis.Get(r.Context(), key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		zap.S().Errorw("Error getting file source", "error", err)
		return
	}

	if err := json.Unmarshal([]byte(fileSrc), &file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		zap.S().Errorw("Error unmarshalling file source", "error", err)
		return
	}

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(img.Bytes())))
	w.Write(img.Bytes())
}