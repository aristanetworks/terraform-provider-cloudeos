# cloudeos_config Resource

cloudeos_config resource sends CloudEOS deployment related info to CVaaS and obtain the bootstrap config.
AWS/Azure instance depends on cloudeos_config for bootstrap config.

CloudEOS can act as Route Reflector, an edge router or a leaf router.

## Example Usage

```hcl
resource "cloudeos_router_config" "cloudeos" {
  cloud_provider = cloudeos_vpc_status.vpc.cloud_provider  // aws/azure
  topology_name = cloudeos_topology.topology.topology_name // Name of  Topology this CloudEOS belongs
  role = cloudeos_vpc_config.vpc.role                      // CloudLeaf/CloudEdge
  cnps = "Dev"                                             // Cloud Network Private Segments Name
  vpc_id = cloudeos_vpc_status.vpc.vpc_id                  // VPC ID in which this CloudEOS is deployed
  tags = {                                                 // A mapping of tags to assign to the resource
    "Name" = "edgeRoutercloudeos2"
    "Cnps" = "Dev"
  }
  region = cloudeos_vpc_config.vpc.region                  // Region of deployment
  is_rr = false                                            // true if this CloudEOS acts as Route Reflector
  ami = "dummy-aws-machine-image"                          // Amazon Machine Image 
  key_name = "foo"                                         // Keypair name
  availability_zone = "us-west-1c"                         // Availability zone
  intf_name = ["edgecloudeos1Intf0", "edgecloudeos1Intf1"] // List of interface name attached to this CloudEOS.
  intf_private_ip = ["15.0.0.101", "15.0.1.101"]           // List of private IP of interfaces.
  intf_type = ["public", "internal"]                       // Type of interfaces
}
```

## Argument Reference

* `cloud_provider` - (Required) aws/azure.
* `cnps` - (Optional) Cloud Network Private Segments Name.
* `region` - (Required) Region of deployment.
* `topology_name` - (Required) Name of topology this CloudEOS is part of.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `vpc_id` - (Required) VPC ID in which this CloudEOS is deployed.
* `role` - (Optional) CloudEdge or CloudLeaf (Same as VPC role).
* `is_rr` - (Optional) true if this CloudEOS acts as a Route Reflector.
* `ami` - (Optional) CloudEOS image.
* `key_name` - (Optional) keypair name.
* `availability_zone` - (Optional) Availability Zone of VPC.
* `intf_name` - (Optional) List of interface names.
* `intf_private_ip` - (Required) List of interface private IPs.
* `intf_type` - (Required) List of Interface type (public, private, internal).
