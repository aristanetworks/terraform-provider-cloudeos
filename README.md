# Terraform Provider for Arista CloudEOS

The Arista CloudEOS provider helps with automating the deployment of a multi-cloud network
fabric using Arista CloudVision as a Service ( CVaaS ). The provider interacts with CVaaS to
create a BGP/EVPN/VxLAN based overlay network between CloudEOS Routers running in various
regions across Cloud Providers.

## Terminology

* CVaaS : Arista [CloudVision](https://www.arista.com/en/products/eos/eos-cloudvision) as-a-Service.
  CloudVision as a Service is the root access point for customers to utilize the CloudEOS solution.
  CVaaS supports a single point of orchestration for multi-cloud, multi-tenant and multi-account management.
* CloudEdge - CloudEdge is a instance of CloudEOS that provides interconnection services with other public clouds
  within the client’s autonomous system. The CloudEdge also interconnects VPCs and VNETs within a cloud provider region.
* CloudLeaf - CloudLeaf is an instance of CloudEOS that is deployed in the VPC and VNETs that hosts the applications VMs.
  It is the gateway for all incoming and outgoing traffic for the VPC.
* Cloud Network Private Segment (CNPS) - The VRF name used for segmentation across your cloud network.
* CLOS topology - EPVN based spine-leaf topology to interconnect all leaf VPCs in a region
    to the CloudEdge routers deployed in the transit/Edge VPC.
* WAN topology - EVPN based full mesh topology to interconnect all the CloudEdges over Internet.
* DPS - [Dynamic Path Selection](https://www.arista.com/en/cg-veos-router/veos-router-dynamic-path-selection-overview)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- [Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

## Usage

### CloudEOS Provider
```
provider "cloudeos" {
  cvaas_domain = "apiserver.arista.io"
  cvaas_server = "arista.io"
  service_account_web_token = "..."
}
```

### Argument Reference
* cvaas_domain - (Required) CVaaS Domain name
* cvaas_server - (Required) CVaaS Server Name
* service_account_web_token - (Required) The access token to authenticate the Terraform client to CVaaS.

## Resources
Documentation for the resources supported by the CloudEOS Provider can be found in the [resources](https://github.com/aristanetworks/terraform-provider-cloudeos/tree/master/docs/resources) folder.
