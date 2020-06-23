// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"fmt"
	"os"
	"testing"

	r "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceSubnet(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceSubnetDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceInitialSubnetConfig,
				Check:  testResourceInitialSubnetConfigCheck,
			},
			{
				Config: testResourceUpdatedSubnetConfig,
				Check:  testResourceUpdatedSubnetConfigCheck,
			},
		},
	})
}

var testResourceInitialSubnetConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test13"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_clos" "clos" {
   name = "clos-test6"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test6"
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
`, os.Getenv("token"))

var resourceSubnetID = ""

func testResourceInitialSubnetConfigCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_subnet.subnet"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_subnet.subnet resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_subnet.subnet resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("cloudeos_subnet.subnet ID not assigned %s", instanceState.ID)
	}
	resourceSubnetID = instanceState.ID

	if got, want := instanceState.Attributes["cloud_provider"], "aws"; got != want {
		return fmt.Errorf("cloudeos_subnet.subnet cloud_provider contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["vpc_id"], "vpc-dummy"; got != want {
		return fmt.Errorf("cloudeos_subnet.subnet vpc_id contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["availability_zone"], "us-west-1b"; got != want {
		return fmt.Errorf("cloudeos_subnet availability_zone contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["subnet_id"], "subnet-id"; got != want {
		return fmt.Errorf("cloudeos_subnet.subnet subnet_id contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["cidr_block"], "11.0.0.0/24"; got != want {
		return fmt.Errorf("cloudeos_subnet.subnet cidr_block contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["subnet_name"], "edgeSubnet0"; got != want {
		return fmt.Errorf("cloudeos_subnet.subnet subnet_name contains %s; want %s", got, want)
	}
	return nil
}

var testResourceUpdatedSubnetConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test13"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_clos" "clos" {
   name = "clos-test6"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test6"
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
  subnet_name = "updatedEdgeSubnet0"
}
`, os.Getenv("token"))

func testResourceUpdatedSubnetConfigCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_subnet.subnet"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_subnet.subnet resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_subnet.subnet resource has no primary instance")
	}

	if instanceState.ID != resourceSubnetID {
		return fmt.Errorf("cloudeos_subnet.subnet ID has changed %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["subnet_name"], "updatedEdgeSubnet0"; got != want {
		return fmt.Errorf("cloudeos_subnet.subnet subnet_name contains %s; want %s", got, want)
	}
	return nil
}

func testResourceSubnetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudeos_subnet" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
