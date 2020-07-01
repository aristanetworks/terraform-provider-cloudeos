# cloudeos_config Resource

cloudeos_config resource sends CloudEOS deployment related info to CVaaS and obtain the bootstrap config
with which the router will be created with. The bootstrap configuration is used by the CloudEOS Router
to start streaming TerminAttr to CVaaS and provision itself to a container.
CloudEOS can act as Route Reflector, an edge router or a leaf router.

## Example Usage

```hcl
resource "aws_vpc" "vpc" {
    cidr_block = "100.0.0.0/16"
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test"          // topology name
   bgp_asn = "65000-65100"              // BGP ASN range
   vtep_ip_cidr = "10.0.0.0/16"          // VTEP CIDR
   terminattr_ip_cidr = "10.1.0.0/16"    // Terminattr CIDR
   dps_controlplane_cidr = "10.2.0.0/16" // DPS control plane cidr
}

resource "cloudeos_router_config" "cloudeos" {
  cloud_provider = "aws"                                   // aws/azure
  topology_name = cloudeos_topology.topology.topology_name // Name of  Topology this CloudEOS belongs
  role = "CloudEdge"                                       // CloudLeaf/CloudEdge
  cnps = "dev"                                             // Cloud Network Private Segments Name
  vpc_id = aws_vpc.vpc.id                                  // VPC/VNET ID in which this CloudEOS is deployed
  region = aws_vpc.vpc.region                              // Region of deployment
  is_rr = false                                            // true if this CloudEOS acts as Route Reflector
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
* `intf_name` - (Required) List of interface names.
* `intf_private_ip` - (Required) List of interface private IPs.
* `intf_type` - (Required) List of Interface type (public, private, internal).
