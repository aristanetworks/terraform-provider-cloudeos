# vpc_config Resource

VPC resource provides info to CVaaS related to AWS VPC (Azure resource group) in which
CloudEOS gets deployed. AWS vpc and Azure resource_group depend on vpc_config to obtain
attributes necessary for VPC creation.

## Example Usage

```hcl
resource "cloudeos_topology" "topology" {
   topology_name = "topo-test"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "4.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_clos" "clos" {
   name = "clos-test"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudLeaf"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test"
   topology_name = cloudeos_topology.topology.topology_name
   cv_container_name = "CloudEdge"
}

resource "cloudeos_vpc_config" "vpc" {
  cloud_provider = "aws"                                     // Cloud Provider "aws/azure"
  topology_name = cloudeos_topology.topology.topology_name   // Topology resource name
  clos_name = cloudeos_clos.clos.name                        // Clos resource name
  wan_name = cloudeos_wan.wan.name                           // Wan resource name (Only needed in "CloudEdge" role)
  role = "CloudEdge"                                         // VPC role, CloudEdge/CloudLeaf
  cnps = "Dev"                                               // Cloud Network Private Segments Name
  tags = {                                                   // A mapping of tags to assign to the resource
       Name = "edgeVpc"
       Cnps = "Dev"
  }
  region = "us-west-1"                                       // region of deployment
}
```

## Argument Reference

* `topology_name` - (Required) Name of topology resource.
* `clos_name` - (Optional) Clos Name this VPC refers to for attributes.
* `wan_name` - (Optional) Wan Name this VPC refers to for attributes.
* `rg_name` - (Optional) Resource group name, only valid for Azure.
* `vnet_name` - (Optional) VNET name, only valid for Azure.
* `role` - (Required) CloudEdge or CloudLeaf.
* `tags` - (Optional) A mapping of tags to assign to the resource.
