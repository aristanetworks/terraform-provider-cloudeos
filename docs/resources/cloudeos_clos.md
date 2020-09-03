# `cloudeos_clos`

The `cloudeos_clos` resource is dependent on the `cloudeos_topology` resource and is used to provide attributes
for the underlay and overlay connectivity for inter-vpc communication between Leaf and Edge routers in the same region.
A `cloudeos_topology` can have multiple `cloudeos_clos` resources, depending on the number of
Leaf-Edge CLOS networks in the entire network topology.

For example, if you want to deploy a Leaf-Edge topology in two AWS regions ( us-east-1 and us-west-1) and one in an Azure region
( westus2 ). You would have to create three `cloudeos_clos` resources, one each for the CLOS network in that region.

To refer to attributes defined in the CLOS resource, leaf VPC and leaf CloudEOS use
the `cloudeos_clos` name in their resource definition.

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

* `name` - (Required) CLOS resource name.
* `topology_name` - (Required) Topology name that this clos resource depends on.
* `cv_container_name` - (Required) CVaaS Configlet Container Name to which the CloudLeaf Routers will be added to.
* `fabric` - (Optional) full_mesh or hub_spoke, default value is `hub_spoke`.
* `leaf_to_edge_peering` - (Optional) Leaf to edge VPC peering, default is `true`.
* `leaf_to_edge_igw` - (Optional) Leaf to edge VPC connection through Internet Gateway, default is `false`.
* `leaf_encryption` - (Optional) Support encryption using Ipsec between Leaf and Edge. Default is `false`.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported

* `ID` - The ID of the cloudeos_clos Resource.

## Timeouts

* `delete` - (Defaults to 5 minutes) Used when deleting the cloudeos_clos Resource.