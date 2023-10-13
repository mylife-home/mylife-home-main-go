package store

import "os"

var _ storeOperations = (*FsOperations)(nil)

type FsOperations struct {
	path string
}

func (operations *FsOperations) Load() ([]byte, error) {
	return os.ReadFile(operations.path)
}

func (operations *FsOperations) Save(data []byte) error {
	return os.WriteFile(operations.path, data, 0644)
}

func init() {
	registerOperations("fs", func(config map[string]any) (storeOperations, error) {
		path, err := getConfigValue[string](config, "path")
		if err != nil {
			return nil, err
		}

		ops := &FsOperations{
			path: path,
		}

		return ops, nil
	})
}
