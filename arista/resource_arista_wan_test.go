// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"fmt"
	"os"
	"testing"

	r "github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceWan(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceWanDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceWanInitialConfig,
				Check:  testResourceWanInitialCheck,
			},
			{
				Config: testResourceWanUpdateConfig,
				Check:  testResourceWanUpdateCheck,
			},
		},
	})
}

var testResourceWanInitialConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test2"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "arista_wan" "wan" {
   name = "wan-test2"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}
`, os.Getenv("token"))

var wanResourceID = ""

func testResourceWanInitialCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_wan.wan"]
	if resourceState == nil {
		return fmt.Errorf("arista_wan.wan resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("arista_wan.wan has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("arista_wan.wan ID not assigned %s", instanceState.ID)
	}
	wanResourceID = instanceState.ID

	if got, want := instanceState.Attributes["name"], "wan-test2"; got != want {
		return fmt.Errorf("arista_wan.wan name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["topology_name"], "topo-test2"; got != want {
		return fmt.Errorf("arista_wan.wan topology_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["cv_container_name"], "CloudEdge"; got != want {
		return fmt.Errorf("arista_wan.wan cv_container_name contains %s; want %s", got, want)
	}
	return nil
}

var testResourceWanUpdateConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test2"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "arista_wan" "wan" {
   name = "wan-test-update2"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}
`, os.Getenv("token"))

func testResourceWanUpdateCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_wan.wan"]
	if resourceState == nil {
		return fmt.Errorf("arista_wan.wan resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("arista_wan.wan resource has no primary instance")
	}

	if instanceState.ID != wanResourceID {
		return fmt.Errorf("arista_wan.wan ID has changed %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["name"], "wan-test-update2"; got != want {
		return fmt.Errorf("arista_wan.wan name contains %s; want %s", got, want)
	}

	return nil
}

func testResourceWanDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arista_wan" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
