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

func TestResourceTopology(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceTopologyDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceTopologyInitialConfig,
				Check:  testResourceTopologyInitialCheck,
			},
			{
				Config: testResourceTopologyUpdateConfig,
				Check:  testResourceTopologyUpdateCheck,
			},
		},
	})
}

var testResourceTopologyInitialConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test1"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}
`, os.Getenv("token"))

var resourceTopoID = ""

func testResourceTopologyInitialCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_topology.topology"]
	if resourceState == nil {
		return fmt.Errorf("arista_topology.topology resource not found in state")
	}

	topoState := resourceState.Primary
	if topoState == nil {
		return fmt.Errorf("arista_topology.topology resource has no primary instance")
	}

	if topoState.ID == "" {
		return fmt.Errorf("arista_topology.topology ID not assigned %s", topoState.ID)
	}
	resourceTopoID = topoState.ID // use this for update testing
	if got, want := topoState.Attributes["topology_name"], "topo-test1"; got != want {
		return fmt.Errorf("topology topology_name contains %s; want %s", got, want)
	}

	if got, want := topoState.Attributes["bgp_asn"], "65000-65100"; got != want {
		return fmt.Errorf("topology bgp_asn contains %s; want %s", got, want)
	}

	if got, want := topoState.Attributes["vtep_ip_cidr"], "1.0.0.0/16"; got != want {
		return fmt.Errorf("topology vtep_ip_cidr contains %s; want %s", got, want)
	}

	if got, want := topoState.Attributes["terminattr_ip_cidr"], "2.0.0.0/16"; got != want {
		return fmt.Errorf("topology terminattr_ip_cidr contains %s; want %s", got, want)
	}

	if got, want := topoState.Attributes["dps_controlplane_cidr"], "3.0.0.0/16"; got != want {
		return fmt.Errorf("topology dps_controlplane_cidr contains %s; want %s", got, want)
	}

	return nil
}

var testResourceTopologyUpdateConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}
`, os.Getenv("token"))

func testResourceTopologyUpdateCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_topology.topology"]
	topoState := resourceState.Primary
	if topoState.ID != resourceTopoID {
		return fmt.Errorf("arista_topology.topology ID has changed during update %s", topoState.ID)
	}

	if got, want := topoState.Attributes["topology_name"], "topo-test"; got != want {
		return fmt.Errorf("topology topology_name contains %s; want %s", got, want)
	}
	return nil
}

func testResourceTopologyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arista_topology" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
