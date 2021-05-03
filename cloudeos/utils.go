package cloudeos

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	api "github.com/terraform-providers/terraform-provider-cloudeos/cloudeos/internal/api"

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

func getCloudProviderType(d *schema.ResourceData) api.CloudProviderType {
	cloudProvider := d.Get("cloud_provider").(string)
	cpType := api.CloudProviderType_CP_UNSPECIFIED
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = api.CloudProviderType_CP_AWS
	case strings.EqualFold("azure", cloudProvider):
		cpType = api.CloudProviderType_CP_AZURE
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

func getCpTypeAndVpcName(d *schema.ResourceData) (string, api.CloudProviderType) {
	var vpcName string
	var cpType api.CloudProviderType
	cloudProvider := d.Get("cloud_provider").(string)
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = api.CloudProviderType_CP_AWS
		vpcName, _ = getAwsVpcName(d)
	case strings.EqualFold("azure", cloudProvider):
		cpType = api.CloudProviderType_CP_AZURE
		vpcName = d.Get("vnet_name").(string)
	}
	return vpcName, cpType
}

func getRoleType(role string) api.RoleType {
	var roleType api.RoleType
	switch {
	case strings.EqualFold("CloudEdge", role):
		roleType = api.RoleType_ROLE_EDGE
	case strings.EqualFold("CloudSpine", role):
		roleType = api.RoleType_ROLE_SPINE
	case strings.EqualFold("CloudLeaf", role):
		roleType = api.RoleType_ROLE_LEAF
	default:
		roleType = api.RoleType_ROLE_UNSPECIFIED
	}
	return roleType
}

func getAndCreateRouteTableIDs(d *schema.ResourceData) *api.RouteTableIds {
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
	var routeTableList api.RouteTableIds
	routeTableList.Public = pub
	routeTableList.Internal = internal
	routeTableList.Private = priv

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
	//No PayG image for Azure.
	cpType := d.Get("cloud_provider")
	if cpType == "azure" {
		ipsecFile := "AristaTesting-IPSec.json"
		content, err := ioutil.ReadFile(ipsecFile)
		if err != nil {
			log.Printf("Problem reading IpSec license file: %s , %v", ipsecFile, err)
		}
		ipsecLicense := string(content)

		bwFile := "AristaTesting-vEOS.json"
		bwContent, err := ioutil.ReadFile(bwFile)
		if err != nil {
			log.Printf("Problem reading vEOS license file: %s , %v", bwFile, err)
		}
		bwLicense := string(bwContent)
		bootstrapCfg := "%EOS-STARTUP-CONFIG-START%\n" +
			cfg +
			"\n" +
			"%EOS-STARTUP-CONFIG-END%\n" +
			"%LICENSE-IPSEC-START%\n" +
			ipsecLicense +
			"%LICENSE-IPSEC-END%\n" +
			"%LICENSE-BANDWIDTH-START%\n" +
			bwLicense +
			"%LICENSE-BANDWIDTH-END%\n"
		if err := d.Set("bootstrap_cfg", bootstrapCfg); err != nil {
			return fmt.Errorf("Error bootstrap_cfg: %v", err)
		}
	} else if cpType == "aws" {
		bootstrapCfg := "%EOS-STARTUP-CONFIG-START%\n" +
			cfg +
			"\n" +
			"%EOS-STARTUP-CONFIG-END%\n"
		if err := d.Set("bootstrap_cfg", bootstrapCfg); err != nil {
			return fmt.Errorf("Error bootstrap_cfg: %v", err)
		}
	}
	return nil
}

func parseRtrResponse(rtr map[string]interface{}, d *schema.ResourceData) error {
	// Parse the bootstrap_cfg, haRtrId, peerRtTable  from response and set
	// in schema
	var bootstrapCfg string
	var haRtrID string
	var peerRtTblID []string // Internal peer route table ID
	var publicRtTblID []string
	var privateRtTblID []string
	var internalRtTblID []string

	for k, v := range rtr {
		if strings.EqualFold(k, "cv_info") {
			if cvInfo, ok := v.(map[string]interface{}); ok {
				for cvKey, cvVal := range cvInfo {
					if strings.EqualFold(cvKey, "bootstrap_cfg") {
						bootstrapCfg = cvVal.(string)
					}
					if strings.EqualFold(cvKey, "ha_rtr_id") {
						haRtrID = cvVal.(string)
					}
					if strings.EqualFold(cvKey, "peer_vpc_rt_table_id") {
						for _, id := range cvVal.([]interface{}) {
							peerRtTblID = append(peerRtTblID, id.(string))
						}
					}
					if strings.EqualFold(cvKey, "ha_rt_table_ids") {
						if rtTblIDs, ok := cvVal.(map[string]interface{}); ok {
							for rtKey, val := range rtTblIDs {
								if strings.EqualFold(rtKey, "public") {
									for _, id := range val.([]interface{}) {
										publicRtTblID = append(publicRtTblID, id.(string))
									}
								}
								if strings.EqualFold(rtKey, "private") {
									for _, id := range val.([]interface{}) {
										privateRtTblID = append(privateRtTblID, id.(string))
									}
								}
								if strings.EqualFold(rtKey, "internal") {
									for _, id := range val.([]interface{}) {
										internalRtTblID = append(internalRtTblID, id.(string))
									}
								}
							}
						}
					}
				}
			}
		}
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
