package log

import (
	stdlog "log"
	"mylife-home-common/log/console"
	"mylife-home-common/log/publish"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/memory"
	"github.com/apex/log/handlers/multi"
)

// Before configuration create an in-memory logger
var rootLogger = &log.Logger{
	Handler: memory.New(),
	Level:   log.DebugLevel,
}

func Init(consoleOutput bool) {
	handlers := []log.Handler{
		publish.New(),
	}

	if consoleOutput {
		handlers = append(handlers, console.New(os.Stdout))
	}

	handler := multi.New(handlers...)

	// Dump memory logs
	// Note: this is not thread-safe, but we are at init-time
	entries := rootLogger.Handler.(*memory.Handler).Entries
	for _, entry := range entries {
		if err := handler.HandleLog(entry); err != nil {
			stdlog.Printf("error logging: %s", err)
		}
	}

	rootLogger.Handler = handler
}

func CreateLogger(name string) Logger {
	return newLoggerImpl(rootLogger.WithField("logger-name", name))
}
