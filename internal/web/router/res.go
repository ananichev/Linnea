package router

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type ApiResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func JSON[T interface{}](w http.ResponseWriter, status int, msg T) {
	w.Header().Set("Content-Type", "application/json")
	j, err := json.MarshalIndent(msg, "", "  ")

	if err != nil {
		zap.S().Warnf("Failed to marshal JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(status)
	w.Write(j)
}

func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func GetBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	b := strings.Split(auth, " ")
	if len(b) != 2 || b[0] != "Bearer" {
		return ""
	}
	return b[1]
}
