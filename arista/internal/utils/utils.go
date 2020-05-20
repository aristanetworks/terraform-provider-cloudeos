// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package hooks

import (
	"fmt"
	"sort"
	"strings"

	cdc "cloudeos-terraform-provider/arista/internal/constants"
	cdm "cloudeos-terraform-provider/arista/internal/models"

	api "cloudeos-terraform-provider/arista/internal/api/clouddeploy.v1"

	"github.com/aristanetworks/glog"
	"github.com/aristanetworks/goarista/key"
	"github.com/google/uuid"
)

// GenerateID -
func GenerateID(prefix string) string {
	return fmt.Sprintf("%s-%s-%s", cdc.ArPrefix, prefix, uuid.New())
}

// GetTopologyInfoTypeFromID -
func GetTopologyInfoTypeFromID(tID string) cdm.TopologyInfoType {
	switch {
	case strings.HasPrefix(tID, fmt.Sprintf("%s-%s-", cdc.ArPrefix, cdc.TopoMetaInfoPrefix)):
		return cdm.TopoInfoMeta
	case strings.HasPrefix(tID, fmt.Sprintf("%s-%s-", cdc.ArPrefix, cdc.TopoClosInfoPrefix)):
		return cdm.TopoInfoClos
	case strings.HasPrefix(tID, fmt.Sprintf("%s-%s-", cdc.ArPrefix, cdc.TopoWanInfoPrefix)):
		return cdm.TopoInfoWan
	default:
		return cdm.TopoInfoTypeUnspecified
	}
}

// GetRouterKey -
func GetRouterKey(rtr *api.Router) key.Key {
	return key.New(rtr.GetId())
}

// ToAerisRouteTableIds -
func ToAerisRouteTableIds(pbrttids *api.RouteTableIds) cdm.RouteTableIDs {
	return cdm.RouteTableIDs{
		Public:   pbrttids.GetPublic(),
		Private:  pbrttids.GetPrivate(),
		Internal: pbrttids.GetInternal(),
	}
}

// ToAerisAzureRouterDetail -
func ToAerisAzureRouterDetail(pbrtr *api.Router) cdm.AzureRouterDetail {
	if pbrtr.GetCpT() != api.CloudProviderType_CP_AZURE {
		return cdm.AzureRouterDetail{}
	}
	pbAzRtr := pbrtr.GetAzRtrDetail()
	return cdm.AzureRouterDetail{
		AvailabilityZone: pbAzRtr.GetAvailZone(),
		ResourceGroup:    pbAzRtr.GetResGroup(),
		InstanceType:     pbAzRtr.GetInstanceType(),
		AvailabilitySet:  pbAzRtr.GetAvailSet(),
	}
}

// ToAerisAwsRouterDetail -
func ToAerisAwsRouterDetail(pbrtr *api.Router) cdm.AwsRouterDetail {
	if pbrtr.GetCpT() != api.CloudProviderType_CP_AWS {
		return cdm.AwsRouterDetail{}
	}
	pbAwsRtr := pbrtr.GetAwsRtrDetail()
	return cdm.AwsRouterDetail{
		AvailabilityZone: pbAwsRtr.GetAvailZone(),
		InstanceType:     pbAwsRtr.GetInstanceType(),
	}
}

// ToAerisNetworkInterface -
func ToAerisNetworkInterface(pbintf *api.NetworkInterface) cdm.NetworkInterface {
	return cdm.NetworkInterface{
		Name:          pbintf.GetName(),
		IntfID:        pbintf.GetIntfId(),
		IntfType:      cdm.NetworkInterfaceType(pbintf.GetIntfType()),
		PrivateIPAddr: pbintf.GetPrivateIpAddr(),
		PublicIPAddr:  pbintf.GetPublicIpAddr(),
		SubnetID:      pbintf.GetSubnet(),
		SecurityGroup: pbintf.GetSecurityGroup(),
	}
}

// ToAerisCvInfo -
func ToAerisCvInfo(pbcvi *api.CVInfo) cdm.CVInfo {
	// There should not be any requirement to read this from pb
	// since cloudVision service is supposed to write this
	// parameter
	return cdm.CVInfo{}
}

// ToAerisRouter -
func ToAerisRouter(pbrtr *api.Router) *cdm.Router {
	azRtrDetail := ToAerisAzureRouterDetail(pbrtr)
	awsRtrDetail := ToAerisAwsRouterDetail(pbrtr)
	cvInfo := ToAerisCvInfo(pbrtr.GetCvInfo())
	rtTableIds := ToAerisRouteTableIds(pbrtr.GetRtTableIds())

	nwIntfs := []cdm.NetworkInterface{}
	pbintfs := pbrtr.GetIntf()
	for _, pbintf := range pbintfs {
		nwIntfs = append(nwIntfs, ToAerisNetworkInterface(pbintf))
	}

	artr := cdm.Router{
		Name:                  pbrtr.GetName(),
		VpcID:                 pbrtr.GetVpcId(),
		CPType:                cdm.CloudProviderType(pbrtr.GetCpT()),
		Region:                pbrtr.GetRegion(),
		InstanceID:            pbrtr.GetInstanceId(),
		HAName:                pbrtr.GetHaName(),
		ID:                    pbrtr.GetId(),
		Tags:                  pbrtr.GetTags(),
		DeviceEnrollmentToken: pbrtr.GetDeviceEnrollmentToken(),
		RouteTableIDs:         rtTableIds,
		RouteReflector:        pbrtr.GetRouteReflector(),
		AwsRtrDetail:          awsRtrDetail,
		AzRtrDetail:           azRtrDetail,
		Intf:                  nwIntfs,
		DepStatus:             cdm.DeploymentStatusCode(pbrtr.GetDepStatus()),
		CVInfo:                cvInfo,
		ModelVersion:          cdm.ModelVersion,
	}

	return &artr
}

