package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type JWTClaimUser struct {
	ID string `json:"u"`

	jwt.RegisteredClaims
}

func AuthenticateJWT(secret, token string, out jwt.Claims) (*jwt.Token, error) {
	claims := &JWTClaimUser{}

	t, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(secret), nil
		},
	)

	out = claims

	if err != nil {
		return nil, err
	}

	if claims.ID == "" {
		return nil, ErrInvalidToken
	}

	return t, nil
}

// Decrypt a JWT token to extract the id of a user
func DecryptJWT(secret, token string) (*JWTClaimUser, error) {
	claims := &JWTClaimUser{}

	_, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	return claims, nil
}

func SignJWT(secret string, claims jwt.Claims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return t.SignedString([]byte(secret))
}
