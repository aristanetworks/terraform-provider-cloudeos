# Clos Resource

Clos contains all the common attributes for leaf VPC and leaf CloudEOS resources.
Clos resource always depends on a topology resource.

To refer to attributes defined in Clos resource, leaf VPC and leaf CloudEOS use
the clos name in their resource definition.

Note: Clos name should be unique across deployment.

## Example Usage

```hcl
resource "cloudeos_clos" "clos" {
   name = "clos-test"                                       // Name of clos resource.
   topology_name = cloudeos_topology.topology.topology_name // Topology name.
   cv_container_name = "CloudLeaf"                          // Container name which CloudEOS is a part of.
}
```

## Argument Reference

* `name` - (Required) Clos resource name.
* `topology_name` - (Required) Topology name this clos resource depends on.
* `fabric` - (Optional) full_mesh or hub_spoke, default value is hub_spoke.
* `leaf_to_edge_peering` - (Optional) Leaf to edge VPC peering, default is true.
* `leaf_to_edge_igw` - (Optional) Leaf to edge VPC connection throught Internet Gateway, default is false.
* `leaf_encryption` - (Optional) Default is false.
* `cv_container_name` - (Optional) Container which CloudEOS is a part of, default is 'CloudLeaf'.