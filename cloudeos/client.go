// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	cdv1_api "github.com/aristanetworks/terraform-provider-cloudeos/cloudeos/arista/clouddeploy.v1"
	fmp "github.com/aristanetworks/cloudvision-go/api/fmp"
	rdr "github.com/aristanetworks/cloudvision-go/api/arista/redirector.v1"

	cvgrpc "github.com/aristanetworks/cloudvision-go/grpc"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"

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

const (
	// Retry attempts for grpc connect
	CVaaSRetryCount = 5
	// Time limit for request timeout 3 min
	requestTimeout = 180
)

//CloudeosProvider configuration
type CloudeosProvider struct {
	srvcAcctToken string
	server        string
	cvaasDomain   string
}

func (p *CloudeosProvider) grpcClient() (*grpc.ClientConn, error) {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithMax(5),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(500 * time.Millisecond)),
		grpc_retry.WithCodes(codes.Unavailable),
	}

	return cvgrpc.DialWithToken(context.Background(), p.server+":443", p.srvcAcctToken,
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(opts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)))

}

func (p *CloudeosProvider) httpClient() (*http.Client, error) {
	return &http.Client{}, nil
}

func (p *CloudeosProvider) getAssignment(target string) (string, error) {
	if strings.ToLower(os.Getenv("CLOUDVISION_REGIONAL_REDIRECT")) == "false" {
		return target, nil
	}

        client, err := p.httpClient()
        if err != nil {
                return "", err
        }

        url := fmt.Sprintf("https://%s/api/v3/services/arista.redirector.v1.AssignmentService/GetOne", target)
        requestBody := strings.NewReader(`{"key":{"system_id":"*"}}`)
        req, err := http.NewRequest("POST", url, requestBody)
        if err != nil {
                return "", err
        }

        var bearer = "Bearer " + p.srvcAcctToken
        req.Header.Add("Authorization", bearer)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

        var intfs []interface{}
        err = json.Unmarshal([]byte(body), &intfs)
        if err != nil {
                return "", fmt.Errorf("Failed to unmarshal to interface: %v", err)
        }

        aResp := &rdr.AssignmentResponse{}
        bytes, err := json.Marshal(intfs[0])
        if err != nil {
                return "", fmt.Errorf("Failed to marshal interface: %v", err)
        }
        opts := &protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}
        err = opts.Unmarshal(bytes, aResp)
        if err != nil {
                return "", fmt.Errorf("Failed to unmarshal with protojson: %v", err)
        }

	fmt.Printf("Clusters returned %+v", aResp.Value.Clusters.Values)
        for _, vals := range aResp.Value.Clusters.Values {
                for _, host := range vals.Hosts.Values {
                        return host, nil
                }
        }
        return "", fmt.Errorf("No assignment found for service account token")
}

func (p *CloudeosProvider) getDeviceEnrollmentToken() (string, error) {
	server, err := p.getAssignment(p.server)
	if err != nil || server == "" {
		return "", fmt.Errorf("Failed to get server assignment: %s", err)
	}

	url := fmt.Sprintf("https://%s/api/v3/services/admin.Enrollment/AddEnrollmentToken",
		strings.Split(server, ":")[0])
	var bearer = "Bearer " + p.srvcAcctToken

	// Create a new request using http
	requestBody := strings.NewReader(`{
		"enrollmentToken":{
			"reenrollDevices":["*"],
			"validFor":"2h",
			"groups":[]}}
	`)

	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return "", errors.New("Error creating AddEnrollmentToken http request")
	}

	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	client, err := p.httpClient()
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error while reading the response bytes: %v", err))
	}

	var data []map[string]interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return "", fmt.Errorf("Failed to get enrollment token : %s (%s)", err, body)
	}

	enrollmentTokenMap, ok := data[0]["enrollmentToken"].(map[string]interface{})
	if !ok {
		return "", errors.New("Token key not found in AddEnrollmentToken response")
	}

	return enrollmentTokenMap["token"].(string), nil
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
		if err = d.Set("peer_rg_name", peerVpcInfo.GetPeerRgName().GetValue()); err != nil {
			return err
		}

		if err = d.Set("peer_vnet_name", peerVpcInfo.GetPeerVnetName().GetValue()); err != nil {
			return err
		}

		if err = d.Set("peer_vnet_id", peerVpcInfo.GetPeerVnetId().GetValue()); err != nil {
			return err
		}

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
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute AddSubnet")
		return err
	}

	defer client.Close()
	subnetClient := cdv1_api.NewSubnetConfigServiceClient(client)
	cpName := getCloudProviderType(d)

	subnetKey := cdv1_api.SubnetKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}
	subnet := &cdv1_api.SubnetConfig{
		Key:       &subnetKey,
		SubnetId:  &wrapperspb.StringValue{Value: d.Get("subnet_id").(string)},
		CpT:       cdv1_api.CloudProviderType(cpName),
		Cidr:      &wrapperspb.StringValue{Value: d.Get("cidr_block").(string)},
		VpcId:     &wrapperspb.StringValue{Value: d.Get("vpc_id").(string)},
		AvailZone: &wrapperspb.StringValue{Value: d.Get("availability_zone").(string)},
	}

	addSubnetRequest := cdv1_api.SubnetConfigSetRequest{
		Value: subnet,
	}
	log.Printf("[CVaaS-INFO] AddSubnetRequest: %v", &addSubnetRequest)

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()
	resp, err := subnetClient.Set(ctx, &addSubnetRequest)
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

//DeleteSubnet deletes subnet resource from Aeris
func (p *CloudeosProvider) DeleteSubnet(d *schema.ResourceData) error {
	client, err := p.grpcClient()
	if err != nil {
		log.Printf("Failed to create new CVaaS Grpc client to execute DeleteSubnet")
		return err
	}

	defer client.Close()
	subnetClient := cdv1_api.NewSubnetConfigServiceClient(client)
	subnetKey := cdv1_api.SubnetKey{
		Id: &wrapperspb.StringValue{Value: d.Get("tf_id").(string)},
	}
	delSubnetRequest := cdv1_api.SubnetConfigDeleteRequest{
		Key: &subnetKey,
	}
	log.Printf("[CVaaS-INFO] DeleteSubnetRequest: %v", delSubnetRequest)
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(requestTimeout*time.Second))
	defer cancel()
	resp, err := subnetClient.Delete(ctx, &delSubnetRequest)
	if err != nil && resp != nil && resp.GetKey().GetId().GetValue() != d.Get("tf_id").(string) {
		return fmt.Errorf("Deleted key %v, tf_id %v", resp.GetKey().GetId().GetValue(),
			d.Get("tf_id").(string))
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
