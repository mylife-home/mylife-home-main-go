package sessions

import (
	"fmt"
	"mylife-home-common/components"
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
	"sync"
)

type stateListener struct {
	registry            components.Registry
	componentChangeChan chan *components.ComponentChange

	state    map[string]any
	stateMux sync.Mutex

	subscriptions map[string]subscription
}

type subscription struct {
	value tools.ObservableValue[any]
	ch    chan any
}

func newStateListener(registry components.Registry) *stateListener {
	l := &stateListener{
		registry:            registry,
		componentChangeChan: make(chan *components.ComponentChange),
		state:               make(map[string]any),
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

	stateNames := l.getStateNames(comp)

	switch change.Action() {
	case components.RegistryAdd:
		for _, name := range stateNames {
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

	case components.RegistryRemove:
		for _, name := range stateNames {
			key := comp.Id() + ":" + name
			sub := l.subscriptions[key]
			delete(l.subscriptions, key)

			sub.value.Unsubscribe(sub.ch)
			close(sub.ch)
		}
	}
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
	l.stateMux.Lock()
	defer l.stateMux.Unlock()

	key := change.componentId + ":" + change.stateName
	l.state[key] = change.value
	fmt.Println("State updated:", key, "=", change.value)
}
