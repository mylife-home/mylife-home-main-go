package engine

import (
	"mylife-home-common/log"
	"sync"
	"time"
)

var logger = log.CreateLogger("mylife:home:core:plugins:logic-selectors:engine")

type InputManager struct {
	config     map[string]func()
	eventStack string
	endWait    *time.Timer
	lastDown   time.Time
	mux        sync.Mutex
}

func NewInputManager() *InputManager {
	return &InputManager{
		config: make(map[string]func()),
	}
}

func (manager *InputManager) Terminate() {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	manager.cancelTimeout()
}

func (manager *InputManager) AddConfig(trigger string, callback func()) {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	manager.config[trigger] = callback
}

func (manager *InputManager) Down() {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	// no input end for now
	manager.cancelTimeout()

	manager.lastDown = time.Now()
}

func (manager *InputManager) Up() {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	// no input end for now
	manager.cancelTimeout()

	// if no down, tchao
	if manager.lastDown.IsZero() {
		manager.eventStack = ""
		return
	}

	// Prise en compte de l'event
	downTs := manager.lastDown
	upTs := time.Now()
	manager.lastDown = time.Time{}

	// Ajout de l'event
	if upTs.Sub(downTs) < 500*time.Millisecond {
		manager.eventStack += "s"
	} else {
		manager.eventStack += "l"
	}

	// Attente de la fin de saisie
	manager.endWait = time.AfterFunc(300*time.Millisecond, manager.onTimeout)

}

func (manager *InputManager) cancelTimeout() {
	if manager.endWait != nil {
		manager.endWait.Stop()
		manager.endWait = nil
	}
}

func (manager *InputManager) onTimeout() {
	manager.mux.Lock()
	defer manager.mux.Unlock()

	manager.executeEvents()

	manager.eventStack = ""
	manager.endWait = nil
}

func (manager *InputManager) executeEvents() {
	logger.Debugf("Execute events : '%s'", manager.eventStack)

	if callback, ok := manager.config[manager.eventStack]; ok {
		callback()
	}
}
