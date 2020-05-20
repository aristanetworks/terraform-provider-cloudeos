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

func TestResourceVpcStatus(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceVpcStatusDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceVpcStatusInitialConfig,
				Check:  testResourceVpcStatusInitialCheck,
			},
			{
				Config: testResourceVpcStatusUpdateConfig,
				Check:  testResourceVpcStatusUpdateCheck,
			},
		},
	})
}

var testResourceVpcStatusInitialConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test5"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "arista_clos" "clos" {
   name = "clos-test5"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "arista_wan" "wan" {
   name = "wan-test5"
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

resource "arista_vpc" "vpc" {
  cloud_provider = arista_vpc_config.vpc.cloud_provider
  vpc_id = "vpc-dummy"
  security_group_id = "sg-dummy"
  cidr_block = "11.0.0.0/16"
  igw = "egdeVpcigw"
  role = arista_vpc_config.vpc.role  
  topology_name = arista_topology.topology.topology_name
  tags = arista_vpc_config.vpc.tags
  clos_name = arista_clos.clos.name
  wan_name = arista_wan.wan.name
  cnps = "Dev"
  region = arista_vpc_config.vpc.region
  account = "dummy_aws_account"
  tf_id = arista_vpc_config.vpc.tf_id
}
`, os.Getenv("token"))
var vpcStatusResourceID = ""

func testResourceVpcStatusInitialCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_vpc.vpc"]
	if resourceState == nil {
		return fmt.Errorf("arista_vpc.vpc resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("arista_vpc.vpc resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("arista_vpc.vpc ID not assigned %s", instanceState.ID)
	}
	vpcStatusResourceID = instanceState.ID

	if got, want := instanceState.Attributes["vpc_id"], "vpc-dummy"; got != want {
		return fmt.Errorf("arista_vpc.vpc vpc_id contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["cidr_block"], "11.0.0.0/16"; got != want {
		return fmt.Errorf("arista_vpc.vpc cidr_block contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["cloud_provider"], "aws"; got != want {
		return fmt.Errorf("arista_vpc.vpc cloud_provider contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["topology_name"], "topo-test5"; got != want {
		return fmt.Errorf("arista_vpc.vpc topology_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["clos_name"], "clos-test5"; got != want {
		return fmt.Errorf("arista_vpc.vpc clos_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["wan_name"], "wan-test5"; got != want {
		return fmt.Errorf("arista_vpc.vpc wan_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["role"], "CloudEdge"; got != want {
		return fmt.Errorf("arista_vpc.vpc role contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["tags.Name"], "edgeVpc"; got != want {
		return fmt.Errorf("arista_vpc.vpc tags contains %s; want %s", got, want)
	}
	return nil
}

var testResourceVpcStatusUpdateConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test5"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "arista_clos" "clos" {
   name = "clos-test5"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "arista_wan" "wan" {
   name = "wan-test5"
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
       Name = "updatedEdgeVpc"
       Cnps = "Dev"
  }
  region = "us-west-1"
}

resource "arista_vpc" "vpc" {
  cloud_provider = arista_vpc_config.vpc.cloud_provider
  vpc_id = "vpc-dummy"
  security_group_id = "sg-dummy"
  cidr_block = "11.0.0.0/16"
  igw = "egdeVpcigw"
  role = arista_vpc_config.vpc.role  
  topology_name = arista_topology.topology.topology_name
  tags = arista_vpc_config.vpc.tags
  clos_name = arista_clos.clos.name
  wan_name = arista_wan.wan.name
  cnps = "Dev"
  region = arista_vpc_config.vpc.region
  account = "dummy_aws_account"
  tf_id = arista_vpc_config.vpc.tf_id
}
`, os.Getenv("token"))

func testResourceVpcStatusUpdateCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_vpc.vpc"]
	if resourceState == nil {
		return fmt.Errorf("arista_vpc.vpc resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("arista_vpc.vpc resource has no primary instance")
	}

	if instanceState.ID != vpcStatusResourceID {
		return fmt.Errorf("arista_vpc.vpc ID has changed %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["tags.Name"], "updatedEdgeVpc"; got != want {
		return fmt.Errorf("arista_vpc.vpc tags contains %s; want %s", got, want)
	}
	return nil
}

func testResourceVpcStatusDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arista_vpc" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
