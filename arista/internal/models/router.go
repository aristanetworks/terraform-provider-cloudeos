// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package models

// CloudProviderType type
type CloudProviderType int32

// Cloud provider names
const (
	CpUnknown CloudProviderType = iota
	Aws
	Azure
	Gcp
)

// TopologyType type
type TopologyType int32

// Topology types
const (
	TopoUnknown TopologyType = iota
	Clos
	Wan
)

// UnderlayConnectionType type
type UnderlayConnectionType int32

// Underlay connection types
const (
	UlNa UnderlayConnectionType = iota
	UlIgw
	UlPeering
	UlTgw
)

// OverlayConnectionType type
type OverlayConnectionType int32

// Overlay connection types
const (
	OlNa OverlayConnectionType = iota
	OlDps
	OlVxlan
	OlIpsec
)

// RoleType for the router
type RoleType int32

// Role type
const (
	RoleUnknown RoleType = iota
	Edge
	Spine
	Leaf
)

// DeploymentStatusCode for the device
type DeploymentStatusCode int32

// Deployment status codes
const (
	DepStatusUnknown DeploymentStatusCode = iota
	DepStatusInProgress
	DepStatusSuccess
	DepStatusErr
)

// CVStatusCode for CloudDeploy service
type CVStatusCode int32

// CloudDeploy service status code
const (
	Unknown        CVStatusCode = iota
	RtrCreated                  // Rtr object created successfully
	RtrDiscovered               // Rtr is streaming, waiting for it to be provisioned
	RtrProvisioned              // Rtr has been moved to a container
	RtrConfigWIP                // Rtr config change in progress
	RtrReady                    // Rtr is ready for next operation
	RtrFailed                   // Rtr could not be created
	RtrInactive                 // Rtr streaming status is inactive after it is provisioned
)

// DeviceStatusCode for UI purpose
type DeviceStatusCode int32

// DeviceStatusCode
const (
	DsUnKnown        DeviceStatusCode = iota
	DsWorkInProgress                  // Router deployment in progress
	DsSuccess                         // Router deployment succeeded
	DsError                           // Router deployment error
)

// NetworkInterfaceType for CloudDeploy
type NetworkInterfaceType int32

// Enum for interface type
const (
	IntfUnknown NetworkInterfaceType = iota
	IntfPrivate
	IntfPublic
	IntfInternal
)

// NetworkInterface on a device deployed in cloud
type NetworkInterface struct {
	IntfID        string
	Name          string
	IntfType      NetworkInterfaceType
	PrivateIPAddr []string
	PublicIPAddr  string
	SubnetID      string
	SecurityGroup string
}

// RouteTableIDs store public/private/internal route table ids associated with a resource
type RouteTableIDs struct {
	Public   []string
	Private  []string
	Internal []string
}

// CVInfo - used by CloudDeploy service
type CVInfo struct {
	StatusCode              CVStatusCode
	BootstrapConfig         string
	HaRtrID                 string
	PeerVpcRouteTableID     []string
	HaRouteTableIDs         RouteTableIDs
	StatusDesc              string
	StatusRecommendedAction string
	DeviceStatus            DeviceStatusCode
}

// AwsRouterDetail parameters
type AwsRouterDetail struct {
	AvailabilityZone string
	InstanceType     string
}

// AzureRouterDetail parameters
type AzureRouterDetail struct {
	AvailabilityZone string
	ResourceGroup    string
	InstanceType     string
	AvailabilitySet  string
}

// DpsPeerInfo object
type DpsPeerInfo struct {
	VtepIP     string
	UnderlayIP string
}

// BgpPeerInfo object
type BgpPeerInfo struct {
	VtepIP string
	Asn    uint32
}

// HAPeerInfo object
type HAPeerInfo struct {
	RtrIP                string   // HA peer IP
	PrivateSubnet        string   // HA peer private subnet
	PrivateRouteTableIDs []string // HA peer private route table IDs
}

// RtrPeerInfo - Various types of peers of a Router
type RtrPeerInfo struct {
	// When a Leaf Router gets deleted, we have to just
	// look at the DpsClosPeers slice for the peer Edge
	// Router to delete the corresponding Leaf Router
	// from there. Keeping DPS Clos and Wan peers separately
	// helps us achieve that
	DpsClosPeers map[string]DpsPeerInfo
	DpsWanPeers  map[string]DpsPeerInfo
	BgpWanPeers  map[string]BgpPeerInfo
	BgpClosPeers map[string]BgpPeerInfo
	HAPeers      map[string]HAPeerInfo
}

// Router object
type Router struct {
	Name                  string
	VpcID                 string
	CPType                CloudProviderType
	Region                string
	InstanceID            string
	DeviceSerialNum       string
	HAName                string
	ID                    string
	Cnps                  map[string]bool
	Tags                  map[string]string
	DeviceEnrollmentToken string
	RouteTableIDs         RouteTableIDs
	RouteReflector        bool
	AwsRtrDetail          AwsRouterDetail
	AzRtrDetail           AzureRouterDetail
	Intf                  []NetworkInterface
	// Attributes written to by CloudDeploy App
	DepStatus DeploymentStatusCode
	CVInfo    CVInfo
	// The following attributes are not part of
	// the Router protobuf
	PeerInfo         RtrPeerInfo
	VtepIP           string
	TerminAttrIP     string
	Lo10IPAddr       string
	Asn              uint32
	Role             RoleType
	PostBootupConfig string
	ModelVersion     string
}

