// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	api "terraform-provider-cloudeos/cloudeos/arista/api"
	cdv1_api "terraform-provider-cloudeos/cloudeos/arista/clouddeploy.v1"
	fmp "terraform-provider-cloudeos/cloudeos/fmp"

	cvgrpc "github.com/aristanetworks/cloudvision-go/grpc"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Constants used to trim tf_id before setting resource ID
const (
	TopoPrefix   = "ar-topo"
	WanPrefix    = "ar-wan"
	ClosPrefix   = "ar-clos"
	VpcPrefix    = "ar-vpc"
	SubnetPrefix = "ar-snet"
	RtrPrefix    = "ar-rtr"
	AwsVpnPrefix = "ar-aws-vpn"
)

// Retry attempts for wss connect
const CVaaSRetryCount = 5

//CloudeosProvider configuration
type CloudeosProvider struct {
	srvcAcctToken string
	server        string
	cvaasDomain   string
}

//Client struct
type Client struct {
	wrpcClient *websocket.Conn
}

type wrpcRequest struct {
	Token   string
	Command string
	Params  map[string]interface{}
}

func aristaCvaasClient(server string, webToken string) (*Client, error) {
	var u = url.URL{Scheme: "wss", Host: server, Path: "/api/v3/wrpc/"}
	req, _ := http.NewRequest("GET", "https://"+server, nil)
	req.Header.Set("Authorization", "Bearer "+webToken)
	req.URL = &u

	var dialer = websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{}

	var respStatus string
	var connectErr error
	var backoffPeriod time.Duration = 4
	for i := 1; i <= CVaaSRetryCount; i++ {
		log.Printf("Connecting to : %s Attempt %d", u.String(), i)

		ws, resp, err := dialer.Dial(u.String(), req.Header)
		if err == nil {
			log.Printf("Created websocket client :%v", resp)
			defer resp.Body.Close()

			client := &Client{
				wrpcClient: ws,
			}
			return client, nil
		}
		// If the APIServer sends back an HTTP response with status != 101
		// (Websocket Upgrade request rejected), check if it's an authorization
		// issue and then fail. For any other err, log the HTTP response if
		// possible and retry with an increasing backoff
		if err == websocket.ErrBadHandshake {
			log.Printf("Failed connecting to CVaaS. Websocket dial failed: %v", err)

			if resp.StatusCode == http.StatusUnauthorized {
				return nil, fmt.Errorf("Failed connecting to CVaaS, error : %v Status : %s",
					err, resp.Status)
			}
			respStatus = resp.Status
			connectErr = err

			responseDump, err := httputil.DumpResponse(resp, true)
			if err == nil {
				log.Printf("CVaaS response: %q", responseDump)
			}

		} else {
			log.Printf("Failed connecting to CVaas, error : %v", err)
			connectErr = err
		}

		log.Printf("Retrying connection to CVaaS in %d seconds", backoffPeriod)
		time.Sleep(backoffPeriod * time.Second)
		backoffPeriod = backoffPeriod * 2
	}

	// All retry attempts have failed
	if respStatus != "" {
		return nil, fmt.Errorf("Failed connecting to CVaaS, error : %v Status : %s",
			connectErr, respStatus)
	}
	return nil, fmt.Errorf("Failed connecting to CVaaS, error : %v", connectErr)

}

func (c *Client) wrpcSend(request *wrpcRequest) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	err := c.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVaaS : %s",
			request.Params["method"].(string), err)
		return resp, err
	}

	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	// Read response from clouddeploy service
	err = c.wrpcClient.ReadJSON(&resp)
	if err != nil {
		log.Printf("Failed to get %s response from CVaaS, Error: %v",
			request.Params["method"].(string), err)
		return resp, err
	}

	if e, ok := resp["error"].(string); ok {
		return resp, errors.New(e)
	}

	// Read "EOF" response from api server
	resp2 := make(map[string]interface{})
	err = c.wrpcClient.ReadJSON(&resp2)
	log.Printf("Received EOF Resp: %v", resp2)
	if (err != nil) || (resp2["error"].(string) != "EOF") {
		log.Printf("Failed to get EOF response from ApiServer for %s, Error: %v",
			request.Params["method"].(string), err)
		return resp, err
	}

	_, ok := resp["result"].(map[string]interface{})
	if !ok {
		errorMsg := "Error reading result from json response for " +
			request.Params["method"].(string)
		log.Println(errorMsg)
		return resp, errors.New(errorMsg)
	}

	log.Printf("Received success response for %s, Response: %v",
		request.Params["method"].(string), resp)
	return resp, nil
}

// Decide what should be time limit for request timeout
// currently 60 sec
const requestTimeout = 60

func (p *CloudeosProvider) grpcClient() (*grpc.ClientConn, error) {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithMax(5),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithCodes(codes.Unavailable),
	}

	return cvgrpc.DialWithToken(context.Background(), p.server+":443", p.srvcAcctToken,
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(opts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)))

}

func (p *CloudeosProvider) getDeviceEnrollmentToken() (string, error) {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute AddEnrollmentToken message")
		return "", err
	}
	defer client.wrpcClient.Close()

	request := &wrpcRequest{
		Token:   "RPC_Token_AddEnrollmentToken",
		Command: "admin",
		Params: map[string]interface{}{
			"service": "admin.Enrollment",
			"method":  "AddEnrollmentToken",
			"body": map[string]interface{}{
				"enrollmentToken": map[string]interface{}{
					"groups":          []string{},    //any groups (in addition to AllDevices)
					"validFor":        "2h",          //duration of token(max 30 days,default:2hrs)
					"reenrollDevices": []string{"*"}, //allows re-enrollment
				},
			},
		},
	}

	resp, err := client.wrpcSend(request)
	if err != nil {
		return "", err
	}

	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key, val := range res {
				if strings.EqualFold(key, "enrollmentToken") {
					if tokenInfo, ok := val.(map[string]interface{}); ok {
						for k, v := range tokenInfo {
							if strings.EqualFold(k, "token") {
								return v.(string), nil
							}
						}
					}
				}
			}
		}
	}

	return "", errors.New("Token key not found in AddEnrollmentToken response")
}

