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

	meta.client.PublishNoWait(topic, Encoding.WriteJson(value), true)
}

func (meta *Metadata) Clear(path string) {
	topic := meta.client.BuildTopic(metadataDomain, path)

	meta.client.PublishNoWait(topic, []byte{}, true)
}

func (meta *Metadata) CreateView(remoteInstanceName string) RemoteMetadataView {
	view := &remoteMetadataView{
		client:       meta.client,
		instanceName: remoteInstanceName,
		onChange:     tools.NewCallbackManager[*ValueChange](),
		registry:     make(map[string]any),
	}

	view.msgToken = view.client.OnMessage().Register(view.onMessage)

	view.client.SubscribeNoWait(view.listenTopic())

	return view
}

func (meta *Metadata) CloseView(view RemoteMetadataView) {
	viewImpl := view.(*remoteMetadataView)
	viewImpl.client.OnMessage().Unregister(viewImpl.msgToken)

	viewImpl.client.UnsubscribeNoWait(viewImpl.listenTopic())
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
