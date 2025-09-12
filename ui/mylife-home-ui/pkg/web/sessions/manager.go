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
	registry                   components.Registry
	model                      *model.ModelManager
	stateListener              *stateListener
	modelUpdateChan            chan struct{}
	requiredComponentStates    map[string]map[string]struct{}
	requiredComponentStatesMux sync.Mutex

	sessions      map[*session]struct{}
	sessionsMux   sync.Mutex
	sessionsIdGen int
}

func NewManager(registry components.Registry, model *model.ModelManager) *Manager {
	m := &Manager{
		registry:                registry,
		model:                   model,
		modelUpdateChan:         make(chan struct{}),
		requiredComponentStates: make(map[string]map[string]struct{}),
		sessions:                make(map[*session]struct{}),
	}

	m.setRequiredComponentStates()
	m.stateListener = newStateListener(registry, m.handleStateChange, m.handleComponentAdd, m.handleComponentRemove)

	go m.modelUpdateWorker()
	m.model.OnUpdate().Subscribe(m.modelUpdateChan)

	return m
}

func (m *Manager) Terminate() {
	logger.Info("Stopping session manager")

	m.model.OnUpdate().Unsubscribe(m.modelUpdateChan)
	close(m.modelUpdateChan)

	m.stateListener.Terminate()

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
	m.sendInitialComponentStates(session)
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
		m.setRequiredComponentStates()
	}
}

func (m *Manager) handleStateChange(componentId string, stateName string, value any) {
	if m.isRequiredComponentState(componentId, stateName) {
		m.sendAll(api.MessageState, &api.StateChange{
			Id:    componentId,
			Name:  stateName,
			Value: value,
		})
	}
}

func (m *Manager) handleComponentAdd(compId string, attributes map[string]any) {
	states := m.getRequiredComponentStates(compId)
	filteredAttributes := make(map[string]any)

	for _, name := range states {
		if v, ok := attributes[name]; ok {
			filteredAttributes[name] = v
		}
	}

	m.sendAll(api.MessageAdd, &api.ComponentAdd{
		Id:         compId,
		Attributes: filteredAttributes,
	})
}

func (m *Manager) handleComponentRemove(compId string) {
	if m.isRequiredComponent(compId) {
		m.sendAll(api.MessageRemove, &api.ComponentRemove{
			Id: compId,
		})
	}
}

func (m *Manager) setRequiredComponentStates() {
	m.requiredComponentStatesMux.Lock()
	defer m.requiredComponentStatesMux.Unlock()

	m.requiredComponentStates = make(map[string]map[string]struct{})

	for _, rcs := range m.model.GetRequiredComponentStates() {
		states, ok := m.requiredComponentStates[rcs.ComponentId]
		if !ok {
			states = make(map[string]struct{})
			m.requiredComponentStates[rcs.ComponentId] = states
		}

		states[rcs.ComponentState] = struct{}{}
	}
}

func (m *Manager) isRequiredComponent(compId string) bool {
	m.requiredComponentStatesMux.Lock()
	defer m.requiredComponentStatesMux.Unlock()

	_, ok := m.requiredComponentStates[compId]
	return ok
}

func (m *Manager) getRequiredComponentStates(compId string) []string {
	m.requiredComponentStatesMux.Lock()
	defer m.requiredComponentStatesMux.Unlock()

	if states, ok := m.requiredComponentStates[compId]; ok {
		names := slices.Collect(maps.Keys(states))
		return names
	}

	return nil
}

func (m *Manager) isRequiredComponentState(componentId string, stateName string) bool {
	m.requiredComponentStatesMux.Lock()
	defer m.requiredComponentStatesMux.Unlock()

	if states, ok := m.requiredComponentStates[componentId]; ok {
		_, ok := states[stateName]
		return ok
	}

	return false
}

func (m *Manager) sendInitialComponentStates(s *session) {
	m.stateListener.stateMux.Lock()
	defer m.stateListener.stateMux.Unlock()

	for compId, states := range m.requiredComponentStates {
		compState := m.stateListener.GetState(compId)
		if compState == nil {
			continue
		}

		msg := &api.ComponentAdd{
			Id:         compId,
			Attributes: make(map[string]any),
		}

		for stateName := range states {
			msg.Attributes[stateName] = compState[stateName]
		}

		m.sendOne(s, api.MessageAdd, msg)
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