//IsValidTopoAddition checks if there already exists an entry in CVaaS by
//the given topo name and that clos topo are not added when deploy mode for the
//corresponding meta topo is provision
func (p *CloudeosProvider) IsValidTopoAddition(d *schema.ResourceData,
	topoType string) (bool, error) {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute IsValidTopoAddition")
		return false, err
	}

	defer client.Close()
	topoInfoClient := cdv1_api.NewTopologyInfoConfigServiceClient(client)
	closName := ""
	wanName := ""
	topoName := d.Get("topology_name").(string)
	if topoName == "" {
		return false, fmt.Errorf("Topology name isn't set")
	}
	if topoType == "TOPO_INFO_CLOS" {
		closName = d.Get("name").(string)
	} else if topoType == "TOPO_INFO_WAN" {
		wanName = d.Get("name").(string)
	}
	topoInfo := &cdv1_api.TopologyInfoConfig{
		Name: &wrapperspb.StringValue{Value: topoName},
	}

	getAllTopoInfoRequest := cdv1_api.TopologyInfoConfigStreamRequest{
		PartialEqFilter: []*cdv1_api.TopologyInfoConfig{topoInfo},
	}

	log.Printf("[CVaaS-INFO] GetAllTopologyInfoRequest: %v", getAllTopoInfoRequest)

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()

	stream, err := topoInfoClient.GetAll(ctx, &getAllTopoInfoRequest)
	if err != nil {
		return false, err
	}

	ents := make([]*cdv1_api.TopologyInfoConfig, 0)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, fmt.Errorf("error reading grpc stream: %v", err)
		}
		ents = append(ents, resp.GetValue())
	}

	for _, ent := range ents {
		if ent.GetName().GetValue() == topoName &&
			ent.GetTopoType().String() == topoType {
			if ent.GetWanInfo().GetWanName().GetValue() == wanName {
				return false, fmt.Errorf("cloudeos_wan %s already exists",
					wanName)
			} else if ent.GetClosInfo().GetClosName().GetValue() == closName {
				return false, fmt.Errorf("cloudeos_clos %s already exists",
					wanName)
			} else {
				return false, fmt.Errorf("cloudeos_topology %s already exists",
					topoName)
			}
		}
		// Find the meta topo for the given clos topo (same name). If the
		// deploy mode for meta is provision, disallow addition of the clos,
		// since we only allow wan topo in provision mode
		if ent.GetName().GetValue() == topoName && ent.GetTopoType().String() == "TOPO_INFO_META" &&
			topoType == "TOPO_INFO_CLOS" && ent.GetDeployMode().GetValue() == "provision" {
			return false, fmt.Errorf("cloudeos_clos cannot be associated with"+
				" a cloudeos_topology resource (%s) that has deploy_mode"+
				" as provision", topoName)
		}
	}
	return true, nil
}

// AddVpcConfig adds VPC resource to Aeris
func (p *CloudeosProvider) AddVpcConfig(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute AddVpcConfig")
		return err
	}

	defer client.Close()
	vpcClient := cdv1_api.NewVpcConfigServiceClient(client)
	vpcName, cpType := getCpTypeAndVpcName(d)
	roleType := getRoleType(d.Get("role").(string))
	vpcKey := &cdv1_api.VpcKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	vpc := &cdv1_api.VpcConfig{
		Name:         &wrapperspb.StringValue{Value: vpcName},
		Key:          vpcKey,
		CpT:          cdv1_api.CloudProviderType(cpType),
		Region:       &wrapperspb.StringValue{Value: d.Get("region").(string)},
		RoleType:     cdv1_api.RoleType(roleType),
		TopologyName: &wrapperspb.StringValue{Value: d.Get("topology_name").(string)},
		ClosName:     &wrapperspb.StringValue{Value: d.Get("clos_name").(string)},
		WanName:      &wrapperspb.StringValue{Value: d.Get("wan_name").(string)},
		Cnps:         &wrapperspb.StringValue{Value: d.Get("cnps").(string)},
		DeployMode:   &wrapperspb.StringValue{Value: strings.ToLower(d.Get("deploy_mode").(string))},
	}

	addVpcRequest := cdv1_api.VpcConfigSetRequest{
		Value: vpc,
	}

	log.Printf("[CVaaS-INFO] AddVpcRequest: %v", &addVpcRequest)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()
	resp, err := vpcClient.Set(ctx, &addVpcRequest)
	if err != nil && resp == nil {
		return err
	}

	if resp.GetValue().GetKey().GetId() != nil {
		tf_id := resp.GetValue().GetKey().GetId().GetValue()
		if err = d.Set("tf_id", tf_id); err != nil {
			return err
		}
	}

	return nil
}

//GetVpc gets vpc which satisfy the filter
func (p *CloudeosProvider) GetVpc(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute GetVpc")
		return err
	}

	defer client.Close()
	vpcClient := cdv1_api.NewVpcConfigServiceClient(client)
	vpcKey := &cdv1_api.VpcKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	getVpcRequest := cdv1_api.VpcConfigRequest{
		Key: vpcKey,
	}

	log.Printf("[CVaaS-INFO] GetVpcRequest: %v", &getVpcRequest)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()

	resp, err := vpcClient.GetOne(ctx, &getVpcRequest)
	log.Printf("Received GetVpc Resp: %v", resp)
	if err != nil && resp == nil {
		return err
	}

	if peerVpcInfo := resp.GetValue().GetPeerVpcInfo(); peerVpcInfo != nil {
		// these fields are only required for azure
		if resp.GetValue().GetCpT() == cdv1_api.CloudProviderType_CLOUD_PROVIDER_TYPE_AZURE {
			if err = d.Set("peer_rg_name", peerVpcInfo.GetPeerRgName().GetValue()); err != nil {
				return err
			}

			if err = d.Set("peer_vnet_name", peerVpcInfo.GetPeerVnetName().GetValue()); err != nil {
				return err
			}

			if err = d.Set("peer_vnet_id", peerVpcInfo.GetPeerVnetId().GetValue()); err != nil {
				return err
			}
		} else {
			// these fields are only required for azure
			peerVpcCidrInfoMap := peerVpcInfo.GetPeerVpcCidr().GetValues()
			for k := range peerVpcCidrInfoMap {
				if err = d.Set("peer_vpc_id", k); err != nil {
					return err
				}

				if err = d.Set("peervpcidr", peerVpcCidrInfoMap[k]); err != nil {
					return err
				}

				if err = d.Set("peer_vpc_cidr", peerVpcCidrInfoMap[k]); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

//CheckVpcDeletionStatus returns nil if Vpc doesn't exist
func (p *CloudeosProvider) CheckVpcDeletionStatus(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute CheckVpcDeletionStatus")
		return err
	}

	defer client.Close()
	vpcClient := cdv1_api.NewVpcConfigServiceClient(client)
	vpcKey := &cdv1_api.VpcKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	getVpcRequest := cdv1_api.VpcConfigRequest{
		Key: vpcKey,
	}

	log.Printf("[CVaaS-INFO] GetVpcRequest: %v", &getVpcRequest)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()

	resp, err := vpcClient.GetOne(ctx, &getVpcRequest)
	if err != nil {
		return err
	}

	log.Printf("Received GetVpc Resp: %v", resp)
	vpcExists := false

	// as we are returning an empty resource in case of no objects in aeris
	// we can assume that if key is emptry then there is no vpc
	// object in aeris
	if resp.GetValue().GetKey().GetId().GetValue() != "" {
		vpcExists = true
	}

	log.Printf("vpcExist: %v", vpcExists)
	if vpcExists {
		return errors.New("Vpc resource exists")
	}
	return nil
}

//CheckVpcPresenceAndGetDeployMode checks if VPC is created in Aeris status
//path and returns deploy_mode set for that vpc
func (p *CloudeosProvider) CheckVpcPresenceAndGetDeployMode(
	d *schema.ResourceData) (string, error) {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute CheckVpcPresenceAndGetDeployMode")
		return "", err
	}

	defer client.Close()
	vpcClient := cdv1_api.NewVpcConfigServiceClient(client)
	vpcID := d.Get("vpc_id").(string)
	cpType := getCloudProviderType(d)
	vpc := &cdv1_api.VpcConfig{
		CpT:    cdv1_api.CloudProviderType(cpType),
		Region: &wrapperspb.StringValue{Value: d.Get("region").(string)},
		VpcId:  &wrapperspb.StringValue{Value: vpcID},
	}

	getAllVpcRequest := &cdv1_api.VpcConfigStreamRequest{
		PartialEqFilter: []*cdv1_api.VpcConfig{vpc},
	}

	log.Printf("[CVaaS-INFO] GetAllVpcRequest : %v", getAllVpcRequest)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()

	stream, err := vpcClient.GetAll(ctx, getAllVpcRequest)
	if err != nil {
		return "", err
	}

	ents := make([]*cdv1_api.VpcConfig, 0)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", fmt.Errorf("error reading grpc stream: %v", err)
		}

		ents = append(ents, resp.GetValue())
	}

	for _, ent := range ents {
		if ent.GetVpcId().GetValue() == vpcID {
			deployMode := strings.ToLower(ent.GetDeployMode().GetValue())
			return deployMode, nil
		}
	}

	return "", errors.New("No response for GetAllVpc")
}

