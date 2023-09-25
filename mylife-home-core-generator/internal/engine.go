package internal

import (
	"path"
	"path/filepath"
	"strings"

	annotation "github.com/YReshetko/go-annotation/pkg"
	"github.com/gookit/goutil/errorx/panics"
)

type Engine struct {
	generators map[string]*Generator
	outputDone bool
}

func MakeEngine() *Engine {
	return &Engine{generators: make(map[string]*Generator)}
}

func (engine *Engine) getOutputPath(node annotation.Node) string {
	filename := node.Meta().FileName()
	filename = filename[:len(filename)-2] + "annotations-generated.go"
	return path.Join(node.Meta().Dir(), filename)
}

func (engine *Engine) getModuleName(node annotation.Node) string {
	const modulePathPrefix = "mylife-home-core-plugins-"

	baseDir := filepath.Base(node.Meta().Root())
	panics.IsTrue(strings.HasPrefix(baseDir, modulePathPrefix), "Module root directory expected: '%s__module_name__', got: '%s'", modulePathPrefix, baseDir)
	return baseDir[len(modulePathPrefix):]
}

func (engine *Engine) getGenerator(node annotation.Node) *Generator {
	outputPath := engine.getOutputPath(node)

	if _, ok := engine.generators[outputPath]; !ok {
		moduleName := engine.getModuleName(node)
		engine.generators[outputPath] = MakeGenerator(node, outputPath, moduleName)
	}

	return engine.generators[outputPath]
}

func (engine *Engine) ProcessPluginAnnotations(node annotation.Node, annotations []Plugin) {
	panics.IsTrue(len(annotations) == 1)

	engine.getGenerator(node).ProcessPluginAnnotation(node, &annotations[0])
}

func (engine *Engine) ProcessStateAnnotations(node annotation.Node, annotations []State) {
	panics.IsTrue(len(annotations) == 1)

	engine.getGenerator(node).ProcessStateAnnotation(node, &annotations[0])
}

func (engine *Engine) ProcessActionAnnotations(node annotation.Node, annotations []Action) {
	panics.IsTrue(len(annotations) == 1)

	engine.getGenerator(node).ProcessActionAnnotation(node, &annotations[0])
}

func (engine *Engine) ProcessConfigAnnotations(node annotation.Node, annotations []Config) {
	panics.IsTrue(len(annotations) == 1)

	engine.getGenerator(node).ProcessConfigAnnotation(node, &annotations[0])
}

func (engine *Engine) Output() map[string][]byte {
	if engine.outputDone {
		return make(map[string][]byte)
	}

	output := make(map[string][]byte)

	for outputPath, generator := range engine.generators {
		output[outputPath] = generator.Output()
	}

	engine.outputDone = true

	return output
}
