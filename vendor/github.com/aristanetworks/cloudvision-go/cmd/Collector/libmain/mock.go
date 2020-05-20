// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package libmain

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/aristanetworks/cloudvision-go/device"
	pgnmi "github.com/aristanetworks/cloudvision-go/provider/gnmi"
	agnmi "github.com/aristanetworks/goarista/gnmi"
	"github.com/fatih/color"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/sirupsen/logrus"
)

type mockInfo struct {
	featureToPath  map[string]string
	seenUpdates    map[string]map[string]bool
	idToInfo       map[string]*device.Info
	lock           sync.Mutex
	seenAllUpdates chan struct{}
}

func (m *mockInfo) processRequest(ctx context.Context,
	req *gnmi.SetRequest) (*gnmi.SetResponse, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	seenAll := true
	for _, updates := range m.seenUpdates {
		if len(updates) < len(m.featureToPath) {
			seenAll = false
			break
		}
	}
	if seenAll {
		go func() {
			m.seenAllUpdates <- struct{}{}
		}()
		return nil, nil
	}
	md, err := device.NewMetadataFromOutgoing(ctx)
	if err != nil {
		return nil, err
	}
	updates := append(req.Replace, req.Update...)
	for _, update := range updates {
		path := agnmi.StrPath(update.Path)
		for _, p := range m.featureToPath {
			if strings.HasPrefix(path, p) {
				m.seenUpdates[md.DeviceID][p] = true
			}
		}
	}
	return nil, nil
}

func (m *mockInfo) printResults(seenAll bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 1, 8, ' ', 0)
	if seenAll {
		if len(m.featureToPath) > 0 {
			color.Green("All features are supported by all devices:")
		} else {
			color.Yellow("Mock mode is set without any paths to check. " +
				"Specify -mockCheckPath to check for feature support.")
		}
		for feature := range m.featureToPath {
			fmt.Fprintln(w, color.GreenString("    %s\tsupported", feature))
		}
	} else {
		color.Red("Some features are not supported by some devices:")
		for id, updates := range m.seenUpdates {
			fmt.Fprintln(w, m.idToInfo[id])
			for feature, path := range m.featureToPath {
				if _, ok := updates[path]; ok {
					fmt.Fprintln(w, color.GreenString("    %s\tsupported", feature))
				} else {
					fmt.Fprintln(w, color.RedString("    %s\tunsupported", feature))
				}
			}
		}
	}
	w.Flush()
}

func (m *mockInfo) initDevice(info *device.Info) {
	m.lock.Lock()
	m.seenUpdates[info.ID] = map[string]bool{}
	m.idToInfo[info.ID] = info
	m.lock.Unlock()
}

func (m *mockInfo) waitForUpdates(errChan chan error, timeout time.Duration) error {
	to := time.After(timeout)
	for {
		select {
		case err := <-errChan:
			return err
		case <-to:
			m.printResults(false)
			return errors.New("Insufficient updates seen within timeout")
		case <-m.seenAllUpdates:
			m.printResults(true)
			return nil
		}
	}
}

func newMockInfo(featureToPath map[string]string) *mockInfo {
	return &mockInfo{
		featureToPath:  featureToPath,
		seenUpdates:    map[string]map[string]bool{},
		lock:           sync.Mutex{},
		seenAllUpdates: make(chan struct{}),
		idToInfo:       map[string]*device.Info{},
	}
}

func runMock(ctx context.Context) {
	mockInfo := newMockInfo(mockFeature)
	inventory := device.NewInventory(ctx,
		pgnmi.NewSimpleGNMIClient(mockInfo.processRequest))
	configs, err := createDeviceConfigs()
	if err != nil {
		logrus.Fatal(err)
	}
	for _, config := range configs {
		info, err := device.NewDeviceInfo(config)
		if err != nil {
			logrus.Fatal(err)
		}
		mockInfo.initDevice(info)
		err = inventory.Add(info)
		if err != nil {
			logrus.Fatalf("Error in inventory.Add(): %v", err)
		}
	}
	logrus.Info("Mock Collector is running")
	errChan := make(chan error)
	err = mockInfo.waitForUpdates(errChan, *mockTimeout)
	if err != nil {
		logrus.Fatal(err)
	}
}
