package model

import "mylife-home-ui/pkg/web/api"

type Window = api.Window
type DefaultWindow = api.DefaultWindow
type Control = api.Control
type ControlDisplay = api.ControlDisplay
type ControlText = api.ControlText
type ControlTextContextItem = api.ControlTextContextItem
type Action = api.Action
type Style = api.Style

type Definition struct {
	Resources     []DefinitionResource `json:"resources"`
	Styles        []DefinitionStyle    `json:"styles"`
	Windows       []Window             `json:"windows"`
	DefaultWindow DefaultWindow        `json:"defaultWindow"`
}

type DefinitionResource struct {
	ID   string `json:"id"`
	Mime string `json:"mime"`
	Data string `json:"data"`
}

type DefinitionStyle struct {
	ID         string         `json:"id"`
	Properties map[string]any `json:"properties"`
}
