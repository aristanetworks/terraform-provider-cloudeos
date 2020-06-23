// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	r "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceClos(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceClosDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceInitialClosConfig,
				Check:  testResourceInitialClosCheck,
			},
			{
				Config:      testResourceClosDuplicateConfig,
				ExpectError: regexp.MustCompile("cloudeos_clos clos-test3 already exists"),
			},
			{
				Config: testResourceUpdatedClosConfig,
				Check:  testResourceUpdatedClosCheck,
			},
		},
	})
}

var testResourceInitialClosConfig = fmt.Sprintf(`
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

resource "cloudeos_clos" "clos" {
   name = "clos-test3"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}
`, os.Getenv("token"))

var testResourceClosDuplicateConfig = fmt.Sprintf(`
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
resource "cloudeos_clos" "clos" {
   name = "clos-test3"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "cloudeos_clos" "clos1" {
   name = "clos-test3"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
   depends_on = [cloudeos_clos.clos]
}
`, os.Getenv("token"))

var closResourceID = ""

func testResourceInitialClosCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_clos.clos"]
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

var testResourceUpdatedClosConfig = fmt.Sprintf(`
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

resource "cloudeos_clos" "clos" {
   name = "clos-test-update3"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}
`, os.Getenv("token"))

func testResourceUpdatedClosCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_clos.clos"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_clos.clos resource not found in state")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_clos.clos resource has no primary instance")
	}

	if instanceState.ID != closResourceID {
		return fmt.Errorf("cloudeos_clos.clos ID has changed %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["name"], "clos-test-update3"; got != want {
		return fmt.Errorf("cloudeos_clos.clos name contains %s; want %s", got, want)
	}
	return nil
}

func testResourceClosDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudeos_clos" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
