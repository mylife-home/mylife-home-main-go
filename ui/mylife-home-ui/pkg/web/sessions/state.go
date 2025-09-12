package sessions

import (
	"mylife-home-common/components"
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
	"sync"
)

type stateListener struct {
	registry            components.Registry
	componentChangeChan chan *components.ComponentChange

	onChange func(componentId string, stateName string, value any)
	state    map[string]map[string]any
	stateMux sync.Mutex

	subscriptions map[string]subscription
}

type subscription struct {
	value tools.ObservableValue[any]
	ch    chan any
}

func newStateListener(registry components.Registry, onChange func(componentId string, stateName string, value any)) *stateListener {
	l := &stateListener{
		registry:            registry,
		componentChangeChan: make(chan *components.ComponentChange),
		onChange:            onChange,
		state:               make(map[string]map[string]any),
		subscriptions:       make(map[string]subscription),
	}

	go l.updateWorker()

	l.registry.OnComponentChange().Subscribe(l.componentChangeChan)

	return l
}

func (l *stateListener) Terminate() {
	l.registry.OnComponentChange().Unsubscribe(l.componentChangeChan)
	close(l.componentChangeChan)
}

// nil if component not present
func (l *stateListener) GetState(componentId string) map[string]any {
	l.stateMux.Lock()
	defer l.stateMux.Unlock()

	stateCopy := make(map[string]any)
	if compState, ok := l.state[componentId]; ok {
		for k, v := range compState {
			stateCopy[k] = v
		}
		return stateCopy
	}

	return nil
}

type stateChange struct {
	componentId string
	stateName   string
	value       any
}

func (l *stateListener) updateWorker() {
	dummy := make(chan stateChange)
	merger := tools.MakeChannelMerger(dummy)

	defer func() {
		// close all subscribed channels
		for _, sub := range l.subscriptions {
			sub.value.Unsubscribe(sub.ch)
			close(sub.ch)
		}

		close(dummy)
	}()

	for {
		select {
		case change, ok := <-l.componentChangeChan:
			if !ok {
				return
			}

			l.handleComponentChange(change, merger)

		case stateChange := <-merger.Out():
			l.handleStateChange(stateChange)
		}
	}
}

func (l *stateListener) handleComponentChange(change *components.ComponentChange, merger *tools.ChannelMerger[stateChange]) {
	comp := change.Component()
	if comp.Plugin().Usage() != metadata.Ui {
		return
	}

	switch change.Action() {
	case components.RegistryAdd:
		l.handleComponentAdd(comp, merger)

	case components.RegistryRemove:
		l.handleComponentRemove(comp)
	}
}

func (l *stateListener) handleComponentAdd(comp components.Component, merger *tools.ChannelMerger[stateChange]) {
	for _, name := range l.getStateNames(comp) {
		key := comp.Id() + ":" + name
		value := comp.StateItem(name)

		sub := subscription{
			value: value,
			ch:    make(chan any, 10), // else subscribe current value below will deadlock
		}

		sub.value.Subscribe(sub.ch, true)
		l.subscriptions[key] = sub

		merger.Add(tools.MapChannel(sub.ch, func(v any) stateChange {
			return stateChange{
				componentId: comp.Id(),
				stateName:   name,
				value:       v,
			}
		}))
	}

	l.stateMux.Lock()
	l.state[comp.Id()] = make(map[string]any)
	l.stateMux.Unlock()
}

func (l *stateListener) handleComponentRemove(comp components.Component) {
	for _, name := range l.getStateNames(comp) {
		key := comp.Id() + ":" + name
		sub := l.subscriptions[key]
		delete(l.subscriptions, key)

		sub.value.Unsubscribe(sub.ch)
		close(sub.ch)
	}

	l.stateMux.Lock()
	l.state[comp.Id()] = make(map[string]any)
	l.stateMux.Unlock()
}

func (l *stateListener) getStateNames(comp components.Component) []string {
	plugin := comp.Plugin()
	stateNames := make([]string, 0)

	for _, name := range plugin.MemberNames() {
		memberMeta := plugin.Member(name)
		if memberMeta.MemberType() != metadata.State {
			continue
		}
		stateNames = append(stateNames, name)
	}

	return stateNames
}

func (l *stateListener) handleStateChange(change stateChange) {
	l.updateState(change)
	l.onChange(change.componentId, change.stateName, change.value)
}

func (l *stateListener) updateState(change stateChange) {
	l.stateMux.Lock()
	defer l.stateMux.Unlock()

	compState := l.state[change.componentId]
	if compState == nil {
		// Late notification for a component that has been removed
		return
	}

	compState[change.stateName] = change.value
}
