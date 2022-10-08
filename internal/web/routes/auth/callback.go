package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/JoachimFlottorp/Linnea/internal/auth"
	"github.com/JoachimFlottorp/Linnea/internal/ctx"
	"github.com/JoachimFlottorp/Linnea/internal/models"
	"github.com/JoachimFlottorp/Linnea/internal/redis"
	"github.com/JoachimFlottorp/Linnea/internal/web/router"
	"github.com/JoachimFlottorp/Linnea/pkg/url"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type TwitchOAuthResponse struct {
	AccessToken 	string `json:"access_token"`
	RefreshToken 	string `json:"refresh_token"`
	ExpiresIn 		int `json:"expires_in"`
	Scope 			[]string `json:"scope"`
	TokenType 		string `json:"token_type"`
}

type TwitchGetUsers struct {
	Users []TwitchUser `json:"data"`
}

type TwitchUser struct {
	ID              string    `json:"id"`
	Login           string    `json:"login"`
	DisplayName     string    `json:"display_name"`
	BroadcasterType string    `json:"broadcaster_type"`
	Description     string    `json:"description"`
	ProfileImageURL string    `json:"profile_image_url"`
	OfflineImageURL string    `json:"offline_image_url"`
	ViewCount       int       `json:"view_count"`
	Email           string    `json:"email"`
	CreatedAt       time.Time `json:"created_at"`
}

const (
	TWITCH_AUTH_URL = "https://id.twitch.tv/oauth2/authorize"
	TWITCH_USER_TOKEN_URL = "https://id.twitch.tv/oauth2/token"
)

type CBRoute struct {
	Ctx ctx.Context
	secret string
}

func newCallback(gCtx ctx.Context) router.Route {
	return &CBRoute{gCtx, gCtx.Config().Http.Jwt.Secret}
}

func (a *CBRoute) Configure() router.RouteConfig {
	return router.RouteConfig{
		URI:        "/callback",
		Method:     []string{http.MethodGet},
		Children:   []router.Route{},
		Middleware: []mux.MiddlewareFunc{},
	}
}

// TODO Csrf check

func (a *CBRoute) Handler(w http.ResponseWriter, r *http.Request) {
	code 		:= r.URL.Query().Get("code")
	error 		:= r.URL.Query().Get("error")
	errorDesc 	:= r.URL.Query().Get("error_description")
	
	if error != "" {
		zap.S().Errorf("Error from auth provider: %s", errorDesc)
		
		router.Redirect(w, r, fmt.Sprintf("/?auth=failure&reason=%s", errorDesc))
		return
	}
	
	if code == "" {
		router.Redirect(w, r, "/?auth=failure&reason=No%20Token")
		return
	}

	oauthRes 	:= TwitchOAuthResponse{}
	users 		:= TwitchGetUsers{}

	{
	url	:= url.NewURLBuilder().
			SetURL(TWITCH_USER_TOKEN_URL).
			AddParam("client_id", a.Ctx.Config().Twitch.ClientID).
			AddParam("client_secret", a.Ctx.Config().Twitch.ClientSecret).
			AddParam("code", code).
			AddParam("grant_type", "authorization_code").
			AddParam("redirect_uri", a.Ctx.Config().Http.PublicAddr + "/auth/callback").
			Build()
		
	req, err := http.NewRequestWithContext(r.Context(), "POST", url, nil)
	if err != nil {
		zap.S().Errorw("Error from auth provider", "error", err)
		
		router.Redirect(w, r, "/?auth=failure&reason=internal")
		return
	}
	

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.S().Errorw("Error from auth provider", "error", err)

		router.Redirect(w, r, "/?auth=failure&reason=internal")
		return
	}
	
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			zap.S().Errorw("Error from auth provider", "error", err)
			
		} else {
			zap.S().Errorw("Error from auth provider", "error", string(body))
		}
		
		router.Redirect(w, r, "/?auth=failure&reason=internal")
		return
	}


	err = json.NewDecoder(resp.Body).Decode(&oauthRes)
	if err != nil {
		zap.S().Errorf("Error JSON decode: %s", err)

		router.Redirect(w, r, "/?auth=failure&reason=internal")
		return
	}

	}

	{
	
	urlBuilder := url.NewURLBuilder()

	urlBuilder.SetURL("https://api.twitch.tv/helix/users")

	req, err := http.NewRequestWithContext(r.Context(), "GET", urlBuilder.Build(), nil)
	if err != nil {
		zap.S().Errorw("Error from auth provider", "error", err)
		
		router.Redirect(w, r, "/?auth=failure&reason=internal")
		return
	}

	req.Header.Add("Authorization", "Bearer " + oauthRes.AccessToken)
	req.Header.Add("Client-ID", a.Ctx.Config().Twitch.ClientID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.S().Errorw("Error from auth provider", "error", err)

		router.Redirect(w, r, "/?auth=failure&reason=internal")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			zap.S().Errorw("Error from auth provider", "error", err)
			
		} else {
			zap.S().Errorw("Error from auth provider", "error", body)
		}
		
		router.Redirect(w, r, "/?auth=failure&reason=internal")
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		zap.S().Errorf("Error JSON decode: %s", err)
		
		router.Redirect(w, r, "/?auth=failure&reason=internal")
		return
	}

	}

	if len(users.Users) == 0 {
		router.Redirect(w, r, "/?auth=failure&reason=bad_request")
		return
	}

	user := users.Users[0]
	/* 3 Months */
	expiry := time.Now().Add(time.Hour * 24 * 30 * 3)

	jwt, err := auth.SignJWT(a.secret, auth.JWTClaimUser{
		ID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "LINNEA_AUTH",
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	})
	if err != nil {
		zap.S().Errorf("Error signing JWT: %s", err)
		
		router.Redirect(w, r, "/?auth=failure&reason=internal")
		return
	}

	rds :=  a.Ctx.Inst().Redis

	{
		u := models.UserFromTwitch(user.Login, user.ID)
		redis.CreateUser(rds, r.Context(), &u)

		if err != nil {
			zap.S().Errorf("Error setting user: %s", err)
			
			router.Redirect(w, r, "/?auth=failure&reason=internal")
			return
		}
	}
	
	{
		key := fmt.Sprintf("user:%s:%s", user.ID, "jwt")
		rds.Set(r.Context(), key, jwt)

		rds.Expire(r.Context(), key, time.Until(expiry))
	}
	
	cookie := http.Cookie{
		Name:     	"token",
		Value:    	jwt,
		Expires:  	expiry,
		HttpOnly: 	true,
		Path: 		"/",
		Domain:  	a.Ctx.Config().Http.Cookie.Domain,
		Secure:  	a.Ctx.Config().Http.Cookie.Secure,
	}

	http.SetCookie(w, &cookie)

	router.Redirect(w, r, "/?auth=success")
}