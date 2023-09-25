package internal

import (
	"fmt"
	"go/ast"
	"mylife-home-common/components/metadata"
	"strings"

	annotation "github.com/YReshetko/go-annotation/pkg"
	"github.com/gookit/goutil/errorx/panics"
	"github.com/iancoleman/strcase"
)

type Generator struct {
	outputPath  string
	moduleName  string
	packageName string

	plugins []*PluginData
	states  []*StateData
	actions []*ActionData
	configs []*ConfigData
}

type PluginData struct {
	typeName string

	typ *ast.TypeSpec
	ann *Plugin

	states  []*StateData
	actions []*ActionData
	configs []*ConfigData

	name        string
	description string
	usage       metadata.PluginUsage
	version     string
}

type StateData struct {
	typeName  string
	fieldName string

	field *ast.Field
	ann   *State

	name        string
	description string
	valueType   metadata.Type
}

type ActionData struct {
	typeName string
	methName string

	fn  *ast.FuncDecl
	ann *Action

	name        string
	description string
	valueType   metadata.Type
}

type ConfigData struct {
	typeName  string
	fieldName string

	field *ast.Field
	ann   *Config

	name        string
	description string
	valueType   metadata.ConfigType
}

func MakeGenerator(node annotation.Node, outputPath string, moduleName string) *Generator {
	return &Generator{
		outputPath:  outputPath,
		moduleName:  moduleName,
		packageName: node.Meta().PackageName(),
		plugins:     make([]*PluginData, 0),
		states:      make([]*StateData, 0),
		actions:     make([]*ActionData, 0),
		configs:     make([]*ConfigData, 0),
	}
}

func (generator *Generator) ProcessPluginAnnotation(node annotation.Node, pluginAnnotation *Plugin) {
	typ, ok := annotation.CastNode[*ast.TypeSpec](node)
	panics.IsTrue(ok)
	panics.IsTrue(typ.TypeParams == nil || len(typ.TypeParams.List) == 0)

	generator.plugins = append(generator.plugins, &PluginData{
		typeName: typ.Name.Name,
		typ:      typ,
		ann:      pluginAnnotation,
	})
}

func (generator *Generator) ProcessStateAnnotation(node annotation.Node, ann *State) {
	field, ok := annotation.CastNode[*ast.Field](node)
	panics.IsTrue(ok)

	typ, ok := annotation.ParentType[*ast.TypeSpec](node)
	panics.IsTrue(ok)
	astType, ok := annotation.CastNode[*ast.TypeSpec](typ)
	panics.IsTrue(ok)

	generator.states = append(generator.states, &StateData{
		typeName:  astType.Name.Name,
		fieldName: field.Names[0].Name,
		field:     field,
		ann:       ann,
	})
}

func (generator *Generator) ProcessActionAnnotation(node annotation.Node, ann *Action) {
	fn, ok := annotation.CastNode[*ast.FuncDecl](node)
	panics.IsTrue(ok)

	astPtr, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
	panics.IsTrue(ok) // Else the function receiver is not a pointer

	generator.actions = append(generator.actions, &ActionData{
		typeName: astPtr.X.(*ast.Ident).Name,
		methName: fn.Name.Name,
		fn:       fn,
		ann:      ann,
	})
}

func (generator *Generator) ProcessConfigAnnotation(node annotation.Node, ann *Config) {
	field, ok := annotation.CastNode[*ast.Field](node)
	panics.IsTrue(ok)

	typ, ok := annotation.ParentType[*ast.TypeSpec](node)
	panics.IsTrue(ok)
	astType, ok := annotation.CastNode[*ast.TypeSpec](typ)
	panics.IsTrue(ok)

	generator.configs = append(generator.configs, &ConfigData{
		typeName:  astType.Name.Name,
		fieldName: field.Names[0].Name,
		field:     field,
		ann:       ann,
	})
}

func (generator *Generator) Output() []byte {
	generator.associate()
	generator.enrich()
	return generator.write()
}

func (generator *Generator) associate() {
	m := make(map[string]*PluginData)

	for _, plugin := range generator.plugins {
		plugin.states = make([]*StateData, 0)
		plugin.actions = make([]*ActionData, 0)
		plugin.configs = make([]*ConfigData, 0)

		m[plugin.typeName] = plugin
	}

	for _, state := range generator.states {
		plugin, ok := m[state.typeName]
		panics.IsTrue(ok)

		plugin.states = append(plugin.states, state)
	}

	for _, action := range generator.actions {
		plugin, ok := m[action.typeName]
		panics.IsTrue(ok)

		plugin.actions = append(plugin.actions, action)
	}

	for _, config := range generator.configs {
		plugin, ok := m[config.typeName]
		panics.IsTrue(ok)

		plugin.configs = append(plugin.configs, config)
	}
}

