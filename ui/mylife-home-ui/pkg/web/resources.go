package web

import "net/http"

func (ws *WebServer) setupResources(mux *http.ServeMux) {
	mux.HandleFunc("/resources/", ws.handleGetResource)
}

func (ws *WebServer) handleGetResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract resource hash from URL path
	// URL format: /resources/{hash}
	hash := r.URL.Path[len("/resources/"):]

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