//AddVpc adds VPC resource to Aeris
func (p *CloudeosProvider) AddVpc(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute AddVpc")
		return err
	}

	defer client.Close()
	vpcClient := cdv1_api.NewVpcConfigServiceClient(client)
	roleType := getRoleType(d.Get("role").(string))
	vpcName, cpType := getCpTypeAndVpcName(d)

	// Note that the deploy_mode for vpc status MUST be the same as vpc config,
	// resource, ensured by the modules, which use the vpc config resource var
	// to set the deployMode var for vpc status
	vpcKey := &cdv1_api.VpcKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	vpc := &cdv1_api.VpcConfig{
		Name:         &wrapperspb.StringValue{Value: vpcName},
		Key:          vpcKey,
		VpcId:        &wrapperspb.StringValue{Value: d.Get("vpc_id").(string)},
		CpT:          cdv1_api.CloudProviderType(cpType),
		Region:       &wrapperspb.StringValue{Value: d.Get("region").(string)},
		RoleType:     cdv1_api.RoleType(roleType),
		TopologyName: &wrapperspb.StringValue{Value: d.Get("topology_name").(string)},
		ClosName:     &wrapperspb.StringValue{Value: d.Get("clos_name").(string)},
		WanName:      &wrapperspb.StringValue{Value: d.Get("wan_name").(string)},
		Cnps:         &wrapperspb.StringValue{Value: d.Get("cnps").(string)},
		Account:      &wrapperspb.StringValue{Value: d.Get("account").(string)},
		DeployMode:   &wrapperspb.StringValue{Value: strings.ToLower(d.Get("deploy_mode").(string))},
	}

	cloudProvider := d.Get("cloud_provider").(string)
	switch {
	case strings.EqualFold("aws", cloudProvider):
		awsVpcInfo := cdv1_api.AwsVpcInfo{
			SecurityGroup: &fmp.RepeatedString{Values: []string{d.Get("security_group_id").(string)}},
			Cidr:          &wrapperspb.StringValue{Value: d.Get("cidr_block").(string)},
		}
		vpc.AwsVpcInfo = &awsVpcInfo
	case strings.EqualFold("azure", cloudProvider):
		azrVnetInfo := cdv1_api.AzureVnetInfo{
			Nsg:           &fmp.RepeatedString{Values: []string{d.Get("security_group_id").(string)}},
			ResourceGroup: &wrapperspb.StringValue{Value: d.Get("rg_name").(string)},
			Cidr:          &wrapperspb.StringValue{Value: d.Get("cidr_block").(string)},
		}
		vpc.AzVnetInfo = &azrVnetInfo
	}

	addVpcRequest := cdv1_api.VpcConfigSetRequest{
		Value: vpc,
	}

	log.Printf("[CVaaS-INFO] AddVpcRequest: %v", &addVpcRequest)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()

	resp, err := vpcClient.Set(ctx, &addVpcRequest)
	if err != nil && resp == nil {
		return err
	}

	return nil
}

//DeleteVpc deletes VPC resource from Aeris
func (p *CloudeosProvider) DeleteVpc(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute DeleteVpc")
		return err
	}

	defer client.Close()
	vpcClient := cdv1_api.NewVpcConfigServiceClient(client)
	vpcKey := cdv1_api.VpcKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	delVpcRequest := cdv1_api.VpcConfigDeleteRequest{
		Key: &vpcKey,
	}

	log.Printf("[CVaaS-INFO] DeleteVpcRequest: %v", &delVpcRequest)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()

	resp, err := vpcClient.Delete(ctx, &delVpcRequest)
	if err != nil && resp != nil && resp.GetKey().GetId().GetValue() != d.Get("tf_id").(string) {
		return fmt.Errorf("Deleted key %v, tf_id %v", resp.GetKey().GetId().GetValue(),
			d.Get("tf_id").(string))
	}
	return nil
}