func (generator *Generator) enrich() {
	for _, plugin := range generator.plugins {
		plugin.name = plugin.ann.Name
		if plugin.name == "" {
			plugin.name = makePluginName(plugin.typeName)
		}

		plugin.description = plugin.ann.Description

		plugin.usage = parsePluginUsage(plugin.ann.Usage)

		panics.IsTrue(plugin.ann.Version != "")
		plugin.version = plugin.ann.Version

		for _, state := range plugin.states {
			state.name = state.ann.Name
			if state.name == "" {
				state.name = makeMemberName(state.fieldName)
			}

			state.description = state.ann.Description

			// Ensure we have "definitions.State[__the_type__]"
			indexExpr := state.field.Type.(*ast.IndexExpr)
			selExpr := indexExpr.X.(*ast.SelectorExpr)
			panics.IsTrue(selExpr.Sel.Name == "State")
			panics.IsTrue(selExpr.X.(*ast.Ident).Name == "definitions")
			nativeTypeName := indexExpr.Index.(*ast.Ident).Name

			state.valueType = parseType(state.ann.Type, nativeTypeName)
		}

		for _, action := range plugin.actions {
			action.name = action.ann.Name
			if action.name == "" {
				action.name = makeMemberName(action.methName)
			}

			action.description = action.ann.Description

			// Ensure that function has no return type, and only one param
			fnType := action.fn.Type
			panics.IsTrue(fnType.Results == nil || len(fnType.Results.List) == 0)
			panics.IsTrue(action.fn.Type.Params != nil && len(action.fn.Type.Params.List) == 1)
			nativeTypeName := action.fn.Type.Params.List[0].Type.(*ast.Ident).Name

			action.valueType = parseType(action.ann.Type, nativeTypeName)
		}

		for _, config := range plugin.configs {
			config.name = config.ann.Name
			if config.name == "" {
				config.name = makeMemberName(config.fieldName)
			}

			config.description = config.ann.Description

			nativeTypeName := config.field.Type.(*ast.Ident).Name
			config.valueType = parseConfigType(nativeTypeName)
		}
	}
}

func parsePluginUsage(value string) metadata.PluginUsage {
	switch strings.ToUpper(value) {
	case "SENSOR":
		return metadata.Sensor
	case "ACTUATOR":
		return metadata.Actuator
	case "LOGIC":
		return metadata.Logic
	case "UI":
		return metadata.Ui
	default:
		panic(fmt.Sprintf("Invalid plugin usage '%s'", value))
	}
}

func parseType(provided string, native string) metadata.Type {
	if provided == "" {
		switch native {
		case "string":
			return metadata.MakeTypeText()
		case "float64":
			return metadata.MakeTypeFloat()
		case "bool":
			return metadata.MakeTypeBool()
		case "any":
			return metadata.MakeTypeComplex()
		default:
			panic(fmt.Sprintf("Cannot infer metadata type from native type '%s'", native))
		}
	}

	providedType, err := metadata.ParseType(provided)
	panics.IsTrue(err != nil, "%f", err)

	switch providedType.(type) {
	case *metadata.RangeType:
		expectNative("int64", native)
	case *metadata.TextType:
		expectNative("string", native)
	case *metadata.FloatType:
		expectNative("float64", native)
	case *metadata.BoolType:
		expectNative("bool", native)
	case *metadata.EnumType:
		expectNative("string", native)
	case *metadata.ComplexType:
		expectNative("any", native)
	default:
		panic(fmt.Sprintf("Unexpected type '%s'", providedType.String()))
	}

	return providedType
}

func expectNative(expected string, native string) {
	if native != expected {
		panic(fmt.Sprintf("Range type: expected '%s', got '%s'", expected, native))
	}
}

func parseConfigType(native string) metadata.ConfigType {
	switch native {
	case "string":
		return metadata.String
	case "bool":
		return metadata.Bool
	case "int64":
		return metadata.Integer
	case "float64":
		return metadata.Float
	default:
		panic(fmt.Sprintf("Cannot infer config type from native type '%s'", native))
	}
}

func makePluginName(name string) string {
	return strcase.ToKebab(name)
}

func makeMemberName(name string) string {
	return strcase.ToLowerCamel(name)
}

func (generator *Generator) write() []byte {
	writer := MakeWrite(generator.packageName)

	for _, plugin := range generator.plugins {
		writer.BeginPlugin(plugin.typeName, generator.moduleName, plugin.name, plugin.description, plugin.usage, plugin.version)

		for _, state := range plugin.states {
			writer.AddState(state.fieldName, state.name, state.description, state.valueType)
		}

		for _, action := range plugin.actions {
			writer.AddAction(action.methName, action.name, action.description, action.valueType)
		}

		for _, config := range plugin.configs {
			writer.AddConfig(config.fieldName, config.name, config.description, config.valueType)
		}

		writer.EndPlugin()
	}

	return writer.Content()
}