// ToResourceRouteTableIDs -
func ToResourceRouteTableIDs(artid *cdm.RouteTableIDs) *api.RouteTableIds {
	prtid, err := api.NewRouteTableIdsFilter(
		api.FilterRouteTableIdsPublic(artid.Public),
		api.FilterRouteTableIdsPrivate(artid.Private),
		api.FilterRouteTableIdsInternal(artid.Internal),
	)

	if err != nil {
		glog.Errorf("error converting route table ids from aeris to resource:%v", err)
		return nil
	}
	return prtid
}

// ToResourceCvInfo -
func ToResourceCvInfo(acvi *cdm.CVInfo) *api.CVInfo {
	prtid := ToResourceRouteTableIDs(&acvi.HaRouteTableIDs)
	pcvi, err := api.NewCVInfoFilter(
		api.FilterCVInfoCvStatusCode(api.CVStatusCode(acvi.StatusCode)),
		api.FilterCVInfoCvStatusDesc(acvi.StatusDesc),
		api.FilterCVInfoCvStatusRecommendedAction(acvi.StatusRecommendedAction),
		api.FilterCVInfoDeviceStatus(api.DeviceStatusCode(acvi.DeviceStatus)),
		api.FilterCVInfoBootstrapCfg(acvi.BootstrapConfig),
		api.FilterCVInfoHaRtrId(acvi.HaRtrID),
		api.FilterCVInfoPeerVpcRtTableId(acvi.PeerVpcRouteTableID),
		api.FilterCVInfoHaRtTableIds(prtid),
	)
	if err != nil {
		glog.Errorf("error converting CvInfo from aeris to resource:%v", err)
		return nil
	}
	return pcvi
}

// ToResourceNetworkIntf -
func ToResourceNetworkIntf(anwi *cdm.NetworkInterface) *api.NetworkInterface {
	pnwi, err := api.NewNetworkInterfaceFilter(
		api.FilterNetworkInterfaceIntfId(anwi.IntfID),
		api.FilterNetworkInterfaceName(anwi.Name),
		api.FilterNetworkInterfaceIntfType(api.NetworkInterfaceType(anwi.IntfType)),
		api.FilterNetworkInterfacePrivateIpAddr(anwi.PrivateIPAddr),
		api.FilterNetworkInterfacePublicIpAddr(anwi.PublicIPAddr),
		api.FilterNetworkInterfaceSubnet(anwi.SubnetID),
		api.FilterNetworkInterfaceSecurityGroup(anwi.SecurityGroup),
	)
	if err != nil {
		glog.Errorf("error converting NwInfo from aeris to resource:%v", err)
		return nil
	}
	return pnwi
}

// ToResourceAzureRouterDetail -
func ToResourceAzureRouterDetail(artr *cdm.Router) *api.AzureRouterDetail {
	if artr.CPType != cdm.Azure {
		return nil
	}
	rtrDetail := artr.AzRtrDetail
	azRtr, err := api.NewAzureRouterDetailFilter(
		api.FilterAzureRouterDetailAvailZone(rtrDetail.AvailabilityZone),
		api.FilterAzureRouterDetailResGroup(rtrDetail.ResourceGroup),
		api.FilterAzureRouterDetailAvailSet(rtrDetail.AvailabilitySet),
		api.FilterAzureRouterDetailInstanceType(rtrDetail.InstanceType),
	)
	if err != nil {
		glog.Errorf("error converting AzureRouterDetail from aeris to resource")
		return nil
	}
	return azRtr
}

// ToResourceAwsRouterDetail -
func ToResourceAwsRouterDetail(artr *cdm.Router) *api.AwsRouterDetail {
	if artr.CPType != cdm.Aws {
		return nil
	}
	rtrDetail := artr.AwsRtrDetail
	awsRtr, err := api.NewAwsRouterDetailFilter(
		api.FilterAwsRouterDetailAvailZone(rtrDetail.AvailabilityZone),
		api.FilterAwsRouterDetailInstanceType(rtrDetail.InstanceType),
	)
	if err != nil {
		glog.Errorf("error converting AwsRouterDetail from aeris to resource:%v", err)
		return nil
	}
	return awsRtr
}

// ToResourceRouter -
func ToResourceRouter(artr *cdm.Router) *api.Router {
	cvInfo := ToResourceCvInfo(&artr.CVInfo)
	rtTableIds := ToResourceRouteTableIDs(&artr.RouteTableIDs)
	nwIntfs := []*api.NetworkInterface{}
	for _, intf := range artr.Intf {
		if pnwi := ToResourceNetworkIntf(&intf); pnwi != nil {
			nwIntfs = append(nwIntfs, pnwi)
		}
	}
	var err error
	var pbrtr *api.Router

	filters := []api.RouterFieldFilter{
		api.FilterRouterName(artr.Name),
		api.FilterRouterVpcId(artr.VpcID),
		api.FilterRouterCpT(api.CloudProviderType(artr.CPType)),
		api.FilterRouterRegion(artr.Region),
		api.FilterRouterInstanceId(artr.InstanceID),
		api.FilterRouterHaName(artr.HAName),
		api.FilterRouterId(artr.ID),
		api.FilterRouterCnps(getCnpsStringFromMap(artr.Cnps)),
		api.FilterRouterTags(artr.Tags),
		api.FilterRouterDeviceEnrollmentToken(artr.DeviceEnrollmentToken),
		api.FilterRouterRtTableIds(rtTableIds),
		api.FilterRouterRouteReflector(artr.RouteReflector),
		api.FilterRouterIntf(nwIntfs),
		api.FilterRouterDepStatus(api.DeploymentStatusCode(artr.DepStatus)),
		api.FilterRouterCvInfo(cvInfo),
		api.FilterRouterDeviceSerialNum(artr.DeviceSerialNum),
	}

	switch artr.CPType {
	case cdm.Aws:
		filters = append(filters, api.FilterRouterAwsRtrDetail(ToResourceAwsRouterDetail(artr)))
	case cdm.Azure:
		filters = append(filters, api.FilterRouterAzRtrDetail(ToResourceAzureRouterDetail(artr)))
	default:
		glog.Errorf("Could not obtain Router resource. Unknown CP_Type:%v", artr.CPType)
		return nil
	}
	pbrtr, err = api.NewRouterFilter(filters...)
	if err != nil {
		glog.Errorf("Could not obtain Router resource. %v", err)
		return nil
	}
	return pbrtr
}

