//go:generate go run ../../mylife-home-core-generator/cmd/main.go .

// @Module(version="1.0.1")
package plugin_entry

import (
	_ "mylife-home-core-plugins-ui-base/plugin"
)
