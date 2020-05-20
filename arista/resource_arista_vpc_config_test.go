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

func TestResourceVpcConfig(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceVpcConfigDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceVpcConfigInitialConfig,
				Check:  testResourceVpcConfigInitialCheck,
			},
			{
				Config: testResourceVpcConfigUpdateConfig,
				Check:  testResourceVpcConfigUpdateCheck,
			},
		},
	})
}

var testResourceVpcConfigInitialConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test4"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "arista_clos" "clos" {
   name = "clos-test4"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "arista_wan" "wan" {
   name = "wan-test4"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}

resource "arista_vpc_config" "vpc" {
  cloud_provider = "aws"
  topology_name = arista_topology.topology.topology_name
  clos_name = arista_clos.clos.name
  wan_name = arista_wan.wan.name
  role = "CloudEdge"
  cnps = "Dev"
  tags = {
       Name = "edgeVpc"
       Cnps = "Dev"
  }
  region = "us-west-1"
}
`, os.Getenv("token"))

func testResourceVpcConfigInitialCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_vpc_config.vpc"]
	if resourceState == nil {
		return fmt.Errorf("arista_vpc_config.vpc resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("arista_vpc_config.vpc resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("arista_vpc_config ID not assigned %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["cloud_provider"], "aws"; got != want {
		return fmt.Errorf("arista_vpc_config cloud_provider contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["topology_name"], "topo-test4"; got != want {
		return fmt.Errorf("arista_vpc_config topology_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["clos_name"], "clos-test4"; got != want {
		return fmt.Errorf("arista_vpc_config clos_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["wan_name"], "wan-test4"; got != want {
		return fmt.Errorf("arista_vpc_config wan_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["role"], "CloudEdge"; got != want {
		return fmt.Errorf("arista_vpc_config role contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["cnps"], "Dev"; got != want {
		return fmt.Errorf("arista_vpc_config cnps contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["tags.Name"], "edgeVpc"; got != want {
		return fmt.Errorf("arista_vpc_config tags contains %s; want %s", got, want)
	}
	return nil
}

var testResourceVpcConfigUpdateConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test4"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "arista_clos" "clos" {
   name = "clos-test4"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "arista_wan" "wan" {
   name = "wan-test4"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}

resource "arista_vpc_config" "vpc" {
  cloud_provider = "aws"
  topology_name = arista_topology.topology.topology_name
  clos_name = arista_clos.clos.name
  wan_name = arista_wan.wan.name
  role = "CloudEdge"
  cnps = "Dev"
  tags = {
       Name = "updatedVpcName"
       Cnps = "Dev"
  }
  region = "us-west-1"
}
`, os.Getenv("token"))

func testResourceVpcConfigUpdateCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_vpc_config.vpc"]
	if resourceState == nil {
		return fmt.Errorf("arista_vpc_config.vpc resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("arista_vpc_config.vpc resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("arista_vpc_config ID not assigned %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["tags.Name"], "updatedVpcName"; got != want {
		return fmt.Errorf("arista_vpc_config tags contains %s; want %s", got, want)
	}
	return nil
}

func testResourceVpcConfigDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arista_vpc_config" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
