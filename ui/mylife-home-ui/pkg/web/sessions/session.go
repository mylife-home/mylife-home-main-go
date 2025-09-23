package sessions

import (
	"context"
	"fmt"
	"mylife-home-ui/pkg/web/api"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const idleTimeout = 30 * time.Second
const pingTimeout = 5 * time.Second

type session struct {
	id        string
	conn      *websocket.Conn
	sendChan  chan *api.SocketMessage
	ctx       context.Context // Cancel when session ends
	cancel    context.CancelFunc
	onClose   func()
	onMessage func(msg *api.SocketMessage)
	timeout   *time.Timer
	termMux   sync.Mutex // Mutex to protect termination
}

func newSession(id string, conn *websocket.Conn) *session {
	ctx, cancel := context.WithCancel(context.Background())
	s := &session{
		id:       id,
		conn:     conn,
		sendChan: make(chan *api.SocketMessage, 2048),
		ctx:      ctx,
		cancel:   cancel,
		timeout:  time.NewTimer(idleTimeout),
	}

	logger.Debugf("New session created: %s", s.id)

	return s
}

func (s *session) setCallbacks(onClose func(), onMessage func(msg *api.SocketMessage)) {
	s.onClose = onClose
	s.onMessage = onMessage
}

func (s *session) start() {
	go s.readWorker()
	go s.writeWorker()
	go s.timeoutWorker()
}

func (s *session) readWorker() {
	for {
		if s.ctx.Err() != nil {
			return
		}

		var msg api.SocketMessage
		err := wsjson.Read(s.ctx, s.conn, &msg)
		if err != nil {
			// Connection closing
			if s.ctx.Err() != nil {
				return
			}

			// Note: cannot assert on real errors since the library does not export them
			if strings.Contains(err.Error(), "received close frame") {
				logger.Debugf("Session %s closed by client", s.id)
				s.Terminate()
				return
			}

			s.error(err)
			continue
		}

		//logger.Debugf("Received message on session %s: %s", s.id, msg.Type)

		// Reset the timeout timer on each message received
		s.timeout.Reset(idleTimeout)

		s.onMessage(&msg)
	}
}

func (s *session) writeWorker() {
	for msg := range s.sendChan {
		err := wsjson.Write(context.Background(), s.conn, msg)
		if err != nil {
			s.error(err)
			continue
		}

		// logger.Debugf("Sent message on session %s: %s", s.id, msg.Type)
	}
}

func (s *session) timeoutWorker() {
	for {
		select {
		case <-s.timeout.C:
			s.sendWebSocketPing()
		case <-s.ctx.Done():
			// Session is being terminated, stop the timeout worker
			s.timeout.Stop()
			return
		}
	}
}

func (s *session) sendWebSocketPing() {
	// logger.Debugf("Sending WebSocket ping to session %s after %s of inactivity", s.id, idleTimeout)

	// Create a context with timeout for the ping operation
	pingCtx, cancel := context.WithTimeout(s.ctx, pingTimeout)
	defer cancel()

	// Use the built-in WebSocket ping which sends a ping and waits for pong
	err := s.conn.Ping(pingCtx)
	if err != nil {
		// Ping failed or timed out
		if pingCtx.Err() == context.DeadlineExceeded {
			s.error(fmt.Errorf("session %s did not respond to WebSocket ping within %s", s.id, pingTimeout))
		} else {
			s.error(fmt.Errorf("WebSocket ping failed for session %s: %v", s.id, err))
		}
		return
	}

	// logger.Debugf("Received WebSocket pong from session %s", s.id)
	s.timeout.Reset(idleTimeout)
}

func (s *session) error(err error) {
	logger.Errorf("websocket error on session %s: %v", s.id, err)
	// s.conn.Close(websocket.StatusInternalError, "internal error")

	s.Terminate()
}

// termCheck checks if the session is already terminated, if not it marks it as terminated.
func (s *session) termCheck() bool {
	s.termMux.Lock()
	defer s.termMux.Unlock()

	if s.ctx.Err() != nil {
		// Already terminated
		return false
	}

	s.cancel()
	return true
}

func (s *session) Terminate() {
	if !s.termCheck() {
		return
	}

	logger.Debugf("Closing session: %s", s.id)

	close(s.sendChan)

	s.timeout.Stop()
	// s.conn.Close(websocket.StatusNormalClosure, "")
	s.conn.CloseNow()
	s.onClose()
}

func (s *session) Send(msg *api.SocketMessage) {
	s.sendChan <- msg
}
