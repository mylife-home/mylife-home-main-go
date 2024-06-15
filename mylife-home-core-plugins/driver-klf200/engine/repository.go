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
		online: tools.MakeSubjectValue[bool](false),
	}
}

func (store *Store) SetClient(client *Client) {
	store.clientMux.Lock()
	defer store.clientMux.Unlock()

	panics.IsTrue(store.client == nil)

	store.client = client

	store.clientOnlineChangedChan = make(chan bool)

	tools.DispatchChannel(store.clientOnlineChangedChan, store.handleOnlineChanged)

	//store.client.Online().Subscribe(store.clientOnlineChangedChan, true)
}

func (store *Store) clearClient() {
	store.clientMux.Lock()
	defer store.clientMux.Unlock()

	panics.IsTrue(store.client != nil)

	//	store.client.Online().Unsubscribe(store.clientOnlineChangedChan)

	close(store.clientOnlineChangedChan)

	store.clientOnlineChangedChan = nil

	store.online.Update(false)

	store.client = nil
}

func (store *Store) UnsetClient() {
	store.clearClient()

	store.mux.Lock()
	defer store.mux.Unlock()
	store.clearDevices()
}

type storeContainer struct {
	store    *Store
	refCount int
}

var repository = make(map[string]*storeContainer)
var repoMux sync.Mutex

func GetStore(boxKey string) *Store {
	repoMux.Lock()
	defer repoMux.Unlock()

	container, ok := repository[boxKey]
	if !ok {
		logger.Infof("Creating new store for box key '%s'", boxKey)
		container = &storeContainer{
			store:    newStore(),
			refCount: 0,
		}

		repository[boxKey] = container
	}

	container.refCount += 1
	return container.store
}

func ReleaseStore(boxKey string) {
	repoMux.Lock()
	defer repoMux.Unlock()

	container := repository[boxKey]
	panics.IsTrue(container.refCount > 0)
	container.refCount -= 1

	if container.refCount == 0 {
		logger.Infof("Removing store for box key '%s'", boxKey)
		delete(repository, boxKey)
	}
}
