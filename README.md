# Terraform Provider for Arista CloudEOS

The Arista CloudEOS provider helps with automating the deployment of a multi-cloud network
fabric using Arista CloudVision as a Service ( CVaaS ). The provider interacts with CVaaS to
create a BGP/EVPN/VxLAN based overlay network between CloudEOS Routers running in various
regions across Cloud Providers.

## Terminology

* CVaaS : Arista [CloudVision](https://www.arista.com/en/products/eos/eos-cloudvision) as a Service
* CloudEdge - A CloudEOS Edge router is typically deployed at the edge of a site (AWS/Azure/GCP region,
    Data center Edge).
* CloudLeaf - A CloudEOS Leaf router is deployed in a application VPC that is connected to a CloudEdge
    using VxLAN/EVPN overlay.
* Cloud Network Private Segment (CNPS) - VRF based segmentation header carried in VXLAN header.
* CLOS topology - EPVN based spine-leaf topology to interconnect all leaf VPCs in a region
    to the CloudEdge routers deployed in transit
* WAN topology - EVPN based full mesh topology to interconnect all the CloudEdges over Internet.
* DPS - [Dynamic Path Selection](https://www.arista.com/en/cg-veos-router/veos-router-dynamic-path-selection-overview)


## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- [Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

## Resources
Documentation regarding the Resources supported by the CloudEOS Provider can be found in the resources/ directory.