// ToResourceRouterClient -
func ToResourceRouterClient(artr *cdm.Router) *api.Router {
	rtTableIds := ToResourceRouteTableIDs(&artr.RouteTableIDs)
	nwIntfs := []*api.NetworkInterface{}
	for _, intf := range artr.Intf {
		if pnwi := ToResourceNetworkIntf(&intf); pnwi != nil {
			nwIntfs = append(nwIntfs, pnwi)
		}
	}
	var err error
	var pbrtr *api.Router

	filters := []api.RouterFieldFilter{
		api.FilterRouterName(artr.Name),
		api.FilterRouterVpcId(artr.VpcID),
		api.FilterRouterCpT(api.CloudProviderType(artr.CPType)),
		api.FilterRouterRegion(artr.Region),
		api.FilterRouterInstanceId(artr.InstanceID),
		api.FilterRouterHaName(artr.HAName),
		api.FilterRouterId(artr.ID),
		api.FilterRouterCnps(getCnpsStringFromMap(artr.Cnps)),
		api.FilterRouterTags(artr.Tags),
		api.FilterRouterDeviceEnrollmentToken(artr.DeviceEnrollmentToken),
		api.FilterRouterRtTableIds(rtTableIds),
		api.FilterRouterRouteReflector(artr.RouteReflector),
		api.FilterRouterIntf(nwIntfs),
		api.FilterRouterDepStatus(api.DeploymentStatusCode(artr.DepStatus)),
	}

	switch artr.CPType {
	case cdm.Aws:
		filters = append(filters, api.FilterRouterAwsRtrDetail(ToResourceAwsRouterDetail(artr)))
	case cdm.Azure:
		filters = append(filters, api.FilterRouterAzRtrDetail(ToResourceAzureRouterDetail(artr)))
	default:
		glog.Errorf("Could not obtain Router resource. Unknown CP_Type:%v", artr.CPType)
		return nil
	}
	pbrtr, err = api.NewRouterFilter(filters...)
	if err != nil {
		glog.Errorf("Could not obtain Router resource. %v", err)
		return nil
	}
	return pbrtr
}

// ToResourceGetRouter -
func ToResourceGetRouter(artr *cdm.Router) *api.Router {
	var err error
	var pbrtr *api.Router

	filters := []api.RouterFieldFilter{
		api.FilterRouterId(artr.ID),
	}

	pbrtr, err = api.NewRouterFilter(filters...)
	if err != nil {
		glog.Errorf("Could not obtain Router resource. %v", err)
		return nil
	}
	return pbrtr
}

// ToResourceListRouter -
func ToResourceListRouter(artr *cdm.Router) *api.Router {
	var err error
	var pbrtr *api.Router

	filters := []api.RouterFieldFilter{
		api.FilterRouterName(artr.Name),
		api.FilterRouterId(artr.ID),
		api.FilterRouterVpcId(artr.VpcID),
		api.FilterRouterCpT(api.CloudProviderType(artr.CPType)),
		api.FilterRouterRegion(artr.Region),
	}

	pbrtr, err = api.NewRouterFilter(filters...)
	if err != nil {
		glog.Errorf("Could not obtain Router resource. %v", err)
		return nil
	}
	return pbrtr
}

// ToResourceListEdgeRouter -
func ToResourceListEdgeRouter(artr *cdm.Router) *api.Router {
	var err error
	var pbrtr *api.Router

	filters := []api.RouterFieldFilter{
		api.FilterRouterVpcId(artr.VpcID),
		api.FilterRouterCpT(api.CloudProviderType(artr.CPType)),
		api.FilterRouterRegion(artr.Region),
		api.FilterRouterRouteReflector(artr.RouteReflector),
	}

	pbrtr, err = api.NewRouterFilter(filters...)
	if err != nil {
		glog.Errorf("Could not obtain Router resource. %v", err)
		return nil
	}
	return pbrtr
}

// GetVpcKey -
func GetVpcKey(vpc *api.Vpc) key.Key {
	return key.New(vpc.GetId())
}

// ToAerisAzureVnetInfo -
func ToAerisAzureVnetInfo(pbvpc *api.Vpc) cdm.AzureVnetInfo {
	if pbvpc.GetCpT() != api.CloudProviderType_CP_AZURE {
		return cdm.AzureVnetInfo{}
	}

	pbAzureVnetInfo := pbvpc.GetAzVnetInfo()
	return cdm.AzureVnetInfo{
		Nsg:           pbAzureVnetInfo.GetNsg(),
		ResourceGroup: pbAzureVnetInfo.GetResourceGroup(),
		Cidr:          pbAzureVnetInfo.GetCidr(),
		AvailSet:      pbAzureVnetInfo.GetAvailSet(),
		PeeringConnID: pbAzureVnetInfo.GetPeeringConnId(),
	}
}

