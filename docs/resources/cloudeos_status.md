# cloudeos_status Resource

cloudeos_status sends back the recently deployed CloudEOS information to CVaaS related.
The information include the public IP, instance ID, interface ID etc.

## Example Usage

```hcl
resource "cloudeos_router_status" "cloudeos" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider             // aws/azure
  cv_container = "dummy"                                              // Container which this CloudEOS belongs
  cnps = "Dev"                                                        // Cloud Network Private Segments Name
  vpc_id = cloudeos_vpc_status.vpc.vpc_id                             // VPC ID in which this CloudEOS is deployed
  instance_id = "i-00000001"                                          // instance ID of deployed CloudEOS
  instance_type = "c5.xlarge"                                         // Type of instance
  region = cloudeos_vpc_config.vpc.region                             // Region of deployment
  tags = cloudeos_router_config.cloudeos.tags                         // A mapping of tags to assign to the resource
  availability_zone = cloudeos_router_config.cloudeos.availability_zone // Availability zone
  primary_network_interface_id = "intf_ID"                            // Interface ID of primary interface
  public_ip = "172.45.67.3"                                           // public IP ( assigned by Aws/Azure )
  intf_name = cloudeos_router_config.cloudeos.intf_name               // List of interface names
  intf_id = ["intf_ID", "intf_ID1"]                                   // List of interface IDs
  intf_private_ip = cloudeos_router_config.cloudeos.intf_private_ip   // List of private IPs
  intf_subnet_id = ["dummy-id1", "dummy-id2"]                         // List of subnet IDs of interfaces
  intf_type = cloudeos_router_config.cloudeos.intf_type               // List of interface types (public, private, internal)
  tf_id = cloudeos_router_config.cloudeos.tf_id
  is_rr = "false"                                                     // true if this CloudEOS acts as Route Reflector
}
```

## Argument Reference

* `cloud_provider` - (Required) aws/azure.
* `cv_container` - (Optional) Container to which cvp should add this device.
* `vpc_id` - (Optional) VPC ID of CloudEOS, only for AWS.
* `rg_name` - (Optional) Resource group name, only for Azure.
* `rg_location` - (Optional) Resource group location, only for Azure.
* `instance_type` - (Required) Instance ID of deployed CloudEOS.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `availability_zone` - (Optional) Availability Zone of VPC.
* `primary_network_interface_id` - (Optional)
* `availability_set_id` - (Optional) Availability Set.
* `public_ip` - (Optional) Public IP of interface.
* `intf_name` - (Required) List of interface names.
* `intf_id` - (Required) List of interface IDs.
* `intf_private_ip` - (Required) List of private IPs.
* `intf_subnet_id` - (Required) List of subnet IDs of interfaces.
* `intf_type` - (Required) List of interface types.
* `private_rt_table_ids` - (Optional) List of private interface route table IDs.
* `internal_rt_table_ids` - (Optional) List of internal interface route table IDs.
* `public_rt_table_ids` - (Optional) List of public route table IDs.
* `ha_name` - (Optional) HA pair name.
* `cnps` - (Optional) Cloud Network Private Segments Name.
* `region` - (Required) Region of deployment.
* `is_rr` - (Optional) true if this CloudEOS acts as a Route Reflector.