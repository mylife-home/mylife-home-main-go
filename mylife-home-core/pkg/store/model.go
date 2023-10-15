package store

// TODO: check config float vs int (for colors)

import (
	"encoding/json"
	"fmt"
)

type storeItemType string

const (
	storeItemTypeComponent storeItemType = "component"
	storeItemTypeBinding   storeItemType = "binding"
)

type storeItem struct {
	Type   storeItemType `json:"type"`
	Config itemConfig    `json:"config"`
}

type itemConfig interface{}

var _ itemConfig = (*ComponentConfig)(nil)
var _ itemConfig = (*BindingConfig)(nil)

type ComponentConfig struct {
	Id     string         `json:"id"`
	Plugin string         `json:"plugin"`
	Config map[string]any `json:"config"`
}

func (config *ComponentConfig) String() string {
	return fmt.Sprintf("%s (type=%s, config=%+v)", config.Id, config.Plugin, config.Config)
}

type BindingConfig struct {
	SourceComponent string `json:"sourceComponent"`
	SourceState     string `json:"sourceState"`
	TargetComponent string `json:"targetComponent"`
	TargetAction    string `json:"targetAction"`
}

func (config *BindingConfig) String() string {
	return fmt.Sprintf("%s.%s -> %s.%s", config.SourceComponent, config.SourceState, config.TargetComponent, config.TargetAction)
}

func modelDeserialize(data []byte) ([]storeItem, error) {
	var items []storeItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func modelSerialize(items []storeItem) ([]byte, error) {
	return json.MarshalIndent(items, "", "  ")
}

func (item *storeItem) UnmarshalJSON(data []byte) error {

	var helper struct {
		Type   storeItemType   `json:"type"`
		Config json.RawMessage `json:"config"`
	}

	if err := json.Unmarshal(data, &helper); err != nil {
		return err
	}

	switch helper.Type {
	case storeItemTypeComponent:
		var config ComponentConfig
		if err := json.Unmarshal(helper.Config, &config); err != nil {
			return err
		}

		item.Type = helper.Type
		item.Config = &config
		return nil

	case storeItemTypeBinding:
		var config BindingConfig
		if err := json.Unmarshal(helper.Config, &config); err != nil {
			return err
		}

		item.Type = helper.Type
		item.Config = &config
		return nil
	}

	return fmt.Errorf("unsupported type: '%s'", helper.Type)
}
