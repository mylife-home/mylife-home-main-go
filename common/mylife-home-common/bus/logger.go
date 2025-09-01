package bus

import (
	"encoding/json"
	"fmt"
	"mylife-home-common/log/publish"
	"mylife-home-common/tools"
	"os"
	"time"
)

const loggerDomain = "logger"

const retention = 1024 * 1024

type Logger struct {
	client   *client
	listener chan *publish.LogEntry
	queue    chan *publish.LogEntry
	exit     chan struct{}
}

func newLogger(client *client) *Logger {
	logger := &Logger{
		client:   client,
		listener: make(chan *publish.LogEntry),
		queue:    make(chan *publish.LogEntry, retention),
		exit:     make(chan struct{}),
	}

	go logger.pump()
	publish.OnEntry().Subscribe(logger.listener)
	go logger.writer()

	return logger
}

func (logger *Logger) terminate() {
	publish.OnEntry().Unsubscribe(logger.listener)
	close(logger.listener)
	close(logger.exit)
}

func (logger *Logger) pump() {
	for entry := range logger.listener {

		// force enqueue. if full, drop message so that enqueue is OK
		enqueued := false

		for !enqueued {
			select {
			case logger.queue <- entry:
				enqueued = true

			default:
				// Note: should log it, but it would cause more entries
				fmt.Println("Logger queue full, dropping message")
				<-logger.queue
			}
		}
	}
}

func (logger *Logger) writer() {
	for {
		select {

		case <-logger.exit:
			return

		case entry := <-logger.queue:
			logger.send(entry)
		}
	}
}

func (logger *Logger) send(entry *publish.LogEntry) {
	payload, err := logger.serialize(entry)
	if err != nil {
		// Note: should log it, but it would cause more entries
		fmt.Printf("Error marshaling log: '%f'\n", err)
		return
	}

	topic := logger.client.BuildTopic(loggerDomain)

	for {
		select {

		case <-logger.exit:
			return

		default:
			if !logger.waitOnline() {
				return
			}

			switch err := logger.client.Publish(topic, payload); {
			case err == errClosing:
				// retry

			case err == nil:
				return

			default:
				fmt.Printf("Error sending log: '%f'\n", err)
				return
			}
		}
	}
}

func (logger *Logger) waitOnline() bool {
	onlineChannel := make(chan bool)
	resultChannel := make(chan bool, 1)

	// We need to to this in a separate goroutine and wait for online chan close
	// else we may block it
	go func() {
		hasResult := false
		exitChannel := logger.exit

		for {
			select {
			case <-exitChannel:
				if !hasResult {
					hasResult = true
					resultChannel <- false
				}
				exitChannel = nil // stop to poll exit, but wait for the onlineChannel to close

			case online, ok := <-onlineChannel:
				// continue to select until
				if !ok {
					return
				}

				if online && !hasResult {
					hasResult = true
					resultChannel <- true
				}
			}
		}
	}()

	logger.client.Online().Subscribe(onlineChannel, true)
	defer func() {
		logger.client.Online().Unsubscribe(onlineChannel)
	}()

	return <-resultChannel
}

func (logger *Logger) serialize(entry *publish.LogEntry) ([]byte, error) {
	data := jsonLog{
		Name:         entry.LoggerName(),
		InstanceName: logger.client.InstanceName(),
		Hostname:     tools.Hostname(),
		Pid:          os.Getpid(),
		Level:        convertLevel(publish.LogLevel(entry.Level())),
		Msg:          entry.Message(),
		Time:         entry.Timestamp().Format(time.RFC3339),
		V:            0,
	}

	if err := entry.Error(); err != nil {
		data.Err = &jsonError{
			Message: err.Message(),
			Name:    "Error", // Go has no error name/type
			Stack:   err.StackTrace(),
		}
	}

	raw, err := json.Marshal(&data)
	return raw, err
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
