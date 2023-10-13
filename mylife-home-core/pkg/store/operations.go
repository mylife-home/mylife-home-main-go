package store

import (
	"fmt"
	"reflect"
)

type storeOperations interface {
	Load() ([]byte, error)
	Save(data []byte) error
}

type storeOperationsFactory = func(config map[string]any) (storeOperations, error)

var operationsRegistry = make(map[string]storeOperationsFactory)

func makeOperations(typ string, config map[string]any) (storeOperations, error) {
	factory, ok := operationsRegistry[typ]
	if !ok {
		return nil, fmt.Errorf("invalid store operations type: '%s'", typ)
	}

	return factory(config)
}

func registerOperations(typ string, factory storeOperationsFactory) {
	operationsRegistry[typ] = factory
}

func getConfigValue[T any](config map[string]any, key string) (T, error) {
	value, ok := config[key]
	if !ok {
		var defaultValue T
		return defaultValue, fmt.Errorf("config is missing field '%s'", key)
	}

	typedValue, ok := value.(T)
	if !ok {
		var defaultValue T
		return defaultValue, fmt.Errorf("config wrong value type for field '%s' (got '%s', expected '%s')", key, reflect.TypeOf(value), reflect.TypeOf(defaultValue))
	}

	return typedValue, nil
}
