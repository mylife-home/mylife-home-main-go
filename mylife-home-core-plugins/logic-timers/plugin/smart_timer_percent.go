package plugin

import (
	"fmt"
	"math"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-logic-timers/engine"
	"strconv"
)

// @Plugin(usage="logic")
type SmartTimerPercent struct {

	// @Config(name="initProgram" description="Programme executé à l'initialisation du composant. Ex: 'o*-0' (Ne doit pas contenir de wait)")
	ConfigInitProgram string

	// @Config(name="triggerProgram" description="Programme principal, lancé sur déclencheur. Ex: 'o0-50 w-1s o0-100 o1-50 w-1s o0-0 o1-0'")
	ConfigTriggerProgram string

	// @Config(name="cancelProgram" description="Programme executé à l'arrêt de triggerProgram. Ex: 'o*-0' (Ne doit pas contenir de wait)")
	ConfigCancelProgram string

	// @State(description="Temps total du programme, en secondes")
	TotalTime definitions.State[float64]

	// @State(description="Temps écoulé du programme, en secondes")
	ProgressTime definitions.State[float64]

	// @State(type="range[0;100]" description="Pourcentage du programme accompli")
	Progress definitions.State[int64]

	// @State()
	Running definitions.State[bool]

	// @State(type="range[0;100]")
	Output0 definitions.State[int64]

	// @State(type="range[0;100]")
	Output1 definitions.State[int64]

	// @State(type="range[0;100]")
	Output2 definitions.State[int64]

	// @State(type="range[0;100]")
	Output3 definitions.State[int64]

	// @State(type="range[0;100]")
	Output4 definitions.State[int64]

	// @State(type="range[0;100]")
	Output5 definitions.State[int64]

	// @State(type="range[0;100]")
	Output6 definitions.State[int64]

	// @State(type="range[0;100]")
	Output7 definitions.State[int64]

	// @State(type="range[0;100]")
	Output8 definitions.State[int64]

	// @State(type="range[0;100]")
	Output9 definitions.State[int64]

	executor       definitions.Executor
	initProgram    *engine.Program[int64]
	triggerProgram *engine.Program[int64]
	cancelProgram  *engine.Program[int64]
	outputs        []definitions.State[int64] // easily address output
}

func (component *SmartTimerPercent) Init(runtime definitions.Runtime) error {
	component.executor = runtime.NewExecutor()

	component.TotalTime.Set(0)
	component.ProgressTime.Set(0)
	component.Progress.Set(0)

	component.outputs = []definitions.State[int64]{
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

	component.initProgram = engine.NewProgram[int64](component.executor, component.parseOutputValue, component.ConfigInitProgram, false)
	component.triggerProgram = engine.NewProgram[int64](component.executor, component.parseOutputValue, component.ConfigTriggerProgram, true)
	component.cancelProgram = engine.NewProgram[int64](component.executor, component.parseOutputValue, component.ConfigCancelProgram, false)

	component.triggerProgram.OnProgress().Register(component.onProgress)
	component.triggerProgram.OnRunning().Register(component.onRunning)
	component.initProgram.OnOutput().Register(component.onOutput)
	component.triggerProgram.OnOutput().Register(component.onOutput)
	component.cancelProgram.OnOutput().Register(component.onOutput)

	component.TotalTime.Set(component.triggerProgram.TotalTime().Seconds())

	component.initProgram.RunSync()

	return nil
}

func (component *SmartTimerPercent) Terminate() {
	component.clear()

	component.executor.Terminate()
}

// @Action
func (component *SmartTimerPercent) Trigger(arg bool) {
	if !arg {
		return
	}

	component.clear()
	component.triggerProgram.Start()
}

// @Action
func (component *SmartTimerPercent) Cancel(arg bool) {
	if !arg {
		return
	}

	component.clear()
}

// @Action
func (component *SmartTimerPercent) Toggle(arg bool) {
	if !arg {
		return
	}

	if component.triggerProgram.Running() {
		component.clear()
	} else {
		component.triggerProgram.Start()
	}
}

func (component *SmartTimerPercent) clear() {
	if component.triggerProgram.Running() {
		component.triggerProgram.Interrupt()
		component.cancelProgram.RunSync()
	}
}

func (component *SmartTimerPercent) onProgress(arg *engine.ProgressArg) {
	component.Progress.Set(int64(math.Round(arg.Percent())))
	component.ProgressTime.Set(arg.ProgressTime().Seconds())
}

func (component *SmartTimerPercent) onRunning(value bool) {
	component.Running.Set(value)
}

func (component *SmartTimerPercent) onOutput(arg *engine.OutputArg[int64]) {
	component.outputs[arg.Index()].Set(arg.Value())
}

func (component *SmartTimerPercent) parseOutputValue(arg string) (int64, error) {
	value, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid percent out value '%s': %w", arg, err)
	}

	if value < 0 || value > 100 {
		return 0, fmt.Errorf("invalid percent out value '%s' (bad range)", arg)
	}

	return int64(value), nil
}
