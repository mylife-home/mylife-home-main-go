package web

import (
	"fmt"
	"net/http"
)

// setupRoutes configures the API routes
func (ws *WebServer) setupRepository(mux *http.ServeMux) {
	// API routes
	mux.HandleFunc("/repository/components", ws.handleGetComponents)
	mux.HandleFunc("/repository/action/", ws.handleExecuteAction)

	// Add more API endpoints as needed
}

// handleGetComponents returns the list of components
func (ws *WebServer) handleGetComponents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	components := ws.registry.GetComponents()
	var componentIds []string

	for _, component := range components {
		componentIds = append(componentIds, component.Id())
	}

	w.Header().Set("Content-Type", "application/json")

	// Simple JSON response - you might want to use a proper JSON library
	fmt.Fprintf(w, `{"components": %v}`, componentIds)
}

// handleExecuteAction handles component action execution
func (ws *WebServer) handleExecuteAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract componentId and actionName from URL path
	// This is a simplified implementation - you might want to use a router library
	// URL format: /api/repository/action/{componentId}/{actionName}

	// TODO: Implement action execution logic
	// component := ws.registry.GetComponent(componentId)
	// component.ExecuteAction(actionName, ...)

	w.WriteHeader(http.StatusOK)
}
