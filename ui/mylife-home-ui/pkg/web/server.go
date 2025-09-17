package web

import (
	"context"
	"fmt"
	"mylife-home-ui/pkg/model"
	"mylife-home-ui/pkg/web/sessions"
	"net/http"
	"time"

	"mylife-home-common/components"
	"mylife-home-common/config"
	"mylife-home-common/log"
)

var logger = log.CreateLogger("mylife:home:ui:web")

type webConfig struct {
	Port int `mapstructure:"port"`
}

type WebServer struct {
	server         *http.Server
	registry       components.Registry
	model          *model.ModelManager
	sessionManager *sessions.Manager
	config         webConfig
}

func NewWebServer(registry components.Registry, model *model.ModelManager) *WebServer {
	ws := &WebServer{
		registry:       registry,
		model:          model,
		sessionManager: sessions.NewManager(registry, model),
	}

	config.BindStructure("web", &ws.config)

	mux := http.NewServeMux()

	ws.setupRepository(mux)
	ws.setupResources(mux)
	ws.setupSessions(mux)
	ws.setupStatic(mux)

	ws.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", ws.config.Port),
		Handler: mux,
	}

	logger.Infof("Starting web server on port %d", ws.config.Port)

	go func() {
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("Failed to start web server: %v", err))
		}
	}()

	return ws
}

func (ws *WebServer) Terminate() {
	logger.Info("Stopping web server")

	ws.sessionManager.Terminate()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ws.server.Shutdown(ctx); err != nil {
		logger.Errorf("Error during web server shutdown: %v", err)
	}
}

func (ws *WebServer) setupSessions(mux *http.ServeMux) {
	mux.HandleFunc("/websocket", ws.sessionManager.HandleWebSocket)
}
