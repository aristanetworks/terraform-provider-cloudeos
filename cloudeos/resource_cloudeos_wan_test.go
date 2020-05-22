// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cloudeos

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	r "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceWan(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceWanDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceInitialWanConfig,
				Check:  testResourceInitialWanCheck,
			},
			{
				Config:      testResourceWanDuplicateConfig,
				ExpectError: regexp.MustCompile("cloudeos_wan wan-test3 already exists"),
			},
			{
				Config: testResourceUpdatedWanConfig,
				Check:  testResourceUpdatedWanCheck,
			},
		},
	})
}

var testResourceInitialWanConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test4"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test2"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}
`, os.Getenv("token"))

var testResourceWanDuplicateConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test3"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test3"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}

resource "cloudeos_wan" "wan1" {
   name = "wan-test3"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudEdge"
   depends_on = [cloudeos_wan.wan]
}
`, os.Getenv("token"))

var wanResourceID = ""

func testResourceInitialWanCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_wan.wan"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_wan.wan resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_wan.wan has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("cloudeos_wan.wan ID not assigned %s", instanceState.ID)
	}
	wanResourceID = instanceState.ID

	if got, want := instanceState.Attributes["name"], "wan-test2"; got != want {
		return fmt.Errorf("cloudeos_wan.wan name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["topology_name"], "topo-test4"; got != want {
		return fmt.Errorf("cloudeos_wan.wan topology_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["cv_container_name"], "CloudEdge"; got != want {
		return fmt.Errorf("cloudeos_wan.wan cv_container_name contains %s; want %s", got, want)
	}
	return nil
}

var testResourceUpdatedWanConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test2"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test-update2"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}
`, os.Getenv("token"))

func testResourceUpdatedWanCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_wan.wan"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_wan.wan resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_wan.wan resource has no primary instance")
	}

	if instanceState.ID != wanResourceID {
		return fmt.Errorf("cloudeos_wan.wan ID has changed %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["name"], "wan-test-update2"; got != want {
		return fmt.Errorf("cloudeos_wan.wan name contains %s; want %s", got, want)
	}

	return nil
}

func testResourceWanDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudeos_wan" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
