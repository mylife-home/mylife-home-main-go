package sessions

import (
	"encoding/json"
	"fmt"
	"maps"
	"mylife-home-ui/pkg/model"
	"mylife-home-ui/pkg/web/api"
	"net/http"
	"slices"
	"sync"
	"time"

	"mylife-home-common/components"
	"mylife-home-common/components/metadata"
	"mylife-home-common/log"

	"github.com/coder/websocket"
)

var logger = log.CreateLogger("mylife:home:ui:web:sessions")

type Manager struct {
	registry        components.Registry
	model           *model.ModelManager
	modelUpdateChan chan struct{}

	sessions      map[*session]struct{}
	sessionsMux   sync.Mutex
	sessionsIdGen int
}

func NewManager(registry components.Registry, model *model.ModelManager) *Manager {
	m := &Manager{
		registry:        registry,
		model:           model,
		modelUpdateChan: make(chan struct{}),
		sessions:        make(map[*session]struct{}),
	}

	go m.modelUpdateWorker()

	m.model.OnUpdate().Subscribe(m.modelUpdateChan)

	return m
}

func (m *Manager) Terminate() {
	logger.Info("Stopping session manager")

	m.model.OnUpdate().Unsubscribe(m.modelUpdateChan)
	close(m.modelUpdateChan)

	// Obtain list then unlock to avoid deadlock in Terminate()
	m.sessionsMux.Lock()
	list := slices.Collect(maps.Keys(m.sessions))
	m.sessionsMux.Unlock()

	for _, session := range list {
		session.Terminate()
	}
}

func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		logger.Errorf("failed to accept websocket connection: %v", err)
		return
	}

	m.sessionsMux.Lock()
	defer m.sessionsMux.Unlock()

	id := fmt.Sprintf("%d", m.sessionsIdGen)
	m.sessionsIdGen++

	session := newSession(id, conn)

	onClose := func() {
		m.removeSession(session)
	}

	onMessage := func(msg *api.SocketMessage) {
		m.processMessage(session, msg)
	}

	session.setCallbacks(onClose, onMessage)

	m.sessions[session] = struct{}{}

	session.start()
	m.sendOne(session, api.MessageModelHash, m.model.GetModelHash())
}

func (m *Manager) removeSession(s *session) {
	m.sessionsMux.Lock()
	defer m.sessionsMux.Unlock()

	delete(m.sessions, s)
}

func (m *Manager) modelUpdateWorker() {
	for range m.modelUpdateChan {
		modelHash := m.model.GetModelHash()
		m.sendAll(api.MessageModelHash, modelHash)
	}
}

func (m *Manager) sendAll(ty api.MessageType, object any) {
	data, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}

	msg := &api.SocketMessage{
		Type: ty,
		Data: data,
	}

	for session := range m.sessions {
		session.Send(msg)
	}
}

func (m *Manager) sendOne(s *session, ty api.MessageType, object any) {
	data, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}

	msg := &api.SocketMessage{
		Type: ty,
		Data: data,
	}

	s.Send(msg)
}

func (m *Manager) processMessage(s *session, msg *api.SocketMessage) {
	switch msg.Type {
	case api.MessagePing:
		m.sendOne(s, api.MessagePong, nil)

	case api.MessageAction:
		var actionMsg api.ActionMessage
		err := json.Unmarshal(msg.Data, &actionMsg)
		if err != nil {
			logger.Errorf("Invalid action message: '%s': %v", string(msg.Data), err)
			return
		}

		m.executeAction(actionMsg.Id, actionMsg.Action)

	default:
		logger.Errorf("Unknown message type: %s", msg.Type)
	}
}

func (m *Manager) executeAction(componentId string, actionName string) {
	comp := m.registry.GetComponent(componentId)
	if comp == nil {
		logger.Warnf("Component not found: %s", componentId)
		return
	}

	plugin := comp.Plugin()
	if plugin.Usage() != metadata.Ui {
		logger.Warnf("Component is not a UI component: %s", componentId)
		return
	}

	actionMember := plugin.Member(actionName)
	if actionMember == nil || actionMember.MemberType() != metadata.Action {
		logger.Warnf("Action not found: %s on component %s", actionName, componentId)
		return
	}

	if !metadata.MakeTypeBool().Equals(actionMember.ValueType()) {
		logger.Warnf("Action type must be boolean: %s on component %s", actionName, componentId)
		return
	}

	action := comp.Action(actionName)

	action <- true
	// FIXME: Give some time for the action to be processed
	// For now actions are emitted async, which may break the order
	time.Sleep(100 * time.Millisecond)
	action <- false
}
