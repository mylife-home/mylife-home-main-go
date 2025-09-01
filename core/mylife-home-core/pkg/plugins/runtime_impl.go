package plugins

import (
	"mylife-home-core-library/definitions"
)

var _ definitions.Runtime = (*runtimeImpl)(nil)

type runtimeImpl struct {
	id string
}

func (rt *runtimeImpl) ComponentId() string {
	return rt.id
}

func makeRuntime(componentId string) definitions.Runtime {
	return &runtimeImpl{
		id: componentId,
	}
}
