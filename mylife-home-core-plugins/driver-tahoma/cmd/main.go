package main

import (
	"fmt"
	"os"

	"github.com/sgrimee/kizcool"
)

// https://ha101-1.overkiz.com/enduser-mobile-web

func main() {
	kiz, err := kizcool.New(os.Getenv("USERNAME"), os.Getenv("PASSWORD"), "https://ha101-1.overkiz.com/enduser-mobile-web", "")
	if err != nil {
		panic(err)
	}

	err = kiz.Login()
	if err != nil {
		panic(err)
	}

	devices, err := kiz.GetDevices()
	if err != nil {
		panic(err)
	}

	for _, device := range devices {
		fmt.Printf("Device name: '%s', URL: '%s'\n", device.Label, device.DeviceURL)
	}
}
