# subnet Resource

Subnet resource provides AWS/Azure subnet deployment info to CVaaS. It depends
on AWS/Azure Subnet.

## Example Usage

```hcl
resource "cloudeos_subnet" "subnet" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider // aws/azure.
  vpc_id = cloudeos_vpc_status.vpc.vpc_id                 // vpc ID in which this subnet is created.
  availability_zone = "us-west-1b"                        // AWS/Azure availability_zone.
  subnet_id = "subnet-id"                                 // Subnet ID of subnet created in AWS/Azure.
  cidr_block = "15.0.0.0/24"                              // Subnet CIDR block.
  subnet_name = "edgeSubnet"                             // Name of the subnet.
}
```

## Argument Reference

* `cloud_provider` - (Required) aws/azure.
* `vpc_id` - (Required) VPC ID in which this subnet is created, equivalent to rg_name in Azure.
* `vnet_name` - (Optional) VNET name, only needed in Azure.
* `availability_zone` - (Optional) Availability zone.
* `subnet_id` - (Required) ID of subnet deployed in AWS/Azure.
* `cidr_block` - (Required) CIDR of the subnet.
* `subnet_name` - (Required) Name of the subnet.