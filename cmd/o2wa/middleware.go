package o2wa

import "net/http"

func (app *Server) AuthMiddleware(requiredGroups []string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userInfo := app.auth(r)
		if userInfo == nil {
			app.oauth2Start(w, r, r.URL.String())
			return
		}

		if !app.requireGroups(userInfo, requiredGroups) {
			http.Error(w, "Unauthorized: insufficient group permissions", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (a *Server) SessionMiddleware(handler http.Handler) http.Handler {
	return a.sessionManager.LoadAndSave(handler)
}
