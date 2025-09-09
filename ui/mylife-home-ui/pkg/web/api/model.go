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
	ID                 string    `json:"id" tstype:",readonly"`
	Style              Style     `json:"style" tstype:",readonly"`
	Height             int       `json:"height" tstype:",readonly"`
	Width              int       `json:"width" tstype:",readonly"`
	BackgroundResource Resource  `json:"backgroundResource" tstype:",readonly"`
	Controls           []Control `json:"controls" tstype:",readonly"`
}

type Control struct {
	ID              string         `json:"id" tstype:",readonly"`
	Style           Style          `json:"style" tstype:",readonly"`
	Height          int            `json:"height" tstype:",readonly"`
	Width           int            `json:"width" tstype:",readonly"`
	X               int            `json:"x" tstype:",readonly"`
	Y               int            `json:"y" tstype:",readonly"`
	Display         ControlDisplay `json:"display" tstype:",readonly"`
	Text            ControlText    `json:"text" tstype:",readonly"`
	PrimaryAction   Action         `json:"primaryAction" tstype:",readonly"`
	SecondaryAction Action         `json:"secondaryAction" tstype:",readonly"`
}

type ControlDisplay struct {
	ComponentID     string                  `json:"componentId" tstype:",readonly"`
	ComponentState  string                  `json:"componentState" tstype:",readonly"`
	DefaultResource Resource                `json:"defaultResource" tstype:",readonly"`
	Map             []ControlDisplayMapItem `json:"map" tstype:",readonly"`
}

type ControlDisplayMapItem struct {
	Min      int      `json:"min" tstype:",readonly"`
	Max      int      `json:"max" tstype:",readonly"`
	Value    any      `json:"value" tstype:"string | boolean,readonly"` // or others ?
	Resource Resource `json:"resource" tstype:",readonly"`
}

type ControlText struct {
	Context []ControlTextContextItem `json:"context" tstype:",readonly"`
	Format  string                   `json:"format" tstype:",readonly"`
}

type ControlTextContextItem struct {
	ID             string `json:"id" tstype:",readonly"`
	ComponentID    string `json:"componentId" tstype:",readonly"`
	ComponentState string `json:"componentState" tstype:",readonly"`
}

type Action struct {
	Component ActionComponent `json:"component" tstype:",readonly"`
	Window    ActionWindow    `json:"window" tstype:",readonly"`
}

type ActionComponent struct {
	ID     string `json:"id" tstype:",readonly"`
	Action string `json:"action" tstype:",readonly"`
}

type ActionWindow struct {
	ID    string `json:"id" tstype:",readonly"`
	Popup bool   `json:"popup" tstype:",readonly"`
}
