package plugin

import (
	"mylife-home-core-library/definitions"
)

// @Plugin(usage="logic")
type ProgramSelector struct {

	// @Config()
	Program0 string

	// @Config()
	Program1 string

	// @Config()
	Program2 string

	// @Config()
	Program3 string

	// @Config()
	Program4 string

	// @Config()
	Program5 string

	// @Config()
	Program6 string

	// @Config()
	Program7 string

	// @Config()
	Program8 string

	// @Config()
	Program9 string

	// @State()
	Program definitions.State[string]
}

func (component *ProgramSelector) Init(runtime definitions.Runtime) error {
	component.Program.Set(component.Program0)

	return nil
}

func (component *ProgramSelector) Terminate() {
	// Noop
}

// @Action()
func (component *ProgramSelector) Set0(arg bool) {
	if arg {
		component.Program.Set(component.Program0)
	}
}

// @Action()
func (component *ProgramSelector) Set1(arg bool) {
	if arg {
		component.Program.Set(component.Program1)
	}
}

// @Action()
func (component *ProgramSelector) Set2(arg bool) {
	if arg {
		component.Program.Set(component.Program2)
	}
}

// @Action()
func (component *ProgramSelector) Set3(arg bool) {
	if arg {
		component.Program.Set(component.Program3)
	}
}

// @Action()
func (component *ProgramSelector) Set4(arg bool) {
	if arg {
		component.Program.Set(component.Program4)
	}
}

// @Action()
func (component *ProgramSelector) Set5(arg bool) {
	if arg {
		component.Program.Set(component.Program5)
	}
}

// @Action()
func (component *ProgramSelector) Set6(arg bool) {
	if arg {
		component.Program.Set(component.Program6)
	}
}

// @Action()
func (component *ProgramSelector) Set7(arg bool) {
	if arg {
		component.Program.Set(component.Program7)
	}
}

// @Action()
func (component *ProgramSelector) Set8(arg bool) {
	if arg {
		component.Program.Set(component.Program8)
	}
}

// @Action()
func (component *ProgramSelector) Set9(arg bool) {
	if arg {
		component.Program.Set(component.Program9)
	}
}
