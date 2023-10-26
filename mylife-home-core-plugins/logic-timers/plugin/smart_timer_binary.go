package plugin

import (
	"fmt"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-logic-timers/engine"
)

// @Plugin(usage="logic")
type SmartTimerBinary struct {

	// @Config(name="initProgram" description="Programme executé à l'initialisation du composant. Ex: 'o*-off' (Ne doit pas contenir de wait)")
	ConfigInitProgram string

	// @Config(name="triggerProgram" description="Programme principal, lancé sur déclencheur. Ex: 'o0-on w-1s o0-off o1-on w-1s o1-off'")
	ConfigTriggerProgram string

	// @Config(name="cancelProgram" description="Programme executé à l'arrêt de triggerProgram. Ex: 'o*-off' (Ne doit pas contenir de wait)")
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

	component.triggerProgram.OnProgress().Register(component.onProgress)
	component.triggerProgram.OnRunning().Register(component.onRunning)
	component.initProgram.OnOutput().Register(component.onOutput)
	component.triggerProgram.OnOutput().Register(component.onOutput)
	component.cancelProgram.OnOutput().Register(component.onOutput)

	component.TotalTime.Set(component.triggerProgram.TotalTime().Seconds())

	component.initProgram.RunSync()

	return nil
}

func (component *SmartTimerBinary) Terminate() {
	component.clear()
}

// @Action
func (component *SmartTimerBinary) Trigger(arg bool) {
	if !arg {
		return
	}

	component.clear()
	component.triggerProgram.Start()
}

// @Action
func (component *SmartTimerBinary) Cancel(arg bool) {
	if !arg {
		return
	}

	component.clear()
}

// @Action
func (component *SmartTimerBinary) Toggle(arg bool) {
	if !arg {
		return
	}

	if component.triggerProgram.Running() {
		component.clear()
	} else {
		component.triggerProgram.Start()
	}
}

func (component *SmartTimerBinary) clear() {
	if component.triggerProgram.Running() {
		component.triggerProgram.Interrupt()
		component.cancelProgram.RunSync()
	}
}

func (component *SmartTimerBinary) onProgress(arg *engine.ProgressArg) {
	component.Progress.Set(int64(arg.Percent()))
	component.ProgressTime.Set(arg.ProgressTime().Seconds())
}

func (component *SmartTimerBinary) onRunning(value bool) {
	component.Running.Set(value)
}

func (component *SmartTimerBinary) onOutput(arg *engine.OutputArg[bool]) {
	component.outputs[arg.Index()].Set(arg.Value())
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
