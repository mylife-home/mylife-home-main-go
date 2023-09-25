package instance_info

import (
	"bufio"
	"fmt"
	"mylife-home-common/defines"
	"mylife-home-common/log"
	"mylife-home-common/tools"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/antichris/go-pirev"
)

var logger = log.CreateLogger("mylife:home:instance-info")

var instanceInfo *InstanceInfo
var listeners = tools.NewCallbackManager[*InstanceInfo]()

func Init() {
	instanceInfo = create()

	go func() {
		for {
			time.Sleep(time.Minute)
			update(extractData(instanceInfo))
		}
	}()
}

func Get() *InstanceInfo {
	return instanceInfo
}

func OnUpdate() tools.CallbackRegistration[*InstanceInfo] {
	return listeners
}

func update(newData *instanceInfoData) {
	newData.SystemUptime = int64(tools.SystemUptime().Seconds())
	newData.InstanceUptime = int64(tools.ApplicationUptime().Seconds())

	instanceInfo = newInstanceInfo(newData)

	listeners.Execute(instanceInfo)
}

func AddComponent(componentName string, version string) {
	newData := extractData(instanceInfo)
	addComponentVersion(newData.Versions, componentName, version)

	update(newData)
}

func AddCapability(capability string) {
	newData := extractData(instanceInfo)
	newData.Capabilities = append(newData.Capabilities, capability)

	update(newData)
}

func create() *InstanceInfo {
	mainComponent := defines.MainComponent()

	data := &instanceInfoData{
		Type:           mainComponent,
		Hardware:       getHardwareInfo(),
		Versions:       make(map[string]string),
		SystemUptime:   int64(tools.SystemUptime().Seconds()),
		InstanceUptime: int64(tools.ApplicationUptime().Seconds()),
		Hostname:       tools.Hostname(),
		Capabilities:   make([]string, 0),
	}

	data.Versions["os"] = runtime.GOOS + "/" + runtime.GOARCH
	data.Versions["golang"] = runtime.Version()

	addComponentVersion(data.Versions, "common", "")
	addComponentVersion(data.Versions, mainComponent, "")

	return newInstanceInfo(data)
}

func addComponentVersion(versions map[string]string, componentName string, version string) {
	name := "mylife-home-" + componentName
	if version == "" {
		// TODO: get build info
		version = "<unknown>"
	}

	versions[name] = version
}

func getHardwareInfo() map[string]string {
	hardware := make(map[string]string)

	rev, model := findRpiData()
	if rev == 0 {
		// not a rpi
		hardware["main"] = runtime.GOARCH
		return hardware
	}

	info := pirev.Identify(pirev.Code(rev))

	hardware["main"] = model
	hardware["processor"] = info.Processor.String()
	hardware["memory"] = fmt.Sprintf("%dMB", info.MemSize)
	hardware["manufacturer"] = info.Manufacturer.String()

	return hardware
}

func findRpiData() (revision uint32, model string) {
	file, err := os.OpenFile("/proc/cpuinfo", os.O_RDONLY, os.ModePerm)
	if err != nil {
		logger.WithError(err).Debugf("could not open /proc/cpuinfo")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			logger.WithError(err).Debugf("invalid line in /proc/cpuinfo : '%s'", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Revision":
			rev, err := strconv.ParseUint(value, 16, 32)
			if err != nil {
				logger.WithError(err).Debugf("invalid revision code : '%s'", value)
			} else {
				revision = uint32(rev)
			}

		case "Model":
			model = value
		}
	}

	err = scanner.Err()
	if err != nil {
		logger.WithError(err).Debugf("could not read /proc/cpuinfo")
	}

	return
}
