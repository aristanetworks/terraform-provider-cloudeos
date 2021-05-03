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

func TestResourceRouterConfig(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceRouterConfigDestroy,
		Steps: []r.TestStep{
			{
				Config:      testLeafRtrConfigInProvisionMode,
				ExpectError: regexp.MustCompile("only applicable to resources with role CloudEdge"),
			},
			{
				Config: testResourceIntialRouterConfig,
				Check:  testResourceIntialRouterCheck,
			},
			{
				Config: testResourceUpdatedRouterConfig,
				Check:  testResourceUpdatedRouterCheck,
			},
		},
	})
}

var testLeafRtrConfigInProvisionMode = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test25"
   deploy_mode = "provision"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test25"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}

resource "cloudeos_vpc_config" "vpc" {
  cloud_provider = "aws"
  topology_name = cloudeos_topology.topology.topology_name
  wan_name = cloudeos_wan.wan.name
  role = "CloudLeaf"
  cnps = ""
  tags = {
       Name = "edgeVpc"
  }
  region = "us-west-1"
  deploy_mode = "provision"
}

resource "cloudeos_vpc_status" "vpc" {
  cloud_provider = cloudeos_vpc_config.vpc.cloud_provider
  vpc_id = "vpc-dummy"
  security_group_id = "sg-dummy"
  cidr_block = "11.0.0.0/16"
  igw = "egdeVpcigw"
  cnps = ""
  role = cloudeos_vpc_config.vpc.role
  topology_name = cloudeos_topology.topology.topology_name
  tags = cloudeos_vpc_config.vpc.tags
  wan_name = cloudeos_wan.wan.name
  region = cloudeos_vpc_config.vpc.region
  account = "dummy_aws_account"
  tf_id = cloudeos_vpc_config.vpc.tf_id
  deploy_mode = "provision"
}

resource "cloudeos_subnet" "subnet" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider
  vpc_id = cloudeos_vpc_status.vpc.vpc_id
  availability_zone = "us-west-1b"
  subnet_id = "subnet-id"
  cidr_block = "11.0.0.0/24"
  subnet_name = "edgeSubnet0"
}

resource "cloudeos_subnet" "subnet1" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider
  vpc_id = cloudeos_vpc_status.vpc.vpc_id
  availability_zone = "us-west-1c"
  subnet_id = "subnet-id1"
  cidr_block = "11.0.1.0/24"
  subnet_name = "edgeSubnet1"
}

resource "cloudeos_router_config" "cloudeos" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider
  topology_name = cloudeos_topology.topology.topology_name
  role = cloudeos_vpc_config.vpc.role
  vpc_id = cloudeos_vpc_status.vpc.vpc_id
  cnps = ""
  tags = {
    "Name" = "edgeRoutercloudeos1"
  }
  region = cloudeos_vpc_config.vpc.region
  is_rr = false
  ami = "dummy-aws-machine-image"
  key_name = "foo"
  availability_zone = "us-west-1c"
  intf_name = ["edgecloudeos1Intf0", "edgecloudeos1Intf1"]
  intf_private_ip = ["11.0.0.101", "11.0.1.101"]
  intf_type = ["public", "internal"]
  deploy_mode = "provision"
}
`, os.Getenv("token"))

var testResourceIntialRouterConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test41"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "4.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_clos" "clos" {
   name = "clos-test41"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test41"
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

resource "cloudeos_vpc_status" "vpc" {
  cloud_provider = cloudeos_vpc_config.vpc.cloud_provider
  vpc_id = "vpc-dummy"
  security_group_id = "sg-dummy"
  cidr_block = "11.0.0.0/16"
  igw = "egdeVpcigw"
  role = cloudeos_vpc_config.vpc.role
  topology_name = cloudeos_topology.topology.topology_name
  tags = cloudeos_vpc_config.vpc.tags
  clos_name = cloudeos_clos.clos.name
  wan_name = cloudeos_wan.wan.name
  cnps = "Dev"
  region = cloudeos_vpc_config.vpc.region
  account = "dummy_aws_account"
  tf_id = cloudeos_vpc_config.vpc.tf_id
}

resource "cloudeos_subnet" "subnet" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider
  vpc_id = cloudeos_vpc_status.vpc.vpc_id
  availability_zone = "us-west-1b"
  subnet_id = "subnet-id"
  cidr_block = "11.0.0.0/24"
  subnet_name = "edgeSubnet0"
}

resource "cloudeos_subnet" "subnet1" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider
  vpc_id = cloudeos_vpc_status.vpc.vpc_id
  availability_zone = "us-west-1c"
  subnet_id = "subnet-id1"
  cidr_block = "11.0.1.0/24"
  subnet_name = "edgeSubnet1"
}

resource "cloudeos_router_config" "cloudeos" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider
  topology_name = cloudeos_topology.topology.topology_name
  role = cloudeos_vpc_config.vpc.role
  cnps = ""
  vpc_id = cloudeos_vpc_status.vpc.vpc_id
  tags = {
    "Name" = "edgeRoutercloudeos1"
    "Cnps" = "Dev"
  }
  region = cloudeos_vpc_config.vpc.region
  is_rr = false
  ami = "dummy-aws-machine-image"
  key_name = "foo"
  availability_zone = "us-west-1c"
  intf_name = ["edgecloudeos1Intf0", "edgecloudeos1Intf1"]
  intf_private_ip = ["11.0.0.101", "11.0.1.101"]
  intf_type = ["public", "internal"]
}
`, os.Getenv("token"))

func testResourceIntialRouterCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_router_config.cloudeos"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_router_config resource not found in state")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_router_config resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("cloudeos_router_config ID not assigned %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["cloud_provider"], "aws"; got != want {
		return fmt.Errorf("cloudeos_router_config cloud_provider contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["vpc_id"], "vpc-dummy"; got != want {
		return fmt.Errorf("cloudeos_router_config vpc_id contains %s; want %s", got, want)
	}
	return nil
}

var testResourceUpdatedRouterConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test41"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "4.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_clos" "clos" {
   name = "clos-test41"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test41"
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

resource "cloudeos_vpc_status" "vpc" {
  cloud_provider = cloudeos_vpc_config.vpc.cloud_provider
  vpc_id = "vpc-dummy"
  security_group_id = "sg-dummy"
  cidr_block = "11.0.0.0/16"
  igw = "egdeVpcigw"
  role = cloudeos_vpc_config.vpc.role  
  topology_name = cloudeos_topology.topology.topology_name
  tags = cloudeos_vpc_config.vpc.tags
  clos_name = cloudeos_clos.clos.name
  wan_name = cloudeos_wan.wan.name
  cnps = "Dev"
  region = cloudeos_vpc_config.vpc.region
  account = "dummy_aws_account"
  tf_id = cloudeos_vpc_config.vpc.tf_id
}

resource "cloudeos_subnet" "subnet" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider
  vpc_id = cloudeos_vpc_status.vpc.vpc_id
  availability_zone = "us-west-1b"
  subnet_id = "subnet-id"
  cidr_block = "11.0.0.0/24"
  subnet_name = "edgeSubnet0"
}

resource "cloudeos_subnet" "subnet1" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider
  vpc_id = cloudeos_vpc_status.vpc.vpc_id
  availability_zone = "us-west-1c"
  subnet_id = "subnet-id1"
  cidr_block = "11.0.1.0/24"
  subnet_name = "edgeSubnet1"
}

resource "cloudeos_router_config" "cloudeos" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider
  topology_name = cloudeos_topology.topology.topology_name
  role = cloudeos_vpc_config.vpc.role
  cnps = ""
  vpc_id = cloudeos_vpc_status.vpc.vpc_id
  tags = {
    "Name" = "Updatedcloudeos1"
    "Cnps" = "Dev"
  }
  region = cloudeos_vpc_config.vpc.region
  is_rr = false
  ami = "dummy-aws-machine-image"
  key_name = "foo"
  availability_zone = "us-west-1c"
  intf_name = ["edgecloudeos1Intf0", "edgecloudeos1Intf1"]
  intf_private_ip = ["11.0.0.101", "11.0.1.101"]
  intf_type = ["public", "internal"]
}
`, os.Getenv("token"))

func testResourceUpdatedRouterCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_router_config.cloudeos"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_router_config.cloudeos resource not found in state")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_router_config.cloudeos resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("cloudeos_router_config.cloudeos ID not assigned %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["tags.Name"], "Updatedcloudeos1"; got != want {
		return fmt.Errorf("cloudeos_router_config.cloudeos tags contains %s; want %s", got, want)
	}
	return nil
}

func testResourceRouterConfigDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudeos_router_config" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
