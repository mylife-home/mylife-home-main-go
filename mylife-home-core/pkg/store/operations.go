package store

import (
	"reflect"

	"github.com/gookit/goutil/errorx/panics"
)

type storeOperations interface {
	Load() ([]byte, error)
	Save(data []byte) error
}

type storeOperationsFactory = func(config map[string]any) storeOperations

var operationsRegistry = make(map[string]storeOperationsFactory)

func makeOperations(typ string, config map[string]any) storeOperations {
	factory, ok := operationsRegistry[typ]
	panics.IsTrue(ok, "invalid store operations type: '%s'", typ)

	return factory(config)
}

func registerOperations(typ string, factory storeOperationsFactory) {
	operationsRegistry[typ] = factory
}

func getConfigValue[T any](config map[string]any, key string) T {
	var defaultValue T

	value, ok := config[key]
	panics.IsTrue(ok, "config is missing field '%s'", key)

	typedValue, ok := value.(T)
	panics.IsTrue(ok, "config wrong value type for field '%s' (got '%s', expected '%s')", key, reflect.TypeOf(value), reflect.TypeOf(defaultValue))

	return typedValue
}
