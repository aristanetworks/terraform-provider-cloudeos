# cloudeos_aws_vpn

The `cloudeos_aws_vpn` resource sends the Ipsec Site-Site VPN connections and attachments information created in AWS to
CVaaS to configure the CloudEOS router with the Ipsec VTI Tunnel in a given CNPS segment ( VRF ). This resource works
for Site-Site VPN connections attached to AWS Transit Gateway Route tables and AWS VPN Gateways.

## Example Usage

```hcl
resource "aws_ec2_transit_gateway" "tgw" {}

resource "aws_customer_gateway" "routerVpnGw" {
  bgp_asn    = 65000
  ip_address = "10.0.0.1"
  type       = "ipsec.1"
}

resource "aws_ec2_transit_gateway_route_table" "rtTable" {
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id
}

resource "aws_vpn_connection" "vpnConn" {
  customer_gateway_id = aws_customer_gateway.routerVpnGw.id
  transit_gateway_id  = aws_ec2_transit_gateway.tgw.id
  type                = "ipsec.1"
}

resource "aws_ec2_transit_gateway_route_table_association" "tgw_rt_association" {
  transit_gateway_attachment_id  = aws_vpn_connection.vpnConn.transit_gateway_attachment_id
  transit_gateway_route_table_id = <tgw-rt-id>
}

resource "cloudeos_aws_vpn" "vpn_config" {
       cgw_id                    = aws_customer_gateway.routerVpnGw.id
       cnps                      = "dev"
       router_id                 = "<cloudeos_router_id>"
       vpn_connection_id         = aws_vpn_connection.vpnConn.id
       tunnel1_aws_endpoint_ip   = aws_vpn_connection.vpnConn.tunnel1_address
       tunnel1_aws_overlay_ip    = aws_vpn_connection.vpnConn.tunnel1_vgw_inside_address
       tunnel1_router_overlay_ip = aws_vpn_connection.vpnConn.tunnel1_cgw_inside_address
       tunnel1_bgp_asn           = aws_vpn_connection.vpnConn.tunnel1_bgp_asn
       tunnel1_bgp_holdtime      = aws_vpn_connection.vpnConn.tunnel1_bgp_holdtime
       tunnel1_preshared_key     = aws_vpn_connection.vpnConn.tunnel1_preshared_key
       tunnel2_aws_endpoint_ip   = aws_vpn_connection.vpnConn.tunnel2_address
       tunnel2_aws_overlay_ip    = aws_vpn_connection.vpnConn.tunnel2_vgw_inside_address
       tunnel2_router_overlay_ip = aws_vpn_connection.vpnConn.tunnel2_cgw_inside_address
       tunnel2_bgp_asn           = aws_vpn_connection.vpnConn.tunnel1_bgp_asn
       tunnel2_bgp_holdtime      = aws_vpn_connection.vpnConn.tunnel2_bgp_holdtime
       tunnel2_preshared_key     = aws_vpn_connection.vpnConn.tunnel2_preshared_key
       tgw_id                    = aws_ec2_transit_gateway.tgw.id
       vpn_gateway_id            = ""
       vpn_tgw_attachment_id     = aws_vpn_connection.vpnConn.transit_gateway_attachment_id

}
```

## Argument Reference
* `cgw_id` - (Required) AWS Customer Gateway ID
* `cnps` - (Required) VRF Segment in which the Ipsec VPN is created.
* `router_id` - (Required) CloudEOS Router to which the AWS Ipsec VPN terminates.
* `vpn_connection_id` - (Required) AWS Site-to-Site VPN Connection ID
* `tunnel1_aws_endpoint_ip` - (Required) AWS Tunnel1 Underlay IP Address
* `tunnel1_aws_overlay_ip` - (Required) VPN Tunnel1 IP address
* `tunnel1_router_overlay_ip` - (Required) CloudEOS Router Tunnel1 IP address
* `tunnel1_bgp_asn` - (Required) AWS VPN Tunnel1 BGP ASN
* `tunnel1_bgp_holdtime` - (Required) VPN Tunnel1 BGP Hold time
* `tunnel1_preshared_key` - (Required) VPN Tunnel1 Ipsec Preshared key
* `tunnel2_aws_endpoint_ip` - (Required) AWS VPN Tunnel2 Underlay IP Address
* `tunnel2_aws_overlay_ip` - (Required) AWS VPN Tunnel2 IP address
* `tunnel2_router_overlay_ip` - (Required) CloudEOS Router Tunnel2 IP address
* `tunnel2_bgp_asn` - (Required) AWS VPN Tunnel2 BGP ASN
* `tunnel2_bgp_holdtime` - (Required) VPN Tunnel2 BGP Hold time
* `tunnel2_preshared_key` - (Required) VPN Tunnel2 Ipsec Preshared key
* `tgw_id` - (Optional) AWS Transit Gateway ID, if the AWS Site-to-Site connection terminates on a TGW.
* `vpn_gateway_id` - (Optional) AWS VPN Gateway ID, if the AWS Site-to-Site connection terminates on a VPN Gateway.
* `vpn_tgw_attachment_id` - (Optional) AWS VPN Transit Gateway Attachment ID

## Attributes Reference

In addition to Arguments listed above - the following Attributes are exported

* `tf_id` - The ID of cloudeos_aws_vpn Resource.