// ValidateTopoInfoForAndGetDeployMode -
func (p *CloudeosProvider) ValidateTopoInfoAndGetDeployMode(
	d *schema.ResourceData) (string, error) {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute ValidateTopoInfoAndGetDeployMode")
		return "", err
	}

	defer client.Close()
	topoInfoClient := cdv1_api.NewTopologyInfoConfigServiceClient(client)
	topoName := d.Get("topology_name").(string)
	closName := d.Get("clos_name").(string)
	wanName := d.Get("wan_name").(string)

	topoInfo := &cdv1_api.TopologyInfoConfig{
		Name: &wrapperspb.StringValue{Value: topoName},
	}

	GetAllTopoInfoRequest := cdv1_api.TopologyInfoConfigStreamRequest{
		PartialEqFilter: []*cdv1_api.TopologyInfoConfig{topoInfo},
	}

	log.Printf("[CVaaS-INFO] GetAllTopologyInfoRequest: %v", &GetAllTopoInfoRequest)
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()

	stream, err := topoInfoClient.GetAll(ctx, &GetAllTopoInfoRequest)
	if err != nil {
		return "", err
	}

	var metaTopoExist bool // true if base topology exists in Aeris
	var closTopoExist bool // true if clos topology exists in Aeris
	var wanTopoExist bool  // true if wan topology exists in Aeris

	var topoDeployMode string

	ents := make([]*cdv1_api.TopologyInfoConfig, 0)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading grpc stream: %v", err)
		}
		ents = append(ents, resp.GetValue())
	}

	for _, ent := range ents {
		if ent.GetName().GetValue() == topoName &&
			ent.GetTopoType().String() == "TOPOLOGY_INFO_TYPE_META" {
			metaTopoExist = true
			topoDeployMode = strings.ToLower(ent.GetDeployMode().GetValue())
		}
		if ent.GetName().GetValue() == topoName &&
			ent.GetTopoType().String() == "TOPOLOGY_INFO_TYPE_WAN" {
			if ent.GetWanInfo().GetWanName().GetValue() == wanName {
				wanTopoExist = true
			}
		}
		if ent.GetName().GetValue() == topoName &&
			ent.GetTopoType().String() == "TOPOLOGY_INFO_TYPE_CLOS" {
			if ent.GetClosInfo().GetClosName().GetValue() == closName {
				closTopoExist = true
			}
		}
	}

	role := d.Get("role").(string)
	var errStr string
	// No vpc with role CloudLeaf can be created in provision deploy mode, so we
	// don't do anything special for it
	if strings.EqualFold("CloudLeaf", role) {
		if metaTopoExist && closTopoExist {
			return topoDeployMode, nil
		}
		if !metaTopoExist {
			errStr = errStr + "Resource cloudeos_topology " + topoName + " does not exist. "
		}
		if !closTopoExist {
			errStr = errStr + "Resource cloudeos_clos " + closName + " does not exist."
		}
	} else if strings.EqualFold("CloudEdge", role) {
		// For vpc with role CloudEdge created with deploy_mode provision, we only
		// allow for wan topo, so we skip the closTopo exist check
		if topoDeployMode == "provision" {
			if metaTopoExist && wanTopoExist {
				return topoDeployMode, nil
			}
		} else if metaTopoExist && wanTopoExist && closTopoExist {
			return topoDeployMode, nil
		}

		if !metaTopoExist {
			errStr = errStr + "Resource cloudeos_topology " + topoName + " does not exist. "
		}

		if !wanTopoExist {
			errStr = errStr + "Resource cloudeos_wan " + wanName + " does not exist."
		}

		// Note that if deploy mode = provision, we'll never get here, as desired
		if !closTopoExist {
			errStr = errStr + "Resource cloudeos_clos " + closName + " does not exist. "
		}
	}
	log.Printf("metaTopoExist: %v", metaTopoExist)
	log.Printf("wanTopoExist: %v", wanTopoExist)
	log.Printf("closTopoExist: %v", closTopoExist)
	return "", errors.New(errStr)
}

//CheckTopologyDeletionStatus returns nil if topology doesn't exist
func (p *CloudeosProvider) CheckTopologyDeletionStatus(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute CheckTopologyDeletionStatus")
		return err
	}
	defer client.Close()
	topoInfoClient := cdv1_api.NewTopologyInfoConfigServiceClient(client)
	topoInfoKey := cdv1_api.TopologyInfoKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	getTopoInfoRequest := cdv1_api.TopologyInfoConfigRequest{
		Key: &topoInfoKey,
	}

	log.Printf("[CVaaS-INFO] GetTopologyInfoRequest: %v", getTopoInfoRequest)

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()
	resp, err := topoInfoClient.GetOne(ctx, &getTopoInfoRequest)
	if err != nil && resp == nil {
		return err
	}

	topologyExists := false
	if resp.GetValue().GetTopoType().String() == "TOPO_INFO_META" ||
		resp.GetValue().GetTopoType().String() == "TOPO_INFO_WAN" ||
		resp.GetValue().GetTopoType().String() == "TOPO_INFO_CLOS" {
		topologyExists = true
	}

	log.Printf("topologyExist: %v", topologyExists)
	if topologyExists {
		return errors.New("Topology resource exists")
	}
	return nil
}

//AddTopology adds Topology resource to Aeris
func (p *CloudeosProvider) AddTopology(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute AddTopology")
		return err
	}
	defer client.Close()
	topoInfoClient := cdv1_api.NewTopologyInfoConfigServiceClient(client)
	// bgp_asn is not needed when deploy_mode = 'provision'
	deployMode := d.Get("deploy_mode").(string)
	asnLow, asnHigh, err := getBgpAsn(d.Get("bgp_asn").(string))
	if err != nil && deployMode != "provision" {
		log.Printf("[CVaaS-ERROR]Failed to parse bgp asn")
		return err
	}

	// get list of managed devices
	var managedDevices []string
	if deviceSet, ok := d.GetOk("eos_managed"); ok {
		for _, device := range deviceSet.(*schema.Set).List() {
			managedDevices = append(managedDevices, device.(string))
		}
	}

	log.Printf("Current provider cloudeos version %s", providerCloudEOSVersion)

	topoInfoKey := cdv1_api.TopologyInfoKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	topoInfo := &cdv1_api.TopologyInfoConfig{
		Version:             &wrapperspb.StringValue{Value: providerCloudEOSVersion},
		Name:                &wrapperspb.StringValue{Value: d.Get("topology_name").(string)},
		Key:                 &topoInfoKey,
		TopoType:            cdv1_api.TopologyInfoType_TOPOLOGY_INFO_TYPE_META,
		BgpAsnLow:           &wrapperspb.Int32Value{Value: int32(asnLow)},
		BgpAsnHigh:          &wrapperspb.Int32Value{Value: int32(asnHigh)},
		VtepIpCidr:          &wrapperspb.StringValue{Value: d.Get("vtep_ip_cidr").(string)},
		TerminattrIpCidr:    &wrapperspb.StringValue{Value: d.Get("terminattr_ip_cidr").(string)},
		DpsControlPlaneCidr: &wrapperspb.StringValue{Value: d.Get("dps_controlplane_cidr").(string)},
		ManagedDevices:      &fmp.RepeatedString{Values: managedDevices},
		CvaasDomain:         &wrapperspb.StringValue{Value: p.cvaasDomain},
		CvaasServer:         &wrapperspb.StringValue{Value: p.server},
		DeployMode:          &wrapperspb.StringValue{Value: deployMode},
	}

	addTopoInfoRequest := cdv1_api.TopologyInfoConfigSetRequest{
		Value: topoInfo,
	}

	log.Printf("[CVaaS-INFO] AddTopologyInfoRequest: %v", &addTopoInfoRequest)

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()

	resp, err := topoInfoClient.Set(ctx, &addTopoInfoRequest)
	if err != nil && resp == nil {
		return err
	}

	if resp.GetValue().GetKey().GetId() != nil {
		tf_id := resp.GetValue().GetKey().GetId().GetValue()
		err = d.Set("tf_id", tf_id)
		if err != nil {
			return err
		}
	}
	return nil
}

