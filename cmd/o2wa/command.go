package o2wa

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"text/template"
	"time"
)

//go:embed defaultPage.html
var defaultPage string

type Endpoint struct {
	Path           string   `json:"path" yaml:"path"`
	Public         bool     `json:"public,omitempty" yaml:"public,omitempty"`
	RequiredGroups []string `json:"requiredGroups,omitempty" yaml:"requiredGroups,omitempty"`
	Command        []string `json:"command" yaml:"command"`
	BinaryOutput   bool     `json:"binaryOutput,omitempty" yaml:"binaryOutput,omitempty"`
	HTMLFile       string   `json:"htmlFile" yaml:"htmlFile"`
}

type templateVariables struct {
	GET                 map[string]string
	POST                map[string]string
	HTTPRequestHeaders  map[string]string
	HTTPResponseHeaders map[string]string
	CSRFToken           string
	AuthClaims          *idTokenClaims
}

type outputLine struct {
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Text      string    `json:"text"`
}

func (app *Server) HandleCommandConfirm(w http.ResponseWriter, r *http.Request, e Endpoint) {
	templateVariables := app.templateVariables(w, r)

	var tmpl *template.Template
	var err error

	if _, err := os.Stat(e.HTMLFile); err == nil {
		tmpl, err = template.ParseFiles(e.HTMLFile)
	} else {
		if e.HTMLFile != "" {
			http.Error(w, "Failed to load HTML file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl, err = template.New("defaultPage").Parse(defaultPage)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, templateVariables)
}

func (app *Server) HandleCommandRun(w http.ResponseWriter, r *http.Request, e Endpoint) {
	if !app.validateCSRFToken(r) {
		http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
		return
	}

	err := app.executeCommand(w, r, e)
	if err != nil {
		http.Error(w, "Failed to execute command: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *Server) executeCommand(w http.ResponseWriter, r *http.Request, e Endpoint) error {
	command, err := app.parseCommand(w, r, e)
	if err != nil {
		return err
	}

	log.Printf("%s requested %s, running: %s\n", app.auth(r).IDTokenClaims.PreferredUsername, e.Path, command)

	cmd := exec.Command(command[0], command[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	var writeMutex sync.Mutex

	if e.BinaryOutput {
		w.Header().Set("Content-Type", "application/octet-stream")
		wg.Add(1)
		go func() {
			defer wg.Done()
			io.Copy(w, stdout) // Only streaming stdout for binary data
		}()
	} else {
		w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
		wg.Add(2)

		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := outputLine{
					Timestamp: time.Now(),
					Source:    "stdout",
					Text:      scanner.Text(),
				}
				jsonLine, _ := json.Marshal(line)

				writeMutex.Lock()
				w.Write(jsonLine)
				w.Write([]byte("\n"))
				writeMutex.Unlock()
			}
		}()

		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := outputLine{
					Timestamp: time.Now(),
					Source:    "stderr",
					Text:      scanner.Text(),
				}
				jsonLine, _ := json.Marshal(line)

				writeMutex.Lock()
				w.Write(jsonLine)
				w.Write([]byte("\n"))
				writeMutex.Unlock()
			}
		}()
	}

	wg.Wait() // Wait for all goroutines to finish

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (app *Server) parseCommand(w http.ResponseWriter, r *http.Request, e Endpoint) ([]string, error) {
	templateVariables := app.templateVariables(w, r)

	parsedCommand := make([]string, len(e.Command))
	for i, v := range e.Command {
		tmpl, err := template.New("command").Parse(v)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, templateVariables)
		if err != nil {
			return nil, err
		}
		parsedCommand[i] = buf.String()
	}

	return parsedCommand, nil
}

func (app *Server) templateVariables(w http.ResponseWriter, r *http.Request) templateVariables {
	templateVariables := templateVariables{
		CSRFToken:  app.csrfToken(r),
		AuthClaims: &app.auth(r).IDTokenClaims,
	}

	templateVariables.GET = make(map[string]string)
	for k, v := range r.URL.Query() {
		templateVariables.GET[k] = v[0]
	}

	templateVariables.POST = make(map[string]string)
	for k, v := range r.PostForm {
		templateVariables.POST[k] = v[0]
	}

	templateVariables.HTTPRequestHeaders = make(map[string]string)
	for k, v := range r.Header {
		templateVariables.HTTPRequestHeaders[k] = v[0]
	}

	templateVariables.HTTPResponseHeaders = make(map[string]string)
	for k, v := range w.Header() {
		templateVariables.HTTPResponseHeaders[k] = v[0]
	}

	return templateVariables
}
