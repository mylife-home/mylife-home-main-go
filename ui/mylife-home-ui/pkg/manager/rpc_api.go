package manager

import (
	"mylife-home-common/bus"
	"mylife-home-common/instance_info"
	"mylife-home-ui/pkg/model"
)

type rpcApi struct {
	transport *bus.Transport
	model     *model.ModelManager
}

func makeRpcApi(transport *bus.Transport, model *model.ModelManager) *rpcApi {
	api := &rpcApi{
		transport: transport,
		model:     model,
	}

	api.transport.Rpc().Serve("definition.set", bus.NewRpcService(api.definitionSet))

	instance_info.AddCapability("ui-api")

	return api
}

func (api *rpcApi) Terminate() {
	api.transport.Rpc().Unserve("definition.set")
}

func (api *rpcApi) definitionSet(definition *model.Definition) (struct{}, error) {
	err := api.model.SetDefinition(definition)
	return struct{}{}, err
}
