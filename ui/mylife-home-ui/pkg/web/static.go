package web

import (
	webapp "mylife-home-ui-webapp"
	"net/http"
)

func (ws *WebServer) setupStatic(mux *http.ServeMux) {
	// Handle all paths - this is the catch-all handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// First, try to serve the file using the standard file server
		fileServer := http.FileServer(http.FS(webapp.FS))

		// Check if the requested file exists
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		if _, err := webapp.FS.Open(path); err != nil {
			// File doesn't exist, serve index.html instead
			r.URL.Path = "/index.html"
		}

		fileServer.ServeHTTP(w, r)
	})
}
