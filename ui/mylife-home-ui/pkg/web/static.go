package web

import (
	webapp "mylife-home-ui-webapp"
	"net/http"
)

func (ws *WebServer) setupStatic(mux *http.ServeMux) {
	// Create a file server for static files only
	fileServer := http.FileServer(http.FS(webapp.FS))

	// Handle static files - this will only serve files that actually exist
	mux.Handle("/", fileServer)
}
