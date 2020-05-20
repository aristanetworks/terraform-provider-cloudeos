// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package models

// TopologyInfoType -
type TopologyInfoType int32

// TopologyInfo types
const (
	TopoInfoTypeUnspecified TopologyInfoType = iota
	TopoInfoMeta
	TopoInfoWan
	TopoInfoClos
)

// FabricType -
type FabricType int32

// Fabric types
const (
	FabricTypeUnspecified FabricType = iota
	FullMesh
	HubSpoke
)

// WanInfo -
type WanInfo struct {
	WanName              string
	CPType               CloudProviderType
	PeerNames            []string
	EdgeEdgePeering      bool
	EdgeEdgeIgw          bool
	EdgeDedicatedConnect bool // DirectConnect
	CvpContainerName     string
}

// ClosInfo -
type ClosInfo struct {
	ClosName         string
	CPType           CloudProviderType
	Fabric           FabricType
	LeafEdgePeering  bool
	LeafEdgeIgw      bool
	LeafEncryption   bool
	CvpContainerName string
}

// TopologyInfo -
type TopologyInfo struct {
	// Topology meta info
	Name                string `index:"true"`
	ID                  string
	TopoType            TopologyInfoType `index:"true"`
	BgpAsnLow           uint32
	BgpAsnHigh          uint32
	VtepIPCidr          string   // CIDR block for VTEP IPs on vEOS
	TerminAttrIPCidr    string   // Loopback IP range on vEOS
	DpsControlPlaneCidr string   // Dps Control plane IP Cidr
	ManagedDevices      []string // Hostnames of existing vEOS instances
	CVaaSDomain         string
	CVaaSServer         string
	CVaaSUserName       string
	Wan                 WanInfo
	Clos                ClosInfo
	Version             string // Version of the client
	ModelVersion        string
}

// TopologyInfoV0 -
type TopologyInfoV0 struct {
	// Topology meta info
	Name                string `index:"true"`
	ID                  string
	TopoType            TopologyInfoType `index:"true"`
	BgpAsnLow           uint32
	BgpAsnHigh          uint32
	VtepIPCidr          string   // CIDR block for VTEP IPs on vEOS
	TerminAttrIPCidr    string   // Loopback IP range on vEOS
	DpsControlPlaneCidr string   // Dps Control plane IP Cidr
	ManagedDevices      []string // Hostnames of existing vEOS instances
	CVaaSDomain         string
	CVaaSServer         string
	CVaaSUserName       string
	Wan                 WanInfo
	Clos                ClosInfo
	Version             string // Version of the client
}
