# Wan Resource

Wan resource contains all the common attributes for edge VPC and edge CloudEOS resources.
Wan resource always depends on a topology resource.

To refer to attributes defined in Wan resource, Edge VPC and edge CloudEOS use
the wan name in their resource definition.

Note: Wan name should be unique across deployment.

## Example Usage

```hcl
resource "cloudeos_wan" "wan" {
   name = "wan-test"                                        // wan name
   topology_name = cloudeos_topology.topology.topology_name // topology name
   cv_container_name = "CloudEdge"                          // Container name which CloudEOS would be a part of.
}
```

## Argument Reference

* `name` - (Require) Name of the wan resource.
* `topology_name` - (Required) Name of the topology this wan resource depends on.
* `edge_to_edge_peering` - (Optional) Peering across edge VPC's, default is false.
* `edge_to_edge_dedicated_connect` - (Optional) Dedicated connection between two edge VP, default is false.
* `edge_to_edge_igw` - (Optional) Internet Gateway between two edge VPC, default is true.
* `cv_container_name` - (Optional) Container which CloudEOS is a part of, default is 'CloudEdge'.