// ToAerisAwsVpcInfo -
func ToAerisAwsVpcInfo(pbvpc *api.Vpc) cdm.AwsVpcInfo {
	if pbvpc.GetCpT() != api.CloudProviderType_CP_AWS {
		return cdm.AwsVpcInfo{}
	}
	pbAwsVpcInfo := pbvpc.GetAwsVpcInfo()
	return cdm.AwsVpcInfo{
		SecurityGroup: pbAwsVpcInfo.GetSecurityGroup(),
		Cidr:          pbAwsVpcInfo.GetCidr(),
		IgwID:         pbAwsVpcInfo.GetIgwId(),
		PeeringConnID: pbAwsVpcInfo.GetPeeringConnId(),
	}
}

// ToAerisPeerVpcInfo -
func ToAerisPeerVpcInfo(pbvpc *api.Vpc) cdm.PeerVpcInfo {
	if pbvpc.GetCpT() != api.CloudProviderType_CP_AZURE {
		return cdm.PeerVpcInfo{}
	}
	pbPeerVpcInfo := pbvpc.GetPeerVpcInfo()
	return cdm.PeerVpcInfo{
		PeerVpcCidr:  pbPeerVpcInfo.GetPeerVpcCidr(),
		PeerRgName:   pbPeerVpcInfo.GetPeerRgName(),
		PeerVnetName: pbPeerVpcInfo.GetPeerVnetName(),
		PeerVnetID:   pbPeerVpcInfo.GetPeerVnetId(),
	}
}

// ToAerisVpc -
func ToAerisVpc(vpc *api.Vpc) *cdm.Vpc {
	awsVpcInfo := ToAerisAwsVpcInfo(vpc)
	azVnetInfo := ToAerisAzureVnetInfo(vpc)
	peerVpcInfo := ToAerisPeerVpcInfo(vpc)

	avpc := cdm.Vpc{
		Name:           vpc.GetName(),
		VpcID:          vpc.GetVpcId(),
		CPType:         cdm.CloudProviderType(vpc.GetCpT()),
		Region:         vpc.GetRegion(),
		ID:             vpc.GetId(),
		RoleType:       cdm.RoleType(vpc.GetRoleType()),
		TopologyName:   vpc.GetTopologyName(),
		ClosName:       vpc.GetClosName(),
		WanName:        vpc.GetWanName(),
		AwsVpcInfo:     awsVpcInfo,
		AzVnetInfo:     azVnetInfo,
		Cnps:           getCnpsMapFromString(vpc.GetCnps()),
		RouteReflector: vpc.GetRouteReflector(),
		Tags:           vpc.GetTags(),
		PeerVpcCidr:    vpc.GetPeerVpcCidr(),
		Account:        vpc.GetAccount(),
		PeerVpcInfo:    peerVpcInfo,
		ModelVersion:   cdm.ModelVersion,
	}
	return &avpc
}

// ToResourceListVpc -
func ToResourceListVpc(avpc *cdm.Vpc) *api.Vpc {
	var pbvpc *api.Vpc
	var errr error

	filters := []api.VpcFieldFilter{
		api.FilterVpcName(avpc.Name),
		api.FilterVpcCpT(api.CloudProviderType(avpc.CPType)),
		api.FilterVpcRegion(avpc.Region),
	}
	pbvpc, errr = api.NewVpcFilter(filters...)
	if errr != nil {
		glog.Errorf("Could not obtain Vpc resource. %v", errr)
		return nil
	}
	return pbvpc
}

// ToResourceGetVpc -
func ToResourceGetVpc(avpc *cdm.Vpc) *api.Vpc {
	var pbvpc *api.Vpc
	var errr error

	filters := []api.VpcFieldFilter{
		api.FilterVpcId(avpc.ID),
		//api.FilterVpcVpcId(avpc.VpcID),
	}
	pbvpc, errr = api.NewVpcFilter(filters...)
	if errr != nil {
		glog.Errorf("Could not obtain Vpc resource. %v", errr)
		return nil
	}
	return pbvpc
}

// ToResourceCheckVpc -
func ToResourceCheckVpc(avpc *cdm.Vpc) *api.Vpc {
	var pbvpc *api.Vpc
	var errr error

	filters := []api.VpcFieldFilter{
		api.FilterVpcCpT(api.CloudProviderType(avpc.CPType)),
		api.FilterVpcRegion(avpc.Region),
		api.FilterVpcVpcId(avpc.VpcID),
	}
	pbvpc, errr = api.NewVpcFilter(filters...)
	if errr != nil {
		glog.Errorf("Could not obtain Vpc resource. %v", errr)
		return nil
	}
	return pbvpc
}

// ToResourceListEdgeVpc -
func ToResourceListEdgeVpc(avpc *cdm.Vpc) *api.Vpc {
	var pbvpc *api.Vpc
	var errr error

	filters := []api.VpcFieldFilter{
		api.FilterVpcCpT(api.CloudProviderType(avpc.CPType)),
		api.FilterVpcRegion(avpc.Region),
		api.FilterVpcTopologyName(avpc.TopologyName),
	}
	pbvpc, errr = api.NewVpcFilter(filters...)
	if errr != nil {
		glog.Errorf("Could not obtain Vpc resource. %v", errr)
		return nil
	}
	return pbvpc
}