//DeleteTopology deletes Topology resource from Aeris
func (p *CloudeosProvider) DeleteTopology(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute DeleteTopology")
		return err
	}
	defer client.Close()
	topoInfoClient := cdv1_api.NewTopologyInfoConfigServiceClient(client)
	topoInfoKey := cdv1_api.TopologyInfoKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}
	delTopoInfoRequest := cdv1_api.TopologyInfoConfigDeleteRequest{
		Key: &topoInfoKey,
	}

	log.Printf("[CVaaS-INFO] DeleteTopologyInfoRequest: %v", &delTopoInfoRequest)

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()

	resp, err := topoInfoClient.Delete(ctx, &delTopoInfoRequest)
	if err != nil && resp != nil && resp.GetKey().GetId().GetValue() != d.Get("tf_id").(string) {
		return fmt.Errorf("Deleted key %v, tf_id %v", resp.GetKey().GetId().GetValue(),
			d.Get("tf_id").(string))
	}
	return nil
}

//AddClosTopology adds clos Topology resource to Aeris
func (p *CloudeosProvider) AddClosTopology(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute AddClosTopology")
		return err
	}
	defer client.Close()
	topoInfoClient := cdv1_api.NewTopologyInfoConfigServiceClient(client)
	fabricName := d.Get("fabric").(string)
	fabric := cdv1_api.FabricType_FABRIC_TYPE_UNSPECIFIED
	if strings.EqualFold("full_mesh", fabricName) {
		fabric = cdv1_api.FabricType_FABRIC_TYPE_FULL_MESH
	} else if strings.EqualFold("hub_spoke", fabricName) {
		fabric = cdv1_api.FabricType_FABRIC_TYPE_HUB_SPOKE
	}

	closInfo := &cdv1_api.ClosInfo{
		ClosName:         &wrapperspb.StringValue{Value: d.Get("name").(string)},
		Fabric:           cdv1_api.FabricType(fabric),
		LeafEdgePeering:  &wrapperspb.BoolValue{Value: d.Get("leaf_to_edge_peering").(bool)},
		LeafEdgeIgw:      &wrapperspb.BoolValue{Value: d.Get("leaf_to_edge_igw").(bool)},
		LeafEncryption:   &wrapperspb.BoolValue{Value: d.Get("leaf_encryption").(bool)},
		CvpContainerName: &wrapperspb.StringValue{Value: d.Get("cv_container_name").(string)},
	}

	topoInfoKey := cdv1_api.TopologyInfoKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	topoInfo := &cdv1_api.TopologyInfoConfig{
		Name:     &wrapperspb.StringValue{Value: d.Get("topology_name").(string)},
		Key:      &topoInfoKey,
		TopoType: cdv1_api.TopologyInfoType(cdv1_api.TopologyInfoType_TOPOLOGY_INFO_TYPE_CLOS),
		ClosInfo: closInfo,
	}
	addTopoInfoRequest := cdv1_api.TopologyInfoConfigSetRequest{
		Value: topoInfo,
	}

	log.Printf("[CVaaS-INFO] AddTopologyInfoRequest: %v", &addTopoInfoRequest)
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()
	resp, err := topoInfoClient.Set(ctx, &addTopoInfoRequest)
	if err != nil && resp == nil {
		return err
	}

	if resp.GetValue().GetKey().GetId() != nil {
		tf_id := resp.GetValue().GetKey().GetId().GetValue()
		err = d.Set("tf_id", tf_id)
		if err != nil {
			return err
		}
	}
	return nil
}

//DeleteClosTopology deletes clos Topology resource from Aeris
func (p *CloudeosProvider) DeleteClosTopology(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute DeleteClosTopology")
		return err
	}
	defer client.Close()
	topoInfoClient := cdv1_api.NewTopologyInfoConfigServiceClient(client)

	topoInfoKey := cdv1_api.TopologyInfoKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}
	delTopoInfoRequest := cdv1_api.TopologyInfoConfigDeleteRequest{
		Key: &topoInfoKey,
	}
	log.Printf("[CVaaS-INFO] DeleteClosTopologyInfoRequest: %v", &delTopoInfoRequest)
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()
	resp, err := topoInfoClient.Delete(ctx, &delTopoInfoRequest)
	if err != nil && resp != nil && resp.GetKey().GetId().GetValue() != d.Get("tf_id").(string) {
		return fmt.Errorf("Deleted key %v, tf_id %v", resp.GetKey().GetId().GetValue(),
			d.Get("tf_id").(string))
	}
	return nil
}

//AddWanTopology adds wan Topology resource to Aeris
func (p *CloudeosProvider) AddWanTopology(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute AddWanTopology")
		return err
	}
	defer client.Close()
	topoInfoClient := cdv1_api.NewTopologyInfoConfigServiceClient(client)
	wanInfo := &cdv1_api.WanInfo{
		WanName:              &wrapperspb.StringValue{Value: d.Get("name").(string)},
		EdgeEdgePeering:      &wrapperspb.BoolValue{Value: d.Get("edge_to_edge_peering").(bool)},
		EdgeEdgeIgw:          &wrapperspb.BoolValue{Value: d.Get("edge_to_edge_igw").(bool)},
		EdgeDedicatedConnect: &wrapperspb.BoolValue{Value: d.Get("edge_to_edge_dedicated_connect").(bool)},
		CvpContainerName:     &wrapperspb.StringValue{Value: d.Get("cv_container_name").(string)},
	}

	topoInfoKey := cdv1_api.TopologyInfoKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	topoInfo := &cdv1_api.TopologyInfoConfig{
		Name:     &wrapperspb.StringValue{Value: d.Get("topology_name").(string)},
		Key:      &topoInfoKey,
		TopoType: cdv1_api.TopologyInfoType_TOPOLOGY_INFO_TYPE_WAN,
		WanInfo:  wanInfo,
	}
	addTopoInfoRequest := cdv1_api.TopologyInfoConfigSetRequest{
		Value: topoInfo,
	}
	log.Printf("[CVaaS-INFO] AddWanTopologyInfoRequest: %v", &addTopoInfoRequest)

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()
	resp, err := topoInfoClient.Set(ctx, &addTopoInfoRequest)
	if err != nil && resp == nil {
		return err
	}

	if resp.GetValue().GetKey().GetId() != nil {
		tf_id := resp.GetValue().GetKey().GetId().GetValue()
		err = d.Set("tf_id", tf_id)
		if err != nil {
			return err
		}
	}
	return nil
}

