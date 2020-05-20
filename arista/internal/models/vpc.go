// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package models

import (
	"fmt"

	"github.com/aristanetworks/glog"
)

// PeerVpcInfo type
type PeerVpcInfo struct {
	// To be used for AWS
	PeerVpcCidr map[string]string // maps peer vpc ID -> cidr

	// To be used for Azure
	PeerRgName   string
	PeerVnetName string
	PeerVnetID   string
}

// AwsVpcInfo type
type AwsVpcInfo struct {
	SecurityGroup []string
	Cidr          string
	IgwID         string
	PeeringConnID []string
}

// AzureVnetInfo description
type AzureVnetInfo struct {
	Nsg           []string
	ResourceGroup string
	Cidr          string
	AvailSet      []string
	PeeringConnID []string
}

// VpcStatusCode describes vpc creation status
type VpcStatusCode int32

// Vpc status codes are used to communicate
// the state that vpc object is in to the client
const (
	// VpcStatusUnspecified - vpc status is unknown
	VpcStatusUnspecified VpcStatusCode = iota
	VpcStatusAddSuccess                // Vpc object was created successfully
	VpcStatusAddFailure                // Vpc object coult not be created
)

// Vpc object
type Vpc struct {
	Name           string
	VpcID          string `index:"true"`
	CPType         CloudProviderType
	Region         string
	ID             string
	RoleType       RoleType
	TopologyName   string
	ClosName       string
	WanName        string
	AwsVpcInfo     AwsVpcInfo
	AzVnetInfo     AzureVnetInfo
	Cnps           map[string]bool
	RouteReflector bool
	Tags           map[string]string
	PeerVpcCidr    map[string]string // maps peer vpc ID -> cidr
	PeerVpcInfo    PeerVpcInfo
	StatusCode     VpcStatusCode
	Account        string
	ModelVersion   string
}

// GetVpcCidr -
func (vpc *Vpc) GetVpcCidr() (string, error) {
	switch vpc.CPType {
	case Aws:
		return vpc.AwsVpcInfo.Cidr, nil
	case Azure:
		return vpc.AzVnetInfo.Cidr, nil
	default:
		glog.Errorf("Unsupported CPType %v for Vpc %s", vpc.CPType, vpc.Name)
		return "", fmt.Errorf("Unsupported CPType")
	}
}

// Clone - Deep clone a PeerVpcInfo object
func (pvi *PeerVpcInfo) Clone() *PeerVpcInfo {
	p := &PeerVpcInfo{
		PeerRgName:   pvi.PeerRgName,
		PeerVnetName: pvi.PeerVnetName,
		PeerVnetID:   pvi.PeerVnetID,
	}
	p.PeerVpcCidr = make(map[string]string)
	for key, value := range pvi.PeerVpcCidr {
		p.PeerVpcCidr[key] = value
	}
	return p
}

// Clone - Deep clone a AzureVnetInfo object
func (avi *AzureVnetInfo) Clone() *AzureVnetInfo {
	a := &AzureVnetInfo{
		ResourceGroup: avi.ResourceGroup,
		Cidr:          avi.Cidr,
	}
	a.Nsg = make([]string, len(avi.Nsg))
	copy(a.Nsg, avi.Nsg)
	a.AvailSet = make([]string, len(avi.AvailSet))
	copy(a.AvailSet, avi.AvailSet)
	a.PeeringConnID = make([]string, len(avi.PeeringConnID))
	copy(a.PeeringConnID, avi.PeeringConnID)
	return a
}

// Clone - Deep clone a AwsVpcInfo object
func (avi *AwsVpcInfo) Clone() *AwsVpcInfo {
	a := &AwsVpcInfo{
		Cidr:  avi.Cidr,
		IgwID: avi.IgwID,
	}
	a.SecurityGroup = make([]string, len(avi.SecurityGroup))
	copy(a.SecurityGroup, avi.SecurityGroup)
	a.PeeringConnID = make([]string, len(avi.PeeringConnID))
	copy(a.PeeringConnID, avi.PeeringConnID)
	return a
}

// Clone - Deep clone a Vpc object
func (vpc *Vpc) Clone() *Vpc {
	v := &Vpc{
		Name:           vpc.Name,
		VpcID:          vpc.VpcID,
		CPType:         vpc.CPType,
		Region:         vpc.Region,
		ID:             vpc.ID,
		RoleType:       vpc.RoleType,
		TopologyName:   vpc.TopologyName,
		ClosName:       vpc.ClosName,
		WanName:        vpc.WanName,
		AwsVpcInfo:     *vpc.AwsVpcInfo.Clone(),
		AzVnetInfo:     *vpc.AzVnetInfo.Clone(),
		RouteReflector: vpc.RouteReflector,
		PeerVpcInfo:    *vpc.PeerVpcInfo.Clone(),
		StatusCode:     vpc.StatusCode,
		Account:        vpc.Account,
		Cnps:           map[string]bool{},
		ModelVersion:   vpc.ModelVersion,
	}
	v.Tags = make(map[string]string)
	for key, value := range vpc.Tags {
		v.Tags[key] = value
	}
	v.PeerVpcCidr = make(map[string]string)
	for key, value := range vpc.PeerVpcCidr {
		v.PeerVpcCidr[key] = value
	}
	v.Cnps = cloneCnps(vpc.Cnps)
	return v
}

// Historical models

// VpcV0 object
type VpcV0 struct {
	Name           string
	VpcID          string `index:"true"`
	CPType         CloudProviderType
	Region         string
	ID             string
	RoleType       RoleType
	TopologyName   string
	ClosName       string
	WanName        string
	AwsVpcInfo     AwsVpcInfo
	AzVnetInfo     AzureVnetInfo
	Cnps           map[string]bool
	RouteReflector bool
	Tags           map[string]string
	PeerVpcCidr    map[string]string // maps peer vpc ID -> cidr
	PeerVpcInfo    PeerVpcInfo
	StatusCode     VpcStatusCode
	Account        string
}
