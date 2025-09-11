package web

import (
	"encoding/json"
	"mylife-home-common/components/metadata"
	"net/http"
	"time"
)

func (ws *WebServer) setupRepository(mux *http.ServeMux) {
	mux.HandleFunc("GET /repository/components", ws.handleGetComponents)
	mux.HandleFunc("GET /repository/state/{componentId}", ws.handleGetState)
	mux.HandleFunc("GET /repository/action/{componentId}/{actionName}", ws.handleExecuteAction)
}

func (ws *WebServer) handleGetComponents(w http.ResponseWriter, r *http.Request) {
	componentIds := make([]string, 0)

	for _, comp := range ws.registry.GetComponents() {
		if comp.Plugin().Usage() == metadata.Ui {
			componentIds = append(componentIds, comp.Id())
		}
	}

	ws.sendJson(w, componentIds)
}

func (ws *WebServer) handleGetState(w http.ResponseWriter, r *http.Request) {
	componentId := r.PathValue("componentId")
	if componentId == "" {
		http.Error(w, "Component ID is required", http.StatusBadRequest)
		return
	}

	comp := ws.registry.GetComponent(componentId)
	if comp == nil {
		http.Error(w, "Component not found", http.StatusNotFound)
		return
	}

	plugin := comp.Plugin()
	if plugin.Usage() != metadata.Ui {
		http.Error(w, "Component is not a UI component", http.StatusBadRequest)
		return
	}

	// Collect state
	state := make(map[string]any)
	for _, name := range plugin.MemberNames() {
		member := plugin.Member(name)
		if member.MemberType() == metadata.State {
			state[name] = comp.StateItem(name).Get()
		}
	}

	ws.sendJson(w, state)
}

func (ws *WebServer) handleExecuteAction(w http.ResponseWriter, r *http.Request) {
	componentId := r.PathValue("componentId")
	actionName := r.PathValue("actionName")
	if componentId == "" || actionName == "" {
		http.Error(w, "Component ID and Action Name are required", http.StatusBadRequest)
		return
	}

	comp := ws.registry.GetComponent(componentId)
	if comp == nil {
		http.Error(w, "Component not found", http.StatusNotFound)
		return
	}

	plugin := comp.Plugin()
	if plugin.Usage() != metadata.Ui {
		http.Error(w, "Component is not a UI component", http.StatusBadRequest)
		return
	}

	actionMember := plugin.Member(actionName)
	if actionMember == nil || actionMember.MemberType() != metadata.Action {
		http.Error(w, "Action not found", http.StatusNotFound)
		return
	}

	if !metadata.MakeTypeBool().Equals(actionMember.ValueType()) {
		http.Error(w, "Action type must be boolean", http.StatusBadRequest)
		return
	}

	action := comp.Action(actionName)
	action <- true
	time.Sleep(100 * time.Millisecond) // FIXME: Give some time for the action to be processed
	action <- false

	w.WriteHeader(http.StatusOK)
}

func (ws *WebServer) sendJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}
