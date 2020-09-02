# cloudeos_router_config

The `cloudeos_router_config` resource sends CloudEOS router deployment information to CVaaS to obtain the bootstrap config
with which the router will be deployed. The bootstrap configuration is used by the CloudEOS Router
to start streaming to CVaaS using TerminAttr, and provision to a CVaaS container.
A CloudEOS router can act as Route Reflector, an edge router or a leaf router.

## Example Usage

```hcl
resource "aws_vpc" "vpc" {
    cidr_block = "100.0.0.0/16"
}

resource "cloudeos_topology" "topology" {
   topology_name = "topo-test"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "192.168.0.0/24"
   terminattr_ip_cidr = "192.168.1.0/24"
   dps_controlplane_cidr = "192.168.2.0/24"
}

resource "cloudeos_router_config" "cloudeos" {
  cloud_provider = "aws"
  topology_name = cloudeos_topology.topology.topology_name
  role = "CloudEdge"
  cnps = "dev"
  vpc_id = aws_vpc.vpc.id
  region = aws_vpc.vpc.region
  is_rr = false
  intf_name = ["publicIntf", "internalIntf"]
  intf_private_ip = ["10.0.0.101", "10.0.1.101"]
  intf_type = ["public", "internal"]
}
```

## Argument Reference

* `cloud_provider` - (Required) Cloud Provider for this deployment. Supports only aws or azure.
* `vpc_id` - (Required) VPC/VNET ID in which this CloudEOS is deployed.
* `region` - (Required) Region of deployment.
* `topology_name` - (Required) Name of the topology in which this CloudEOS router is deployed in.
* `intf_name` - (Required) List of interface names.
* `intf_private_ip` - (Required) List of interface private IPs. Currently, only supports 1 IP address per interface.
* `intf_type` - (Required) List of Interface type (public, private, internal). A `public` interface has a public IP
                 associated with it. An `internal` interface is the interface which connects the Leaf and Edge routers.
                 And a `private` interface is the default GW interface for all host traffic.
* `cnps` - (Optional) Cloud Network Private Segments Name. ( VRF name )
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `role` - (Optional) CloudEdge or CloudLeaf (Same as VPC role).
* `is_rr` - (Optional) true if this CloudEOS acts as a Route Reflector.
* `ami` - (Optional) CloudEOS image. ( AWS only )
* `key_name` - (Optional) keypair name ( AWS only )
* `availability_zone` - (Optional) Availability Zone of VPC.

## Attributes Reference

In addition to Arguments listed above - the following Attributes are exported

* `ID` - The ID of cloudeos_router_config Resource.
* `bootstrap_cfg` - Bootstrap configuration for the CloudEOS router.
* `peer_routetable_id` - Router table ID of peer.

## Timeouts

* `create` - (Default of 5 minute) Used when creating the cloudeos_config Resource.
* `delete` - (Defaults to 10 minutes) Used when deleting the cloudeos_config Resource.