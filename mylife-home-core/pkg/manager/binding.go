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
	registry             *components.Registry
	config               *store.BindingConfig
	source               *bindingComponent
	target               *bindingComponent
	errors               []string
	registryChangeToken  tools.RegistrationToken
	componentChangeToken tools.RegistrationToken
}

func makeBinding(registry *components.Registry, config *store.BindingConfig) *binding {
	b := &binding{
		registry: registry,
		config:   config,
		errors:   make([]string, 0),
	}

	b.registryChangeToken = b.registry.OnComponentChange().Register(b.onComponentChange)

	for _, data := range b.registry.GetComponentsData().Clone() {
		b.onComponentAdd(data.InstanceName(), data.Component())
	}

	logger.Infof("Binding '%s' created", b.config)

	return b
}

func (b *binding) Terminate() {
	b.registry.OnComponentChange().Unregister(b.registryChangeToken)

	source := b.source
	if source != nil {
		b.onComponentRemove(source.instanceName, source.component)
	}

	target := b.target
	if target != nil {
		b.onComponentRemove(target.instanceName, target.component)
	}

	logger.Infof("Binding '%s' closed")
}

func (b *binding) Error() bool {
	return len(b.errors) > 0
}

func (b *binding) Errors() tools.ReadonlySlice[string] {
	return tools.NewReadonlySlice(b.errors)
}

func (b *binding) Active() bool {
	return b.source != nil && b.target != nil && !b.Error()
}

func (b *binding) onComponentChange(change *components.ComponentChange) {
	switch change.Action() {
	case components.RegistryAdd:
		b.onComponentAdd(change.InstanceName(), change.Component())
	case components.RegistryRemove:
		b.onComponentRemove(change.InstanceName(), change.Component())
	}
}

func (b *binding) onComponentAdd(instanceName string, component components.Component) {
	switch component.Id() {
	case b.config.SourceComponent:
		b.source = &bindingComponent{
			instanceName: instanceName,
			component:    component,
		}
		b.initBinding()

	case b.config.TargetComponent:
		b.target = &bindingComponent{
			instanceName: instanceName,
			component:    component,
		}
		b.initBinding()
	}
}

func (b *binding) onComponentRemove(instanceName string, component components.Component) {
	var sourceComponent components.Component
	var targetComponent components.Component

	if b.source != nil {
		sourceComponent = b.source.component
	}
	if b.target != nil {
		targetComponent = b.target.component
	}

	switch component {
	case sourceComponent:
		b.terminateBinding()
		b.source = nil

	case targetComponent:
		b.terminateBinding()
		b.target = nil
	}
}

func (b *binding) initBinding() {
	if b.source == nil || b.target == nil {
		return
	}

	sourceState := b.config.SourceState
	targetAction := b.config.TargetAction

	// assert that props exists and type matches
	sourcePlugin := b.source.component.Plugin()
	targetPlugin := b.target.component.Plugin()
	sourceMember := b.findMember(sourcePlugin, sourceState, metadata.State)
	targetMember := b.findMember(targetPlugin, targetAction, metadata.Action)

	errors := make([]string, 0)

	if sourceMember == nil {
		err := fmt.Sprintf("State '%s' does not exist on component %s", sourceState, b.buildComponentFullId(b.source))
		errors = append(errors, err)
	}

	if targetMember == nil {
		err := fmt.Sprintf("Action '%s' does not exist on component %s", targetAction, b.buildComponentFullId(b.target))
		errors = append(errors, err)
	}

	if sourceMember != nil && targetMember != nil {
		sourceType := sourceMember.ValueType()
		targetType := targetMember.ValueType()
		if !metadata.TypeEquals(sourceType, targetType) {
			sourceDesc := fmt.Sprintf("State '%s' on component %s", sourceState, b.buildComponentFullId(b.source))
			targetDesc := fmt.Sprintf("action '%s' on component %s", targetAction, b.buildComponentFullId(b.target))
			err := fmt.Sprintf("%s has type '%s', which is different from type '%s' for %s", sourceDesc, sourceType, targetType, targetDesc)
			errors = append(errors, err)
		}
	}

	b.errors = errors
	if len(errors) > 0 {
		// we have errors, do not activate binding
		logger.Errorf("Binding '%s' errors: %s", b.config, strings.Join(errors, ", "))
		return
	}

	sourceComponent := b.source.component

	b.componentChangeToken = sourceComponent.OnStateChange().Register(b.onSourceStateChange)

	value := sourceComponent.GetStateItem(sourceState)
	if value != nil {
		// else not provided yet, don't bind null values
		b.target.component.ExecuteAction(targetAction, value)
	}

	logger.Debugf("Binding '%s' activated", b.config)
}

func (b *binding) terminateBinding() {
	if b.Active() {
		b.source.component.OnStateChange().Unregister(b.componentChangeToken)
	}

	b.errors = make([]string, 0)

	logger.Debugf("Binding '%s' deactivated", b.config)
}

func (b *binding) onSourceStateChange(change *components.StateChange) {
	sourceState := b.config.SourceState
	targetAction := b.config.TargetAction

	if change.Name() != sourceState {
		return
	}

	if b.target != nil {
		b.target.component.ExecuteAction(targetAction, change.Value())
	}
}

func (b *binding) findMember(plugin *metadata.Plugin, name string, memberType metadata.MemberType) *metadata.Member {
	member := plugin.Member(name)
	if member != nil && member.MemberType() == memberType {
		return member
	} else {
		return nil
	}
}

func (b *binding) buildComponentFullId(comp *bindingComponent) string {
	instanceName := "local"
	if comp.instanceName != "" {
		instanceName = comp.instanceName
	}

	return fmt.Sprintf("'%s' (plugin='%s:%s')", comp.component.Id(), instanceName, comp.component.Plugin().Id())
}

type bindingComponent struct {
	instanceName string
	component    components.Component
}
