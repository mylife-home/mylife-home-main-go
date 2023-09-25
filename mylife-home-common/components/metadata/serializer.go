package metadata

import (
	"encoding/json"
	"fmt"

	"github.com/gookit/goutil/errorx/panics"
)

type serializerImpl struct{}

var Seralizer = serializerImpl{}

type netPlugin struct {
	Module      string               `json:"module"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Usage       PluginUsage          `json:"usage"`
	Version     string               `json:"version"`
	Config      map[string]netConfig `json:"config"`
	Members     map[string]netMember `json:"members"`
}

type netConfig struct {
	Description string     `json:"description,omitempty"`
	ValueType   ConfigType `json:"valueType"`
}

type netMember struct {
	Description string     `json:"description,omitempty"`
	MemberType  MemberType `json:"memberType"`
	ValueType   string     `json:"valueType"`
}

type netComponent struct {
	Id     string `json:"id"`
	Plugin string `json:"plugin"`
}

func (e *serializerImpl) SerializeComponent(component *Component) any {
	return safeSerialize(netComponent{
		Id:     component.Id(),
		Plugin: component.Plugin(),
	})
}

func (e *serializerImpl) DeserializeComponent(data any) *Component {
	var net netComponent
	safeDeserialize(data, &net)

	panics.NotEmpty(net.Id)
	panics.NotEmpty(net.Plugin)

	return MakeComponent(net.Id, net.Plugin)
}

func (e *serializerImpl) SerializePlugin(plugin *Plugin) any {
	net := netPlugin{
		Module:      plugin.Module(),
		Name:        plugin.Name(),
		Description: plugin.Description(),
		Usage:       plugin.Usage(),
		Version:     plugin.Version(),
		Config:      make(map[string]netConfig),
		Members:     make(map[string]netMember),
	}

	for _, name := range plugin.ConfigNames() {
		configItem := plugin.Config(name)
		net.Config[name] = netConfig{
			Description: configItem.Description(),
			ValueType:   configItem.ValueType(),
		}
	}

	for _, name := range plugin.MemberNames() {
		member := plugin.Member(name)
		net.Members[name] = netMember{
			Description: member.Description(),
			MemberType:  member.MemberType(),
			ValueType:   member.ValueType().String(),
		}
	}

	return safeSerialize(net)
}

func (e *serializerImpl) DeserializePlugin(data any) *Plugin {
	var net netPlugin
	safeDeserialize(data, &net)

	panics.NotEmpty(net.Module)
	panics.NotEmpty(net.Name)
	panics.NotEmpty(net.Usage)
	panics.NotEmpty(net.Version)

	builder := MakePluginBuilder(net.Module, net.Name, net.Description, net.Usage, net.Version)

	for name, configItem := range net.Config {
		panics.NotEmpty(name)
		panics.NotEmpty(configItem.ValueType)

		builder.AddConfig(name, configItem.Description, configItem.ValueType)
	}

	for name, member := range net.Members {
		panics.NotEmpty(name)
		panics.NotEmpty(member.MemberType)
		panics.NotEmpty(member.ValueType)

		valueType, err := ParseType(member.ValueType)
		if err != nil {
			panic(err)
		}

		switch member.MemberType {
		case State:
			builder.AddState(name, member.Description, valueType)
		case Action:
			builder.AddAction(name, member.Description, valueType)
		default:
			panic(fmt.Errorf("unsupported member type '%s'", member.MemberType))
		}
	}

	return builder.Build()
}

func safeSerialize(net any) any {
	raw, err := json.Marshal(net)
	if err != nil {
		panic(err)
	}

	var value any

	if err := json.Unmarshal(raw, &value); err != nil {
		panic(err)
	}

	return value
}

func safeDeserialize(data any, net any) {
	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(raw, net); err != nil {
		panic(err)
	}
}
