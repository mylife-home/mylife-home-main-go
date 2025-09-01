package plugin

import (
	"mylife-home-common/log"
	"mylife-home-core-library/definitions"
	"mylife-home-core-plugins-logic-selectors/engine"
	"strings"
)

var logger = log.CreateLogger("mylife:home:core:plugins:logic-selectors:smart-input")

// @Plugin(usage="logic")
type SmartInput struct {

	// @Config(description="Séquence pour déclencher l'action. Ex: \"l ss\" ou \"l|ss\"")
	Triggers0 string
	// @Config(description="Séquence pour déclencher l'action. Ex: \"l ss\" ou \"l|ss\"")
	Triggers1 string
	// @Config(description="Séquence pour déclencher l'action. Ex: \"l ss\" ou \"l|ss\"")
	Triggers2 string
	// @Config(description="Séquence pour déclencher l'action. Ex: \"l ss\" ou \"l|ss\"")
	Triggers3 string

	// @State()
	Output0 definitions.State[bool]

	// @State()
	Output1 definitions.State[bool]

	// @State()
	Output2 definitions.State[bool]

	// @State()
	Output3 definitions.State[bool]

	manager *engine.InputManager
}

func (component *SmartInput) processTrigger(input string, callback func()) {
	triggers := strings.FieldsFunc(input, func(r rune) bool {
		return r == '|' || r == ' '
	})

	if len(triggers) == 0 {
		return
	}

	logger.Debugf("Configuring triggers: %+v", triggers)
	for _, trigger := range triggers {
		component.manager.AddConfig(trigger, callback)
	}
}

func (component *SmartInput) Init(runtime definitions.Runtime) error {
	component.manager = engine.NewInputManager()

	component.processTrigger(component.Triggers0, component.executeOutput0)
	component.processTrigger(component.Triggers1, component.executeOutput1)
	component.processTrigger(component.Triggers2, component.executeOutput2)
	component.processTrigger(component.Triggers3, component.executeOutput3)

	return nil
}

func (component *SmartInput) Terminate() {
	component.manager.Terminate()
	// Noop
}

// @Action()
func (component *SmartInput) Action(arg bool) {
	if arg {
		component.manager.Down()
	} else {
		component.manager.Up()
	}
}

func (component *SmartInput) executeOutput0() {
	component.Output0.Set(true)
	component.Output0.Set(false)
}

func (component *SmartInput) executeOutput1() {
	component.Output1.Set(true)
	component.Output1.Set(false)
}

func (component *SmartInput) executeOutput2() {
	component.Output2.Set(true)
	component.Output2.Set(false)
}

func (component *SmartInput) executeOutput3() {
	component.Output3.Set(true)
	component.Output3.Set(false)
}
