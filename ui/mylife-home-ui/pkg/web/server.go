package web

import (
	"context"
	"fmt"
	"mylife-home-ui/pkg/model"
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
	server   *http.Server
	registry components.Registry
	model    *model.ModelManager
	config   webConfig
}

func NewWebServer(registry components.Registry, model *model.ModelManager) *WebServer {
	conf := webConfig{}
	config.BindStructure("web", &conf)

	ws := &WebServer{
		registry: registry,
		model:    model,
		config:   conf,
	}

	if err := ws.start(); err != nil {
		panic(fmt.Sprintf("Failed to start web server: %v", err))
	}

	return ws
}

// Start starts the web server
func (ws *WebServer) start() error {
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
			logger.Errorf("Web server failed to run: %v", err)
		}
	}()

	return nil
}

// Stop gracefully stops the web server
func (ws *WebServer) Terminate() {
	logger.Info("Stopping web server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ws.server.Shutdown(ctx); err != nil {
		logger.Errorf("Error during web server shutdown: %v", err)
	}
}
