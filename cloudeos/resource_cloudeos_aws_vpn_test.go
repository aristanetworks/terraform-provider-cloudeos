// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"fmt"
	"os"
	"testing"

	r "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceAwsVpn(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceAwsVpnDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceInitialAwsVpnConfig,
				Check:  testResourceInitialAwsVpnCheck,
			},
		},
	})
}

var testResourceInitialAwsVpnConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_aws_vpn" "vpn_config" {
       cgw_id                   = "cgw-0e93e4eef31c42466"
       cnps                     = "dev"
       router_id                = "rtr1"
       tgw_id                   = "tgw-0a5856fd8cb6fbee6"
	   vpn_tgw_attachment_id    = "tgw-attach-a1234576"
       tunnel1_aws_overlay_ip    = "169.254.244.201"
       tunnel1_bgp_asn          = "64512"
       tunnel1_bgp_holdtime     = "30"
       tunnel1_aws_endpoint_ip      = "3.230.55.101"
       tunnel1_router_overlay_ip = "169.254.244.202"
	   tunnel1_preshared_key    = "key1"
       tunnel2_bgp_asn          = "64512"
       tunnel2_bgp_holdtime     = "30"
       tunnel2_aws_endpoint_ip      = "34.224.224.37"
       tunnel2_router_overlay_ip = "169.254.36.110"
       tunnel2_aws_overlay_ip    = "169.254.36.11"
	   tunnel2_preshared_key    = "key1"
       vpc_id                   = "vpc-0d981c28a83c3fe55"
       vpn_connection_id        = "vpn-091b0a507e134a329"
	   vpn_gateway_id			= ""
}`, os.Getenv("token"))

func testResourceInitialAwsVpnCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_aws_vpn.vpn_config"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_aws_vpn resource not found in state")
	}
	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_aws_vpn instance not found in state")
	}
	if instanceState.ID == "" {
		return fmt.Errorf("cloudeos_router_config ID not assigned %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["cnps"], "dev"; got != want {
		return fmt.Errorf("cloudeos_router_config cloud_provider contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["vpc_id"], "vpc-0d981c28a83c3fe55"; got != want {
		return fmt.Errorf("cloudeos_router_config vpc_id contains %s; want %s", got, want)
	}
	return nil
}

func testResourceAwsVpnDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudeos_aws_vpn" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
