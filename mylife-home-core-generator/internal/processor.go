package internal

import (
	annotation "github.com/YReshetko/go-annotation/pkg"
)

type AnnotationType string

const (
	AnnotationTypePlugin AnnotationType = "plugin"
	AnnotationTypeState  AnnotationType = "state"
	AnnotationTypeAction AnnotationType = "action"
	AnnotationTypeConfig AnnotationType = "config"
)

func init() {
	engine := MakeEngine()

	annotation.Register[Plugin](&PluginAnnotationProcessor{engine})
	annotation.Register[State](&StateAnnotationProcessor{engine})
	annotation.Register[Action](&ActionAnnotationProcessor{engine})
	annotation.Register[Config](&ConfigAnnotationProcessor{engine})
}

var _ annotation.AnnotationProcessor = (*PluginAnnotationProcessor)(nil)
var _ annotation.AnnotationProcessor = (*StateAnnotationProcessor)(nil)
var _ annotation.AnnotationProcessor = (*ActionAnnotationProcessor)(nil)
var _ annotation.AnnotationProcessor = (*ConfigAnnotationProcessor)(nil)

type PluginAnnotationProcessor struct {
	engine *Engine
}

func (processor *PluginAnnotationProcessor) Process(node annotation.Node) error {
	annotations := annotation.FindAnnotations[Plugin](node.Annotations())
	processor.engine.ProcessPluginAnnotations(node, annotations)

	return nil
}

func (processor *PluginAnnotationProcessor) Output() map[string][]byte {
	return processor.engine.Output()
}

func (processor *PluginAnnotationProcessor) Version() string {
	return "0.0.1"
}

func (processor *PluginAnnotationProcessor) Name() string {
	return "MylifeHomePluginAnnotationProcessor"
}

type StateAnnotationProcessor struct {
	engine *Engine
}

func (processor *StateAnnotationProcessor) Process(node annotation.Node) error {
	annotations := annotation.FindAnnotations[State](node.Annotations())
	processor.engine.ProcessStateAnnotations(node, annotations)

	return nil
}

func (processor *StateAnnotationProcessor) Output() map[string][]byte {
	return processor.engine.Output()
}

func (processor *StateAnnotationProcessor) Version() string {
	return "0.0.1"
}

func (processor *StateAnnotationProcessor) Name() string {
	return "MylifeHomeStateAnnotationProcessor"
}

type ActionAnnotationProcessor struct {
	engine *Engine
}

func (processor *ActionAnnotationProcessor) Process(node annotation.Node) error {
	annotations := annotation.FindAnnotations[Action](node.Annotations())
	processor.engine.ProcessActionAnnotations(node, annotations)

	return nil
}

func (processor *ActionAnnotationProcessor) Output() map[string][]byte {
	return processor.engine.Output()
}

func (processor *ActionAnnotationProcessor) Version() string {
	return "0.0.1"
}

func (processor *ActionAnnotationProcessor) Name() string {
	return "MylifeHomeActionAnnotationProcessor"
}

type ConfigAnnotationProcessor struct {
	engine *Engine
}

func (processor *ConfigAnnotationProcessor) Process(node annotation.Node) error {
	annotations := annotation.FindAnnotations[Config](node.Annotations())
	processor.engine.ProcessConfigAnnotations(node, annotations)

	return nil
}

func (processor *ConfigAnnotationProcessor) Output() map[string][]byte {
	return processor.engine.Output()
}

func (processor *ConfigAnnotationProcessor) Version() string {
	return "0.0.1"
}

func (processor *ConfigAnnotationProcessor) Name() string {
	return "MylifeHomeConfigAnnotationProcessor"
}
