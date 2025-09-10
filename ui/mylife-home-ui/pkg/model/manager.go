package model

import (
	"encoding/json"
	"fmt"
	"mylife-home-ui/pkg/web/api"
	"os"

	"mylife-home-common/config"
	"mylife-home-common/log"
	"mylife-home-common/tools"
)

var logger = log.CreateLogger("mylife:home:ui:model")

type modelConfig struct {
	StorePath string `mapstructure:"storePath"`
}

type Resource struct {
	Mime string
	Data []byte
}

type RequiredComponentState struct {
	ComponentId    string
	ComponentState string
}

type ModelManager struct {
	modelHash               tools.SubjectValue[string]
	resources               map[string]*Resource
	requiredComponentStates []RequiredComponentState
	config                  modelConfig
}

func NewModelManager() *ModelManager {
	conf := modelConfig{}
	config.BindStructure("model", &conf)

	mm := &ModelManager{
		modelHash:               tools.MakeSubjectValue[string](""),
		resources:               make(map[string]*Resource),
		requiredComponentStates: []RequiredComponentState{},
		config:                  conf,
	}

	mm.load()

	return mm
}

func (mm *ModelManager) ModelHash() tools.ObservableValue[string] {
	return mm.modelHash
}

// Given without copy, do not modify
func (mm *ModelManager) GetRequiredComponentStates() []RequiredComponentState {
	return mm.requiredComponentStates
}

// Given without copy, do not modify
func (mm *ModelManager) GetResource(hash string) (*Resource, error) {
	resource, exists := mm.resources[hash]
	if !exists {
		return nil, fmt.Errorf("resource with hash '%s' does not exist", hash)
	}
	return resource, nil
}

func (mm *ModelManager) load() {
	content, err := os.ReadFile(mm.config.StorePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("Using default empty model")
			mm.setDefinition(mm.buildDefaultDefinition())
			return
		}

		panic(err)
	}

	logger.Infof("Load model from store: '%s'", mm.config.StorePath)

	var definition Definition
	if err := json.Unmarshal(content, &definition); err != nil {
		panic(err)
	}

	if err := mm.setDefinition(&definition); err != nil {
		panic(err)
	}
}

func (mm *ModelManager) buildDefaultDefinition() *Definition {
	return &Definition{
		Resources: []DefinitionResource{},
		Windows: []Window{
			{
				Id:                 "default-window",
				Style:              []string{},
				Width:              300,
				Height:             100,
				BackgroundResource: "",
				Controls: []Control{
					{
						Id:      "default-control",
						Style:   []string{},
						X:       0,
						Y:       0,
						Width:   300,
						Height:  100,
						Display: nil,
						Text: &ControlText{
							Format:  `return "No definition has been set";`,
							Context: []ControlTextContextItem{},
						},
						PrimaryAction:   nil,
						SecondaryAction: nil,
					},
				},
			},
		},
		DefaultWindow: DefaultWindow{
			"desktop": "default-window",
			"mobile":  "default-window",
		},
		Styles: []DefinitionStyle{},
	}
}

func (mm *ModelManager) setDefinition(definition *Definition) error {
	builder := newBuilder()

	if err := builder.BuildModel(definition); err != nil {
		return fmt.Errorf("failed to build model: %w", err)
	}

	mm.extractRequiredComponentStates(builder.Model)
	mm.resources = builder.Resources
	mm.modelHash.Update(builder.ModelHash)

	logger.Infof("Updated model : %s", mm.modelHash)

	logger.Infof("Save model to store: '%s'", mm.config.StorePath)
	content, err := json.Marshal(definition)
	if err != nil {
		return fmt.Errorf("failed to marshal definition: %w", err)
	}

	if err := os.WriteFile(mm.config.StorePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write model to store: %w", err)
	}

	return nil
}

func (mm *ModelManager) extractRequiredComponentStates(model *api.Model) {

	list := make([]RequiredComponentState, 0)

	for _, window := range model.Windows {
		for _, control := range window.Controls {
			display := control.Display
			text := control.Text

			if display != nil && display.ComponentId != "" && display.ComponentState != "" {
				list = append(list, RequiredComponentState{
					ComponentId:    display.ComponentId,
					ComponentState: display.ComponentState,
				})
			}

			if text != nil && text.Context != nil {
				for _, contextItem := range text.Context {
					list = append(list, RequiredComponentState{
						ComponentId:    contextItem.ComponentId,
						ComponentState: contextItem.ComponentState,
					})
				}
			}
		}
	}

	mm.requiredComponentStates = list
}