// ToResourceVpc -
func ToResourceVpc(avpc *cdm.Vpc) *api.Vpc {
	var pbvpc *api.Vpc
	var errr error

	filters := []api.VpcFieldFilter{
		api.FilterVpcName(avpc.Name),
		api.FilterVpcVpcId(avpc.VpcID),
		api.FilterVpcCpT(api.CloudProviderType(avpc.CPType)),
		api.FilterVpcRegion(avpc.Region),
		api.FilterVpcId(avpc.ID),
		api.FilterVpcRoleType(api.RoleType(avpc.RoleType)),
		api.FilterVpcTopologyName(avpc.TopologyName),
		api.FilterVpcClosName(avpc.ClosName),
		api.FilterVpcWanName(avpc.WanName),
		api.FilterVpcCnps(getCnpsStringFromMap(avpc.Cnps)),
		api.FilterVpcRouteReflector(avpc.RouteReflector),
		api.FilterVpcTags(avpc.Tags),
		api.FilterVpcPeerVpcCidr(avpc.PeerVpcCidr),
		api.FilterVpcStatusCode(api.VpcStatusCode(avpc.StatusCode)),
		api.FilterVpcAccount(avpc.Account),
	}
	peerVpcInfo := avpc.PeerVpcInfo
	pbPeerVpcInfo, err := api.NewPeerVpcInfoFilter(
		api.FilterPeerVpcInfoPeerVpcCidr(peerVpcInfo.PeerVpcCidr),
		api.FilterPeerVpcInfoPeerRgName(peerVpcInfo.PeerRgName),
		api.FilterPeerVpcInfoPeerVnetName(peerVpcInfo.PeerVnetName),
		api.FilterPeerVpcInfoPeerVnetId(peerVpcInfo.PeerVnetID),
	)
	if err != nil {
		glog.Errorf("error converting PeerVpcInfo from aeris to resource:%v", err)
	}
	filters = append(filters, api.FilterVpcPeerVpcInfo(pbPeerVpcInfo))

	switch avpc.CPType {
	case cdm.Aws:
		{
			aCloudVpcInfo := avpc.AwsVpcInfo
			pbAwsVpcInfo, err := api.NewAwsVpcInfoFilter(
				api.FilterAwsVpcInfoSecurityGroup(aCloudVpcInfo.SecurityGroup),
				api.FilterAwsVpcInfoCidr(aCloudVpcInfo.Cidr),
				api.FilterAwsVpcInfoIgwId(aCloudVpcInfo.IgwID),
				api.FilterAwsVpcInfoPeeringConnId(aCloudVpcInfo.PeeringConnID),
			)
			if err != nil {
				glog.Errorf("error converting cloudVpcInfo from aeris to resource:%v", err)
				return nil
			}
			filters = append(filters, api.FilterVpcAwsVpcInfo(pbAwsVpcInfo))
		}
	case cdm.Azure:
		{
			aCloudVnetInfo := avpc.AzVnetInfo
			pbAzureVnetInfo, err := api.NewAzureVnetInfoFilter(
				api.FilterAzureVnetInfoNsg(aCloudVnetInfo.Nsg),
				api.FilterAzureVnetInfoResourceGroup(aCloudVnetInfo.ResourceGroup),
				api.FilterAzureVnetInfoCidr(aCloudVnetInfo.Cidr),
				api.FilterAzureVnetInfoAvailSet(aCloudVnetInfo.AvailSet),
				api.FilterAzureVnetInfoPeeringConnId(aCloudVnetInfo.PeeringConnID),
			)
			if err != nil {
				glog.Errorf("error converting cloudVnetInfo from aeris to resource:%v", err)
				return nil
			}
			filters = append(filters, api.FilterVpcAzVnetInfo(pbAzureVnetInfo))
		}
	default:
		{
			glog.Errorf("unknown cloud provider type: %v", avpc.CPType)
			return nil
		}
	}

	pbvpc, errr = api.NewVpcFilter(filters...)
	if errr != nil {
		glog.Errorf("Could not obtain Vpc resource. %v", errr)
		return nil
	}
	return pbvpc
}

// ToResourceVpcClient -
func ToResourceVpcClient(avpc *cdm.Vpc) *api.Vpc {
	var pbvpc *api.Vpc
	var errr error

	filters := []api.VpcFieldFilter{
		api.FilterVpcName(avpc.Name),
		api.FilterVpcVpcId(avpc.VpcID),
		api.FilterVpcCpT(api.CloudProviderType(avpc.CPType)),
		api.FilterVpcRegion(avpc.Region),
		api.FilterVpcId(avpc.ID),
		api.FilterVpcRoleType(api.RoleType(avpc.RoleType)),
		api.FilterVpcTopologyName(avpc.TopologyName),
		api.FilterVpcClosName(avpc.ClosName),
		api.FilterVpcWanName(avpc.WanName),
		api.FilterVpcCnps(getCnpsStringFromMap(avpc.Cnps)),
		api.FilterVpcRouteReflector(avpc.RouteReflector),
		api.FilterVpcTags(avpc.Tags),
		api.FilterVpcAccount(avpc.Account),
	}

	switch avpc.CPType {
	case cdm.Aws:
		{
			aCloudVpcInfo := avpc.AwsVpcInfo
			pbAwsVpcInfo, err := api.NewAwsVpcInfoFilter(
				api.FilterAwsVpcInfoSecurityGroup(aCloudVpcInfo.SecurityGroup),
				api.FilterAwsVpcInfoCidr(aCloudVpcInfo.Cidr),
				api.FilterAwsVpcInfoIgwId(aCloudVpcInfo.IgwID),
				api.FilterAwsVpcInfoPeeringConnId(aCloudVpcInfo.PeeringConnID),
			)
			if err != nil {
				glog.Errorf("error converting cloudVpcInfo from aeris to resource:%v", err)
				return nil
			}
			filters = append(filters, api.FilterVpcAwsVpcInfo(pbAwsVpcInfo))
		}
	case cdm.Azure:
		{
			aCloudVnetInfo := avpc.AzVnetInfo
			pbAzureVnetInfo, err := api.NewAzureVnetInfoFilter(
				api.FilterAzureVnetInfoNsg(aCloudVnetInfo.Nsg),
				api.FilterAzureVnetInfoResourceGroup(aCloudVnetInfo.ResourceGroup),
				api.FilterAzureVnetInfoCidr(aCloudVnetInfo.Cidr),
				api.FilterAzureVnetInfoAvailSet(aCloudVnetInfo.AvailSet),
				api.FilterAzureVnetInfoPeeringConnId(aCloudVnetInfo.PeeringConnID),
			)
			if err != nil {
				glog.Errorf("error converting cloudVnetInfo from aeris to resource:%v", err)
				return nil
			}
			filters = append(filters, api.FilterVpcAzVnetInfo(pbAzureVnetInfo))
		}
	default:
		{
			glog.Errorf("unknown cloud provider type: %v", avpc.CPType)
			return nil
		}
	}

	pbvpc, errr = api.NewVpcFilter(filters...)
	if errr != nil {
		glog.Errorf("Could not obtain Vpc resource. %v", errr)
		return nil
	}
	return pbvpc
}

