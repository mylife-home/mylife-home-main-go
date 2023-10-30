package plugins

import (
	"context"
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
)

var _ definitions.Runtime = (*runtimeImpl)(nil)

type runtimeImpl struct {
	id           string
	ctx          context.Context
	mainLoopChan *tools.ChannelMerger[func()]
}

func (rt *runtimeImpl) ComponentId() string {
	return rt.id
}

func (rt *runtimeImpl) Context() context.Context {
	return rt.ctx
}

func (rt *runtimeImpl) NewExecutor() definitions.Executor {
	return &executorImpl{
		channel: rt.mainLoopChan.Create(),
	}
}

func makeRuntime(componentId string, terminateContext context.Context, mainLoopChan *tools.ChannelMerger[func()]) definitions.Runtime {
	return &runtimeImpl{
		id:           componentId,
		ctx:          terminateContext,
		mainLoopChan: mainLoopChan,
	}
}

var _ definitions.Executor = (*executorImpl)(nil)

type executorImpl struct {
	channel chan<- func()
}

func (exec *executorImpl) Execute(callback func()) {
	exec.channel <- callback
}

func (exec *executorImpl) Terminate() {
	close(exec.channel)
}
