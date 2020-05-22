// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cloudeos

import (
	"fmt"
	"os"
	"testing"

	r "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceVpcConfig(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceVpcConfigDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceInitialVpcConfig,
				Check:  testResourceInitialVpcConfigCheck,
			},
			{
				Config: testResourceUpdatedVpcConfig,
				Check:  testResourceUpdatedVpcConfigCheck,
			},
		},
	})
}

var testResourceInitialVpcConfig = fmt.Sprintf(`
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

resource "cloudeos_clos" "clos" {
   name = "clos-test4"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test4"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}

resource "cloudeos_vpc_config" "vpc" {
  cloud_provider = "aws"
  topology_name = cloudeos_topology.topology.topology_name
  clos_name = cloudeos_clos.clos.name
  wan_name = cloudeos_wan.wan.name
  role = "CloudEdge"
  cnps = "Dev"
  tags = {
       Name = "edgeVpc"
       Cnps = "Dev"
  }
  region = "us-west-1"
}
`, os.Getenv("token"))

func testResourceInitialVpcConfigCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_vpc_config.vpc"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_vpc_config.vpc resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_vpc_config.vpc resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("cloudeos_vpc_config ID not assigned %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["cloud_provider"], "aws"; got != want {
		return fmt.Errorf("cloudeos_vpc_config cloud_provider contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["topology_name"], "topo-test4"; got != want {
		return fmt.Errorf("cloudeos_vpc_config topology_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["clos_name"], "clos-test4"; got != want {
		return fmt.Errorf("cloudeos_vpc_config clos_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["wan_name"], "wan-test4"; got != want {
		return fmt.Errorf("cloudeos_vpc_config wan_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["role"], "CloudEdge"; got != want {
		return fmt.Errorf("cloudeos_vpc_config role contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["cnps"], "Dev"; got != want {
		return fmt.Errorf("cloudeos_vpc_config cnps contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["tags.Name"], "edgeVpc"; got != want {
		return fmt.Errorf("cloudeos_vpc_config tags contains %s; want %s", got, want)
	}
	return nil
}

var testResourceUpdatedVpcConfig = fmt.Sprintf(`
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

resource "cloudeos_clos" "clos" {
   name = "clos-test4"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test4"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}

resource "cloudeos_vpc_config" "vpc" {
  cloud_provider = "aws"
  topology_name = cloudeos_topology.topology.topology_name
  clos_name = cloudeos_clos.clos.name
  wan_name = cloudeos_wan.wan.name
  role = "CloudEdge"
  cnps = "Dev"
  tags = {
       Name = "updatedVpcName"
       Cnps = "Dev"
  }
  region = "us-west-1"
}
`, os.Getenv("token"))

func testResourceUpdatedVpcConfigCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_vpc_config.vpc"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_vpc_config.vpc resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_vpc_config.vpc resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("cloudeos_vpc_config ID not assigned %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["tags.Name"], "updatedVpcName"; got != want {
		return fmt.Errorf("cloudeos_vpc_config tags contains %s; want %s", got, want)
	}
	return nil
}

func testResourceVpcConfigDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudeos_vpc_config" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
