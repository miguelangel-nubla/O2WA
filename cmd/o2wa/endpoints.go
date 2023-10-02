package o2wa

import (
	"encoding/json"
	"net/http"
)

func (app *Server) EndpointValues(w http.ResponseWriter, r *http.Request) {
	userInfo := app.templateVariables(w, r)

	// Serialize to JSON
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		http.Error(w, "Failed to serialize user info to JSON", http.StatusInternalServerError)
		return
	}

	// Set the content type and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(userInfoJSON)
}

// create a EndpointLogout function that will be used to log out the user
func (app *Server) EndpointLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	err := app.sessionManager.Destroy(r.Context())
	if err != nil {
		http.Error(w, "Could not end session", http.StatusInternalServerError)
		return
	}

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusFound)
}