//DeleteWanTopology deletes wan Topology resource from Aeris
func (p *CloudeosProvider) DeleteWanTopology(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute DeleteWanTopology")
		return err
	}
	defer client.Close()
	topoInfoClient := cdv1_api.NewTopologyInfoConfigServiceClient(client)
	topoInfoKey := cdv1_api.TopologyInfoKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}
	delTopoInfoRequest := cdv1_api.TopologyInfoConfigDeleteRequest{
		Key: &topoInfoKey,
	}
	log.Printf("[CVaaS-INFO] DeleteWanTopologyInfoRequest: %v", &delTopoInfoRequest)

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()
	resp, err := topoInfoClient.Delete(ctx, &delTopoInfoRequest)
	if err != nil && resp != nil && resp.GetKey().GetId().GetValue() != d.Get("tf_id").(string) {
		return fmt.Errorf("Deleted key %v, tf_id %v", resp.GetKey().GetId().GetValue(),
			d.Get("tf_id").(string))
	}
	return nil
}

//AddSubnet adds subnet resource to Aeris
func (p *CloudeosProvider) AddSubnet(d *schema.ResourceData) error {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute AddSubnet message")
		return err
	}
	defer client.wrpcClient.Close()

	cpName := getCloudProviderType(d)
	subnet := &api.Subnet{
		SubnetId:  d.Get("subnet_id").(string),
		CpT:       cpName,
		Id:        d.Get("tf_id").(string),
		Cidr:      d.Get("cidr_block").(string),
		VpcId:     d.Get("vpc_id").(string),
		AvailZone: d.Get("availability_zone").(string),
	}

	fieldMask, err := getOuterFieldMask(subnet)
	if err != nil {
		log.Print("AddSubnet: Failed to get field mask")
		return err
	}
	subnet.FieldMask = fieldMask

	addSubnetRequest := api.AddSubnetRequest{
		Subnet: subnet,
	}

	log.Printf("AddSubnetRequestPb:%s", subnet)
	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + d.Get("subnet_id").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Subnets",
			"method":  "AddSubnet",
			"body":    &addSubnetRequest,
		},
	}

	resp, err := client.wrpcSend(&request)
	if err != nil {
		return err
	}

	// Get the primary key, id, from response and set tf_id = id
	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key, val := range res {
				if strings.EqualFold(key, "subnet") {
					if subnet, ok := val.(map[string]interface{}); ok {
						for k, v := range subnet {
							if strings.EqualFold(k, "id") {
								err = d.Set("tf_id", v)
								if err != nil {
									return err
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

//DeleteSubnet deletes subnet resource from Aeris
func (p *CloudeosProvider) DeleteSubnet(d *schema.ResourceData) error {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute DeleteSubnet message")
		return err
	}
	defer client.wrpcClient.Close()

	cpName := getCloudProviderType(d)
	subnet := &api.Subnet{
		SubnetId: d.Get("subnet_id").(string),
		CpT:      cpName,
		Id:       d.Get("tf_id").(string),
		VpcId:    d.Get("vpc_id").(string),
	}

	fieldMask, err := getOuterFieldMask(subnet)
	if err != nil {
		log.Print("DeleteSubnet: Failed to get field mask")
		return err
	}
	subnet.FieldMask = fieldMask

	delSubnetRequest := api.DeleteSubnetRequest{
		Subnet: subnet,
	}

	request := wrpcRequest{
		Token:   "RPC_Token_Delete_" + d.Get("subnet_id").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Subnets",
			"method":  "DeleteSubnet",
			"body":    &delSubnetRequest,
		},
	}

	_, err = client.wrpcSend(&request)
	if err != nil {
		return err
	}

	return nil
}

//GetRouterResponse - Common function to get the response from Clouddeploy.
func (p *CloudeosProvider) GetRouterResponse(d *schema.ResourceData) (*cdv1_api.RouterConfigResponse,
	error) {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("GetRouterResponse: Failed to create new CVaaS Grpc client, err: %v", err)
		return nil, err
	}

	defer client.Close()
	rtrClient := cdv1_api.NewRouterConfigServiceClient(client)

	routerKey := cdv1_api.RouterKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}
	getRouterRequest := cdv1_api.RouterConfigRequest{
		Key: &routerKey,
	}

	log.Printf("[CVaaS-INFO] GetRouterRequest: %v", &getRouterRequest)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()
	return rtrClient.GetOne(ctx, &getRouterRequest)
}

//GetRouter gets router details from CloudDeploy
func (p *CloudeosProvider) GetRouter(d *schema.ResourceData) error {
	// create new client
	resp, err := p.GetRouterResponse(d)
	if err != nil {
		log.Printf("GetRouterResponse returned error: %v", err)
		return err
	}

	log.Printf("[CVaaS-INFO] Received GetRouterResponse: %v", resp)

	if resp.GetValue() != nil {
		if err = parseRtrResponse(resp.GetValue(), d); err != nil {
			return err
		}
	} else {
		// bootstrap_cfg can't be null. This will result in not
		// creation of aws_instance.cloudeosVm
		if err := setBootStrapCfg(d, ""); err != nil {
			return err
		}
	}

	return nil
}

func (p *CloudeosProvider) GetRouterStatusAndSetBgpAsn(d *schema.ResourceData) error {
	resp, err := p.GetRouterResponse(d)
	if err != nil {
		log.Printf("GetRouterResponse returned error: %v", err)
		return err
	}

	log.Printf("[CVaaS-INFO] Received GetRouterResponse: %v", resp)
	routerBgpAsn := fmt.Sprint(resp.GetValue().GetBgpAsn().GetValue())
	if err = d.Set("router_bgp_asn", routerBgpAsn); err != nil {
		return err
	}

	return nil
}

//CheckRouterDeletionStatus returns nil if Router doesn't exist
func (p *CloudeosProvider) CheckRouterDeletionStatus(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("CheckRouterDeletionStatus: Failed to create new CVaaS Grpc client, err: %v", err)
		return err
	}
	defer client.Close()
	rtrClient := cdv1_api.NewRouterConfigServiceClient(client)

	routerKey := cdv1_api.RouterKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}
	getRouterRequest := cdv1_api.RouterConfigRequest{
		Key: &routerKey,
	}
	log.Printf("[CVaaS-INFO] GetRouterRequest: %v", &getRouterRequest)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()

	resp, err := rtrClient.GetOne(ctx, &getRouterRequest)
	if err != nil {
		log.Printf("GetRouterRequest failed, error: %v", err)
		return err
	}

	log.Printf("[CVaaS-INFO] Received GetRouter Resp: %v", resp)
	// In case of object not existing in aeris, the server returns an empty response
	// i.e router protobuf with all fields empty, so checking if key is not present
	// should be sufficient to confirm that router is deleted from aeris
	if resp.GetValue().GetKey().GetId().GetValue() != "" {
		log.Printf("Router Exists")
		return errors.New("Router resource exists")

	}

	return nil
}