// ToResourceLeafVpc -
func ToResourceLeafVpc(avpc *cdm.Vpc) *api.Vpc {
	var pbvpc *api.Vpc
	var errr error

	filters := []api.VpcFieldFilter{
		api.FilterVpcName(avpc.Name),
		api.FilterVpcVpcId(avpc.VpcID),
		api.FilterVpcCpT(api.CloudProviderType(avpc.CPType)),
		api.FilterVpcRegion(avpc.Region),
		api.FilterVpcId(avpc.ID),
		api.FilterVpcRoleType(api.RoleType(avpc.RoleType)),
		api.FilterVpcTopologyName(avpc.TopologyName),
		api.FilterVpcClosName(avpc.ClosName),
		api.FilterVpcCnps(getCnpsStringFromMap(avpc.Cnps)),
		api.FilterVpcRouteReflector(avpc.RouteReflector),
		api.FilterVpcTags(avpc.Tags),
		api.FilterVpcAccount(avpc.Account),
	}

	switch avpc.CPType {
	case cdm.Aws:
		{
			aCloudVpcInfo := avpc.AwsVpcInfo
			pbAwsVpcInfo, err := api.NewAwsVpcInfoFilter(
				api.FilterAwsVpcInfoSecurityGroup(aCloudVpcInfo.SecurityGroup),
				api.FilterAwsVpcInfoCidr(aCloudVpcInfo.Cidr),
				api.FilterAwsVpcInfoIgwId(aCloudVpcInfo.IgwID),
				api.FilterAwsVpcInfoPeeringConnId(aCloudVpcInfo.PeeringConnID),
			)
			if err != nil {
				glog.Errorf("error converting cloudVpcInfo from aeris to resource:%v", err)
				return nil
			}
			filters = append(filters, api.FilterVpcAwsVpcInfo(pbAwsVpcInfo))
		}
	case cdm.Azure:
		{
			aCloudVnetInfo := avpc.AzVnetInfo
			pbAzureVnetInfo, err := api.NewAzureVnetInfoFilter(
				api.FilterAzureVnetInfoNsg(aCloudVnetInfo.Nsg),
				api.FilterAzureVnetInfoResourceGroup(aCloudVnetInfo.ResourceGroup),
				api.FilterAzureVnetInfoCidr(aCloudVnetInfo.Cidr),
				api.FilterAzureVnetInfoAvailSet(aCloudVnetInfo.AvailSet),
				api.FilterAzureVnetInfoPeeringConnId(aCloudVnetInfo.PeeringConnID),
			)
			if err != nil {
				glog.Errorf("error converting cloudVnetInfo from aeris to resource:%v", err)
				return nil
			}
			filters = append(filters, api.FilterVpcAzVnetInfo(pbAzureVnetInfo))
		}
	default:
		{
			glog.Errorf("unknown cloud provider type: %v", avpc.CPType)
			return nil
		}
	}

	pbvpc, errr = api.NewVpcFilter(filters...)
	if errr != nil {
		glog.Errorf("Could not obtain Vpc resource. %v", errr)
		return nil
	}
	return pbvpc
}

// GetSubnetKey -
func GetSubnetKey(s *api.Subnet) key.Key {
	return key.New(s.GetId())
}

// ToAerisSubnet -
func ToAerisSubnet(pbs *api.Subnet) *cdm.Subnet {
	as := cdm.Subnet{
		SubnetID:     pbs.GetSubnetId(),
		CPType:       cdm.CloudProviderType(pbs.GetCpT()),
		ID:           pbs.GetId(),
		CidrBlock:    pbs.GetCidr(),
		VpcID:        pbs.GetVpcId(),
		Zone:         pbs.GetAvailZone(),
		PrimaryGW:    pbs.GetPrimGw(),
		SecondaryGW:  pbs.GetSecGw(),
		ModelVersion: cdm.ModelVersion,
	}
	return &as
}

// ToResourceSubnet -
func ToResourceSubnet(as *cdm.Subnet) *api.Subnet {
	pbs, err := api.NewSubnetFilter(
		api.FilterSubnetSubnetId(as.SubnetID),
		api.FilterSubnetCpT(api.CloudProviderType(as.CPType)),
		api.FilterSubnetId(as.ID),
		api.FilterSubnetCidr(as.CidrBlock),
		api.FilterSubnetVpcId(as.VpcID),
		api.FilterSubnetAvailZone(as.Zone),
		api.FilterSubnetPrimGw(as.PrimaryGW),
		api.FilterSubnetSecGw(as.SecondaryGW),
	)

	if err != nil {
		glog.Errorf("Could not obtain Subnet resource. %v", err)
		return nil
	}
	return pbs
}

// GetTopologyInfoKey -
func GetTopologyInfoKey(ti *api.TopologyInfo) key.Key {
	return key.New(ti.GetId())
}

