package plugin

import (
	"context"
	"fmt"
	"math"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-logic-timers/engine"
	"sync"
)

// @Plugin(usage="logic")
type SmartTimerBinary struct {

	// @Config(name="initProgram" description="Programme executé à l'initialisation du composant. Ex: \"o*-off\" (Ne doit pas contenir de wait)")
	ConfigInitProgram string

	// @Config(name="triggerProgram" description="Programme principal, lancé sur déclencheur. Ex: \"o0-on w-1s o0-off o1-on w-1s o1-off\"")
	ConfigTriggerProgram string

	// @Config(name="cancelProgram" description="Programme executé à l'arrêt de triggerProgram. Ex: \"o*-off\" (Ne doit pas contenir de wait)")
	ConfigCancelProgram string

	// @State(description="Temps total du programme, en secondes")
	TotalTime definitions.State[float64]

	// @State(description="Temps écoulé du programme, en secondes")
	ProgressTime definitions.State[float64]

	// @State(type="range[0;100]" description="Pourcentage du programme accompli")
	Progress definitions.State[int64]

	// @State()
	Running definitions.State[bool]

	// @State()
	Output0 definitions.State[bool]

	// @State()
	Output1 definitions.State[bool]

	// @State()
	Output2 definitions.State[bool]

	// @State()
	Output3 definitions.State[bool]

	// @State()
	Output4 definitions.State[bool]

	// @State()
	Output5 definitions.State[bool]

	// @State()
	Output6 definitions.State[bool]

	// @State()
	Output7 definitions.State[bool]

	// @State()
	Output8 definitions.State[bool]

	// @State()
	Output9 definitions.State[bool]

	initProgram    *engine.Program[bool]
	triggerProgram *engine.Program[bool]
	cancelProgram  *engine.Program[bool]
	outputs        []definitions.State[bool] // easily address output
	onProgressChan chan *engine.ProgressArg
	onRunningChan  chan bool
	onOutputChan   chan *engine.OutputArg[bool]
	pumpExited     chan struct{}
	runMux         sync.Mutex
	run            *runData
}

func (component *SmartTimerBinary) Init(runtime definitions.Runtime) error {

	component.TotalTime.Set(0)
	component.ProgressTime.Set(0)
	component.Progress.Set(0)

	component.outputs = []definitions.State[bool]{
		component.Output0,
		component.Output1,
		component.Output2,
		component.Output3,
		component.Output4,
		component.Output5,
		component.Output6,
		component.Output7,
		component.Output8,
		component.Output9,
	}

	component.initProgram = engine.NewProgram[bool](component.parseOutputValue, component.ConfigInitProgram, false)
	component.triggerProgram = engine.NewProgram[bool](component.parseOutputValue, component.ConfigTriggerProgram, true)
	component.cancelProgram = engine.NewProgram[bool](component.parseOutputValue, component.ConfigCancelProgram, false)

	component.onProgressChan = make(chan *engine.ProgressArg)
	component.onRunningChan = make(chan bool)
	component.onOutputChan = make(chan *engine.OutputArg[bool])

	component.pumpExited = make(chan struct{})
	go component.pump()

	component.triggerProgram.OnProgress().Subscribe(component.onProgressChan)
	component.triggerProgram.OnRunning().Subscribe(component.onRunningChan, false)
	component.initProgram.OnOutput().Subscribe(component.onOutputChan)
	component.triggerProgram.OnOutput().Subscribe(component.onOutputChan)
	component.cancelProgram.OnOutput().Subscribe(component.onOutputChan)

	component.TotalTime.Set(component.triggerProgram.TotalTime().Seconds())

	component.initProgram.Run(context.Background())

	return nil
}

func (component *SmartTimerBinary) Terminate() {
	component.runMux.Lock()
	component.clear()
	component.runMux.Unlock()

	component.triggerProgram.OnProgress().Unsubscribe(component.onProgressChan)
	component.triggerProgram.OnRunning().Unsubscribe(component.onRunningChan)
	component.initProgram.OnOutput().Unsubscribe(component.onOutputChan)
	component.triggerProgram.OnOutput().Unsubscribe(component.onOutputChan)
	component.cancelProgram.OnOutput().Unsubscribe(component.onOutputChan)

	close(component.onProgressChan)
	close(component.onRunningChan)
	close(component.onOutputChan)

	<-component.pumpExited
}

func (component *SmartTimerBinary) pump() {
	defer close(component.pumpExited)

	onProgressChan := component.onProgressChan
	onRunningChan := component.onRunningChan
	onOutputChan := component.onOutputChan

	for onProgressChan != nil || onRunningChan != nil || onOutputChan != nil {
		select {
		case progress, ok := <-onProgressChan:
			if !ok {
				onProgressChan = nil
				continue
			}

			component.Progress.Set(int64(math.Round(progress.Percent())))
			component.ProgressTime.Set(progress.ProgressTime().Seconds())

		case running, ok := <-onRunningChan:
			if !ok {
				onRunningChan = nil
				continue
			}

			component.Running.Set(running)

		case output, ok := <-onOutputChan:
			if !ok {
				onOutputChan = nil
				continue
			}

			component.outputs[output.Index()].Set(output.Value())
		}
	}
}

// @Action
func (component *SmartTimerBinary) Trigger(arg bool) {
	if !arg {
		return
	}

	component.runMux.Lock()
	defer component.runMux.Unlock()

	component.clear()
	component.start()
}

// @Action
func (component *SmartTimerBinary) Cancel(arg bool) {
	if !arg {
		return
	}

	component.runMux.Lock()
	defer component.runMux.Unlock()

	component.clear()
}

// @Action
func (component *SmartTimerBinary) Toggle(arg bool) {
	if !arg {
		return
	}

	component.runMux.Lock()
	defer component.runMux.Unlock()

	if component.run != nil {
		component.clear()
	} else {
		component.start()
	}
}

// Call inside the lock
func (component *SmartTimerBinary) start() {
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan struct{})

	go component.entry(ctx, exit)

	component.run = &runData{
		cancel: cancel,
		exit:   exit,
	}
}

// Call inside the lock
func (component *SmartTimerBinary) clear() {
	run := component.run
	if run == nil {
		return
	}

	run.cancel()
	<-run.exit

	// Each lock wants to clear run if started
	// This means we can set to nil only if no lock is currently taken
	// Which means nobody is trying to clear us (in which case it will
	// set itself component.run to nil)
	if component.runMux.TryLock() {
		component.run = nil
		component.runMux.Unlock()
	}
}

func (component *SmartTimerBinary) entry(ctx context.Context, exit chan<- struct{}) {
	defer close(exit)

	ret := component.triggerProgram.Run(ctx)

	// interrupted
	if !ret {
		component.cancelProgram.Run(context.Background())
	}

	// set run as nil
	// FIXME: lock !?
	component.run = nil
}

func (component *SmartTimerBinary) parseOutputValue(arg string) (bool, error) {
	switch arg {
	case "on":
		return true, nil
	case "off":
		return false, nil
	}

	var def bool
	return def, fmt.Errorf("invalid binary out value: '%s'", arg)
}
