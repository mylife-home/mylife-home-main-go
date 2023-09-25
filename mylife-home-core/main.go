package main

import (
	"mylife-home-core/cmd"

	// Plugin list here
	_ "mylife-home-core-plugins-logic-base"
)

func main() {
	cmd.Execute()
}
