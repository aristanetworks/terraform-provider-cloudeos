package cloudeos

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	cdv1_api "terraform-provider-cloudeos/cloudeos/arista/clouddeploy.v1"

	fmp "github.com/aristanetworks/cloudvision-go/api/fmp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Given a ResourceData with 'role' and 'deploy_mode' attributes,
// this shall specify whether the deploy_mode value is valid
// for the given role. Currently, when deploy_mode ='provision'
// we only accept resources with role = 'CloudEdge'
func validateDeployModeWithRole(d *schema.ResourceData) error {

	deployMode := strings.ToLower(d.Get("deploy_mode").(string))
	role := d.Get("role").(string)
	if deployMode == "provision" && strings.EqualFold("CloudLeaf", role) {
		return errors.New("Deploy mode provision is only applicable to " +
			"resources with role CloudEdge")
	}
	return nil
}

func getCloudProviderType(d *schema.ResourceData) cdv1_api.CloudProviderType {
	cloudProvider := d.Get("cloud_provider").(string)
	cpType := cdv1_api.CloudProviderType_CLOUD_PROVIDER_TYPE_UNSPECIFIED
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = cdv1_api.CloudProviderType_CLOUD_PROVIDER_TYPE_AWS
	case strings.EqualFold("azure", cloudProvider):
		cpType = cdv1_api.CloudProviderType_CLOUD_PROVIDER_TYPE_AZURE
	}
	return cpType
}

func getAwsVpcName(d *schema.ResourceData) (string, error) {
	var vpcName string
	if value, ok := d.GetOk("tags"); ok {
		tags := value.(map[string]interface{})
		for k, v := range tags {
			if strings.EqualFold("Name", k) {
				vpcName = v.(string)
			}
		}
	} else {
		return "", fmt.Errorf("Router name not configured in tags")
	}

	return vpcName, nil
}

func getCpTypeAndVpcName(d *schema.ResourceData) (string, cdv1_api.CloudProviderType) {
	var vpcName string
	var cpType cdv1_api.CloudProviderType
	cloudProvider := d.Get("cloud_provider").(string)
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = cdv1_api.CloudProviderType_CLOUD_PROVIDER_TYPE_AWS
		vpcName, _ = getAwsVpcName(d)
	case strings.EqualFold("azure", cloudProvider):
		cpType = cdv1_api.CloudProviderType_CLOUD_PROVIDER_TYPE_AZURE
		vpcName = d.Get("vnet_name").(string)
	}
	return vpcName, cpType
}

func getRoleType(role string) cdv1_api.RoleType {
	var roleType cdv1_api.RoleType
	switch {
	case strings.EqualFold("CloudEdge", role):
		roleType = cdv1_api.RoleType_ROLE_TYPE_EDGE
	case strings.EqualFold("CloudSpine", role):
		roleType = cdv1_api.RoleType_ROLE_TYPE_SPINE
	case strings.EqualFold("CloudLeaf", role):
		roleType = cdv1_api.RoleType_ROLE_TYPE_LEAF
	default:
		roleType = cdv1_api.RoleType_ROLE_TYPE_UNSPECIFIED
	}
	return roleType
}

func getAndCreateRouteTableIDs(d *schema.ResourceData) *cdv1_api.RouteTableIds {
	privateRtTblList := d.Get("private_rt_table_ids").([]interface{})
	internalRtTblList := d.Get("internal_rt_table_ids").([]interface{})
	publicRtTblList := d.Get("public_rt_table_ids").([]interface{})

	priv := make([]string, len(privateRtTblList))
	for i, v := range privateRtTblList {
		priv[i] = fmt.Sprint(v)
	}
	pub := make([]string, len(publicRtTblList))
	for i, v := range publicRtTblList {
		pub[i] = fmt.Sprint(v)
	}
	internal := make([]string, len(internalRtTblList))
	for i, v := range internalRtTblList {
		internal[i] = fmt.Sprint(v)
	}
	routeTableList := cdv1_api.RouteTableIds{
		Public:   &fmp.RepeatedString{Values: pub},
		Private:  &fmp.RepeatedString{Values: priv},
		Internal: &fmp.RepeatedString{Values: internal},
	}

	return &routeTableList
}

func getBgpAsn(bgpAsnRange string) (uint32, uint32, error) {
	if bgpAsnRange == "" {
		log.Printf("[CVaaS-ERROR] bgp_asn cannot be empty")
		return uint32(0), uint32(0), errors.New("bgp_asn is empty")
	}
	asnRange := strings.Split(bgpAsnRange, "-")
	if len(asnRange) != 2 {
		log.Printf("[CVaaS-ERROR] Can't parse bgp_asn")
		return uint32(0), uint32(0), errors.New("Can't parse bgp_asn")
	}
	asnLow, err := strconv.ParseUint(asnRange[0], 10, 32)
	if err != nil {
		log.Printf("[CVaaS-ERROR]Can't parse bgp asn")
		return uint32(0), uint32(0), err
	}
	asnHigh, err := strconv.ParseUint(asnRange[1], 10, 32)
	if err != nil {
		log.Printf("[CVaaS-ERROR]Can't parse bgp asn")
		return uint32(0), uint32(0), err
	}
	log.Printf("[CVaaS-INFO]Bgp Asn Range %v - %v", asnLow, asnHigh)
	return uint32(asnLow), uint32(asnHigh), err
}

