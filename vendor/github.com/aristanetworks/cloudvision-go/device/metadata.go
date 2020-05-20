// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package device

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

// Metadata represents all grpc metadata about a device.
type Metadata struct {
	DeviceID string
	// TypeCheck indicates whether the gNMI server should perform OpenConfig type-checking.
	TypeCheck bool
	// OpenConfig indicates whether transmitted gNMI data is OpenConfig-modeled.
	OpenConfig       bool
	DeviceType       *string
	Alive            *bool
	CollectorVersion string
}

const (
	deviceIDMetadata         = "deviceID"
	openConfigMetadata       = "openConfig"
	typeCheckMetadata        = "typeCheck"
	deviceTypeMetadata       = "deviceType"
	deviceLivenessMetadata   = "deviceLiveness"
	collectorVersionMetadata = "collectorVersion"
)

// NewMetadataFromOutgoing returns a metadata from an outgoing context.
func NewMetadataFromOutgoing(ctx context.Context) (Metadata, error) {
	ret := Metadata{}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return ret, errors.Errorf("Unable to get metadata from outgoing context")
	}
	return newMetadata(md)
}

// NewMetadataFromIncoming returns a metadata from an incoming context.
func NewMetadataFromIncoming(ctx context.Context) (Metadata, error) {
	ret := Metadata{}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ret, errors.Errorf("Unable to get metadata from incoming context")
	}
	return newMetadata(md)
}

func newMetadata(md metadata.MD) (Metadata, error) {
	ret := Metadata{}
	var err error
	deviceIDVal := md.Get(deviceIDMetadata)
	if len(deviceIDVal) != 1 {
		return ret, errors.Errorf("Context should have device ID metadata")
	}
	ret.DeviceID = deviceIDVal[0]

	openConfigVal := md.Get(openConfigMetadata)
	if len(openConfigVal) != 1 {
		return ret, errors.Errorf("Context should have openConfig metadata")
	}
	ret.OpenConfig, err = strconv.ParseBool(openConfigVal[0])
	if err != nil {
		return ret, errors.Errorf("Error parsing openConfig value: %v", err)
	}

	typeCheckVal := md.Get(typeCheckMetadata)
	if len(typeCheckVal) != 1 {
		return ret, errors.Errorf("Context should have typeCheck metadata")
	}
	ret.TypeCheck, err = strconv.ParseBool(typeCheckVal[0])
	if err != nil {
		return ret, errors.Errorf("Error parsing typeCheck value: %v", err)
	}

	deviceTypeVal := md.Get(deviceTypeMetadata)
	if len(deviceTypeVal) != 0 {
		ret.DeviceType = &deviceTypeVal[0]
	}

	deviceLivenessVal := md.Get(deviceLivenessMetadata)
	if len(deviceLivenessVal) != 0 {
		t := (deviceLivenessVal[0] == "true")
		ret.Alive = &t
	}

	collectorVersionVal := md.Get(collectorVersionMetadata)
	if len(collectorVersionVal) != 0 {
		ret.CollectorVersion = collectorVersionVal[0]
	}

	return ret, nil
}