//AddRouterConfig adds Router resource to Aeris
func (p *CloudeosProvider) AddRouterConfig(d *schema.ResourceData) error {
	enrollmentToken, err := p.getDeviceEnrollmentToken()
	if err != nil {
		log.Printf("Error getting device enrollment token, error: %v", err)
		return err
	}

	client, err := p.grpcClient()
	if err != nil {
		log.Printf("AddRouterConfig: Failed to create new CVaaS Grpc client, err: %v", err)
		return err
	}
	defer client.Close()

	rtrClient := cdv1_api.NewRouterConfigServiceClient(client)
	routerName, err := getRouterNameFromSchema(d)
	if err != nil {
		log.Printf("Error getting router name from schema, error: %v", err)
		return err
	}

	//Adding Intf Private IP, Type and Name to the first message.
	var intfs []*cdv1_api.NetworkInterface
	intfNameList := d.Get("intf_name").([]interface{})
	privateIPList := d.Get("intf_private_ip").([]interface{})
	intfTypeList := d.Get("intf_type").([]interface{})
	intfCount := len(intfNameList)
	for i := 0; i < intfCount; i++ {
		intf := &cdv1_api.NetworkInterface{
			Name:          &wrapperspb.StringValue{Value: intfNameList[i].(string)},
			PrivateIpAddr: &fmp.RepeatedString{Values: []string{privateIPList[i].(string)}},
		}

		switch {
		case strings.EqualFold(intfTypeList[i].(string), "public"):
			intf.IntfType = cdv1_api.NetworkInterfaceType_NETWORK_INTERFACE_TYPE_PUBLIC
		case strings.EqualFold(intfTypeList[i].(string), "private"):
			intf.IntfType = cdv1_api.NetworkInterfaceType_NETWORK_INTERFACE_TYPE_PRIVATE
		case strings.EqualFold(intfTypeList[i].(string), "internal"):
			intf.IntfType = cdv1_api.NetworkInterfaceType_NETWORK_INTERFACE_TYPE_INTERNAL
		}
		intfs = append(intfs, intf)
	}

	cpType := getCloudProviderType(d)

	rtrKey := &cdv1_api.RouterKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	rtr := &cdv1_api.RouterConfig{
		Name:                  &wrapperspb.StringValue{Value: routerName},
		Key:                   rtrKey,
		VpcId:                 &wrapperspb.StringValue{Value: d.Get("vpc_id").(string)},
		CpT:                   cdv1_api.CloudProviderType(cpType),
		Region:                &wrapperspb.StringValue{Value: d.Get("region").(string)},
		Cnps:                  &wrapperspb.StringValue{Value: d.Get("cnps").(string)},
		DeviceEnrollmentToken: &wrapperspb.StringValue{Value: enrollmentToken},
		RouteReflector:        &wrapperspb.BoolValue{Value: d.Get("is_rr").(bool)},
		Intf:                  &cdv1_api.RepeatedNetworkInterfaces{Values: intfs},
		DeployMode:            &wrapperspb.StringValue{Value: strings.ToLower(d.Get("deploy_mode").(string))},
	}

	addRouterRequest := cdv1_api.RouterConfigSetRequest{
		Value: rtr,
	}
	log.Printf("[CVaaS-INFO] AddRouterRequest: %v", addRouterRequest)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()
	resp, err := rtrClient.Set(ctx, &addRouterRequest)
	if err != nil {
		log.Printf("AddRouterRequestFailed, error: %v", err)
		return err
	}

	log.Printf("[CVaaS-INFO] AddRouterResponse: %v", resp)
	if resp.GetValue().GetKey().GetId() != nil {
		tf_id := resp.GetValue().GetKey().GetId().GetValue()
		if err = d.Set("tf_id", tf_id); err != nil {
			return err
		}
	}
	return nil
}

//CheckEdgeRouterPresence checks if a edge router is present
func (p *CloudeosProvider) CheckEdgeRouterPresence(d *schema.ResourceData) error {
	// Logic to check edge router presence
	//  - Call GetAllVpc with region, cp_type and role=Edge and get vpc_ids
	//    of all edge vpc's.
	//  - Call ListRouter with edge vpc_id and check if there is any router.
	//  - If we found a router then that router is an edge router.

	// create new client
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute CheckEdgeRouterPresence")
		return err
	}

	defer client.Close()
	vpcClient := cdv1_api.NewVpcConfigServiceClient(client)
	cpType := getCloudProviderType(d)
	// Code for GetAllVpc request
	vpc := &cdv1_api.VpcConfig{
		CpT:          cdv1_api.CloudProviderType(cpType),
		Region:       &wrapperspb.StringValue{Value: d.Get("region").(string)},
		TopologyName: &wrapperspb.StringValue{Value: d.Get("topology_name").(string)},
	}

	getAllRequest := &cdv1_api.VpcConfigStreamRequest{
		PartialEqFilter: []*cdv1_api.VpcConfig{vpc},
	}

	log.Printf("[CVaaS-INFO] GetAllRequest: %v", getAllRequest)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()
	stream, err := vpcClient.GetAll(ctx, getAllRequest)
	if err != nil {
		return err
	}

	ents := make([]*cdv1_api.VpcConfig, 0)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("error reading grpc stream: %v", err)
		}

		ents = append(ents, resp.GetValue())
	}

	var edgeVpcIDs []string // store the vpc_id of all edge VPC's
	for _, ent := range ents {
		if ent.GetRoleType().String() == "ROLE_TYPE_EDGE" {
			edgeVpcIDs = append(edgeVpcIDs, ent.GetVpcId().GetValue())
		}
	}

	if len(edgeVpcIDs) == 0 {
		return errors.New("no edge VPC exists")
	}

	// for list routers
	clientRtr, err := p.grpcClient()
	if err != nil {
		log.Printf("CheckEdgeRouterPresence: Failed to create new CVaaS Grpc client, err: %v", err)
		return err
	}
	defer clientRtr.Close()
	cRtr := cdv1_api.NewRouterConfigServiceClient(clientRtr)

	// for each edge VPC check if a leaf router exist
	for _, edgeVpcID := range edgeVpcIDs {
		// Code for ListRouter request
		rtr := &cdv1_api.RouterConfig{
			VpcId:          &wrapperspb.StringValue{Value: edgeVpcID},
			CpT:            cdv1_api.CloudProviderType(cpType),
			Region:         &wrapperspb.StringValue{Value: d.Get("region").(string)},
			RouteReflector: &wrapperspb.BoolValue{Value: false},
		}

		GetAllRouterRequest := &cdv1_api.RouterConfigStreamRequest{
			PartialEqFilter: []*cdv1_api.RouterConfig{rtr},
		}

		log.Printf("[CVaaS-INFO] GetAllRouterRequest: %v", GetAllRouterRequest)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
		defer cancel()
		stream, err := cRtr.GetAll(ctx, GetAllRouterRequest)
		if err != nil {
			return err
		}

		ents := make([]*cdv1_api.RouterConfig, 0)
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error reading grpc stream: %v", err)
			}
			ents = append(ents, resp.GetValue())
		}

		var rtrVpcIDs []string // stores vpc_id's of all routers in this region

		for _, ent := range ents {
			rtrVpcIDs = append(rtrVpcIDs, ent.GetVpcId().GetValue())
		}
		// check if any rtrVpcIDs is present
		log.Printf("Checking for edge router")
		edgeRtrCount := len(rtrVpcIDs)
		if edgeRtrCount > 0 {
			log.Printf("Found an edge router")
			return nil
		}
	}

	return errors.New("no edge router exists")
}

