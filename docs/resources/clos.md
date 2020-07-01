# `cloudeos_clos`

`cloudeos_clos` is a dependent on the `cloudeos_topology` resource and is used to provide attributes
for the underlay and overlay connectivity for inter-vpc communication in the same region.
A `cloudeos_topology` can consist of multiple `cloudeos_clos` resources dependent on the number of
Leaf-Edge CLOS networks there are across multiple regions or Cloud Providers.

For example, if the customer has 2 AWS regions ( us-east-1 and us-west-1) and 1 Azure region
( westus2 ) in which they want to deploy this CloudEOS network. You would have to create
3 `cloudeos_clos` resources, one each of the CLOS networks in that region.

To refer to attributes defined in Clos resource, leaf VPC and leaf CloudEOS use
the clos name in their resource definition.

## Example Usage

```hcl
resource "cloudeos_topology" "topology1" {
   topology_name = "topo-test"
   bgp_asn = "65000-65100"
   vtep_ip_cidr = "1.0.0.0/16"
   terminattr_ip_cidr = "4.0.0.0/16"
   dps_controlplane_cidr = "3.0.0.0/16"
}
resource "cloudeos_clos" "clos" {
   name = "clos-test"
   topology_name = cloudeos_topology.topology1.topology_name
   cv_container_name = "CloudLeaf"
}
```

## Argument Reference

* `name` - (Required) Clos resource name.
* `topology_name` - (Required) Topology name this clos resource depends on.
* `cv_container_name` - (Required) CVaaS Container Name to which the CloudLeaf Routers will be added to.
* `fabric` - (Optional) full_mesh or hub_spoke, default value is `hub_spoke`.
* `leaf_to_edge_peering` - (Optional) Leaf to edge VPC peering, default is `true`.
* `leaf_to_edge_igw` - (Optional) Leaf to edge VPC connection through Internet Gateway, default is `false`.
* `leaf_encryption` - (Optional) Support encryption using Ipsec between Leaf and Edge. Default is `false`.
