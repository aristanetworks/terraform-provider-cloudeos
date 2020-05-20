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

func TestResourceClos(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceClosDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceClosInitialConfig,
				Check:  testResourceClosInitialCheck,
			},
			{
				Config: testResourceClosUpdateConfig,
				Check:  testResourceClosUpdateCheck,
			},
		},
	})
}

var testResourceClosInitialConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test3"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "arista_clos" "clos" {
   name = "clos-test3"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}
`, os.Getenv("token"))
var closResourceID = ""

func testResourceClosInitialCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_clos.clos"]
	if resourceState == nil {
		return fmt.Errorf("resource not found in state")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("ID not assigned %s", instanceState.ID)
	}
	closResourceID = instanceState.ID

	if got, want := instanceState.Attributes["name"], "clos-test3"; got != want {
		return fmt.Errorf("clos contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["topology_name"], "topo-test3"; got != want {
		return fmt.Errorf("clos contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["cv_container_name"], "CloudLeaf"; got != want {
		return fmt.Errorf("clos contains %s; want %s", got, want)
	}
	return nil
}

var testResourceClosUpdateConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test3"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "arista_clos" "clos" {
   name = "clos-test-update3"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}
`, os.Getenv("token"))

func testResourceClosUpdateCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_clos.clos"]
	if resourceState == nil {
		return fmt.Errorf("arista_clos.clos resource not found in state")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("arista_clos.clos resource has no primary instance")
	}

	if instanceState.ID != closResourceID {
		return fmt.Errorf("arista_clos.clos ID has changed %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["name"], "clos-test-update3"; got != want {
		return fmt.Errorf("arista_clos.clos name contains %s; want %s", got, want)
	}
	return nil
}

func testResourceClosDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arista_clos" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
