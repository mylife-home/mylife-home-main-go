package defines

import (
	"fmt"
	"os"

	"github.com/gookit/goutil/errorx/panics"
)

var mainComponentValue string
var mainComponentVersionValue string
var instanceNameValue string

func Init(mainComponent string, mainComponentVersion string) {
	mainComponentValue = mainComponent
	mainComponentVersionValue = mainComponentVersion

	hostname, err := os.Hostname()
	if err != nil {
		panic(fmt.Errorf("could not get hostname: %f", err))
	}
	instanceNameValue = hostname + "-" + mainComponentValue
}

func MainComponentVersion() string {
	panics.IsTrue(mainComponentVersionValue != "", "MainComponentVersion value has not been set")
	return mainComponentVersionValue
}

func MainComponent() string {
	panics.IsTrue(mainComponentValue != "", "MainComponent value has not been set")
	return mainComponentValue
}

func InstanceName() string {
	panics.IsTrue(instanceNameValue != "", "InstanceName value has not been set")
	return instanceNameValue
}
