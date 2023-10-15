package manager

import (
	"mylife-home-common/bus"
	"mylife-home-common/config"
	"mylife-home-common/instance_info"
	"mylife-home-common/log"
	"mylife-home-core/pkg/plugins"
	"mylife-home-core/pkg/store"
)

var logger = log.CreateLogger("mylife:home:core:manager")

type managerConfig struct {
	SupportsBindings bool `mapstructure:"supportsBindings"`
}

type Manager struct {
	transport *bus.Transport
	cm        *componentManager
}

func MakeManager() (*Manager, error) {
	conf := managerConfig{}
	config.BindStructure("manager", &conf)

	supportsBindings := conf.SupportsBindings
	transport := bus.NewTransport(bus.NewOptions().SetPresenceTracking(supportsBindings))
	cm, err := makeComponentManager(transport)
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		transport: transport,
		cm:        cm,
	}

	manager.addPluginsInstanceInfo()

	if err := manager.transport.Rpc().Serve("components.add", bus.NewRpcService(manager.rpcComponentAdd)); err != nil {
		return manager, err
	}

	if err := manager.transport.Rpc().Serve("components.remove", bus.NewRpcService(manager.rpcComponentRemove)); err != nil {
		return manager, err
	}

	if err := manager.transport.Rpc().Serve("components.list", bus.NewRpcService(manager.rpcComponentList)); err != nil {
		return manager, err
	}

	instance_info.AddCapability("components-api")

	if manager.cm.SupportsBindings() {
		if err := manager.transport.Rpc().Serve("bindings.add", bus.NewRpcService(manager.rpcBindingAdd)); err != nil {
			return manager, err
		}

		if err := manager.transport.Rpc().Serve("bindings.remove", bus.NewRpcService(manager.rpcBindingRemove)); err != nil {
			return manager, err
		}

		if err := manager.transport.Rpc().Serve("bindings.list", bus.NewRpcService(manager.rpcBindingList)); err != nil {
			return manager, err
		}

		instance_info.AddCapability("bindings-api")
	}

	if err := manager.transport.Rpc().Serve("store.save", bus.NewRpcService(manager.rpcStoreSave)); err != nil {
		return manager, err
	}

	instance_info.AddCapability("store-api")

	return manager, nil
}

func (manager *Manager) Terminate() {

	manager.transport.Rpc().Unserve("components.add")
	manager.transport.Rpc().Unserve("components.remove")
	manager.transport.Rpc().Unserve("components.list")

	if manager.cm.SupportsBindings() {
		manager.transport.Rpc().Unserve("bindings.add")
		manager.transport.Rpc().Unserve("bindings.remove")
		manager.transport.Rpc().Unserve("bindings.list")
	}

	manager.transport.Rpc().Unserve("store.save")

	manager.cm.Terminate()
	manager.transport.Terminate()
}

func (manager *Manager) addPluginsInstanceInfo() {
	modules := make(map[string]string)
	for _, id := range plugins.Ids() {
		meta := plugins.GetPlugin(id).Metadata()
		modules[meta.Module()] = meta.Version()
	}

	for module, version := range modules {
		instance_info.AddComponent("core-plugins-"+module, version)
	}
}

func (manager *Manager) rpcComponentAdd(config *store.ComponentConfig) (struct{}, error) {
	err := manager.cm.AddComponent(config.Id, config.Plugin, config.Config)
	return struct{}{}, err
}

func (manager *Manager) rpcComponentRemove(input struct {
	Id string `json:"id"`
}) (struct{}, error) {
	err := manager.cm.RemoveComponent(input.Id)
	return struct{}{}, err
}

func (manager *Manager) rpcComponentList(input struct{}) ([]*store.ComponentConfig, error) {
	list := manager.cm.GetComponents()
	return list, nil
}

func (manager *Manager) rpcBindingAdd(config *store.BindingConfig) (struct{}, error) {
	err := manager.cm.AddBinding(config)
	return struct{}{}, err
}

func (manager *Manager) rpcBindingRemove(config *store.BindingConfig) (struct{}, error) {
	err := manager.cm.RemoveBinding(config)
	return struct{}{}, err
}

func (manager *Manager) rpcBindingList(input struct{}) ([]*store.BindingConfig, error) {
	list := manager.cm.GetBindings()
	return list, nil
}

func (manager *Manager) rpcStoreSave(input struct{}) (struct{}, error) {
	err := manager.cm.Save()
	return struct{}{}, err
}
