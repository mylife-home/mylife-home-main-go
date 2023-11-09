package manager

import (
	"fmt"
	"mylife-home-common/components"
	"mylife-home-common/components/metadata"
	"mylife-home-common/tools"
	"mylife-home-core/pkg/store"
	"strings"
)

type binding struct {
	registry            components.Registry
	config              *store.BindingConfig
	componentChangeChan chan *components.ComponentChange
	workerExited        chan struct{}

	// updated by worker
	sourceInstance string
	source         components.Component
	targetInstance string
	target         components.Component
	sourceState    tools.ObservableValue[any]
	targetAction   chan<- any
}

func makeBinding(registry components.Registry, config *store.BindingConfig) *binding {
	b := &binding{
		registry:            registry,
		config:              config,
		componentChangeChan: make(chan *components.ComponentChange),
		workerExited:        make(chan struct{}),
	}

	go b.worker()

	b.registry.OnComponentChange().Subscribe(b.componentChangeChan)

	logger.Infof("Binding '%s' created", b.config)

	return b
}

func (b *binding) Terminate() {
	b.registry.OnComponentChange().Unsubscribe(b.componentChangeChan)
	close(b.componentChangeChan)
	<-b.workerExited

	logger.Infof("Binding '%s' closed", b.config)
}

func (b *binding) worker() {
	b.onInit()
	defer b.onClose()

	for {
		select {
		case change, ok := <-b.componentChangeChan:
			if !ok {
				return
			}

			b.onComponentChange(change)
		}
	}
}

func (b *binding) onInit() {
	sourceData := b.registry.GetComponentData(b.config.SourceComponent)
	targetData := b.registry.GetComponentData(b.config.TargetComponent)

	if sourceData != nil {
		b.sourceInstance = sourceData.InstanceName()
		b.source = sourceData.Component()
	}

	if targetData != nil {
		b.targetInstance = targetData.InstanceName()
		b.target = targetData.Component()
	}

	b.refreshBinding()
}

func (b *binding) onClose() {
	b.sourceInstance = ""
	b.source = nil
	b.targetInstance = ""
	b.target = nil

	b.refreshBinding()
}

func (b *binding) onComponentChange(change *components.ComponentChange) {
	comp := change.Component()

	switch comp.Id() {
	case b.config.SourceComponent:
		switch change.Action() {
		case components.RegistryAdd:
			b.sourceInstance = change.InstanceName()
			b.source = comp
		case components.RegistryRemove:
			b.sourceInstance = ""
			b.source = nil
		}

		b.refreshBinding()

	case b.config.TargetComponent:
		switch change.Action() {
		case components.RegistryAdd:
			b.targetInstance = change.InstanceName()
			b.target = comp
		case components.RegistryRemove:
			b.targetInstance = ""
			b.target = nil
		}

		b.refreshBinding()
	}
}

func (b *binding) refreshBinding() {
	// check if the state is already consistent
	shouldActivate := b.source != nil && b.target != nil
	active := b.targetAction != nil && b.sourceState != nil

	if shouldActivate == active {
		return
	}

	if shouldActivate {
		if !b.validate() {
			return
		}

		// enable binding
		b.sourceState = b.source.StateItem(b.config.SourceState)
		b.targetAction = b.target.Action(b.config.TargetAction)

		// nil value indicate that the state is not fetched yet.
		// Do not transmit in case.
		value := b.sourceState.Get()
		if value != nil {
			b.targetAction <- value
		}

		// Note: we may miss a state change here

		b.sourceState.Subscribe(b.targetAction)

	} else {
		// disable binding
		b.sourceState.Unsubscribe(b.targetAction)
		b.sourceState = nil
		b.targetAction = nil
	}
}

func (b *binding) validate() bool {

	errors := make([]string, 0)

	sourceState := b.findMember(b.source, b.config.SourceState, metadata.State)
	targetAction := b.findMember(b.source, b.config.TargetAction, metadata.Action)

	if sourceState == nil {
		err := fmt.Sprintf("State '%s' does not exist on component %s", b.config.SourceState, b.buildComponentFullId(b.sourceInstance, b.source))
		errors = append(errors, err)
	}

	if targetAction == nil {
		err := fmt.Sprintf("Action '%s' does not exist on component %s", b.config.TargetAction, b.buildComponentFullId(b.targetInstance, b.target))
		errors = append(errors, err)
	}

	if sourceState != nil && targetAction != nil {
		sourceType := sourceState.ValueType()
		targetType := targetAction.ValueType()
		if !metadata.TypeEquals(sourceType, targetType) {
			sourceDesc := fmt.Sprintf("State '%s' on component %s", sourceState.Name(), b.buildComponentFullId(b.sourceInstance, b.source))
			targetDesc := fmt.Sprintf("action '%s' on component %s", targetAction.Name(), b.buildComponentFullId(b.targetInstance, b.target))
			err := fmt.Sprintf("%s has type '%s', which is different from type '%s' for %s", sourceDesc, sourceType, targetType, targetDesc)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		logger.Errorf("Binding '%s' errors: %s", b.config, strings.Join(errors, ", "))
		return false
	}

	return true
}

func (b *binding) buildComponentFullId(instanceName string, comp components.Component) string {
	if instanceName == "" {
		instanceName = "local"
	}

	return fmt.Sprintf("'%s' (plugin='%s:%s')", comp.Id(), instanceName, comp.Plugin().Id())
}

func (b *binding) findMember(comp components.Component, name string, memberType metadata.MemberType) *metadata.Member {
	member := comp.Plugin().Member(name)
	if member != nil && member.MemberType() == memberType {
		return member
	} else {
		return nil
	}
}
