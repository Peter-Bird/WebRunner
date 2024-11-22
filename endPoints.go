package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Endpoint represents the configuration for each HTTP endpoint
type Endpoint struct {
	Path        string
	HandlerFunc http.HandlerFunc
	Description string
	Type        string // "text" or "json"
	Data        string // For text endpoints, the data to be written
}

// endpoints is a slice containing all the HTTP endpoints
var endpoints = []Endpoint{
	{
		Path:        "/whoru",
		Description: "Returns 'wf-run :- I am WebRunner a Web Services Driver Program.'",
		Type:        "text",
		Data:        "wf-run :- Web Services Driver Program.",
	},
	{
		Path:        "/",
		Description: "Displays the main page.",
		Type:        "page",
		Data:        "WebRunner.html",
	},
	{
		Path:        "/favicon.ico",
		Description: "Favicon for the application.",
		Type:        "page",
		Data:        "favicon.ico",
	},
	{
		Path:        "/version",
		Description: "Returns the version of the program.",
		Type:        "text",
		Data:        "1.0.0",
	},
	{
		Path:        "/eps",
		Description: "Returns the list of endpoints.",
		Type:        "json",
	},
}

// getStaticDir dynamically constructs the full path to the static directory
func getStaticDir() string {
	// Get the directory of the currently running program
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}

	// Resolve the directory where the executable resides
	exeDir := filepath.Dir(exePath)

	// Combine the executable directory with the relative static directory path
	staticDir := filepath.Join(exeDir, "static")

	log.Printf("Static directory resolved to: %s", staticDir)
	return staticDir
}

// setupHTTPHandlers sets up all the HTTP endpoints
func setupHTTPHandlers() {
	// Register all endpoints
	for i, ep := range endpoints {
		if ep.HandlerFunc == nil {
			// Generate handler based on type
			switch ep.Type {
			case "page":
				ep.HandlerFunc = createPageHandler(ep.Data)
			case "text":
				// For text endpoints, write the Data variable
				ep.HandlerFunc = createTextHandler(ep.Data)
			case "json":
				// For JSON endpoints, handle /eps separately
				if ep.Path == "/eps" {
					// Special handling for /eps to ensure it captures all endpoints
					ep.HandlerFunc = createJSONHandler(func() interface{} {
						return getEndpointsList(endpoints)
					})
				} else {
					// For other JSON endpoints, parse the Data field
					var jsonData interface{}
					if err := json.Unmarshal([]byte(ep.Data), &jsonData); err != nil {
						log.Printf("Failed to parse JSON data for endpoint %s: %v", ep.Path, err)
						// Assign a handler that returns an error
						ep.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
							http.Error(w, "Invalid JSON data", http.StatusInternalServerError)
						}
					} else {
						// Use createJSONHandler with the parsed data
						ep.HandlerFunc = createJSONHandler(func() interface{} {
							return jsonData
						})
					}
				}
			default:
				// Default handler for unknown types
				ep.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotImplemented)
					w.Write([]byte("Endpoint type not implemented"))
				}
			}
			// Update the endpoint in the slice
			endpoints[i] = ep
		}
		http.HandleFunc(ep.Path, ep.HandlerFunc)
	}
}

// createTextHandler returns an http.HandlerFunc that writes the provided data
func createPageHandler(data string) http.HandlerFunc {

	//log.Println(getStaticDir() + "/" + data)
	return func(w http.ResponseWriter, r *http.Request) {
		//log.Println(getStaticDir() + "/" + data)
		if data == "favicon.ico" {
			w.Header().Set("Content-Type", "image/x-icon")
		}
		http.ServeFile(w, r, getStaticDir()+"/"+data)
	}
}

// createTextHandler returns an http.HandlerFunc that writes the provided data
func createTextHandler(data string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(data))
	}
}

// createJSONHandler returns an http.HandlerFunc that writes the provided data as JSON.
// The dataFunc is a function that returns the data to be encoded in JSON.
func createJSONHandler(dataFunc func() interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := dataFunc()
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		}
	}
}

// getEndpointsList returns a slice of maps containing endpoint information.
func getEndpointsList(endpoints []Endpoint) []map[string]string {
	eps := []map[string]string{}
	for _, ep := range endpoints {
		eps = append(eps, map[string]string{
			"endpoint":    ep.Path,
			"description": ep.Description,
			"type":        ep.Type,
		})
	}
	return eps
}
