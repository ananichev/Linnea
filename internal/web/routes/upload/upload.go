package upload

import (
	"fmt"
	"net/http"

	"github.com/JoachimFlottorp/Linnea/internal/ctx"
	"github.com/JoachimFlottorp/Linnea/internal/image"
	"github.com/JoachimFlottorp/Linnea/internal/models"
	"github.com/JoachimFlottorp/Linnea/internal/redis"
	"github.com/JoachimFlottorp/Linnea/internal/web/middleware"
	"github.com/JoachimFlottorp/Linnea/internal/web/router"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Route struct {
	Ctx ctx.Context
}

func NewUpload(gCtx ctx.Context) router.Route {
	return &Route{gCtx}
}

func (a *Route) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI:      "/upload",
		Method:   []string{http.MethodPost},
		Children: []router.Route{},
		Middleware: []mux.MiddlewareFunc{
			middleware.Auth(a.Ctx),
		},
	}
}

// 128 MB
const MAX_SIZE = 128 * 1024 * 1024

func (a *Route) Handler(w http.ResponseWriter, r *http.Request) {
	secret := a.Ctx.Config().Http.Jwt.Secret
	auth := router.GetBearerToken(r)
	if auth == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := redis.GetUser(r.Context(), a.Ctx.Inst().Redis, secret, auth)
	if err != nil {
		zap.S().Warnw("Failed to get user from redis", "error", err)

		router.JSON(w, http.StatusInternalServerError, router.ApiResult{
			Success: false,
			Data:    "Internal server error",
		})
		return
	}

	if err := r.ParseMultipartForm(MAX_SIZE); err != nil {
		router.JSON(w, http.StatusBadRequest, router.ApiResult{
			Success: false,
			Data:    err.Error(),
		})
		return
	}

	f, multiHeader, err := r.FormFile("file")
	if err != nil {
		router.JSON(w, http.StatusBadRequest, router.ApiResult{
			Success: false,
			Data:    "Missing 'file' form field",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MAX_SIZE)

	file, err := image.Read(f, multiHeader, r)
	if err != nil {
		router.JSON(w, http.StatusBadRequest, router.ApiResult{
			Success: false,
			Data:    "Invalid file",
		})
		return
	}

	if err := a.Ctx.Inst().Storage.UploadFile(r.Context(), &s3manager.UploadInput{
		Body:         aws.ReadSeekCloser(file.Data),
		Key:          aws.String(file.Name),
		ContentType:  aws.String(file.ContentType),
		CacheControl: aws.String("public, max-age=15552000"),
	}); err != nil {
		router.JSON(w, http.StatusInternalServerError, router.ApiResult{
			Success: false,
			Data:    "Internal server error",
		})

		zap.S().Errorw("Failed to upload file", "error", err)
		return
	}

	// TODO: Add to database

	{
		file := models.File{
			OwnerID:     user.TwitchUID,
			Name:        file.Name,
			ContentType: file.ContentType,
		}

		fileSer, err := file.ToString()
		if err != nil {
			zap.S().Errorw("Failed to serialize file", "error", err)
			router.JSON(w, http.StatusInternalServerError, router.ApiResult{
				Success: false,
				Data:    "Internal server error",
			})
			return
		}

		key := fmt.Sprintf("file:%s", file.Name)

		if err := a.Ctx.Inst().Redis.Set(r.Context(), key, fileSer); err != nil {
			zap.S().Errorw("Failed to add file to user", "error", err)
			router.JSON(w, http.StatusInternalServerError, router.ApiResult{
				Success: false,
				Data:    "Internal server error",
			})
			return
		}
	}

	url := fmt.Sprintf("%s/i/%s", a.Ctx.Config().Http.PublicAddr, file.Name)

	// Write to response for ShareX / Chatterino
	w.Write([]byte(url))
}
