package engine

import (
	"mylife-home-common/tools"
	"sync"
)

var data = make(map[string]State)
var mux sync.Mutex

func GetState(key string) State {
	mux.Lock()
	defer mux.Unlock()

	state, ok := data[key]
	if !ok {
		state = tools.MakeSubjectValue(initialStateValue)
		data[key] = state
	}

	return state
}
