# cloudeos_vpc_status Resource

The `cloudeos_vpc_status` resource provides AWS VPC or Azure Resource Group/VNET deployment information to CVaaS.

## Example Usage

### AWS example

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

resource "aws_vpc" "vpc" {
    cidr_block = "100.0.0.0/16"
}

resource "aws_security_group" "sg" {
  name = "example_sg"
}

resource "cloudeos_vpc_status" "vpc" {
  cloud_provider = cloudeos_vpc_config.vpc.cloud_provider     // Provider name
  vpc_id = aws_vpc.vpc.id                                     // ID of the aws vpc
  security_group_id = aws_security_group.sg.id                // security group associated with VPC
  cidr_block = aws_vpc.vpc.cidr_block                         // VPC CIDR block
  igw = aws_security_group.sg.name                            // IGW name
  role = cloudeos_vpc_config.vpc.role                         // VPC role (CloudEdge/CloudLeaf)
  topology_name = cloudeos_topology.topology.topology_name    // Topology Name
  tags = cloudeos_vpc_config.vpc.tags                         // A mapping of tags to assign to the resource
  clos_name = cloudeos_clos.clos.name                         // Clos Name
  wan_name = cloudeos_wan.wan.name                            // Wan Name 
  cnps = cloudeos_vpc_config.vpc.cnps                         // Cloud Network Private Segments Name
  region = cloudeos_vpc_config.vpc.region                     // Region of deployment
  account = "dummy_aws_account"                               // The unique identifier of the account
  tf_id = cloudeos_vpc_config.vpc.tf_id
}
```

### Azure example

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
  cloud_provider = "azure"
  topology_name = cloudeos_topology.topology.topology_name
  clos_name = cloudeos_clos.clos.name
  wan_name = cloudeos_wan.wan.name
  role = "CloudEdge"
  cnps = "Dev"
  tags = {
       Name = "azureEdgeVpc"
       Cnps = "Dev"
  }
  vnet_name = "edge1Vnet"
  region = "westus2"
}

resource "azurerm_resource_group" "rg" {
  name = "edge1RG"
  location = cloudeos_vpc_config.vpc.region
}

resource "azurerm_virtual_network" "vnet" {
  name                = cloudeos_vpc_config.vpc.vnet_name
  address_space       = ["100.0.0.0/16"]
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  tags                = cloudeos_vpc_config.vpc.tags
}

resource "azurerm_network_security_group" "sg" {
  depends_on          = [azurerm_resource_group.rg]
  name                = "example_sg"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
}

data "azurerm_client_config" "current" {}

resource "cloudeos_vpc_status" "vpc" {
  cloud_provider = cloudeos_vpc_config.vpc.cloud_provider      // Provider name
  rg_name = azurerm_resource_group.rg.name                     // Azure resource group name
  vpc_id = azurerm_virtual_network.vnet.id                     // ID of the azure virtual network
  security_group_id = azurerm_network_security_group.sg[0].id  // security group associated with virtual network
  cidr_block = azurerm_virtual_network.vnet.address_space[0]   // VPC CIDR block
  role = cloudeos_vpc_config.vpc.role                          // VPC role (CloudEdge/CloudLeaf)
  topology_name = cloudeos_topology.topology.topology_name     // Topology Name
  tags = cloudeos_vpc_config.vpc.tags                          // A mapping of tags to assign to the resource
  clos_name = cloudeos_clos.clos.name                          // Clos Name
  wan_name = cloudeos_wan.wan.name                             // Wan Name 
  cnps = cloudeos_vpc_config.vpc.cnps                          // Cloud Network Private Segments Name
  region = cloudeos_vpc_config.vpc.region                      // Region of deployment
  account = data.azurerm_client_config.current.subscription_id // The unique identifier of the account
  tf_id = cloudeos_vpc_config.vpc.tf_id
}
```

## Argument Reference

* `cloud_provider` - (Required) The Cloud Provider in which the VPC/VNET is deployed.
* `cnps` - (Required) Cloud Network Private Segments Name. ( VRF Name )
* `topology_name` - (Required) Name of topology resource.
* `region` - (Required) Region of deployment.
* `vpc_id` - (Required) VPC ID, this is equiv to vnet_id in Azure.
* `role` - (Required) CloudEdge or CloudLeaf.
* `role` - (Required) VPC role, CloudEdge/CloudLeaf.
* `account` - (Required) The unique identifier of the account.
* `rg_name` - (Optional) Resource group name, only valid for Azure.
* `vnet_name` - (Optional) VNET name, only valid for Azure.
* `clos_name` - (Optional) Clos Name this VPC refers to for attributes.
* `wan_name` - (Optional) Wan Name this VPC refers to for attributes.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `cidr_block` - (Optional) CIDR Block for VPC.
* `igw`- (Optional) Internet gateway id, only valid for AWS.

## Attributes Reference

In addition to Arguments listed above - the following Attributes are exported

* `ID` - The ID of cloudeos_vpc_status Resource.

## Timeouts

* `delete` - (Defaults to 5 minutes) Used when deleting the cloudeos_vpc_status Resource.