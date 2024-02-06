package main

import (
	"mylife-home-core/cmd"

	// Plugin list here
	_ "mylife-home-core-plugins-driver-absoluta"
	_ "mylife-home-core-plugins-driver-notifications"
	_ "mylife-home-core-plugins-driver-tahoma"
	_ "mylife-home-core-plugins-logic-base"
	_ "mylife-home-core-plugins-logic-clim"
	_ "mylife-home-core-plugins-logic-colors"
	_ "mylife-home-core-plugins-logic-selectors"
	_ "mylife-home-core-plugins-logic-timers"
	_ "mylife-home-core-plugins-ui-base"
)

func main() {
	cmd.Execute()
}
