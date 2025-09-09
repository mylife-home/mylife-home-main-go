package api

type ComponentStates = map[string]any

type Reset = map[string]ComponentStates

type ComponentAdd struct {
	ID         string          `json:"id" tstype:",readonly"`
	Attributes ComponentStates `json:"attributes" tstype:",readonly"`
}

type ComponentRemove struct {
	ID string `json:"id" tstype:",readonly"`
}

type StateChange struct {
	ID    string `json:"id" tstype:",readonly"`
	Name  string `json:"name" tstype:",readonly"`
	Value any    `json:"value" tstype:",readonly"`
}
