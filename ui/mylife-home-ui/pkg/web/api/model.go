package api

type Style = []string // static for now
type Resource = string

type Model struct {
	Windows       []Window      `json:"windows"`
	DefaultWindow DefaultWindow `json:"defaultWindow"`
	StyleHash     string        `json:"styleHash"` // css file to fetch
}

type DefaultWindow = map[string]string

type Window struct {
	ID                 string    `json:"id" tsype:",readonly"`
	Style              Style     `json:"style" tsype:",readonly"`
	Height             int       `json:"height" tsype:",readonly"`
	Width              int       `json:"width" tsype:",readonly"`
	BackgroundResource Resource  `json:"backgroundResource" tsype:",readonly"`
	Controls           []Control `json:"controls" tsype:",readonly"`
}

type Control struct {
	ID              string         `json:"id" tsype:",readonly"`
	Style           Style          `json:"style" tsype:",readonly"`
	Height          int            `json:"height" tsype:",readonly"`
	Width           int            `json:"width" tsype:",readonly"`
	X               int            `json:"x" tsype:",readonly"`
	Y               int            `json:"y" tsype:",readonly"`
	Display         ControlDisplay `json:"display" tsype:",readonly"`
	Text            ControlText    `json:"text" tsype:",readonly"`
	PrimaryAction   Action         `json:"primaryAction" tsype:",readonly"`
	SecondaryAction Action         `json:"secondaryAction" tsype:",readonly"`
}

type ControlDisplay struct {
	ComponentID     string                  `json:"componentId" tsype:",readonly"`
	ComponentState  string                  `json:"componentState" tsype:",readonly"`
	DefaultResource Resource                `json:"defaultResource" tsype:",readonly"`
	Map             []ControlDisplayMapItem `json:"map" tsype:",readonly"`
}

type ControlDisplayMapItem struct {
	Min      int      `json:"min" tsype:",readonly"`
	Max      int      `json:"max" tsype:",readonly"`
	Value    any      `json:"value" tsype:"string | boolean,readonly"` // or others ?
	Resource Resource `json:"resource" tsype:",readonly"`
}

type ControlText struct {
	Context []ControlTextContextItem `json:"context" tsype:",readonly"`
	Format  string                   `json:"format" tsype:",readonly"`
}

type ControlTextContextItem struct {
	ID             string `json:"id" tsype:",readonly"`
	ComponentID    string `json:"componentId" tsype:",readonly"`
	ComponentState string `json:"componentState" tsype:",readonly"`
}

type Action struct {
	Component ActionComponent `json:"component" tsype:",readonly"`
	Window    ActionWindow    `json:"window" tsype:",readonly"`
}

type ActionComponent struct {
	ID     string `json:"id" tsype:",readonly"`
	Action string `json:"action" tsype:",readonly"`
}

type ActionWindow struct {
	ID    string `json:"id" tsype:",readonly"`
	Popup bool   `json:"popup" tsype:",readonly"`
}
