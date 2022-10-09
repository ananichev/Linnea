package middleware

import (
	"net/http"
	"strings"

	"github.com/JoachimFlottorp/Linnea/internal/ctx"
	"github.com/JoachimFlottorp/Linnea/internal/redis"
	"github.com/JoachimFlottorp/Linnea/internal/web/router"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/golang-jwt/jwt/v4"
)

// Validates the users JWT session token
func Auth(gCtx ctx.Context) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a := r.Header.Get("Authorization")
			if a == "" {
				router.JSON(w, http.StatusUnauthorized, router.ApiResult{
					Success: false,
					Data:    "Missing Authorization header",
				})
				return
			}

			b := strings.Split(a, " ")
			if len(b) != 2 || b[0] != "Bearer" {
				router.JSON(w, http.StatusUnauthorized, router.ApiResult{
					Success: false,
					Data:    "Invalid Authorization header",
				})
				return
			}
			t := b[1]

			secret := gCtx.Config().Http.Jwt.Secret

			_, err := redis.GetUser(r.Context(), gCtx.Inst().Redis, secret, t)

			if err != nil {
				var msg string
				switch err.(*jwt.ValidationError).Errors {
				case jwt.ValidationErrorExpired:
					msg = "Token has expired, please login again"
				default:
					{
						zap.S().Warnw("Failed to validate JWT token", "error", err)
						msg = "Invalid token or user"
					}
				}

				router.JSON(w, http.StatusUnauthorized, router.ApiResult{
					Success: false,
					Data:    msg,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
