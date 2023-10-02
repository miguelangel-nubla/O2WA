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
