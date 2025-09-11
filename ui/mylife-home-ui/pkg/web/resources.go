package web

import "net/http"

func (ws *WebServer) setupResources(mux *http.ServeMux) {
	mux.HandleFunc("GET /resources/{hash}", ws.handleGetResource)
}

func (ws *WebServer) handleGetResource(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	if hash == "" {
		http.Error(w, "Resource hash is required", http.StatusBadRequest)
		return
	}

	resource, err := ws.model.GetResource(hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", resource.Mime)
	// Since resources are immutable, we can set long cache headers
	w.Header().Set("Cache-Control", "public, max-age=31557600, s-maxage=31557600, immutable") // 31557600 seconds = 1 year
	w.Write(resource.Data)
}
