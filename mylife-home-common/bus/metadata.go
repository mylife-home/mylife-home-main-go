package bus

import (
	"maps"
	"mylife-home-common/tools"
	"sync"
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
	OnChange() tools.Observable[*ValueChange]

	InstanceName() string
	Values() map[string]any
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

func (meta *Metadata) Set(path string, value any) error {
	topic := meta.client.BuildTopic(metadataDomain, path)
	return meta.client.PublishRetain(topic, Encoding.WriteJson(value))
}

func (meta *Metadata) Clear(path string) error {
	topic := meta.client.BuildTopic(metadataDomain, path)
	return meta.client.ClearRetain(topic)
}

func (meta *Metadata) CreateView(remoteInstanceName string) (RemoteMetadataView, error) {
	view := &remoteMetadataView{
		client:       meta.client,
		instanceName: remoteInstanceName,
		onChange:     tools.MakeSubject[*ValueChange](),
		registry:     make(map[string]any),
	}

	if err := view.client.Subscribe(view.listenTopic(), view.onMessage); err != nil {
		return nil, err
	}

	return view, nil
}

func (meta *Metadata) CloseView(view RemoteMetadataView) error {
	viewImpl := view.(*remoteMetadataView)
	return viewImpl.client.Unsubscribe(viewImpl.listenTopic())
}

var _ RemoteMetadataView = (*remoteMetadataView)(nil)

type remoteMetadataView struct {
	client       *client
	msgToken     tools.RegistrationToken
	instanceName string
	onChange     tools.Subject[*ValueChange]

	registry map[string]any
	mux      sync.RWMutex
}

func (view *remoteMetadataView) onMessage(m *message) {
	view.mux.Lock()
	defer view.mux.Unlock()

	if len(m.Payload()) == 0 {
		delete(view.registry, m.Path())
		view.onChange.Notify(&ValueChange{ValueClear, m.Path(), nil})
	} else {
		value := Encoding.ReadJson(m.Payload())
		view.registry[m.Path()] = value
		view.onChange.Notify(&ValueChange{ValueSet, m.Path(), value})
	}
}

func (view *remoteMetadataView) listenTopic() string {
	return view.client.BuildRemoteTopic(view.instanceName, metadataDomain, "#")
}

func (view *remoteMetadataView) OnChange() tools.Observable[*ValueChange] {
	return view.onChange
}

func (view *remoteMetadataView) InstanceName() string {
	return view.instanceName
}

func (view *remoteMetadataView) Values() map[string]any {
	view.mux.RLock()
	defer view.mux.RUnlock()

	return maps.Clone(view.registry)
}
