# cloudeos_wan

The `cloudeos_wan` resource is used to provide attributes for the underlay and overlay connectivity
amongst Edges and Route Reflectors which are part of single WAN Network. In a traditional network topology,
a WAN Network includes multiple site/branch/cloud Edges connected through Ipsec VPN and/or private connects.
Similarly here, we have extended that concept which allows you to connect multiple clouds and regions
in a single WAN fabric.

The Edge/RR VPC and Edge CloudEOS router associate with a WAN using its `name`.
It is also possible to create multiple isolated WAN fabrics by creating multiple `cloudeos_wan` resources
and then associating the WAN name to the corresponding resource.


## Example Usage

```hcl
resource "cloudeos_topology" "topology1" {
   topology_name = "topo-test"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "4.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}

resource "cloudeos_wan" "wan" {
   name = "wan-test"                                         // wan name
   topology_name = cloudeos_topology.topology1.topology_name // topology name
   cv_container_name = "CloudEdge"                           // Container name on CVaaS

}
```

## Argument Reference

* `name` - (Required) Name of the wan resource.
* `topology_name` - (Required) Name of the topology this wan resource depends on.
* `cv_container_name` - (Required) CVaaS Container Name to which the CloudEdge Routers
    will be added to.
* `edge_to_edge_igw` - (Optional) Edge to Edge Connectivity through the Internet Gateway.
* `edge_to_edge_peering` - (Optional) Peering across Edge VPC's, default is false.
    ( Not supported yet )
* `edge_to_edge_dedicated_connect` - (Optional) Dedicated connection between two Edge VPC,
    default is false. ( Not Supported yet )

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported

* `ID` - The ID of the cloudeos_wan resource.

## Timeouts

* `delete` - (Defaults to 5 minutes) Used when deleting the Wan Resource.