# CloudEOS Provider

CloudEOS provider is used to automate deployment and configuration of CloudEOS devices in AWS and Azure.

The goals are to
1. Deploy CloudEOS in various roles across AWS and Azure.
2. Automatically configure network connectivity between CloudEOS instances using CNPS (Cloud Native Private Segment).
3. Connect the CloudEOS instances back to CVAAS for management.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- [Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

## Usage Example

CloudEOS Provider Usage
```
# Configure the CloudEOS Provider
provider "cloudeos" {
  cvaas_domain = "..."
  cvaas_server = "..."
  service_account_web_token = "..."
}
```

### Argument Reference
cvaas_domain - (Required) Domain name "apiserver.arista.io".
cvaas_server - (Required) <description>
service_account_web_token - (Required) <decription>


## Resources
Documentation regarding the Resources supported by the CloudEOS Provider can be found in resources directory.

### Resource Dependencies
![CloudEOSResource Dependencies](graph.svg)

