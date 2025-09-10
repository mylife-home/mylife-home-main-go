package web

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func (ws *WebServer) setupSessions(mux *http.ServeMux) {
	mux.HandleFunc("/ws", ws.handleWebSocket)
}

func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		// ...
	}
	defer c.CloseNow()

	// Set the context as needed. Use of r.Context() is not recommended
	// to avoid surprising behavior (see http.Hijacker).
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var v any
	err = wsjson.Read(ctx, c, &v)
	if err != nil {
		// ...
	}

	log.Printf("received: %v", v)

	c.Close(websocket.StatusNormalClosure, "")
}

type sessionsManager struct {
}

type session struct {
}
