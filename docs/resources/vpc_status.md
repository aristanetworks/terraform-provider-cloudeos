# vpc_status Resource

VPC resource provides AWS VPC (Azure resource group) deployment info to CVaaS.
It depends on AWS vpc and Azure resource_group.

## Example Usage

```hcl
resource "cloudeos_vpc_status" "vpc" {
  cloud_provider = cloudeos_vpc_config.vpc.cloud_provider     // Provider name (aws/azure)
  vpc_id = "vpc-dummy-id"                                     // ID of the aws vpc or azure resource group
  security_group_id = "sg-dummy-id"                           // security group associated with VPC
  cidr_block = "15.0.0.0/16"                                  // VPC CIDR block
  igw = "egdeVpcigwName"                                      // IGW name
  role = cloudeos_vpc_config.vpc.role                         // VPC role (CloudEdge/CloudLeaf)
  topology_name = cloudeos_topology.topology.topology_name    // Topology Name
  tags = cloudeos_vpc_config.vpc.tags                         // A mapping of tags to assign to the resource
  clos_name = cloudeos_clos.clos.name                         // Clos Name
  wan_name = cloudeos_wan.wan.name                            // Wan Name 
  cnps = "Dev"                                                // Cloud Network Private Segments Name
  region = cloudeos_vpc_config.vpc.region                     // Region of deployment
  account = "dummy_aws_account"                               // The unique identifier of the account
  tf_id = cloudeos_vpc_config.vpc.tf_id
}
``

## Argument Reference

* `cloud_provider` - (Required) aws/azure.
* `cnps` - (Required) Cloud Network Private Segments Name.
* `region` - (Required) Region of deployment.
* `rg_name` - (Optional) Resource group name, only valid for Azure.
* `vnet_name` - (Optional) VNET name, only valid for Azure.
* `vpc_id` - (Required) VPC ID, this is equiv to vnet_id in Azure.
* `topology_name` - (Required) Name of topology resource.
* `clos_name` - (Optional) Clos Name this VPC refers to for attributes.
* `wan_name` - (Optional) Wan Name this VPC refers to for attributes.
* `role` - (Required) CloudEdge or CloudLeaf.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `cidr_block` - (Optional) CIDR Block for VPC.
* `igw`- (Optional) Internet gateway id.
* `resource_group` - (Optional) Azure resource group.
* `role` - (Required) VPC role, CloudEdge/CloudLeaf.
* `account` - (Required) The unique identifier of the account.