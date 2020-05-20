// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package models

// ModelVersion - version number of the current model
const (
	ModelVersion = "V0"
)

func cloneCnps(src map[string]bool) map[string]bool {
	dst := make(map[string]bool)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneBgpPeerInfoMap(src map[string]BgpPeerInfo) map[string]BgpPeerInfo {
	dst := make(map[string]BgpPeerInfo)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneHAPeerInfoMap(src map[string]HAPeerInfo) map[string]HAPeerInfo {
	dst := make(map[string]HAPeerInfo, len(src))
	for k, v := range src {
		dst[k] = *v.Clone()
	}
	return dst
}

// Clone - Deep clone a HAPeerInfo object
func (hpi *HAPeerInfo) Clone() *HAPeerInfo {
	p := &HAPeerInfo{
		RtrIP:         hpi.RtrIP,
		PrivateSubnet: hpi.PrivateSubnet,
	}
	p.PrivateRouteTableIDs = make([]string, len(hpi.PrivateRouteTableIDs))
	copy(p.PrivateRouteTableIDs, hpi.PrivateRouteTableIDs)
	return p
}

func cloneDpsPeerInfoMap(src map[string]DpsPeerInfo) map[string]DpsPeerInfo {
	dst := make(map[string]DpsPeerInfo, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
