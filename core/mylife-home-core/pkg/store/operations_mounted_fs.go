package store

import (
	"os"
	"os/exec"
)

var _ storeOperations = (*MountedFsOperations)(nil)

type MountedFsOperations struct {
	path       string
	mountPoint string
}

func (operations *MountedFsOperations) Load() ([]byte, error) {
	return os.ReadFile(operations.path)
}

func (operations *MountedFsOperations) Save(data []byte) error {
	if mountErr := operations.remount("rw"); mountErr != nil {
		return mountErr
	}

	err := os.WriteFile(operations.path, data, 0644)
	mountErr := operations.remount("ro")

	if err != nil {
		return err
	}

	return mountErr
}

func (operations *MountedFsOperations) remount(mountType string) error {
	return exec.Command("mount", "-o", "remount,"+mountType, operations.mountPoint).Run()
}

func init() {
	registerOperations("mounted-fs", func(config map[string]any) storeOperations {
		return &MountedFsOperations{
			path:       getConfigValue[string](config, "path"),
			mountPoint: getConfigValue[string](config, "mountPoint"),
		}
	})
}
