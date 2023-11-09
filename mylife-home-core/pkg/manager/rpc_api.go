package manager

import (
	"mylife-home-common/bus"
	"mylife-home-common/instance_info"
	"mylife-home-core/pkg/store"
)

type rpcApi struct {
	transport        *bus.Transport
	cm               *componentManager
	supportsBindings bool
}

func makeRpcApi(transport *bus.Transport, cm *componentManager, supportsBindings bool) *rpcApi {
	api := &rpcApi{
		transport:        transport,
		cm:               cm,
		supportsBindings: supportsBindings,
	}

	api.transport.Rpc().Serve("components.add", bus.NewRpcService(api.componentAdd))
	api.transport.Rpc().Serve("components.remove", bus.NewRpcService(api.componentRemove))
	api.transport.Rpc().Serve("components.list", bus.NewRpcService(api.componentList))

	instance_info.AddCapability("components-api")

	if api.supportsBindings {
		api.transport.Rpc().Serve("bindings.add", bus.NewRpcService(api.bindingAdd))
		api.transport.Rpc().Serve("bindings.remove", bus.NewRpcService(api.bindingRemove))
		api.transport.Rpc().Serve("bindings.list", bus.NewRpcService(api.bindingList))

		instance_info.AddCapability("bindings-api")
	}

	api.transport.Rpc().Serve("store.save", bus.NewRpcService(api.storeSave))

	instance_info.AddCapability("store-api")

	return api
}

func (api *rpcApi) Terminate() {
	api.transport.Rpc().Unserve("components.add")
	api.transport.Rpc().Unserve("components.remove")
	api.transport.Rpc().Unserve("components.list")

	if api.supportsBindings {
		api.transport.Rpc().Unserve("bindings.add")
		api.transport.Rpc().Unserve("bindings.remove")
		api.transport.Rpc().Unserve("bindings.list")
	}

	api.transport.Rpc().Unserve("store.save")
}

func (api *rpcApi) componentAdd(config *store.ComponentConfig) (struct{}, error) {
	err := api.cm.AddComponent(config.Id, config.Plugin, config.Config)
	return struct{}{}, err
}

func (api *rpcApi) componentRemove(input struct {
	Id string `json:"id"`
}) (struct{}, error) {
	err := api.cm.RemoveComponent(input.Id)
	return struct{}{}, err
}

func (api *rpcApi) componentList(input struct{}) ([]*store.ComponentConfig, error) {
	list := api.cm.GetComponents()
	return list, nil
}

func (api *rpcApi) bindingAdd(config *store.BindingConfig) (struct{}, error) {
	err := api.cm.AddBinding(config)
	return struct{}{}, err
}

func (api *rpcApi) bindingRemove(config *store.BindingConfig) (struct{}, error) {
	err := api.cm.RemoveBinding(config)
	return struct{}{}, err
}

func (api *rpcApi) bindingList(input struct{}) ([]*store.BindingConfig, error) {
	list := api.cm.GetBindings()
	return list, nil
}

func (api *rpcApi) storeSave(input struct{}) (struct{}, error) {
	err := api.cm.Save()
	return struct{}{}, err
}
