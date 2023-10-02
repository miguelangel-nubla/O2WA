package o2wa

import (
	"context"
	"encoding/gob"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/grokify/go-pkce"
	"golang.org/x/oauth2"

	"github.com/alexedwards/scs/v2"
)

type idTokenClaims struct {
	Iss               string   `json:"iss"`
	Sub               string   `json:"sub"`
	Aud               string   `json:"aud"`
	Exp               int64    `json:"exp"`
	Nbf               int64    `json:"nbf"`
	Iat               int64    `json:"iat"`
	Nonce             string   `json:"nonce"`
	Azp               string   `json:"azp"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	Scopes            []string `json:"scopes"`
	Groups            []string `json:"groups"`
}

type authData struct {
	OAuth2Token   oauth2.Token
	IDTokenClaims idTokenClaims
}

type oauthFlowData struct {
	State    string
	Nonce    string
	Redirect string
}

type Server struct {
	ctx            context.Context
	config         oauth2.Config
	verifier       *oidc.IDTokenVerifier
	codeVerifier   string
	sessionManager *scs.SessionManager
	sessionKey     string
	oauthKey       string
	csrfKey        string
}

func init() {
	gob.Register(authData{})
	gob.Register(oauthFlowData{})
}

func NewServer(clientID, clientSecret, oidcIssuer, callbackURL string) *Server {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, oidcIssuer)
	if err != nil {
		log.Fatal(err)
	}

	oidcConfig := &oidc.Config{ClientID: clientID}
	verifier := provider.Verifier(oidcConfig)

	codeVerifier, err := pkce.NewCodeVerifier(96)
	if err != nil {
		log.Fatal(err)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode
	sessionManager.Cookie.Secure = true

	return &Server{
		ctx:          ctx,
		verifier:     verifier,
		codeVerifier: codeVerifier,
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  callbackURL,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "groups"},
		},
		sessionManager: sessionManager,
		sessionKey:     "authData",
		oauthKey:       "oauthFlowData",
	}
}

func (app *Server) Oauth2Callback(w http.ResponseWriter, r *http.Request) {
	oauthFlowDataValue := app.sessionManager.Get(r.Context(), app.oauthKey)
	if oauthFlowDataValue == nil {
		handleError(w, "oauth flow session data not found", http.StatusBadRequest)
		return
	}

	oauthFlowData, ok := oauthFlowDataValue.(oauthFlowData)
	if !ok {
		handleError(w, "oauth flow session data invalid", http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("state") != oauthFlowData.State {
		handleError(w, "oauth flow state did not match", http.StatusBadRequest)
		return
	}

	oauth2Token, err := app.config.Exchange(
		app.ctx,
		r.URL.Query().Get("code"),
		oauth2.SetAuthURLParam(pkce.ParamCodeVerifier, app.codeVerifier),
	)
	if err != nil {
		handleError(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	idToken, err := app.verifier.Verify(app.ctx, oauth2Token.Extra("id_token").(string))
	if err != nil {
		handleError(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if idToken.Nonce != oauthFlowData.Nonce {
		handleError(w, "nonce did not match", http.StatusBadRequest)
		return
	}

	app.sessionManager.Remove(r.Context(), app.oauthKey)

	var claims idTokenClaims
	if err := idToken.Claims(&claims); err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authData := authData{
		OAuth2Token:   *oauth2Token,
		IDTokenClaims: claims,
	}

	app.setAuth(r, &authData)
	http.Redirect(w, r, oauthFlowData.Redirect, http.StatusFound)
}
