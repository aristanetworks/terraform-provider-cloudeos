# cloudeos_vpc_config

The `cloudeos_vpc_config` resource sends the deployment information about the AWS VPC and Azure VNET to CVaaS.
CVaaS returns the peering information required by the Leaf VPC/VNETs to create a VPC/VNET Peering connection with its
corresponding Edge.

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
* `clos_name` - (Optional) CLOS Name this VPC refers to for attributes.
* `wan_name` - (Optional) WAN Name this VPC refers to for attributes.
* `rg_name` - (Optional) Resource group name, only valid for Azure.
* `vnet_name` - (Optional) VNET name, only valid for Azure.
* `role` - (Required) CloudEdge or CloudLeaf.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

In addition to Arguments listed above - the following Attributes are exported

* `ID` - The ID of cloudeos_vpc_config Resource.

A CloudLeaf VPC peers with the CloudEdge VPC to enable communication between instances between them.
The following Attributes are exported in CloudLeaf VPC that provides information about the peer CloudEdge VPC.

* `peer_vpc_id` - ID of the CloudEdge peer VPC, only valid for AWS.
* `peer_vpc_cidr` - CIDR of the CloudEdge peer VPC, only valid for AWS.
* `peer_vnet_id` - ID of the CloudEdge peer VNET, only valid for Azure.
* `peer_rg_name` - Resource Group name of the peer CloudEdge, only valid for Azure.
* `peer_vnet_name` - VNET name of the peer CloudEdge, only valid for Azure.

## Timeouts

* `create` - (Default of 3 minute) Used when creating the cloudeos_vpc_config Resource.
* `delete` - (Defaults to 5 minutes) Used when deleting the cloudeos_vpc_config Resource.