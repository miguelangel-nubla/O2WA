package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/miguelangel-nubla/o2wa/cmd/o2wa"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var configSchema []byte

type Config struct {
	ClientID     string          `json:"clientID"`
	ClientSecret string          `json:"clientSecret"`
	OidcIssuer   string          `json:"oidcIssuer"`
	CallbackURL  string          `json:"callbackURL"`
	Endpoints    []o2wa.Endpoint `json:"endpoints"`
	TLSCert      string          `json:"tlsCert,omitempty"`
	TLSKey       string          `json:"tlsKey,omitempty"`
	ListenAddr   string          `json:"listenAddr"`
}

func main() {
	// Load configuration from JSON
	config := loadConfig("config.json")

	app := o2wa.NewServer(config.ClientID, config.ClientSecret, config.OidcIssuer, config.CallbackURL)

	for _, endpoint := range config.Endpoints {
		handler := func(e o2wa.Endpoint) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// If it's a GET request, serve the HTML file
				if r.Method == http.MethodGet {
					app.HandleCommandConfirm(w, r, e)
					return
				}

				// If it's a POST request, handle the command
				if r.Method == http.MethodPost {
					app.HandleCommandRun(w, r, e)
					return
				}
			}
		}(endpoint)

		if endpoint.Public {
			http.HandleFunc(endpoint.Path, handler)
		} else {
			http.HandleFunc(endpoint.Path, app.AuthMiddleware(endpoint.RequiredGroups, handler))
		}
	}

	u, err := url.Parse(config.CallbackURL)
	if err != nil {
		log.Fatalf("Failed to parse callbackURL: %s", err)
	}

	http.HandleFunc("/values", app.AuthMiddleware([]string{}, app.EndpointValues))
	http.HandleFunc(u.Path, app.Oauth2Callback)

	server := &http.Server{
		Addr:    config.ListenAddr,
		Handler: app.SessionMiddleware(http.DefaultServeMux),
	}

	if config.TLSCert != "" && config.TLSKey != "" {
		log.Printf("listening on https://%s/", config.ListenAddr)
		log.Fatal(server.ListenAndServeTLS(config.TLSCert, config.TLSKey))
	} else {
		log.Printf("listening on http://%s/", config.ListenAddr)
		log.Fatal(server.ListenAndServe())
	}
}

func loadConfig(filename string) Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open config file: %s", err)
	}
	defer file.Close()

	// Load the JSON content
	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Failed to decode config file: %s", err)
	}

	// Load the schema and document
	schemaLoader := gojsonschema.NewBytesLoader(configSchema)
	documentLoader := gojsonschema.NewGoLoader(config)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		log.Fatalf("Failed to validate config: %s", err)
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			// Print error description, for example
			log.Printf("- %s\n", desc)
		}
		log.Fatal("Config validation failed")
	}

	return config
}
