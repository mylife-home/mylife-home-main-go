package engine

import (
	"mylife-home-common/tools"
	"sync"

	"github.com/gookit/goutil/errorx/panics"
)

type Store struct {
	client                  *Client
	clientOnlineChangedChan chan bool
	clientMux               sync.Mutex

	online tools.SubjectValue[bool]

	mux sync.Mutex
}

func newStore() *Store {
	return &Store{
		online: tools.MakeSubjectValue(false),
	}
}

func (store *Store) SetClient(client *Client) {
	store.clientMux.Lock()
	defer store.clientMux.Unlock()

	panics.IsTrue(store.client == nil)

	store.client = client

	store.clientOnlineChangedChan = make(chan bool)

	tools.DispatchChannel(store.clientOnlineChangedChan, store.handleOnlineChanged)

	store.client.Online().Subscribe(store.clientOnlineChangedChan, true)
}

func (store *Store) clearClient() {
	store.clientMux.Lock()
	defer store.clientMux.Unlock()

	panics.IsTrue(store.client != nil)

	store.client.Online().Unsubscribe(store.clientOnlineChangedChan)

	close(store.clientOnlineChangedChan)

	store.clientOnlineChangedChan = nil

	store.online.Update(false)

	store.client = nil
}

func (store *Store) UnsetClient() {
	store.clearClient()

	store.mux.Lock()
	defer store.mux.Unlock()
	// store.clearDevices()
}

func (store *Store) Online() tools.ObservableValue[bool] {
	return store.online
}

func (store *Store) handleOnlineChanged(online bool) {
	store.online.Update(online)
	// for now we consider devices stay even if offline (and states will stay accurate ...)
}
