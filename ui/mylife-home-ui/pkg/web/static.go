package web

import (
	"io/fs"
	webapp "mylife-home-ui-webapp"
	"net/http"
)

func (ws *WebServer) setupStatic(mux *http.ServeMux) {
	// Create a file server for static files only
	fileServer := http.FileServer(http.FS(webapp.FS))

	// Print the files being served
	// This is useful for debugging and ensuring the files are correctly embedded
	logger.Debug("Serving static files:")
	fs.WalkDir(webapp.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		if err != nil {
			panic(err)
		}
		logger.Debugf("  %s", path)
		return nil
	})

	// Handle static files - this will only serve files that actually exist
	mux.Handle("/", fileServer)
}
