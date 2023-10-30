package store

import (
	"fmt"
	"strings"
	"sync"

	"mylife-home-common/config"
	"mylife-home-common/log"

	"golang.org/x/exp/maps"
)

var logger = log.CreateLogger("mylife:home:core:store")

type storeConfig struct {
	Type             string         `mapstructure:"type"`
	OperationsConfig map[string]any `mapstructure:",remain"`
}

type Store struct {
	operations storeOperations
	components map[string]*ComponentConfig
	bindings   map[string]*BindingConfig
	mux        sync.Mutex // Need to sync because Save() is executed in its own goroutine
}

func MakeStore() *Store {
	conf := storeConfig{}
	config.BindStructure("store", &conf)

	operations := makeOperations(conf.Type, conf.OperationsConfig)

	store := &Store{
		operations: operations,
		components: make(map[string]*ComponentConfig),
		bindings:   make(map[string]*BindingConfig),
	}

	return store
}

func (store *Store) Load() error {
	store.mux.Lock()
	defer store.mux.Unlock()

	data, err := store.operations.Load()
	if err != nil {
		return err
	}

	items, err := modelDeserialize(data)
	if err != nil {
		return err
	}

	for _, item := range items {
		switch item.Type {
		case storeItemTypeComponent:
			config := item.Config.(*ComponentConfig)
			store.components[config.Id] = config

		case storeItemTypeBinding:
			config := item.Config.(*BindingConfig)
			key := store.buildBindingKey(config)
			store.bindings[key] = config

		default:
			return fmt.Errorf("unsupported type: '%s'", item.Type)
		}
	}

	logger.Infof("%d items loaded", len(items))

	return nil
}

func (store *Store) Save() error {
	store.mux.Lock()
	defer store.mux.Unlock()

	items := make([]storeItem, 0, len(store.components)+len(store.bindings))

	for _, config := range store.components {
		items = append(items, storeItem{
			Type:   storeItemTypeComponent,
			Config: config,
		})
	}

	for _, config := range store.bindings {
		items = append(items, storeItem{
			Type:   storeItemTypeBinding,
			Config: config,
		})
	}

	data, err := modelSerialize(items)
	if err != nil {
		return err
	}

	if err := store.operations.Save(data); err != nil {
		return err
	}

	logger.Infof("%d items saved", len(items))

	return nil
}

func (store *Store) SetComponent(config *ComponentConfig) {
	store.mux.Lock()
	defer store.mux.Unlock()

	store.components[config.Id] = config
}

func (store *Store) RemoveComponent(id string) {
	store.mux.Lock()
	defer store.mux.Unlock()

	delete(store.components, id)
}

func (store *Store) AddBinding(config *BindingConfig) {
	key := store.buildBindingKey(config)

	store.mux.Lock()
	defer store.mux.Unlock()

	store.bindings[key] = config
}

func (store *Store) RemoveBinding(config *BindingConfig) {
	key := store.buildBindingKey(config)

	store.mux.Lock()
	defer store.mux.Unlock()

	delete(store.bindings, key)
}

func (store *Store) buildBindingKey(config *BindingConfig) string {
	return strings.Join([]string{config.SourceComponent, config.SourceState, config.TargetComponent, config.TargetAction}, ":")
}

func (store *Store) GetComponents() []*ComponentConfig {
	store.mux.Lock()
	defer store.mux.Unlock()

	return maps.Values(store.components)
}

func (store *Store) GetBindings() []*BindingConfig {
	store.mux.Lock()
	defer store.mux.Unlock()

	return maps.Values(store.bindings)
}

func (store *Store) HasBindings() bool {
	store.mux.Lock()
	defer store.mux.Unlock()

	return len(store.bindings) > 0
}
