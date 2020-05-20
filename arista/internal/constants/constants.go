// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package constants

const (
	// LogCritical is reserverd for critical messages
	LogCritical = 0
	// LogEntry is used for priting logs on function entry
	LogEntry = 2
	// LogDetail is used for printing additional information
	LogDetail = 2
	// LogStats : To be used by stats collector
	LogStats = 8
	// VpcPrefix -
	VpcPrefix = "vpc"
	// RtrPrefix -
	RtrPrefix = "rtr"
	// SubnetPrefix -
	SubnetPrefix = "snet"
	// TopoMetaInfoPrefix -
	TopoMetaInfoPrefix = "topo"
	// TopoClosInfoPrefix -
	TopoClosInfoPrefix = "clos"
	// TopoWanInfoPrefix -
	TopoWanInfoPrefix = "wan"
	// ArPrefix -
	ArPrefix = "ar"
	//ClientVersion -
	ClientVersion = "0.0.2"
)

// Constants for secondary index fields in Router service aeris model
const (
	DeviceSerialNum = "DeviceSerialNum"
)

// Constants for secondary index fields in Vpc service aeris model
const (
	VpcID = "VpcID"
)

// Constants for secondary index fields in TopoInfo service aeris model
const (
	TopoName = "Name"
	TopoType = "TopoType"
)
