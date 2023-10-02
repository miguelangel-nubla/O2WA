package o2wa

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc"
	"github.com/grokify/go-pkce"
	"golang.org/x/oauth2"
)

func (app *Server) oauth2Start(w http.ResponseWriter, r *http.Request, redirectURL string) {
	state, err := randString(16)
	if err != nil {
		handleError(w, "Internal error", http.StatusInternalServerError)
		return
	}
	nonce, err := randString(16)
	if err != nil {
		handleError(w, "Internal error", http.StatusInternalServerError)
		return
	}

	flowData := oauthFlowData{State: state, Nonce: nonce, Redirect: redirectURL}
	app.sessionManager.Put(r.Context(), app.oauthKey, flowData)

	codeChallenge := pkce.CodeChallengeS256(app.codeVerifier)

	authURL := app.config.AuthCodeURL(
		state,
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam(pkce.ParamCodeChallenge, codeChallenge),
		oauth2.SetAuthURLParam(pkce.ParamCodeChallengeMethod, pkce.MethodS256),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (app *Server) setAuth(r *http.Request, data *authData) error {
	app.sessionManager.Put(r.Context(), app.sessionKey, data)
	return nil
}

func (app *Server) auth(r *http.Request) *authData {
	value := app.sessionManager.Get(r.Context(), app.sessionKey)
	if value == nil {
		// errors.New("no authData found in session")
		return nil
	}

	data, ok := value.(authData)
	if !ok {
		// errors.New("session data invalid")
		return nil
	}

	return &data
}

func (app *Server) csrfToken(r *http.Request) string {
	value := app.sessionManager.Get(r.Context(), app.csrfKey)
	if value != nil {
		csrfToken, ok := value.(string)
		if ok {
			return csrfToken
		}
	}

	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("Failed to generate CSRF token:", err)
	}

	csrfToken := base64.StdEncoding.EncodeToString(b)
	app.sessionManager.Put(r.Context(), app.csrfKey, csrfToken)

	return csrfToken
}

func (app *Server) validateCSRFToken(r *http.Request) bool {
	headerToken := r.Header.Get("X-CSRF-Token")

	value := app.sessionManager.Get(r.Context(), app.csrfKey)
	if value != nil {
		token, ok := value.(string)
		if ok && token == headerToken {
			return true
		}
	}

	return false
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func constructURL(basePath string, redirectParameter string) (string, error) {
	u, err := url.Parse(basePath)
	if err != nil {
		return "", err
	}

	query := u.Query()
	query.Set("redirect", redirectParameter)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func handleError(w http.ResponseWriter, err string, code int) {
	http.Error(w, err, code)
}
