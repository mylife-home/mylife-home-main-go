package cmd

import (
	"mylife-home-core/pkg/manager"
	"mylife-home-core/pkg/plugins"
	"mylife-home-core/pkg/version"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"mylife-home-common/config"
	"mylife-home-common/defines"
	"mylife-home-common/instance_info"
	"mylife-home-common/log"
)

var logger = log.CreateLogger("mylife:home:core:main")

var configFile string
var logConsole bool

var rootCmd = &cobra.Command{
	Use:   "mylife-home-core",
	Short: "mylife-home-core - Mylife Home Core",
	Run:   run,
}

func run(_ *cobra.Command, _ []string) {
	log.Init(logConsole)
	config.Init(configFile)
	defines.Init("core", version.Value)
	instance_info.Init()
	plugins.Build()

	m := manager.MakeManager()

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)
	<-channel

	m.Terminate()
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
