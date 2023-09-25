package bus

import (
	"mylife-home-common/tools"
)

type ValueChangeType int

const (
	ValueSet ValueChangeType = iota
	ValueClear
)

type ValueChange struct {
	typ   ValueChangeType
	path  string
	value any
}

func (vs *ValueChange) Type() ValueChangeType {
	return vs.typ
}

func (vs *ValueChange) Path() string {
	return vs.path
}

func (vs *ValueChange) Value() any {
	return vs.value
}

type RemoteMetadataView interface {
	OnChange() tools.CallbackRegistration[*ValueChange]

	InstanceName() string
	Values() tools.ReadonlyMap[string, any]
}

const metadataDomain = "metadata"

type Metadata struct {
	client *client
}

func newMetadata(client *client) *Metadata {
	return &Metadata{
		client: client,
	}
}

func (meta *Metadata) Set(path string, value any) {
	topic := meta.client.BuildTopic(metadataDomain, path)

	fireAndForget(func() error {
		return meta.client.Publish(topic, Encoding.WriteJson(value), true)
	})
}

func (meta *Metadata) Clear(path string) {
	topic := meta.client.BuildTopic(metadataDomain, path)

	fireAndForget(func() error {
		return meta.client.Publish(topic, []byte{}, true)
	})
}

func (meta *Metadata) CreateView(remoteInstanceName string) (RemoteMetadataView, error) {
	view := &remoteMetadataView{
		client:       meta.client,
		instanceName: remoteInstanceName,
		onChange:     tools.NewCallbackManager[*ValueChange](),
		registry:     make(map[string]any),
	}

	view.msgToken = view.client.OnMessage().Register(view.onMessage)

	if err := view.client.Subscribe(view.listenTopic()); err != nil {
		view.client.OnMessage().Unregister(view.msgToken)
		return nil, err
	}

	return view, nil
}

func (meta *Metadata) CloseView(view RemoteMetadataView) {
	viewImpl := view.(*remoteMetadataView)
	viewImpl.client.OnMessage().Unregister(viewImpl.msgToken)

	if err := viewImpl.client.Unsubscribe(viewImpl.listenTopic()); err != nil {
		logger.WithError(err).Warnf("Error closing view to '%s'", view.InstanceName())
	}
}

type remoteMetadataView struct {
	client       *client
	msgToken     tools.RegistrationToken
	instanceName string
	onChange     *tools.CallbackManager[*ValueChange]
	registry     map[string]any
}

func (view *remoteMetadataView) onMessage(m *message) {

	if m.InstanceName() != view.instanceName || m.Domain() != metadataDomain {
		return
	}

	// Note: onMessage is called from one goroutine, no need for map sync
	if len(m.Payload()) == 0 {
		delete(view.registry, m.Path())
		view.onChange.Execute(&ValueChange{ValueClear, m.Path(), nil})
	} else {
		value := Encoding.ReadJson(m.Payload())
		view.registry[m.Path()] = value
		view.onChange.Execute(&ValueChange{ValueSet, m.Path(), value})
	}
}

func (view *remoteMetadataView) listenTopic() string {
	return view.client.BuildRemoteTopic(view.instanceName, metadataDomain, "#")
}

func (view *remoteMetadataView) OnChange() tools.CallbackRegistration[*ValueChange] {
	return view.onChange
}

func (view *remoteMetadataView) InstanceName() string {
	return view.instanceName
}

func (view *remoteMetadataView) Values() tools.ReadonlyMap[string, any] {
	return tools.NewReadonlyMap[string, any](view.registry)
}
