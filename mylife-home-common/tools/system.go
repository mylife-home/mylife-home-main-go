package tools

import (
	"fmt"
	"os"
	"time"

	"github.com/mackerelio/go-osstat/uptime"
)

var startTime = time.Now()

func ApplicationUptime() time.Duration {
	return time.Since(startTime)
}

func SystemUptime() time.Duration {
	value, err := uptime.Get()
	if err != nil {
		panic(fmt.Errorf("could not get system uptime: %f", err))
	}

	return value
}

func Hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(fmt.Errorf("could not get hostname: %f", err))
	}

	return hostname
}
