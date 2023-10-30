package cmd

import (
	"fmt"
	"mylife-home-core/pkg/manager"
	"mylife-home-core/pkg/plugins"
	"mylife-home-core/pkg/version"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"mylife-home-common/config"
	"mylife-home-common/defines"
	"mylife-home-common/executor"
	"mylife-home-common/instance_info"
	"mylife-home-common/log"
)

var logger = log.CreateLogger("mylife:home:core:main")

var configFile string
var logConsole bool

var m *manager.Manager

var rootCmd = &cobra.Command{
	Use:   "mylife-home-core",
	Short: "mylife-home-core - Mylife Home Core",
	Run:   run,
}

func run(_ *cobra.Command, _ []string) {
	defines.Init("core", version.Value)
	log.Init(logConsole)
	config.Init(configFile)
	plugins.Build()

	fmt.Printf("TOT\n")

	executor.Run(
		setupSignals,
		instance_info.Init,
		setupManager)
}

func setupSignals() {
	fmt.Printf("setupSignals\n")
	channel := make(chan os.Signal, 1)

	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-channel
		signal.Reset(syscall.SIGINT, syscall.SIGTERM)
		logger.Debugf("Got signal %s", s)

		exit()
	}()
}

func setupManager() {
	var err error

	m, err = manager.MakeManager()
	if err != nil {
		logger.WithError(err).Error("Failed to initialize manager")
		exit()
		return
	}
}

func exit() {
	logger.Debug("Exit initiated")

	if m != nil {
		m.Terminate()
	}

	executor.Stop(false)
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