// ToAerisWanInfo -
func ToAerisWanInfo(pbti *api.TopologyInfo) cdm.WanInfo {
	if pbti.GetTopoType() != api.TopologyInfoType_TOPO_INFO_WAN {
		return cdm.WanInfo{}
	}
	pbWanInfo := pbti.GetWanInfo()
	return cdm.WanInfo{
		WanName:              pbWanInfo.GetWanName(),
		CPType:               cdm.CloudProviderType(pbWanInfo.GetCpType()),
		PeerNames:            pbWanInfo.GetPeerNames(),
		EdgeEdgePeering:      pbWanInfo.GetEdgeEdgePeering(),
		EdgeEdgeIgw:          pbWanInfo.GetEdgeEdgeIgw(),
		EdgeDedicatedConnect: pbWanInfo.GetEdgeDedicatedConnect(),
		CvpContainerName:     pbWanInfo.GetCvpContainerName(),
	}
}

// ToAerisClosInfo -
func ToAerisClosInfo(pbti *api.TopologyInfo) cdm.ClosInfo {
	if pbti.GetTopoType() != api.TopologyInfoType_TOPO_INFO_CLOS {
		return cdm.ClosInfo{}
	}
	pbClosInfo := pbti.GetClosInfo()
	return cdm.ClosInfo{
		ClosName:         pbClosInfo.GetClosName(),
		CPType:           cdm.CloudProviderType(pbClosInfo.GetCpType()),
		Fabric:           cdm.FabricType(pbClosInfo.GetFabric()),
		LeafEdgePeering:  pbClosInfo.GetLeafEdgePeering(),
		LeafEdgeIgw:      pbClosInfo.GetLeafEdgeIgw(),
		LeafEncryption:   pbClosInfo.GetLeafEncryption(),
		CvpContainerName: pbClosInfo.GetCvpContainerName(),
	}
}

// ToAerisTopologyInfo -
func ToAerisTopologyInfo(pbti *api.TopologyInfo) *cdm.TopologyInfo {
	atopoInfo := cdm.TopologyInfo{
		Name:                pbti.GetName(),
		ID:                  pbti.GetId(),
		TopoType:            cdm.TopologyInfoType(pbti.GetTopoType()),
		BgpAsnLow:           pbti.GetBgpAsnLow(),
		BgpAsnHigh:          pbti.GetBgpAsnHigh(),
		VtepIPCidr:          pbti.GetVtepIpCidr(),
		TerminAttrIPCidr:    pbti.GetTerminattrIpCidr(),
		DpsControlPlaneCidr: pbti.GetDpsControlPlaneCidr(),
		ManagedDevices:      pbti.GetManagedDevices(),
		CVaaSDomain:         pbti.GetCvaasDomain(),
		CVaaSServer:         pbti.GetCvaasServer(),
		Wan:                 ToAerisWanInfo(pbti),
		Clos:                ToAerisClosInfo(pbti),
		Version:             pbti.GetVersion(),
		ModelVersion:        cdm.ModelVersion,
	}
	return &atopoInfo
}

// ToResourceWanInfo -
func ToResourceWanInfo(ati *cdm.TopologyInfo) *api.WanInfo {
	if ati.TopoType != cdm.TopoInfoWan {
		return nil
	}
	aWanInfo := ati.Wan
	pbWanInfo, err := api.NewWanInfoFilter(
		api.FilterWanInfoWanName(aWanInfo.WanName),
		api.FilterWanInfoCpType(api.CloudProviderType(aWanInfo.CPType)),
		api.FilterWanInfoPeerNames(aWanInfo.PeerNames),
		api.FilterWanInfoEdgeEdgePeering(aWanInfo.EdgeEdgePeering),
		api.FilterWanInfoEdgeEdgeIgw(aWanInfo.EdgeEdgeIgw),
		api.FilterWanInfoEdgeDedicatedConnect(aWanInfo.EdgeDedicatedConnect),
		api.FilterWanInfoCvpContainerName(aWanInfo.CvpContainerName),
	)
	if err != nil {
		glog.Errorf("error converting WanInfo from aeris to resource:%v", err)
		return nil
	}
	return pbWanInfo
}

// ToResourceClosInfo -
func ToResourceClosInfo(ati *cdm.TopologyInfo) *api.ClosInfo {
	if ati.TopoType != cdm.TopoInfoClos {
		return nil
	}
	aClosInfo := ati.Clos
	pbClosInfo, err := api.NewClosInfoFilter(
		api.FilterClosInfoClosName(aClosInfo.ClosName),
		api.FilterClosInfoCpType(api.CloudProviderType(aClosInfo.CPType)),
		api.FilterClosInfoFabric(api.FabricType(aClosInfo.Fabric)),
		api.FilterClosInfoLeafEdgePeering(aClosInfo.LeafEdgePeering),
		api.FilterClosInfoLeafEdgeIgw(aClosInfo.LeafEdgeIgw),
		api.FilterClosInfoLeafEncryption(aClosInfo.LeafEncryption),
		api.FilterClosInfoCvpContainerName(aClosInfo.CvpContainerName),
	)
	if err != nil {
		glog.Errorf("error converting ClosInfo from aeris to resource:%v", err)
		return nil
	}
	return pbClosInfo
}

