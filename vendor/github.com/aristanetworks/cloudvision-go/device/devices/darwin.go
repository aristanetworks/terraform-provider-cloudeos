// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package devices

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/aristanetworks/cloudvision-go/device"
	"github.com/aristanetworks/cloudvision-go/provider"
	pdarwin "github.com/aristanetworks/cloudvision-go/provider/darwin"
)

// Register this device with its options.
func init() {
	options := map[string]device.Option{
		"pollInterval": device.Option{
			Description: "Polling interval, with unit suffix (s/m/h)",
			Default:     "20s",
		},
	}
	device.Register("darwin", NewDarwinDevice, options)
}

type darwin struct {
	deviceID string
	provider provider.GNMIProvider
}

func (d *darwin) Alive() (bool, error) {
	// Runs on the device itself, so if the method is called, it's alive.
	return true, nil
}

// Use the device's serial number as its ID.
func (d *darwin) deviceSerial() (string, error) {
	profiler := exec.Command("system_profiler", "SPHardwareDataType")
	awk := exec.Command("awk", "/Serial/ {print $4}")
	pipe, err := profiler.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer pipe.Close()

	awk.Stdin = pipe

	err = profiler.Start()
	if err != nil {
		return "", err
	}
	out, err := awk.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

func (d *darwin) DeviceID() (string, error) {
	return d.deviceID, nil
}

func (d *darwin) Providers() ([]provider.Provider, error) {
	return []provider.Provider{d.provider}, nil
}

// NewDarwinDevice instantiates a MacBook device.
func NewDarwinDevice(options map[string]string) (device.Device, error) {
	pollInterval, err := device.GetDurationOption("pollInterval", options)
	if err != nil {
		return nil, err
	}

	device := darwin{}
	did, err := device.deviceSerial()
	if err != nil {
		return nil, fmt.Errorf("Failure getting device ID: %v", err)
	}
	device.deviceID = did
	device.provider = pdarwin.NewDarwinProvider(pollInterval)

	return &device, nil
}