// Clone - Deep clone a RouteTableIDs object
func (rti *RouteTableIDs) Clone() *RouteTableIDs {
	r := &RouteTableIDs{}
	r.Public = make([]string, len(rti.Public))
	copy(r.Public, rti.Public)
	r.Private = make([]string, len(rti.Private))
	copy(r.Private, rti.Private)
	r.Internal = make([]string, len(rti.Internal))
	copy(r.Internal, rti.Internal)
	return r
}

// Clone - Deep clone a CVInfo object
func (cvi *CVInfo) Clone() *CVInfo {
	c := &CVInfo{
		StatusCode:              cvi.StatusCode,
		BootstrapConfig:         cvi.BootstrapConfig,
		HaRtrID:                 cvi.HaRtrID,
		HaRouteTableIDs:         *cvi.HaRouteTableIDs.Clone(),
		StatusDesc:              cvi.StatusDesc,
		StatusRecommendedAction: cvi.StatusRecommendedAction,
		DeviceStatus:            cvi.DeviceStatus,
	}
	c.PeerVpcRouteTableID = make([]string, len(cvi.PeerVpcRouteTableID))
	copy(c.PeerVpcRouteTableID, cvi.PeerVpcRouteTableID)
	return c
}

// Clone - Deep clone a RtrPeerInfo object
func (rpi *RtrPeerInfo) Clone() *RtrPeerInfo {
	r := &RtrPeerInfo{}

	r.DpsClosPeers = cloneDpsPeerInfoMap(rpi.DpsClosPeers)
	r.DpsWanPeers = cloneDpsPeerInfoMap(rpi.DpsWanPeers)
	r.BgpWanPeers = cloneBgpPeerInfoMap(rpi.BgpWanPeers)
	r.BgpClosPeers = cloneBgpPeerInfoMap(rpi.BgpClosPeers)
	r.HAPeers = cloneHAPeerInfoMap(rpi.HAPeers)

	return r
}

// Clone - Deep clone a NetworkInterface object
func (ni *NetworkInterface) Clone() *NetworkInterface {
	n := &NetworkInterface{
		IntfID:        ni.IntfID,
		Name:          ni.Name,
		IntfType:      ni.IntfType,
		PublicIPAddr:  ni.PublicIPAddr,
		SubnetID:      ni.SubnetID,
		SecurityGroup: ni.SecurityGroup,
	}
	n.PrivateIPAddr = make([]string, len(ni.PrivateIPAddr))
	copy(n.PrivateIPAddr, ni.PrivateIPAddr)
	return n
}

// Clone - Deep clone a Router object
func (rtr *Router) Clone() *Router {
	r := &Router{
		Name:             rtr.Name,
		VpcID:            rtr.VpcID,
		CPType:           rtr.CPType,
		Region:           rtr.Region,
		InstanceID:       rtr.InstanceID,
		DeviceSerialNum:  rtr.DeviceSerialNum,
		HAName:           rtr.HAName,
		ID:               rtr.ID,
		RouteTableIDs:    *rtr.RouteTableIDs.Clone(),
		RouteReflector:   rtr.RouteReflector,
		AwsRtrDetail:     rtr.AwsRtrDetail,
		AzRtrDetail:      rtr.AzRtrDetail,
		DepStatus:        rtr.DepStatus,
		CVInfo:           *rtr.CVInfo.Clone(),
		PeerInfo:         *rtr.PeerInfo.Clone(),
		VtepIP:           rtr.VtepIP,
		TerminAttrIP:     rtr.TerminAttrIP,
		Lo10IPAddr:       rtr.Lo10IPAddr,
		Asn:              rtr.Asn,
		Role:             rtr.Role,
		Cnps:             map[string]bool{},
		PostBootupConfig: rtr.PostBootupConfig,
		ModelVersion:     rtr.ModelVersion,
	}
	r.Tags = make(map[string]string)
	for key, value := range rtr.Tags {
		r.Tags[key] = value
	}
	for _, intf := range rtr.Intf {
		r.Intf = append(r.Intf, *intf.Clone())
	}
	r.Cnps = cloneCnps(rtr.Cnps)
	return r
}

// Historical models

// RouterV0 -
type RouterV0 struct {
	Name                  string
	VpcID                 string
	CPType                CloudProviderType
	Region                string
	InstanceID            string
	DeviceSerialNum       string
	HAName                string
	ID                    string
	Cnps                  map[string]bool
	Tags                  map[string]string
	DeviceEnrollmentToken string
	RouteTableIDs         RouteTableIDs
	RouteReflector        bool
	AwsRtrDetail          AwsRouterDetail
	AzRtrDetail           AzureRouterDetail
	Intf                  []NetworkInterface
	// Attributes written to by CloudDeploy App
	DepStatus DeploymentStatusCode
	CVInfo    CVInfo
	// The following attributes are not part of
	// the Router protobuf
	PeerInfo         RtrPeerInfo
	VtepIP           string
	TerminAttrIP     string
	Lo10IPAddr       string
	Asn              uint32
	Role             RoleType
	PostBootupConfig string
}
