package metadata

type Component struct {
	id     string
	plugin string
}

func MakeComponent(id string, plugin string) *Component {
	return &Component{
		id:     id,
		plugin: plugin,
	}
}

func (comp *Component) Id() string {
	return comp.id
}

func (comp *Component) Plugin() string {
	return comp.plugin
}
