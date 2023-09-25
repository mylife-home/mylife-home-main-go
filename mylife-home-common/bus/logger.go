package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"mylife-home-common/log/publish"
	"mylife-home-common/tools"
	"os"
	"sync"
	"time"
)

const loggerDomain = "logger"

const offlineRetention = 1024 * 1024

type Logger struct {
	client    *client
	queue     chan *publish.LogEntry
	onOffline func() // to set when offline, to cancel logsender routine
}

func newLogger(client *client) *Logger {
	logger := &Logger{
		client: client,
		queue:  make(chan *publish.LogEntry, offlineRetention),
	}

	logger.client.OnOnlineChanged().Register(logger.onOnlineChange)
	publish.OnEntry().Register(logger.onEntry)

	return logger
}

func (logger *Logger) onOnlineChange(online bool) {
	if online {
		ctx, cancel := context.WithCancel(context.Background())
		logger.onOffline = cancel
		go logger.logSender(ctx)
	} else {
		// Stop routine when we go offline
		logger.onOffline()
	}
}

func (logger *Logger) logSender(ctx context.Context) {
	for {
		select {
		case e := <-logger.queue:
			logger.send(e)
		case <-ctx.Done():
			return
		}
	}
}

func (logger *Logger) onEntry(e *publish.LogEntry) {
	logger.queue <- e
}

type jsonLog struct {
	Name         string     `json:"name"`
	InstanceName string     `json:"instanceName"`
	Hostname     string     `json:"hostname"`
	Pid          int        `json:"pid"`
	Level        int        `json:"level"`
	Msg          string     `json:"msg"`
	Err          *jsonError `json:"err"`
	Time         string     `json:"time"`
	V            int        `json:"v"` // 0
}

type jsonError struct {
	Message string `json:"message"`
	Name    string `json:"name"`
	Stack   string `json:"stack"`
}

func (logger *Logger) send(e *publish.LogEntry) {
	data := jsonLog{
		Name:         e.LoggerName(),
		InstanceName: logger.client.InstanceName(),
		Hostname:     tools.Hostname(),
		Pid:          os.Getpid(),
		Level:        convertLevel(publish.LogLevel(e.Level())),
		Msg:          e.Message(),
		Time:         e.Timestamp().Format(time.RFC3339),
		V:            0,
	}

	if ee := e.Error(); ee != nil {
		data.Err = &jsonError{
			Message: ee.Message(),
			Name:    "Error", // Go has no error name/type
			Stack:   ee.StackTrace(),
		}
	}

	raw, err := json.Marshal(&data)
	if err != nil {
		// Note: should log it, but it would cause more entries
		fmt.Printf("Error marshaling log: '%f'\n", err)
	}

	err = logger.client.Publish(logger.client.BuildTopic(loggerDomain), raw, false)
	if err != nil {
		// Note: should log it, but it would cause more entries
		fmt.Printf("Error sending log: '%f'\n", err)
	}
}

func convertLevel(level publish.LogLevel) int {
	/* from bunyan doc:
	"fatal" (60): The service/app is going to stop or become unusable now. An operator should definitely look into this soon.
	"error" (50): Fatal for a particular request, but the service/app continues servicing other requests. An operator should look at this soon(ish).
	"warn" (40): A note on something that should probably be looked at by an operator eventually.
	"info" (30): Detail on regular operation.
	"debug" (20): Anything else, i.e. too verbose to be included in "info" level.
	"trace" (10): Logging from external libraries used by your app or very detailed application logging.
	*/

	switch level {
	case publish.Debug:
		return 20
	case publish.Info:
		return 30
	case publish.Warn:
		return 40
	case publish.Error:
		return 50
	}

	panic(fmt.Errorf("unknown level %s", level))
}

type offlineQueue struct {
	buffer     []*publish.LogEntry
	bufferSync sync.Mutex
}

func newOfflineQueue() *offlineQueue {
	return &offlineQueue{
		buffer: make([]*publish.LogEntry, 0),
	}
}

func (queue *offlineQueue) Enqueue(e *publish.LogEntry) {
	fmt.Printf("Enqueue\n")
	queue.bufferSync.Lock()
	defer queue.bufferSync.Unlock()

	if len(queue.buffer) >= offlineRetention {
		// Note: should log it, but it would cause more entries to arrive
		fmt.Println("Log entry dropped because offline queue is full")
		return
	}

	queue.buffer = append(queue.buffer, e)
}

func (queue *offlineQueue) Dequeue() *publish.LogEntry {
	fmt.Printf("Dequeue\n")
	queue.bufferSync.Lock()
	defer queue.bufferSync.Unlock()

	if len(queue.buffer) == 0 {
		return nil
	}

	e := queue.buffer[0]
	queue.buffer = queue.buffer[1:]
	return e
}
