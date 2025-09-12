package sessions

import (
	"context"
	"errors"
	"fmt"
	"mylife-home-ui/pkg/web/api"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const idleTimeout = 30 * time.Second

type session struct {
	id        string
	conn      *websocket.Conn
	sendChan  chan *api.SocketMessage
	ctx       context.Context // Cancel when session ends
	cancel    context.CancelFunc
	onClose   func()
	onMessage func(msg *api.SocketMessage)
	timeout   *time.Timer
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
		var msg api.SocketMessage
		fmt.Println("Reading message...")
		err := wsjson.Read(s.ctx, s.conn, &msg)
		fmt.Println("Read complete.")
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			s.error(err)
			continue
		}

		logger.Debugf("Received message on session %s: %s", s.id, msg.Type)

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

		logger.Debugf("Sent message on session %s: %s", s.id, msg.Type)
	}
}

func (s *session) timeoutWorker() {
	select {
	case <-s.timeout.C:
		s.error(fmt.Errorf("session %s timed out after %s of inactivity", s.id, idleTimeout))
	case <-s.ctx.Done():
		// Session is being terminated, stop the timeout worker
		s.timeout.Stop()
		return
	}
}

func (s *session) error(err error) {
	logger.Errorf("websocket error on session %s: %v", s.id, err)
	// s.conn.Close(websocket.StatusInternalError, "internal error")

	s.Terminate()
}

func (s *session) Terminate() {
	if s.ctx.Err() != nil {
		// Already terminated
		return
	}

	logger.Debugf("Closing session: %s", s.id)

	close(s.sendChan)

	s.cancel()
	s.timeout.Stop()
	// s.conn.Close(websocket.StatusNormalClosure, "")
	s.conn.CloseNow()
	s.onClose()
}

func (s *session) Send(msg *api.SocketMessage) {
	s.sendChan <- msg
}
