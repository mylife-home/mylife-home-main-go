package store

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var _ storeOperations = (*LbuOperations)(nil)

// LbuOperations keeps the store in the Alpine diskless overlay. The root
// filesystem is a tmpfs rebuilt at every boot, so the file is written like any
// other and "lbu commit" then packs the overlay onto the boot media, which is
// what makes the write outlive a reboot.
//
// The path must be somewhere lbu includes, such as /etc.
type LbuOperations struct {
	path string
}

func (operations *LbuOperations) Load() ([]byte, error) {
	return os.ReadFile(operations.path)
}

func (operations *LbuOperations) Save(data []byte) error {
	if err := os.WriteFile(operations.path, data, 0644); err != nil {
		return err
	}

	return operations.commit()
}

// commit packs the overlay onto the boot media. A failure is returned to the
// caller because until it succeeds the data exists only in tmpfs, and the save
// would be lost at the next reboot without anything else reporting it.
func (operations *LbuOperations) commit() error {
	output, err := exec.Command("lbu", "commit", "-d").CombinedOutput()
	if err != nil {
		return fmt.Errorf("lbu commit failed: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func init() {
	registerOperations("lbu", func(config map[string]any) storeOperations {
		return &LbuOperations{
			path: getConfigValue[string](config, "path"),
		}
	})
}
