// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package device

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aristanetworks/cloudvision-go/log"
	"github.com/aristanetworks/cloudvision-go/provider"
	"github.com/aristanetworks/cloudvision-go/version"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const heartbeatInterval = 10 * time.Second

// An Inventory maintains a set of devices.
type Inventory interface {
	Add(deviceInfo *Info) error
	Delete(key string) error
	Get(key string) (*Info, error)
	List() []*Info
}

// deviceConn contains a device and its gNMI connections.
type deviceConn struct {
	info              *Info
	ctx               context.Context
	cancel            context.CancelFunc
	rawGNMIClient     gnmi.GNMIClient
	wrappedGNMIClient *gNMIClientWrapper
	group             sync.WaitGroup
}

// inventory implements the Inventory interface.
type inventory struct {
	ctx           context.Context
	rawGNMIClient gnmi.GNMIClient
	devices       map[string]*deviceConn
	lock          sync.Mutex
}

func (dc *deviceConn) sendPeriodicUpdates() error {
	ticker := time.NewTicker(heartbeatInterval)
	ctx := metadata.AppendToOutgoingContext(dc.ctx,
		collectorVersionMetadata, version.Version)
	if _, ok := dc.info.Device.(Manager); ok {
		// ManagementSystem is a system managing other devices which itself
		// shouldn't be treated as an actual streaming device in CloudVision.
		ctx = metadata.AppendToOutgoingContext(ctx,
			deviceTypeMetadata, "managementSystem")
	} else {
		// Target is an ordinary device streaming to CloudVision.
		ctx = metadata.AppendToOutgoingContext(ctx,
			deviceTypeMetadata, "target")
	}
	did, _ := dc.info.Device.DeviceID()

	if _, err := dc.wrappedGNMIClient.Set(ctx, &gnmi.SetRequest{}); err != nil {
		if err != nil {
			log.Log(dc.info.Device).Infof("Failed to send periodic "+
				"update for device %v", did)
		}
	}
	for {
		select {
		case <-dc.ctx.Done():
			return nil
		case <-ticker.C:
			if alive, err := dc.info.Device.Alive(); err == nil {
				if alive {
					ctx := metadata.AppendToOutgoingContext(dc.ctx,
						deviceLivenessMetadata, "true")
					_, err = dc.wrappedGNMIClient.Set(ctx, &gnmi.SetRequest{})
					if err != nil {
						// Don't give up if an update fails for some reason.
						log.Log(dc.info.Device).Infof("Failed to send periodic "+
							"update for device %v", did)
					}
				} else {
					log.Log(dc.info.Device).Infof("Device %s is not alive", did)
				}
			} else {
				log.Log(dc.info.Device).Infof("Device %s is not alive", did)
			}
		}
	}
}

func (i *inventory) newDeviceConn(info *Info) *deviceConn {
	dc := &deviceConn{info: info}
	dc.ctx, dc.cancel = context.WithCancel(i.ctx)
	dc.rawGNMIClient = i.rawGNMIClient
	dc.wrappedGNMIClient = newGNMIClientWrapper(dc.rawGNMIClient, nil,
		info.ID, false)
	return dc
}

func (dc *deviceConn) runProviders() error {
	providers, err := dc.info.Device.Providers()
	if err != nil {
		return err
	}
	logFileName := dc.info.ID + ".log"
	err = log.InitLogging(logFileName, dc.info.Device)
	if err != nil {
		return fmt.Errorf("Error setting up logging for device %s: %v", dc.info.ID, err)
	}

	for _, p := range providers {
		err = log.InitLogging(logFileName, p)
		if err != nil {
			return fmt.Errorf("Error setting up logging for provider %#v: %v", p, err)
		}

		pt, ok := p.(provider.GNMIProvider)
		if !ok {
			return errors.New("unexpected provider type; need GNMIProvider")
		}

		pt.InitGNMI(newGNMIClientWrapper(dc.rawGNMIClient, pt, dc.info.ID, pt.OpenConfig()))

		// Start the providers.
		dc.group.Add(1)
		go func(p provider.Provider) {
			err := p.Run(dc.ctx)
			if err != nil {
				log.Log(p).Errorf("Provider exiting with error %v", err)
			}
			dc.group.Done()
		}(p)
	}
	return nil
}

// Add adds a device to the inventory, opens up any gNMI connections
// required by the device's providers, and then starts its providers.
func (i *inventory) Add(info *Info) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	if info.ID == "" {
		return fmt.Errorf("ID in device.Info cannot be empty")
	}
	if dev, ok := i.devices[info.ID]; ok {
		if info.Config.Device != dev.info.Config.Device {
			return fmt.Errorf("Cannot add device '%s' (type '%s') to inventory; "+
				"it already exists with a different type ('%s')",
				info.ID, info.Config.Device, dev.info.Config.Device)
		}
		log.Log(info.Device).Infof("Device %s already exists (type '%s')\n",
			info.ID, info.Config.Device)
		return nil
	}

	dc := i.newDeviceConn(info)
	i.devices[info.ID] = dc

	if err := dc.runProviders(); err != nil {
		return err
	}

	// Send periodic updates of device-level metadata.
	dc.group.Add(1)
	go func() {
		err := dc.sendPeriodicUpdates()
		if err != nil {
			log.Log(info.Device).Errorf("Error updating device metadata: %v", err)
		}
		dc.group.Done()
	}()

	if manager, ok := info.Device.(Manager); ok {
		dc.group.Add(1)
		go func() {
			err := manager.Manage(i)
			if err != nil {
				log.Log(info.Device).Errorf("Error in manager.Manage: %v", err)
			}
			dc.group.Done()
		}()
	}

	log.Log(info.Device).Infof("Added device %s", info.ID)
	return nil
}

func (i *inventory) Delete(key string) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	if key == "" {
		return fmt.Errorf("key in inventory.Delete cannot be empty")
	}
	dc, ok := i.devices[key]
	if !ok {
		return nil
	}

	// Cancel the device context and delete the device from the device
	// map. We need to make sure this device's providers are finished
	// before deleting the device.
	dc.cancel()
	dc.group.Wait()
	delete(i.devices, key)
	log.Log(dc.info.Device).Infof("Deleted device %s", key)
	return nil
}

func (i *inventory) Get(key string) (*Info, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	if key == "" {
		return nil, fmt.Errorf("key in inventory.Get cannot be empty")
	}
	d, ok := i.devices[key]
	if !ok {
		return nil, fmt.Errorf("Device %s not found", key)
	}
	return d.info, nil
}

func (i *inventory) List() []*Info {
	var ret []*Info
	for _, conn := range i.devices {
		ret = append(ret, conn.info)
	}
	return ret
}

// NewInventory creates an Inventory.
func NewInventory(ctx context.Context, gnmiClient gnmi.GNMIClient) Inventory {
	inv := &inventory{
		ctx:           ctx,
		devices:       make(map[string]*deviceConn),
		rawGNMIClient: gnmiClient,
	}
	return inv
}
