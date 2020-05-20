// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package provider

import (
	"github.com/openconfig/gnmi/proto/gnmi"
)

// A GNMIProvider emits updates as gNMI SetRequests.
type GNMIProvider interface {
	Provider

	// InitGNMI initializes the provider with a gNMI client.
	InitGNMI(client gnmi.GNMIClient)

	// OpenConfig indicates whether the provider wants OpenConfig
	// type-checking.
	OpenConfig() bool
}
