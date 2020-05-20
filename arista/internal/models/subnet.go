// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package models

// Subnet object
type Subnet struct {
	SubnetID     string
	CPType       CloudProviderType
	ID           string
	CidrBlock    string
	VpcID        string
	AvailZone    string
	Zone         string
	PrimaryGW    string // Status
	SecondaryGW  string // Status
	ModelVersion string
}

// SubnetV0 object
type SubnetV0 struct {
	SubnetID    string
	CPType      CloudProviderType
	ID          string
	CidrBlock   string
	VpcID       string
	AvailZone   string
	Zone        string
	PrimaryGW   string // Status
	SecondaryGW string // Status
}
