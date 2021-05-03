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

func TestResourceTopology(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceTopologyDestroy,
		Steps: []r.TestStep{
			{
				Config:      testResourceDuplicateTopology,
				ExpectError: regexp.MustCompile("cloudeos_topology topo-test1 already exists"),
			},
			{
				Config:      testInvalidVtepCidr,
				ExpectError: regexp.MustCompile("is not a valid CIDR"),
			},
			{
				Config:      testInvalidTACidr,
				ExpectError: regexp.MustCompile("is not a valid CIDR"),
			},
			{
				Config:      testInvalidDPSCidr,
				ExpectError: regexp.MustCompile("is not a valid CIDR"),
			},
			{
				Config:      testInvaliddDeployModeValue,
				ExpectError: regexp.MustCompile("Valid options for deploy mode in"),
			},
			{
				Config:      testBgpAsnWithProvisionDeployMode,
				ExpectError: regexp.MustCompile("bgp_asn should not be specified"),
			},
			{
				Config:      testMissingAttributesForDefaultDeployMode,
				ExpectError: regexp.MustCompile("is a required variable for cloudeos_topology"),
			},
			{
				Config: testResourceInitialTopologyConfig,
				Check:  testResourceInitialTopologyCheck,
			},
			{
				Config: testResourceUpdatedTopologyConfig,
				Check:  testResourceUpdatedTopologyCheck,
			},
		},
	})
}

var testResourceInitialTopologyConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology2" {
   topology_name = "topo-test26"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

`, os.Getenv("token"))

// When deploy_mode is unspecified/ defaults to empty, ensure
// that terminattr_ip_cidr, dps_controlplane_cidr, bgp_asn,
// vtep_ip_cidr are specified, since they are required
var testMissingAttributesForDefaultDeployMode = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology2" {
   topology_name = "topo-test2"
   bgp_asn = "65000-65100"
}

`, os.Getenv("token"))

// When deploy_mode = provision, bgp_asn, vtep_ip_cidr,
// terminattr_ip_cidr, dps_controlplane_cidr are not allowed
var testBgpAsnWithProvisionDeployMode = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology2" {
   topology_name = "topo-test2"
   bgp_asn = "65000-65100"
   deploy_mode = "PROvision"
}

`, os.Getenv("token"))

// Only 'provision' for deploy_mode is supported,
// validated at the backend
var testInvaliddDeployModeValue = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology2" {
   topology_name = "topo-test97"
   deploy_mode = "randommode"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

`, os.Getenv("token"))

var testResourceDuplicateTopology = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}
resource "cloudeos_topology" "topology0" {
   topology_name = "topo-test1"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "4.0.0.0/16"
   terminattr_ip_cidr = "5.0.0.0/16"
   dps_controlplane_cidr = "6.0.0.0/16"
}

resource "cloudeos_topology" "topology1" {
   topology_name = "topo-test1"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "10.0.0.0/16"
   terminattr_ip_cidr = "11.0.0.0/16"
   dps_controlplane_cidr = "12.0.0.0/16"
   depends_on = [cloudeos_topology.topology0]
}
`, os.Getenv("token"))

var testInvalidVtepCidr = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}
resource "cloudeos_topology" "topology3" {
   topology_name = "topo-test3"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "4.0.0.0"
   terminattr_ip_cidr = "5.0.0.0/35"
   dps_controlplane_cidr = "6.0.0.0/16"
}
`, os.Getenv("token"))

var testInvalidTACidr = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}
resource "cloudeos_topology" "topology3" {
   topology_name = "topo-test3"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "4.0.0.0/8"
   terminattr_ip_cidr = "5.0.0.0/35"
   dps_controlplane_cidr = "6.0.0.0/16"
}
`, os.Getenv("token"))

var testInvalidDPSCidr = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}
resource "cloudeos_topology" "topology3" {
   topology_name = "topo-test3"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "4.0.0.0/8"
   terminattr_ip_cidr = "5.0.0.0/35"
   dps_controlplane_cidr = "256.0.0.0/16"
}
`, os.Getenv("token"))

var resourceTopoID = ""

func testResourceInitialTopologyCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_topology.topology2"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_topology.topology resource not found in state")
	}

	topoState := resourceState.Primary
	if topoState == nil {
		return fmt.Errorf("cloudeos_topology.topology resource has no primary instance")
	}

	if topoState.ID == "" {
		return fmt.Errorf("cloudeos_topology.topology ID not assigned %s", topoState.ID)
	}
	resourceTopoID = topoState.ID // use this for update testing
	if got, want := topoState.Attributes["topology_name"], "topo-test2"; got != want {
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

var testResourceUpdatedTopologyConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology2" {
   topology_name = "topo-test2"
   bgp_asn = "65000-65500"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}
`, os.Getenv("token"))

func testResourceUpdatedTopologyCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_topology.topology2"]
	topoState := resourceState.Primary
	if topoState.ID != resourceTopoID {
		return fmt.Errorf("cloudeos_topology.topology ID has changed during update %s to %s",
			resourceTopoID, topoState.ID)
	}

	if got, want := topoState.Attributes["bgp_asn"], "65000-65500"; got != want {
		return fmt.Errorf("topology topology_name contains %s; want %s", got, want)
	}
	return nil
}

func testResourceTopologyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudeos_topology" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
