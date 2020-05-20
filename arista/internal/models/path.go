// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION

package models

import (
	"encoding/json"

	"github.com/aristanetworks/goarista/key"
)

// PathKey -
type PathKey struct {
	SrcVpcUUID    string
	LocalRtrUUID  string
	DstVpcUUID    string
	RemoteRtrUUID string
	UlConn        UnderlayConnectionType
}

func (pk PathKey) String() string {
	b, _ := pk.MarshalJSON()
	return string(b)
}

// MarshalJSON -
func (pk PathKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(pk.ToBuiltin())
}

// ToBuiltin -
func (pk PathKey) ToBuiltin() interface{} {
	return map[string]interface{}{
		"SrcVpcUUID":    pk.SrcVpcUUID,
		"LocalRtrUUID":  pk.LocalRtrUUID,
		"DstVpcUUID":    pk.DstVpcUUID,
		"RemoteRtrUUID": pk.RemoteRtrUUID,
		"UlConn":        pk.UlConn,
	}
}

// PathCharacteristics -
type PathCharacteristics struct {
	LatencyMs uint64
	JitterMs  uint64
	PktLossPc uint64
	BwMbps    uint64
	Up        bool
	Uptime    uint64
}

// Path -
type Path struct {
	SrcVpcCloudID    string
	SrcVpcName       string
	SrcVpcUUID       string
	LocalIntfIPAddr  string
	LocalRtrCloudID  string
	LocalRtrName     string
	LocalRtrUUID     string
	SrcRegion        string
	SrcCpType        CloudProviderType
	DstVpcCloudID    string
	DstVpcName       string
	DstVpcUUID       string
	RemoteIntfIPAddr string
	RemoteRtrCloudID string
	RemoteRtrName    string
	RemoteRtrUUID    string
	DstRegion        string
	DstCpType        CloudProviderType
	TopologyName     string
	UlConn           UnderlayConnectionType
	PathChars        PathCharacteristics
	ModelVersion     string
}

// Key -
func (p Path) Key() key.Key {
	return key.New(PathKey{
		p.SrcVpcUUID,
		p.LocalRtrUUID,
		p.DstVpcUUID,
		p.RemoteRtrUUID,
		p.UlConn,
	})
}

// Historical models

// PathV0 -
type PathV0 struct {
	SrcVpcCloudID    string
	SrcVpcName       string
	SrcVpcUUID       string
	LocalIntfIPAddr  string
	LocalRtrCloudID  string
	LocalRtrName     string
	LocalRtrUUID     string
	SrcRegion        string
	SrcCpType        CloudProviderType
	DstVpcCloudID    string
	DstVpcName       string
	DstVpcUUID       string
	RemoteIntfIPAddr string
	RemoteRtrCloudID string
	RemoteRtrName    string
	RemoteRtrUUID    string
	DstRegion        string
	DstCpType        CloudProviderType
	TopologyName     string
	UlConn           UnderlayConnectionType
	PathChars        PathCharacteristics
}
