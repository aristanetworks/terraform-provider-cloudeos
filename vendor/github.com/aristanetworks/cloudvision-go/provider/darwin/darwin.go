// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package darwin

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aristanetworks/cloudvision-go/provider"
	pgnmi "github.com/aristanetworks/cloudvision-go/provider/gnmi"
	"github.com/aristanetworks/cloudvision-go/provider/openconfig"
	"github.com/openconfig/gnmi/proto/gnmi"
)

type darwin struct {
	client       gnmi.GNMIClient
	errc         chan error
	pollInterval time.Duration
}

// Return a set of gNMI updates for in/out bytes, packets, and errors
// for a given interface, as reported by netstat.
func updatesFromNetstatLine(fields []string) ([]*gnmi.Update, error) {
	intfName := fields[0]
	inBytes, err := strconv.ParseUint(fields[6], 10, 64)
	if err != nil {
		return nil, err
	}
	inPkts, err := strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		return nil, err
	}
	inErrs, err := strconv.ParseUint(fields[5], 10, 64)
	if err != nil {
		return nil, err
	}
	outBytes, err := strconv.ParseUint(fields[9], 10, 64)
	if err != nil {
		return nil, err
	}
	outPkts, err := strconv.ParseUint(fields[7], 10, 64)
	if err != nil {
		return nil, err
	}
	outErrs, err := strconv.ParseUint(fields[8], 10, 64)
	if err != nil {
		return nil, err
	}

	return []*gnmi.Update{
		pgnmi.Update(pgnmi.IntfStatePath(intfName, "name"),
			pgnmi.Strval(intfName)),
		pgnmi.Update(pgnmi.IntfPath(intfName, "name"),
			pgnmi.Strval(intfName)),
		pgnmi.Update(pgnmi.IntfConfigPath(intfName, "name"),
			pgnmi.Strval(intfName)),
		pgnmi.Update(pgnmi.IntfStatePath(intfName, "type"),
			pgnmi.Strval(openconfig.InterfaceType(6))),
		pgnmi.Update(pgnmi.IntfStatePath(intfName, "admin-status"),
			pgnmi.Strval(openconfig.IntfAdminStatus(1))),
		pgnmi.Update(pgnmi.IntfStatePath(intfName, "oper-status"),
			pgnmi.Strval(openconfig.IntfOperStatus(1))),
		pgnmi.Update(pgnmi.IntfStateCountersPath(intfName, "in-octets"),
			pgnmi.Uintval(inBytes)),
		pgnmi.Update(pgnmi.IntfStateCountersPath(intfName, "in-unicast-pkts"),
			pgnmi.Uintval(inPkts)),
		pgnmi.Update(pgnmi.IntfStateCountersPath(intfName, "in-errors"),
			pgnmi.Uintval(inErrs)),
		pgnmi.Update(pgnmi.IntfStateCountersPath(intfName, "out-octets"),
			pgnmi.Uintval(outBytes)),
		pgnmi.Update(pgnmi.IntfStateCountersPath(intfName, "out-unicast-pkts"),
			pgnmi.Uintval(outPkts)),
		pgnmi.Update(pgnmi.IntfStateCountersPath(intfName, "out-errors"),
			pgnmi.Uintval(outErrs)),
	}, nil
}

func (d *darwin) updateInterfaces() ([]*gnmi.SetRequest, error) {
	ns := exec.Command("netstat", "-ibn")
	out, err := ns.CombinedOutput()
	if err != nil {
		return nil, err
	}

	setRequest := new(gnmi.SetRequest)
	updates := make([]*gnmi.Update, 0)
	interfaceList := make(map[string]bool)

	// Iterate over lines in output, selecting only ethernet interfaces
	// to stream out updates for.
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && strings.HasPrefix(fields[0], "en") {
			intfName := fields[0]
			if _, ok := interfaceList[intfName]; ok {
				continue
			}
			interfaceList[intfName] = true
			u, err := updatesFromNetstatLine(fields)
			if err != nil {
				return nil, err
			}
			updates = append(updates, u...)
		}
	}

	setRequest.Delete = []*gnmi.Path{pgnmi.Path("interfaces")}
	setRequest.Replace = updates
	return []*gnmi.SetRequest{setRequest}, nil
}

func (d *darwin) handleErrors(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-d.errc:
			return fmt.Errorf("Error in darwin provider: %v", err)
		}
	}
}

func (d *darwin) Run(ctx context.Context) error {
	// Run updateInterfaces at the specified polling interval,
	// forever. PollForever sends the updates produced by
	// updateInterfaces to the gNMI client and sends any
	// resulting errors to the error channel to be handled by
	// handleErrors.
	go pgnmi.PollForever(ctx, d.client, d.pollInterval,
		d.updateInterfaces, d.errc)

	// handleErrors only returns if it sees an error.
	return d.handleErrors(ctx)
}

func (d *darwin) InitGNMI(client gnmi.GNMIClient) {
	d.client = client
}

func (d *darwin) OpenConfig() bool {
	return true
}

// NewDarwinProvider returns a darwin provider that registers a
// Darwin device and streams interface statistics.
func NewDarwinProvider(pollInterval time.Duration) provider.GNMIProvider {
	return &darwin{
		errc:         make(chan error),
		pollInterval: pollInterval,
	}
}
