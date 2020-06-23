# Topology Resource

Topology resource contains attributes common across resources in a deployment.
To refer to attributes defined in Topology resource, VPC and CloudEOS resource use
the topology name in their resource definition.

Note: Topology name should be unique across deployment.

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
* `bgp_asn` - (Required) BGP ASN Range.
* `vtep_ip_cidr` - (Required) CIDR block for VTEP IPs on CloudEOS.
* `terminattr_ip_cidr` - (Required) Loopback IP range on CloudEOS.
* `dps_controlplane_cidr` - (Required) CIDR block for TerminAttr IPs on CloudEOS.
* `eos_managed` - (Optional) List of CloudEOS devices already deployed.