func getRouterNameFromSchema(d *schema.ResourceData) (string, error) {
	var routerName string
	if value, ok := d.GetOk("tags"); ok {
		tags := value.(map[string]interface{})
		for k, v := range tags {
			if strings.EqualFold("Name", k) {
				routerName = v.(string)
			}
		}
	} else {
		return "", fmt.Errorf("Router name not configured in tags")
	}

	return routerName, nil
}

func setBootStrapCfg(d *schema.ResourceData, cfg string) error {
	if strings.EqualFold(cfg, "") {
		log.Printf("[WARN]The CloudEOS Router is deployed but without bootstrap configuration")
	}
	bootstrapCfg := "%EOS-STARTUP-CONFIG-START%\n" +
		cfg +
		"\n" +
		"%EOS-STARTUP-CONFIG-END%\n"

	imageOffer := d.Get("cloudeos_image_offer")
	value, licensesExist := d.GetOk("licenses")
	if licensesExist && imageOffer == "cloudeos-router-byol" {
		licenses := value.(*schema.Set).List()
		for _, v := range licenses {
			license := v.(map[string]interface{})
			marker := ""
			switch license["type"] {
			case "ipsec":
				marker = "IPSEC"
			case "bandwidth":
				marker = "BANDWIDTH"
			default:
				return fmt.Errorf("Unrecognised license type : %s", license["type"])
			}
			if marker != "" {
				filePath := license["path"].(string)
				content, err := ioutil.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("Problem reading %s license file: %s , %v", license["type"], filePath, err)
				} else {
					bootstrapCfg = bootstrapCfg +
						"%LICENSE-" + marker + "-START%\n" +
						string(content) +
						"%LICENSE-" + marker + "-END%\n"
				}
			}
		}
	}
	if err := d.Set("bootstrap_cfg", bootstrapCfg); err != nil {
		return fmt.Errorf("Error bootstrap_cfg: %v", err)
	}
	return nil
}

func parseRtrResponse(ent *cdv1_api.RouterConfig, d *schema.ResourceData) error {
	// Parse the bootstrap_cfg, haRtrId, peerRtTable  from response and set
	// in schema
	var bootstrapCfg string
	var haRtrID string
	var peerRtTblID []string // Internal peer route table ID
	var publicRtTblID []string
	var privateRtTblID []string
	var internalRtTblID []string

	bootstrapCfg = ent.GetCvInfo().GetBootstrapCfg().GetValue()
	haRtrID = ent.GetCvInfo().GetHaRtrId().GetValue()
	for _, id := range ent.GetCvInfo().GetPeerVpcRtTableId().GetValues() {
		peerRtTblID = append(peerRtTblID, id)
	}

	for _, id := range ent.GetCvInfo().GetHaRtTableIds().GetInternal().GetValues() {
		internalRtTblID = append(internalRtTblID, id)
	}

	for _, id := range ent.GetCvInfo().GetHaRtTableIds().GetPublic().GetValues() {
		publicRtTblID = append(publicRtTblID, id)
	}
	for _, id := range ent.GetCvInfo().GetHaRtTableIds().GetPrivate().GetValues() {
		privateRtTblID = append(privateRtTblID, id)
	}

	// set bootstrap_cfg
	if err := setBootStrapCfg(d, bootstrapCfg); err != nil {
		return err
	}
	if err := d.Set("ha_rtr_id", haRtrID); err != nil {
		return fmt.Errorf("Not able to set ha_rtr_id: %v", err)
	}
	if err := d.Set("peerroutetableid1", peerRtTblID); err != nil {
		return fmt.Errorf("Not able to set peer route table ID: %v ", err)
	}
	if err := d.Set("peer_routetable_id", peerRtTblID); err != nil {
		return fmt.Errorf("Not able to set peer route table ID: %v ", err)
	}
	if err := d.Set("public_rt_table_id", publicRtTblID); err != nil {
		return fmt.Errorf("Not able to set public route table id: %v", err)
	}
	if err := d.Set("private_rt_table_id", privateRtTblID); err != nil {
		return fmt.Errorf("Not able to set private route table ID: %v", err)
	}
	if err := d.Set("internal_rt_table_id", internalRtTblID); err != nil {
		return fmt.Errorf("Not able to set internal route table ID: %v", err)
	}
	return nil
}
