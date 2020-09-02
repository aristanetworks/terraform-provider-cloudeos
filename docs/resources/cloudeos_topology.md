# `cloudeos_topology`

The CloudEOS provider is used to create a network fabric ( topology ) which spans across multiple
cloud providers and regions. The solution deploys the network fabric using BGP-EVPN, VxLAN, DPS and Ipsec.

To get the desired parameters and requirements from the user about this fabric we have the following
resources: `cloudeos_topology`, `cloudeos_clos` and `cloudeos_wan`.

For example, a deployment which spans across two AWS regions ( us-east-1 and us-west-1 )
and one Azure region ( westus2 ) will need the user to create: one `cloudeos_topology` resource,
one `cloudeos_wan` resource and three `cloudeos_clos` resource.

The `cloudeos_topology` resource created above is then referenced by other `CloudEOS` resources to associate with
a given topology.

#### Note: Two `cloudeos_topology` with the same topology_name cannot be created.

## Example Usage

```hcl
resource "cloudeos_topology" "topology" {
   topology_name = "topo-test"          // topology name
   bgp_asn = "65000-65100"              // BGP ASN range
   vtep_ip_cidr = "1.0.0.0/16"          // VTEP CIDR
   terminattr_ip_cidr = "4.0.0.0/16"    // Terminattr CIDR
   dps_controlplane_cidr = "3.0.0.0/16" // DPS control plane cidr
}
```

## Argument Reference

* `topology_name` - (Required) Name of the topology.
* `bgp_asn` - (Required) A range of BGP ASN’s which would be used to configure CloudEOS instances,
    based on the role and region in which they are being deployed. For example, a CloudEdge and CloudLeaf
    instance in the same region and CLOS will use iBGP and will have the same ASN. Whereas 2 CloudEdge’s
    in different regions use eBGP and will have different ASNs.
* `vtep_ip_cidr` - (Required) CIDR block for VTEP IPs for CloudEOS Routers
* `terminattr_ip_cidr` - (Required) TerminAttr is used by Arista devices to stream Telemetry to CVaaS.
    Every CloudEOS Router needs a unique TerminAttr local IP.
* `dps_controlplane_cidr` - (Required) Each CloudEOS router needs a unique IP for Dynamic Path Selection.
* `eos_managed` - (Optional) List of CloudEOS devices already deployed.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported:

* `ID` - The ID of the cloudeos_topology Resource.

## Timeouts

* `delete` - (Defaults to 5 minutes) Used when deleting the Topology Resource.
