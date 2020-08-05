package cloudeos

import (
	"context"
	"errors"
	"fmt"

	cvgrpc "github.com/aristanetworks/cloudvision-go/grpc"
	"github.com/golang/protobuf/ptypes/wrappers"
	api "github.com/terraform-providers/terraform-provider-cloudeos/cloudeos/internal/api"
	fmp "github.com/terraform-providers/terraform-provider-cloudeos/cloudeos/internal/fmp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func (p *CloudeosProvider) DeleteAwsVpnConfig(d *schema.ResourceData) error {
	client, err := cvgrpc.DialWithToken(context.Background(), p.server+":443", p.srvcAcctToken)
	if err != nil || client == nil {
		return err
	}
	c := api.NewAWSVpnConfigServiceClient(client)
	defer client.Close()

	awsVpnKey := &api.AWSVpnKey{
		TfId: &wrappers.StringValue{Value: d.Get("tf_id").(string)},
	}
	awsVpnConfigDeleteRequest := api.AWSVpnConfigDeleteRequest{
		Key: awsVpnKey,
	}

	resp, err := c.Delete(context.Background(), &awsVpnConfigDeleteRequest)
	if err != nil && resp != nil && resp.Key.GetTfId().GetValue() != d.Get("tf_id").(string) {
		return fmt.Errorf("Deleted key %v, tf_id %v", resp.GetKey().GetTfId().GetValue(),
			d.Get("tf_id").(string))
	}
	return nil

}

func (p *CloudeosProvider) AddAwsVpnConfig(d *schema.ResourceData) error {
	client, err := cvgrpc.DialWithToken(context.Background(), p.server+":443", p.srvcAcctToken)
	if err != nil {
		return err
	}
	c := api.NewAWSVpnConfigServiceClient(client)
	if c == nil {
		return errors.New("Failed to create the CVaaS GRPC client")
	}
	defer client.Close()

	var tunnels []*api.TunnelInfo
	var tunnel1 api.TunnelInfo
	var tunnel2 api.TunnelInfo
	tunnel1Endpoint := d.Get("tunnel1_aws_endpoint_ip").(string)
	tunnel2Endpoint := d.Get("tunnel2_aws_endpoint_ip").(string)
	tunnel1RouterTunnelIp := d.Get("tunnel1_router_overlay_ip").(string)
	tunnel2RouterTunnelIp := d.Get("tunnel2_router_overlay_ip").(string)
	tunnel1BgpAsn := d.Get("tunnel1_bgp_asn").(string)
	tunnel2BgpAsn := d.Get("tunnel2_bgp_asn").(string)
	tunnel1AwsTunnelIp := d.Get("tunnel1_aws_overlay_ip").(string)
	tunnel2AwsTunnelIp := d.Get("tunnel2_aws_overlay_ip").(string)
	tunnel1BgpHoldTime := d.Get("tunnel1_bgp_holdtime").(string)
	tunnel2BgpHoldTime := d.Get("tunnel2_bgp_holdtime").(string)
	tunnel1PresharedKey := d.Get("tunnel1_preshared_key").(string)
	tunnel2PresharedKey := d.Get("tunnel2_preshared_key").(string)
	tunnel1.TunnelAwsEndpointIp = &fmp.IPAddress{Value: tunnel1Endpoint}
	tunnel2.TunnelAwsEndpointIp = &fmp.IPAddress{Value: tunnel2Endpoint}
	tunnel1.TunnelBgpAsn = &wrappers.StringValue{Value: tunnel1BgpAsn}
	tunnel2.TunnelBgpAsn = &wrappers.StringValue{Value: tunnel2BgpAsn}
	tunnel1.TunnelRouterOverlayIp = &fmp.IPAddress{Value: tunnel1RouterTunnelIp}
	tunnel2.TunnelRouterOverlayIp = &fmp.IPAddress{Value: tunnel2RouterTunnelIp}
	tunnel1.TunnelAwsOverlayIp = &fmp.IPAddress{Value: tunnel1AwsTunnelIp}
	tunnel2.TunnelAwsOverlayIp = &fmp.IPAddress{Value: tunnel2AwsTunnelIp}
	tunnel1.TunnelBgpHoldtime = &wrappers.StringValue{Value: tunnel1BgpHoldTime}
	tunnel2.TunnelBgpHoldtime = &wrappers.StringValue{Value: tunnel2BgpHoldTime}
	tunnel1.TunnelPresharedKey = &wrappers.StringValue{Value: tunnel1PresharedKey}
	tunnel2.TunnelPresharedKey = &wrappers.StringValue{Value: tunnel2PresharedKey}
	//Ipsec Info is default and we aren't passing that for now

	tunnels = append(tunnels, &tunnel1)
	tunnels = append(tunnels, &tunnel2)

	tunnelInfoList := &api.TunnelInfoList{
		TunnelInfo: tunnels,
	}
	awsVpnKey := &api.AWSVpnKey{
		TfId: &wrappers.StringValue{Value: d.Get("tf_id").(string)},
	}
	tgwId := d.Get("tgw_id").(string)
	vpnConnectionId := d.Get("vpn_connection_id").(string)
	awsVpnConfigInfo := &api.AWSVpnConfig{
		Key:                awsVpnKey,
		TgwId:              &wrappers.StringValue{Value: tgwId},
		VpnConnectionId:    &wrappers.StringValue{Value: vpnConnectionId},
		CgwId:              &wrappers.StringValue{Value: d.Get("cgw_id").(string)},
		CloudeosRouterId:   &wrappers.StringValue{Value: d.Get("router_id").(string)},
		CloudeosVpcId:      &wrappers.StringValue{Value: d.Get("vpc_id").(string)},
		VpnTgwAttachmentId: &wrappers.StringValue{Value: d.Get("vpn_tgw_attachment_id").(string)},
		Cnps:               &wrappers.StringValue{Value: d.Get("cnps").(string)},
		VpnGatewayId:       &wrappers.StringValue{Value: d.Get("vpn_gateway_id").(string)},
		TunnelInfoList:     tunnelInfoList,
	}

	awsVpnConfigSetRequest := api.AWSVpnConfigSetRequest{
		Value: awsVpnConfigInfo,
	}
	resp, err := c.Set(context.Background(), &awsVpnConfigSetRequest)
	if err != nil && resp == nil {
		return err
	}

	value := resp.Value
	if value != nil && value.GetKey() != nil && value.GetKey().GetTfId() != nil {
		tf_id := value.GetKey().GetTfId().GetValue()
		d.Set("tf_id", tf_id)
	}
	return nil
}
