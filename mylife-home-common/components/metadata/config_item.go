package metadata

type ConfigType string

const (
	String  ConfigType = "string"
	Bool    ConfigType = "bool"
	Integer ConfigType = "integer"
	Float   ConfigType = "float"
)

type ConfigItem struct {
	name        string
	description string
	valueType   ConfigType
}

func (config *ConfigItem) Name() string {
	return config.name
}

func (config *ConfigItem) Description() string {
	return config.description
}

func (config *ConfigItem) ValueType() ConfigType {
	return config.valueType
}

func (ctype ConfigType) Validate(value any) bool {
	ok := false

	switch ctype {
	case String:
		_, ok = value.(string)
	case Bool:
		_, ok = value.(bool)
	case Integer:
		_, ok = value.(int64)
	case Float:
		_, ok = value.(float64)
	}

	return ok
}