// AddRouter adds Router resource to Aeris
func (p *CloudeosProvider) AddRouter(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("AddRouter: Failed to create new CVaaS Grpc client, err: %v", err)
		return err
	}

	defer client.Close()
	rtrClient := cdv1_api.NewRouterConfigServiceClient(client)

	routerName, err := getRouterNameFromSchema(d)
	if err != nil {
		log.Printf("Error getting router name from schema, err: %v", err)
		return err
	}

	var intfs []*cdv1_api.NetworkInterface
	publicIP, isPublicIP := d.GetOk("public_ip")
	intfNameList := d.Get("intf_name").([]interface{})
	intfIDList := d.Get("intf_id").([]interface{})
	privateIPList := d.Get("intf_private_ip").([]interface{})
	subnetIDList := d.Get("intf_subnet_id").([]interface{})
	intfTypeList := d.Get("intf_type").([]interface{})
	routeTableList := getAndCreateRouteTableIDs(d)

	intfCount := len(intfNameList)
	for i := 0; i < intfCount; i++ {
		intf := &cdv1_api.NetworkInterface{
			Name:          &wrapperspb.StringValue{Value: intfNameList[i].(string)},
			IntfId:        &wrapperspb.StringValue{Value: intfIDList[i].(string)},
			PrivateIpAddr: &fmp.RepeatedString{Values: []string{privateIPList[i].(string)}},
			Subnet:        &wrapperspb.StringValue{Value: subnetIDList[i].(string)},
		}

		if i == 0 && isPublicIP {
			intf.PublicIpAddr = &wrapperspb.StringValue{Value: publicIP.(string)}
		} else {
			intf.PublicIpAddr = &wrapperspb.StringValue{Value: ""}
		}
		switch {
		case strings.EqualFold(intfTypeList[i].(string), "public"):
			intf.IntfType = cdv1_api.NetworkInterfaceType_NETWORK_INTERFACE_TYPE_PUBLIC
		case strings.EqualFold(intfTypeList[i].(string), "private"):
			intf.IntfType = cdv1_api.NetworkInterfaceType_NETWORK_INTERFACE_TYPE_PRIVATE
		case strings.EqualFold(intfTypeList[i].(string), "internal"):
			intf.IntfType = cdv1_api.NetworkInterfaceType_NETWORK_INTERFACE_TYPE_INTERNAL
		}

		intfs = append(intfs, intf)
	}

	cpType := getCloudProviderType(d)
	rtrKey := &cdv1_api.RouterKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}

	rtr := &cdv1_api.RouterConfig{
		Name:       &wrapperspb.StringValue{Value: routerName},
		Key:        rtrKey,
		VpcId:      &wrapperspb.StringValue{Value: d.Get("vpc_id").(string)},
		CpT:        cdv1_api.CloudProviderType(cpType),
		Cnps:       &wrapperspb.StringValue{Value: d.Get("cnps").(string)},
		Region:     &wrapperspb.StringValue{Value: d.Get("region").(string)},
		InstanceId: &wrapperspb.StringValue{Value: d.Get("instance_id").(string)},
		//Tag: d.Get("tag_id"),
		DepStatus:      cdv1_api.DeploymentStatusCode(cdv1_api.DeploymentStatusCode_DEPLOYMENT_STATUS_CODE_SUCCESS),
		Intf:           &cdv1_api.RepeatedNetworkInterfaces{Values: intfs},
		RtTableIds:     routeTableList,
		RouteReflector: &wrapperspb.BoolValue{Value: d.Get("is_rr").(bool)},
		HaName:         &wrapperspb.StringValue{Value: d.Get("ha_name").(string)},
		DeployMode:     &wrapperspb.StringValue{Value: strings.ToLower(d.Get("deploy_mode").(string))},
	}

	cloudProvider := d.Get("cloud_provider").(string)
	switch {
	case strings.EqualFold("aws", cloudProvider):
		awsRtrDetail := cdv1_api.AwsRouterDetail{
			AvailZone:    &wrapperspb.StringValue{Value: d.Get("availability_zone").(string)},
			InstanceType: &wrapperspb.StringValue{Value: d.Get("instance_type").(string)},
		}
		rtr.AwsRtrDetail = &awsRtrDetail
	case strings.EqualFold("azure", cloudProvider):
		azrRtrDetail := cdv1_api.AzureRouterDetail{
			AvailZone:    &wrapperspb.StringValue{Value: d.Get("rg_location").(string)},
			ResGroup:     &wrapperspb.StringValue{Value: d.Get("rg_name").(string)},
			InstanceType: &wrapperspb.StringValue{Value: d.Get("instance_type").(string)},
		}
		rtr.AzRtrDetail = &azrRtrDetail
	}

	addRouterRequest := cdv1_api.RouterConfigSetRequest{
		Value: rtr,
	}

	log.Printf("[CVaaS-INFO] AddRouterRequest: %v", &addRouterRequest)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()
	resp, err := rtrClient.Set(ctx, &addRouterRequest)
	if err != nil {
		log.Printf("AddRouterRequest failed, error: %v", err)
		return err
	}

	log.Printf("[CVaaS-INFO] AddRouterResponse: %v", resp)
	return nil
}

// DeleteRouter deletes Router resource from Aeris
func (p *CloudeosProvider) DeleteRouter(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("DeleteRouter: Failed to create new CVaaS Grpc client, err: %v", err)
		return err
	}
	defer client.Close()
	rtrClient := cdv1_api.NewRouterConfigServiceClient(client)

	routerKey := cdv1_api.RouterKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}
	delRouterRequest := cdv1_api.RouterConfigDeleteRequest{
		Key: &routerKey,
	}
	log.Printf("[CVaaS-INFO] DeleteRouterRequest : %v", delRouterRequest)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(requestTimeout*time.Second))
	defer cancel()

	resp, err := rtrClient.Delete(ctx, &delRouterRequest)
	if err != nil {
		log.Printf("DeleteRouterRequest failed, error: %v", err)
		return err
	}

	log.Printf("[CVaaS-INFO] DeleteRouterResponse: %v", resp)
	// check if deleted resource matches with terraform resource
	if resp.GetKey().GetId().GetValue() != d.Get("tf_id").(string) {
		return fmt.Errorf("Deleted key %v, tf_id %v", resp.GetKey().GetId().GetValue(),
			d.Get("tf_id").(string))
	}
	return nil
}
