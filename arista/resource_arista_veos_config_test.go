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

func TestResourceVeosConfig(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceVeosConfigDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceVeosConfigInitialConfig,
				Check:  testResourceVeosConfigInitialCheck,
			},
			{
				Config: testResourceVeosConfigUpdateConfig,
				Check:  testResourceVeosConfigUpdateCheck,
			},
		},
	})
}

var testResourceVeosConfigInitialConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test8"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "4.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "arista_clos" "clos" {
   name = "clos-test8"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "arista_wan" "wan" {
   name = "wan-test8"
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

resource "arista_subnet" "subnet" {
  cloud_provider = arista_vpc.vpc.cloud_provider
  vpc_id = arista_vpc.vpc.vpc_id
  availability_zone = "us-west-1b"
  subnet_id = "subnet-id"
  cidr_block = "11.0.0.0/24"
  subnet_name = "edgeSubnet0"
}

resource "arista_subnet" "subnet1" {
  cloud_provider = arista_vpc.vpc.cloud_provider
  vpc_id = arista_vpc.vpc.vpc_id
  availability_zone = "us-west-1c"
  subnet_id = "subnet-id1"
  cidr_block = "11.0.1.0/24"
  subnet_name = "edgeSubnet1"
}

resource "arista_veos_config" "veos" {
  cloud_provider = arista_vpc.vpc.cloud_provider
  topology_name = arista_topology.topology.topology_name
  role = arista_vpc_config.vpc.role
  cnps = ""
  vpc_id = arista_vpc.vpc.vpc_id
  tags = {
    "Name" = "edgeRouterveos1"
    "Cnps" = "Dev"
  }
  region = arista_vpc_config.vpc.region
  is_rr = false
  ami = "dummy-aws-machine-image"
  key_name = "foo"
  availability_zone = "us-west-1c"
  intf_name = ["edgeveos1Intf0", "edgeveos1Intf1"]
  intf_private_ip = ["11.0.0.101", "11.0.1.101"]
  intf_type = ["public", "internal"]
}
`, os.Getenv("token"))

func testResourceVeosConfigInitialCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_veos_config.veos"]
	if resourceState == nil {
		return fmt.Errorf("arista_veos_config.veos resource not found in state")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("arista_veos_config.veos resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("arista_veos_config.veos ID not assigned %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["cloud_provider"], "aws"; got != want {
		return fmt.Errorf("arista_veos_config.veos cloud_provider contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["vpc_id"], "vpc-dummy"; got != want {
		return fmt.Errorf("arista_veos_config.veos vpc_id contains %s; want %s", got, want)
	}
	return nil
}

var testResourceVeosConfigUpdateConfig = fmt.Sprintf(`
provider "arista" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "arista_topology" "topology" {
   topology_name = "topo-test8"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "4.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "arista_clos" "clos" {
   name = "clos-test8"
   topology_name = arista_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "arista_wan" "wan" {
   name = "wan-test8"
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

resource "arista_subnet" "subnet" {
  cloud_provider = arista_vpc.vpc.cloud_provider
  vpc_id = arista_vpc.vpc.vpc_id
  availability_zone = "us-west-1b"
  subnet_id = "subnet-id"
  cidr_block = "11.0.0.0/24"
  subnet_name = "edgeSubnet0"
}

resource "arista_subnet" "subnet1" {
  cloud_provider = arista_vpc.vpc.cloud_provider
  vpc_id = arista_vpc.vpc.vpc_id
  availability_zone = "us-west-1c"
  subnet_id = "subnet-id1"
  cidr_block = "11.0.1.0/24"
  subnet_name = "edgeSubnet1"
}

resource "arista_veos_config" "veos" {
  cloud_provider = arista_vpc.vpc.cloud_provider
  topology_name = arista_topology.topology.topology_name
  role = arista_vpc_config.vpc.role
  cnps = ""
  vpc_id = arista_vpc.vpc.vpc_id
  tags = {
    "Name" = "UpdatedEdgeRouterveos1"
    "Cnps" = "Dev"
  }
  region = arista_vpc_config.vpc.region
  is_rr = false
  ami = "dummy-aws-machine-image"
  key_name = "foo"
  availability_zone = "us-west-1c"
  intf_name = ["edgeveos1Intf0", "edgeveos1Intf1"]
  intf_private_ip = ["11.0.0.101", "11.0.1.101"]
  intf_type = ["public", "internal"]
}
`, os.Getenv("token"))

func testResourceVeosConfigUpdateCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["arista_veos_config.veos"]
	if resourceState == nil {
		return fmt.Errorf("arista_veos_config.veos resource not found in state")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("arista_veos_config.veos resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("arista_veos_config.veos ID not assigned %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["tags.Name"], "UpdatedEdgeRouterveos1"; got != want {
		return fmt.Errorf("arista_veos_config.veos tags contains %s; want %s", got, want)
	}
	return nil
}

func testResourceVeosConfigDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arista_veos_config" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
