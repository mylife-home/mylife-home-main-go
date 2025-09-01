package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"mylife-home-common/config"
	"mylife-home-common/defines"
	"mylife-home-common/instance_info"
	"mylife-home-common/log"
	"mylife-home-common/version"
)

var logger = log.CreateLogger("mylife:home:ui:main")

var configFile string
var logConsole bool

var rootCmd = &cobra.Command{
	Use:   "mylife-home-ui",
	Short: "mylife-home-ui - Mylife Home UI",
	Run:   run,
}

func run(_ *cobra.Command, _ []string) {
	log.Init(logConsole)
	config.Init(configFile)
	defines.Init("ui", version.Value)
	instance_info.Init()

	// m := manager.MakeManager()

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)
	<-channel

	// m.Terminate()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "config.yaml", "config file (default is $(PWD)/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&logConsole, "log-console", false, "Log to console")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
