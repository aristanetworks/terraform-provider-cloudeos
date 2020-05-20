// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package devices

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aristanetworks/cloudvision-go/device"
	"github.com/aristanetworks/cloudvision-go/provider"
	pgnmi "github.com/aristanetworks/cloudvision-go/provider/gnmi"

	"github.com/aristanetworks/goarista/gnmi"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

func init() {
	// Set options
	options := map[string]device.Option{
		"address": device.Option{
			Description: "gNMI server host/port",
			Required:    true,
		},
		"paths": device.Option{
			Description: "gNMI subscription path (comma-separated if multiple)",
			Default:     "/",
			Required:    false,
		},
		"username": device.Option{
			Description: "gNMI subscription username",
			Default:     "",
			Required:    false,
		},
		"password": device.Option{
			Description: "gNMI subscription password",
			Default:     "",
			Required:    false,
		},
		"cafile": device.Option{
			Description: "Path to server TLS certificate file",
			Default:     "",
			Required:    false,
		},
		"certfile": device.Option{
			Description: "Path to client TLS certificate file",
			Default:     "",
			Required:    false,
		},
		"keyfile": device.Option{
			Description: "Path to client TLS private key file",
			Default:     "",
			Required:    false,
		},
		"compression": device.Option{
			Description: "Compression method (Supported options: \"\" and \"gzip\")",
			Default:     "",
			Required:    false,
		},
		"tls": device.Option{
			Description: "Enable TLS",
			Default:     "false",
			Required:    false,
		},
		"device_id": device.Option{
			Description: "device ID",
			Default:     "",
			Required:    false,
		},
	}

	// Register
	device.Register("openconfig", newOpenConfig, options)
}

type openconfigDevice struct {
	gNMIProvider provider.GNMIProvider
	gNMIClient   pb.GNMIClient
	config       *gnmi.Config
	deviceID     string
}

func (o *openconfigDevice) Alive() (bool, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = gnmi.NewContext(ctx, o.config)
	livenessPath := "/system/processes/process/state"
	req, err := gnmi.NewGetRequest(gnmi.SplitPaths([]string{livenessPath}), "")
	if err != nil {
		return false, err
	}
	resp, err := o.gNMIClient.Get(ctx, req)
	return err == nil && resp != nil && len(resp.Notification) > 0, nil
}

func (o *openconfigDevice) Providers() ([]provider.Provider, error) {
	return []provider.Provider{o.gNMIProvider}, nil
}

func (o *openconfigDevice) DeviceID() (string, error) {
	if o.deviceID != "" {
		return o.deviceID, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = gnmi.NewContext(ctx, o.config)
	// TODO: Use /components/component/state/serial-no after checking state/type is Chassis.
	// Not doing that now because EOS doesn't support it so I don't even know what the gNMI
	// response looks like..
	livenessPath := "/system/config"
	req, err := gnmi.NewGetRequest(gnmi.SplitPaths([]string{livenessPath}), "")
	if err != nil {
		return "", err
	}
	resp, err := o.gNMIClient.Get(ctx, req)
	if err != nil || len(resp.Notification) == 0 {
		return "", fmt.Errorf("Unable to get request to %v: %v", livenessPath, err)
	}
	var config map[string]string
	val := pgnmi.Unmarshal(resp.Notification[0].Update[0].Val)
	err = json.Unmarshal(val.([]byte), &config)
	if err != nil {
		return "", err
	}
	return config["openconfig-system:hostname"] + "." + config["openconfig-system:domain-name"], nil
}

func parseGNMIOptions(opt map[string]string) (*gnmi.Config, error) {
	config := &gnmi.Config{}
	var err error
	config.Addr, err = device.GetStringOption("address", opt)
	if err != nil {
		return nil, err
	}
	config.Username, err = device.GetStringOption("username", opt)
	if err != nil {
		return nil, err
	}
	config.Password, err = device.GetStringOption("password", opt)
	if err != nil {
		return nil, err
	}
	config.CAFile, err = device.GetStringOption("cafile", opt)
	if err != nil {
		return nil, err
	}
	config.CertFile, err = device.GetStringOption("certfile", opt)
	if err != nil {
		return nil, err
	}
	config.KeyFile, err = device.GetStringOption("keyfile", opt)
	if err != nil {
		return nil, err
	}
	config.Compression, err = device.GetStringOption("compression", opt)
	if err != nil {
		return nil, err
	}
	config.TLS, err = device.GetBoolOption("tls", opt)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// newOpenConfig returns an openconfig device.
func newOpenConfig(opt map[string]string) (device.Device, error) {
	deviceID, err := device.GetStringOption("device_id", opt)
	if err != nil {
		return nil, err
	}
	gNMIPaths, err := device.GetStringOption("paths", opt)
	if err != nil {
		return nil, err
	}
	openconfig := &openconfigDevice{}
	config, err := parseGNMIOptions(opt)
	if err != nil {
		return nil, err
	}
	client, err := gnmi.Dial(config)
	if err != nil {
		return nil, err
	}
	openconfig.gNMIClient = client
	openconfig.config = config
	openconfig.deviceID = deviceID

	openconfig.gNMIProvider = pgnmi.NewGNMIProvider(client, config, strings.Split(gNMIPaths, ","))

	return openconfig, nil
}
