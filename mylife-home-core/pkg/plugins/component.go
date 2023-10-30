package plugins

import (
	"context"
	"mylife-home-common/components"
	"mylife-home-common/components/metadata"
	"mylife-home-common/executor"
	"mylife-home-common/tools"
	"mylife-home-core-library/definitions"
	"sync"
)

var _ components.Component = (*Component)(nil)

type Component struct {
	// metadata/direct component management
	id            string
	plugin        *Plugin
	target        definitions.Plugin
	actions       map[string]func(any)                            // direct on component
	state         map[string]any                                  // state on main loop
	onStateChange *tools.CallbackManager[*components.StateChange] // on main loop

	// component loop management
	wg             sync.WaitGroup
	closeCtxCancel func()
	closeCtx       context.Context
	mainLoopChan   *tools.ChannelMerger[func()]
	actionExecutor definitions.Executor
}

func newComponent(id string, plugin *Plugin, target definitions.Plugin, actions map[string]func(any)) *Component {
	comp := &Component{
		id:            id,
		plugin:        plugin,
		target:        target,
		actions:       actions,
		state:         make(map[string]any),
		onStateChange: tools.NewCallbackManager[*components.StateChange](),
	}

	return comp
}

func (comp *Component) Id() string {
	return comp.id
}

func (comp *Component) Plugin() *metadata.Plugin {
	return comp.plugin.Metadata()
}

func (comp *Component) OnStateChange() tools.CallbackRegistration[*components.StateChange] {
	return comp.onStateChange
}

func (comp *Component) GetStateItem(name string) any {
	return comp.state[name]
}

func (comp *Component) GetState() tools.ReadonlyMap[string, any] {
	return tools.NewReadonlyMap(comp.state)
}

func (comp *Component) ExecuteAction(actionName string, arg any) {
	action := comp.actions[actionName]

	comp.actionExecutor.Execute(func() {
		action(arg)
	})
}

// Called from component loop on state change
func (comp *Component) stateChanged(stateName string, value any) {
	executor.Execute(func() {
		comp.state[stateName] = value
		comp.onStateChange.Execute(components.NewStateChange(stateName, value))
	})
}

func (comp *Component) Init() error {
	ctx, cancel := context.WithCancel(context.Background())
	comp.closeCtxCancel = cancel
	comp.closeCtx = ctx

	comp.startLoop()
	return comp.runTargetInit()
}

func (comp *Component) Terminate() {
	comp.closeCtxCancel()
	comp.runTargetTerminate()
	comp.stopLoop()
}

func (comp *Component) runTargetInit() error {

	rt := makeRuntime(comp.id, comp.closeCtx, comp.mainLoopChan)
	retc := make(chan error)

	comp.actionExecutor.Execute(func() {
		retc <- comp.target.Init(rt)
		close(retc)
	})

	// wait init response and return
	return <-retc

}

func (comp *Component) runTargetTerminate() {

	retc := make(chan struct{})

	comp.actionExecutor.Execute(func() {
		comp.target.Terminate()
		close(retc)
	})

	// wait response and return
	<-retc
}

func (comp *Component) startLoop() {
	// since actionExecutor is the initial channel, we cannot build it the usual way
	actionChan := make(chan func())
	comp.actionExecutor = &executorImpl{channel: actionChan}
	comp.mainLoopChan = tools.MakeChannelMerger[func()](actionChan)

	comp.wg.Add(1)
	go comp.loop()
}

func (comp *Component) stopLoop() {
	comp.actionExecutor.Terminate()
	// Note: all other executors have to terminate to make it exit
	comp.wg.Wait() // wait pluginLoop to exit
}

func (comp *Component) loop() {
	defer comp.wg.Done()

	bufferedOut, bufferedIn := tools.BufferedChannel[func()]()
	tools.PipeChannel(comp.mainLoopChan.Out(), bufferedOut)

	for callback := range bufferedIn {
		callback()
	}
}
