# cloudeos_wan

The `cloudeos_wan` resource is used to provide attributes for the underlay and overlay connectivity
amongst Edges and Route Reflectors. To refer to attributes defined in Wan resource, Edge VPC and edge CloudEOS use
the wan name in their resource definition.

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

* `ID` - The ID of the Wan Resource.

## Timeouts

* `delete` - (Defaults to 5 minutes) Used when deleting the Wan Resource.