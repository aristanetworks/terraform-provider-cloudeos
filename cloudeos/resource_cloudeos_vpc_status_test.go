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

func TestResourceVpcStatus(t *testing.T) {
	r.Test(t, r.TestCase{
		Providers:    testProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testResourceVpcStatusDestroy,
		Steps: []r.TestStep{
			{
				Config: testResourceInitialVpcStatusConfig,
				Check:  testResourceInitialVpcStatusConfigCheck,
			},
			{
				Config: testResourceUpdatedVpcStatusConfigConfig,
				Check:  testResourceUpdatedVpcStatusConfigCheck,
			},
		},
	})
}

var testResourceInitialVpcStatusConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test5"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_clos" "clos" {
   name = "clos-test5"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test5"
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
`, os.Getenv("token"))
var vpcStatusResourceID = ""

func testResourceInitialVpcStatusConfigCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_vpc_status.vpc"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_vpc_status.vpc resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_vpc_status.vpc resource has no primary instance")
	}

	if instanceState.ID == "" {
		return fmt.Errorf("cloudeos_vpc_status.vpc ID not assigned %s", instanceState.ID)
	}
	vpcStatusResourceID = instanceState.ID

	if got, want := instanceState.Attributes["vpc_id"], "vpc-dummy"; got != want {
		return fmt.Errorf("cloudeos_vpc_status.vpc vpc_id contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["cidr_block"], "11.0.0.0/16"; got != want {
		return fmt.Errorf("cloudeos_vpc_status.vpc cidr_block contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["cloud_provider"], "aws"; got != want {
		return fmt.Errorf("cloudeos_vpc_status.vpc cloud_provider contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["topology_name"], "topo-test5"; got != want {
		return fmt.Errorf("cloudeos_vpc_status.vpc topology_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["clos_name"], "clos-test5"; got != want {
		return fmt.Errorf("cloudeos_vpc_status.vpc clos_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["wan_name"], "wan-test5"; got != want {
		return fmt.Errorf("cloudeos_vpc_status.vpc wan_name contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["role"], "CloudEdge"; got != want {
		return fmt.Errorf("cloudeos_vpc_status.vpc role contains %s; want %s", got, want)
	}

	if got, want := instanceState.Attributes["tags.Name"], "edgeVpc"; got != want {
		return fmt.Errorf("cloudeos_vpc_status.vpc tags contains %s; want %s", got, want)
	}
	return nil
}

var testResourceUpdatedVpcStatusConfigConfig = fmt.Sprintf(`
provider "cloudeos" {
  cvaas_domain = "apiserver.cv-play.corp.arista.io"
  cvaas_server = "www.cv-play.corp.arista.io"
  // clouddeploy token
  service_account_web_token = %q
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test5"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "2.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_clos" "clos" {
   name = "clos-test5"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test5"
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
       Name = "updatedEdgeVpc"
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
`, os.Getenv("token"))

func testResourceUpdatedVpcStatusConfigCheck(s *terraform.State) error {
	resourceState := s.Modules[0].Resources["cloudeos_vpc_status.vpc"]
	if resourceState == nil {
		return fmt.Errorf("cloudeos_vpc_status.vpc resource not found")
	}

	instanceState := resourceState.Primary
	if instanceState == nil {
		return fmt.Errorf("cloudeos_vpc_status.vpc resource has no primary instance")
	}

	if instanceState.ID != vpcStatusResourceID {
		return fmt.Errorf("cloudeos_vpc_status.vpc ID has changed %s", instanceState.ID)
	}

	if got, want := instanceState.Attributes["tags.Name"], "updatedEdgeVpc"; got != want {
		return fmt.Errorf("cloudeos_vpc_status.vpc tags contains %s; want %s", got, want)
	}
	return nil
}

func testResourceVpcStatusDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudeos_vpc_status" {
			continue
		}
		// TODO
		return nil
	}
	return nil
}
