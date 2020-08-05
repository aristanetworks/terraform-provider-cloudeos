// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cloudeos

import (
	//"errors"
	"strings"
	//"time"

	//"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

//cloudeosAwsVpnStatus: Define the cloudeosAwsVpnStatus schema ( input and output variables )
func cloudeosAwsVpn() *schema.Resource {
	return &schema.Resource{
		Create: cloudeosAwsVpnCreate,
		Read:   cloudeosAwsVpnRead,
		Update: cloudeosAwsVpnUpdate,
		Delete: cloudeosAwsVpnDelete,

		Schema: map[string]*schema.Schema{
			"cnps": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Segment/VRF ID",
			},
			"tgw_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Transit Gateway ID",
			},
			"router_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "tf_id of the CloudEOS Router",
			},
			"vpn_gateway_id": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "VPN Gateway ID",
			},
			"vpn_connection_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Vpn connection ID",
			},
			"vpn_tgw_attachment_id": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "TGW Attachment ID for the VPN connection",
			},
			"cgw_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "AWS Customer Gateway ID",
			},
			"tunnel1_aws_endpoint_ip": { //tunnel1_address
				Required:    true,
				Type:        schema.TypeString,
				Description: "Public IP address of the AWS VPN Connection endpoint",
			},
			"tunnel1_bgp_asn": { //tunnel1_bgp_asn
				Required:    true,
				Type:        schema.TypeString,
				Description: "BGP ASN",
			},
			"tunnel1_router_overlay_ip": { //tunnel1_cgw_inside_address
				Required:    true,
				Type:        schema.TypeString,
				Description: "Tunnel Interface overlay IP address for the router",
			},
			"tunnel1_aws_overlay_ip": { //tunnel1_vgw_inside_address
				Required:    true,
				Type:        schema.TypeString,
				Description: "Tunnel IP address of the AWS VPN Connection",
			},
			"tunnel1_bgp_holdtime": { //tunnel1_bgp_hold_timer
				Required:    true,
				Type:        schema.TypeString,
				Description: "Hold timer value for BGP",
			},
			"tunnel1_preshared_key": { //tunnel1_preshared_key
				Required:    true,
				Type:        schema.TypeString,
				Description: "Ipsec Preshared key for Tunnel1",
				Sensitive:   true,
			},
			"tunnel2_aws_endpoint_ip": { //tunnel2_address
				Required:    true,
				Type:        schema.TypeString,
				Description: "Public IP address of the AWS VPN Connection endpoint",
			},
			"tunnel2_bgp_asn": { //tunnel2_bgp_asn
				Required:    true,
				Type:        schema.TypeString,
				Description: "BGP ASN",
			},
			"tunnel2_router_overlay_ip": { //tunnel2_cgw_inside_address
				Required:    true,
				Type:        schema.TypeString,
				Description: "Tunnel interface IP address for the router",
			},
			"tunnel2_aws_overlay_ip": { //tunnel2_vgw_inside_address
				Required:    true,
				Type:        schema.TypeString,
				Description: "Tunnel IP address of the AWS VPN Connection",
			},
			"tunnel2_bgp_holdtime": { //tunnel2_bgp_hold_timer
				Required:    true,
				Type:        schema.TypeString,
				Description: "Hold timer value for BGP",
			},
			"tunnel2_preshared_key": { //tunnel2_preshared_key
				Required:    true,
				Type:        schema.TypeString,
				Description: "Pre shared key for Tunnel1",
				Sensitive:   true,
			},
			"vpc_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "VPC ID for the Router given in \"router_id\"",
			},
			"tf_id": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "Unique resource ID",
			},
		},
	}
}
func cloudeosAwsVpnRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func cloudeosAwsVpnUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.AddAwsVpnConfig(d)
	if err != nil {
		return err
	}

	return nil
}

func cloudeosAwsVpnDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)

	err := provider.DeleteAwsVpnConfig(d)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func cloudeosAwsVpnCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)

	err := provider.AddAwsVpnConfig(d)
	if err != nil {
		return err
	}
	uuid := "cloudeos-aws-vpn" + strings.TrimPrefix(d.Get("tf_id").(string), AwsVpnPrefix)
	d.SetId(uuid)
	return nil
}