// ToResourceTopologyInfo -
func ToResourceTopologyInfo(ati *cdm.TopologyInfo) *api.TopologyInfo {
	filters := []api.TopologyInfoFieldFilter{
		api.FilterTopologyInfoName(ati.Name),
		api.FilterTopologyInfoId(ati.ID),
		api.FilterTopologyInfoTopoType(api.TopologyInfoType(ati.TopoType)),
		api.FilterTopologyInfoBgpAsnLow(ati.BgpAsnLow),
		api.FilterTopologyInfoBgpAsnHigh(ati.BgpAsnHigh),
		api.FilterTopologyInfoVtepIpCidr(ati.VtepIPCidr),
		api.FilterTopologyInfoTerminattrIpCidr(ati.TerminAttrIPCidr),
		api.FilterTopologyInfoDpsControlPlaneCidr(ati.DpsControlPlaneCidr),
		api.FilterTopologyInfoManagedDevices(ati.ManagedDevices),
		api.FilterTopologyInfoCvaasDomain(ati.CVaaSDomain),
		api.FilterTopologyInfoCvaasServer(ati.CVaaSServer),
		api.FilterTopologyInfoVersion(ati.Version),
	}

	switch ati.TopoType {
	case cdm.TopoInfoWan:
		filters = append(filters, api.FilterTopologyInfoWanInfo(ToResourceWanInfo(ati)))
	case cdm.TopoInfoClos:
		filters = append(filters, api.FilterTopologyInfoClosInfo(ToResourceClosInfo(ati)))
	}
	pbtopoInfo, err := api.NewTopologyInfoFilter(filters...)
	if err != nil {
		glog.Errorf("Could not obtain TopologyInfo resource. %v", err)
		return nil
	}
	return pbtopoInfo
}

// ToResourceListTopologyInfo -
func ToResourceListTopologyInfo(ati *cdm.TopologyInfo) *api.TopologyInfo {
	filters := []api.TopologyInfoFieldFilter{
		api.FilterTopologyInfoName(ati.Name),
	}
	pbtopoInfo, err := api.NewTopologyInfoFilter(filters...)
	if err != nil {
		glog.Errorf("Could not obtain TopologyInfo resource. %v", err)
		return nil
	}
	return pbtopoInfo
}

// GetPathKey -
func GetPathKey(path *api.Path) key.Key {
	pk := cdm.PathKey{
		SrcVpcUUID:    path.GetSrcVpcUuid(),
		LocalRtrUUID:  path.GetLocalRtrUuid(),
		DstVpcUUID:    path.GetDstVpcUuid(),
		RemoteRtrUUID: path.GetRemoteRtrUuid(),
		UlConn:        cdm.UnderlayConnectionType(path.GetUlT()),
	}
	return key.New(pk)
}

// ToResourcePathCharacteristics -
func ToResourcePathCharacteristics(apc *cdm.PathCharacteristics) *api.PathCharacteristics {
	rpc, err := api.NewPathCharacteristicsFilter(
		api.FilterPathCharacteristicsBwMbps(apc.BwMbps),
		api.FilterPathCharacteristicsJitterMs(apc.JitterMs),
		api.FilterPathCharacteristicsLatencyMs(apc.LatencyMs),
		api.FilterPathCharacteristicsPktLossPc(apc.PktLossPc),
		api.FilterPathCharacteristicsUp(apc.Up),
		api.FilterPathCharacteristicsUptime(apc.Uptime),
	)
	if err != nil {
		glog.Errorf("Could not obtain resource PathCharacteristics")
		return nil
	}
	return rpc
}

// ToResourcePath -
func ToResourcePath(apath *cdm.Path) *api.Path {
	pathChar := ToResourcePathCharacteristics(&apath.PathChars)
	rpath, err := api.NewPathFilter(
		api.FilterPathSrcVpcCloudId(apath.SrcVpcCloudID),
		api.FilterPathSrcVpcName(apath.SrcVpcName),
		api.FilterPathSrcVpcUuid(apath.SrcVpcUUID),
		api.FilterPathLocalRtrCloudId(apath.LocalRtrCloudID),
		api.FilterPathLocalRtrName(apath.LocalRtrName),
		api.FilterPathLocalRtrUuid(apath.LocalRtrUUID),
		api.FilterPathLocalIntfIpAddr(apath.LocalIntfIPAddr),
		api.FilterPathSrcRegion(apath.SrcRegion),
		api.FilterPathSrcCpT(api.CloudProviderType(apath.SrcCpType)),
		api.FilterPathDstVpcCloudId(apath.DstVpcCloudID),
		api.FilterPathDstVpcName(apath.DstVpcName),
		api.FilterPathDstVpcUuid(apath.DstVpcUUID),
		api.FilterPathRemoteRtrCloudId(apath.RemoteRtrCloudID),
		api.FilterPathRemoteRtrName(apath.RemoteRtrName),
		api.FilterPathRemoteRtrUuid(apath.RemoteRtrUUID),
		api.FilterPathRemoteIntfIpAddr(apath.RemoteIntfIPAddr),
		api.FilterPathDstRegion(apath.DstRegion),
		api.FilterPathDstCpT(api.CloudProviderType(apath.DstCpType)),
		api.FilterPathTopologyName(apath.TopologyName),
		api.FilterPathUlT(api.UnderlayConnectionType(apath.UlConn)),
		api.FilterPathPathChar(pathChar),
	)

	if err != nil {
		glog.Errorf("Could not obtain Path resource: %v", err)
		return nil
	}
	return rpath
}

// ToResourceListPathByTopo -
func ToResourceListPathByTopo(apath *cdm.Path) *api.Path {
	rpath, err := api.NewPathFilter(
		api.FilterPathTopologyName(apath.TopologyName),
	)
	if err != nil {
		glog.Errorf("Could not obtain Path resource. %v", err)
		return nil
	}
	return rpath
}

func getCnpsStringFromMap(cnps map[string]bool) string {
	cnpsSlice := []string{}
	for k := range cnps {
		if k != "" {
			cnpsSlice = append(cnpsSlice, k)
		}
	}
	sort.Strings(cnpsSlice)
	return strings.Join(cnpsSlice, ", ")
}

func getCnpsMapFromString(cnps string) map[string]bool {
	cnpsMap := map[string]bool{}
	if cnps != "" {
		cnpsMap[cnps] = true
	}
	return cnpsMap
}
