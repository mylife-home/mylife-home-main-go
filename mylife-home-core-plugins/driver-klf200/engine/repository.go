package engine

import (
	"sync"

	"github.com/gookit/goutil/errorx/panics"
)

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
