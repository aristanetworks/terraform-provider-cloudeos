// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	cdm "cloudeos-terraform-provider/arista/internal/models"
	cdu "cloudeos-terraform-provider/arista/internal/utils"

	clouddeploy_v1 "cloudeos-terraform-provider/arista/internal/api/clouddeploy.v1"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/terraform/helper/schema"
)

//AristaProvider configuration
type AristaProvider struct {
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

func aristaCvpClient(server string, webToken string) (*Client, error) {
	var u = url.URL{Scheme: "wss", Host: server, Path: "/api/v3/wrpc/"}
	req, _ := http.NewRequest("GET", "https://"+server, nil)
	req.Header.Set("Authorization", "Bearer "+webToken)
	req.URL = &u

	log.Printf("Connecting to : %s", u.String())

	var dialer = websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	ws, resp, err := dialer.Dial(u.String(), req.Header)
	if err != nil {
		log.Printf("Websocket dial failed: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, fmt.Errorf("Received unexpected http response, statusCode: %v", resp.StatusCode)
	}

	log.Printf("Created websocket client :%v", resp)

	defer resp.Body.Close()
	client := &Client{
		wrpcClient: ws,
	}

	return client, nil
}

func getCloudProviderType(d *schema.ResourceData) cdm.CloudProviderType {
	cloudProvider := d.Get("cloud_provider").(string)
	cpType := cdm.CpUnknown
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = cdm.Aws
	case strings.EqualFold("azure", cloudProvider):
		cpType = cdm.Azure
	}
	return cpType
}

func (c *Client) wrpcSend(request *wrpcRequest) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	err := c.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVP : %s",
			request.Params["method"].(string), err)
		return resp, err
	}

	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	// Read response from clouddeploy service
	err = c.wrpcClient.ReadJSON(&resp)
	if err != nil {
		log.Printf("Failed to get %s response from CVP, Error: %v",
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

func getAwsVpcName(d *schema.ResourceData) (string, error) {
	var vpcName string
	if value, ok := d.GetOk("tags"); ok {
		tags := value.(map[string]interface{})
		for k, v := range tags {
			if strings.EqualFold("Name", k) {
				vpcName = v.(string)
			}
		}
	} else {
		return "", fmt.Errorf("Router name not configured in tags")
	}

	return vpcName, nil
}

func getCpTypeAndVpcName(d *schema.ResourceData) (string, cdm.CloudProviderType) {
	var vpcName string
	var cpType cdm.CloudProviderType
	cloudProvider := d.Get("cloud_provider").(string)
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = cdm.Aws
		vpcName, _ = getAwsVpcName(d)
	case strings.EqualFold("azure", cloudProvider):
		cpType = cdm.Azure
		vpcName = d.Get("vnet_name").(string)
	}
	return vpcName, cpType
}

func getRoleType(role string) cdm.RoleType {
	var roleType cdm.RoleType
	switch {
	case strings.EqualFold("CloudEdge", role):
		roleType = cdm.Edge
	case strings.EqualFold("CloudSpine", role):
		roleType = cdm.Spine
	case strings.EqualFold("CloudLeaf", role):
		roleType = cdm.Leaf
	default:
		roleType = cdm.RoleUnknown
	}
	return roleType
}

func (p *AristaProvider) getDeviceEnrollmentToken() (string, error) {
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
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

//AddVpcConfig adds VPC resource to Aeris
func (p *AristaProvider) AddVpcConfig(d *schema.ResourceData) error {
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute AddVpc message")
		return err
	}
	defer client.wrpcClient.Close()

	vpcName, cpType := getCpTypeAndVpcName(d)
	roleType := getRoleType(d.Get("role").(string))

	vpc := cdm.Vpc{
		Name:         vpcName,
		CPType:       cpType,
		Region:       d.Get("region").(string),
		RoleType:     roleType,
		TopologyName: d.Get("topology_name").(string),
		ClosName:     d.Get("clos_name").(string),
		WanName:      d.Get("wan_name").(string),
		Cnps:         map[string]bool{d.Get("cnps").(string): true},
	}

	vpcpb := cdu.ToResourceVpcClient(&vpc)
	if strings.EqualFold(d.Get("role").(string), "CloudLeaf") {
		vpcpb = cdu.ToResourceLeafVpc(&vpc)
	}
	log.Printf("[CVP-INFO]AddVpcDataRequestPb:%#v", vpcpb)

	addVpcRequest := clouddeploy_v1.AddVpcRequest{
		Vpc: vpcpb,
	}

	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + vpcName + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Vpcs",
			"method":  "AddVpc",
			"body":    &addVpcRequest,
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
				if strings.EqualFold(key, "vpc") {
					if vpc, ok := val.(map[string]interface{}); ok {
						for k, v := range vpc {
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

//GetVpc gets vpc which satisfy the filter
func (p *AristaProvider) GetVpc(d *schema.ResourceData) error {
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute GetVpc message")
		return err
	}
	defer client.wrpcClient.Close()

	vpc := cdm.Vpc{
		ID: d.Get("tf_id").(string),
	}

	vpcpb := cdu.ToResourceGetVpc(&vpc)
	getVpcRequest := clouddeploy_v1.GetVpcRequest{
		Vpc: vpcpb,
	}
	log.Printf("[CVP-INFO]GetVpcRequestPb:%s", vpcpb)

	request := wrpcRequest{
		Token:   "RPC_Token_Get_" + d.Get("tf_id").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Vpcs",
			"method":  "GetVpc",
			"body":    &getVpcRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVP : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		log.Printf("Failed to get %s response from CVP, Error: %v",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Received GetVpc Resp: %v", resp)
	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key, val := range res {
				if strings.EqualFold(key, "vpc") {
					if vpc, ok := val.(map[string]interface{}); ok {
						for k, v := range vpc {
							if strings.EqualFold(k, "peer_vpc_cidr") {
								log.Printf("GetVpc peer_vpc_cidr:%s", v)
								if peer, ok := v.(map[string]interface{}); ok {
									for k := range peer {
										err = d.Set("peer_vpc_id", k)
										if err != nil {
											return err
										}
										err = d.Set("peervpcidr", peer[k])
										if err != nil {
											return err
										}
									}
								}
							} else if strings.EqualFold(k, "peer_vpc_info") {
								if peerVpcInfo, ok := v.(map[string]interface{}); ok {
									for k := range peerVpcInfo {
										if strings.EqualFold(k, "peer_rg_name") {
											err = d.Set("peer_rg_name", peerVpcInfo[k])
											if err != nil {
												return err
											}
										} else if strings.EqualFold(k, "peer_vnet_name") {
											err = d.Set("peer_vnet_name", peerVpcInfo[k])
											if err != nil {
												return err
											}
										} else if strings.EqualFold(k, "peer_vnet_id") {
											err = d.Set("peer_vnet_id", peerVpcInfo[k])
											if err != nil {
												return err
											}
										} else if strings.EqualFold(k, "peer_vpc_cidr") {
											if peer, ok :=
												peerVpcInfo[k].(map[string]interface{}); ok {
												for k := range peer {
													err = d.Set("peer_vpc_id", k)
													if err != nil {
														return err
													}
													err = d.Set("peervpcidr", peer[k])
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
					}
				}
			}
		}
	}
	return nil
}

//ListVpc gets all vpc which satisfy the filter
func (p *AristaProvider) ListVpc(d *schema.ResourceData) error {
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute ListVpc message")
		return err
	}
	defer client.wrpcClient.Close()

	vpcName, cpType := getCpTypeAndVpcName(d)
	//roleType := getRoleType(d.Get("role").(string))
	vpc := cdm.Vpc{
		Name:   vpcName,
		CPType: cpType,
		Region: d.Get("region").(string),
		//RoleType: cdm.RoleUnknown, // BUG in clouddeploy
		//RoleType: roleType,
	}

	vpcpb := cdu.ToResourceListVpc(&vpc)
	listVpcRequest := clouddeploy_v1.ListVpcRequest{
		Filter: []*clouddeploy_v1.Vpc{vpcpb},
	}
	log.Printf("[CVP-INFO]ListVpcRequestPb:%s", vpcpb)

	request := wrpcRequest{
		Token:   "RPC_Token_List_" + vpcName + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Vpcs",
			"method":  "ListVpc",
			"body":    &listVpcRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVP : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		log.Printf("Failed to get %s response from CVP, Error: %v",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Received Resp: %v", resp)
	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key, val := range res {
				if strings.EqualFold(key, "vpc") {
					log.Printf("ListVpc vpc:%s", val)
					if vpc, ok := val.(map[string]interface{}); ok {
						for k, v := range vpc {
							//This check will be removed as soon as the App is updated
							//with PeerVpcInfo.
							if strings.EqualFold(k, "peer_vpc_cidr") {
								log.Printf("ListVpc peer_vpc_cidr:%s", v)
								// TODO: Read peer_vpc_cidr from map
								if peer, ok := v.(map[string]interface{}); ok {
									for k := range peer {
										err = d.Set("peer_vpc_id", k)
										if err != nil {
											return err
										}
										err = d.Set("peervpcidr", peer[k])
										if err != nil {
											return err
										}
									}
								}
							} else if strings.EqualFold(k, "peer_vpc_info") {
								if peerVpcInfo, ok := v.(map[string]interface{}); ok {
									for k := range peerVpcInfo {
										if strings.EqualFold(k, "peer_rg_name") {
											err = d.Set("peer_rg_name", peerVpcInfo[k])
											if err != nil {
												return err
											}
										} else if strings.EqualFold(k, "peer_vnet_name") {
											err = d.Set("peer_vnet_name", peerVpcInfo[k])
											if err != nil {
												return err
											}
										} else if strings.EqualFold(k, "peer_vnet_id") {
											err = d.Set("peer_vnet_id", peerVpcInfo[k])
											if err != nil {
												return err
											}
										} else if strings.EqualFold(k, "peer_vpc_cidr") {
											if peer, ok :=
												peerVpcInfo[k].(map[string]interface{}); ok {
												for k := range peer {
													err = d.Set("peer_vpc_id", k)
													if err != nil {
														return err
													}
													err = d.Set("peervpcidr", peer[k])
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
					}
				}
			}
		}
	}
	return nil
}

//CheckVpcPresence check if VPC is created in Aeris status path
func (p *AristaProvider) CheckVpcPresence(d *schema.ResourceData) error {
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute CheckVpcPresence")
		return err
	}
	defer client.wrpcClient.Close()

	vpcID := d.Get("vpc_id").(string)
	cloudProvider := d.Get("cloud_provider").(string)
	var cpType cdm.CloudProviderType
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = cdm.Aws
	case strings.EqualFold("azure", cloudProvider):
		cpType = cdm.Azure
	}
	vpc := cdm.Vpc{
		CPType: cpType,
		Region: d.Get("region").(string),
		VpcID:  vpcID,
	}

	vpcpb := cdu.ToResourceCheckVpc(&vpc)
	listVpcRequest := clouddeploy_v1.ListVpcRequest{
		Filter: []*clouddeploy_v1.Vpc{vpcpb},
	}
	log.Printf("[CVP-INFO]CheckVpcRequestPb:%s", vpcpb)

	request := wrpcRequest{
		Token:   "RPC_Token_List_" + vpcID + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Vpcs",
			"method":  "ListVpc",
			"body":    &listVpcRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVP : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent CheckVpcPresence %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		log.Printf("Failed to get %s response from CVP, Error: %v",
			request.Params["method"].(string), err)
		return err
	}

	log.Printf("Received Resp: %v", resp)
	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key, val := range res {
				if strings.EqualFold(key, "vpc") {
					if vpc, ok := val.(map[string]interface{}); ok {
						for k, v := range vpc {
							if strings.EqualFold(k, "vpc_id") {
								if v.(string) == vpcID {
									return nil
								}
							}
						}
					}
				}
			}
		}
	}
	return errors.New("No response for ListVpc")
}

//AddVpc adds VPC resource to Aeris
func (p *AristaProvider) AddVpc(d *schema.ResourceData) error {
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute AddVpc message")
		return err
	}
	defer client.wrpcClient.Close()

	roleType := getRoleType(d.Get("role").(string))
	var vpcName string

	cloudProvider := d.Get("cloud_provider").(string)
	var cpType cdm.CloudProviderType
	var awsVpcInfo cdm.AwsVpcInfo
	var azrVnetInfo cdm.AzureVnetInfo
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = cdm.Aws
		awsVpcInfo.SecurityGroup = []string{d.Get("security_group_id").(string)}
		awsVpcInfo.Cidr = d.Get("cidr_block").(string)
		awsVpcName, err := getAwsVpcName(d)
		if err != nil {
			return err
		}
		vpcName = awsVpcName
	case strings.EqualFold("azure", cloudProvider):
		cpType = cdm.Azure
		azrVnetInfo.Nsg = []string{d.Get("security_group_id").(string)}
		azrVnetInfo.ResourceGroup = d.Get("rg_name").(string)
		azrVnetInfo.Cidr = d.Get("cidr_block").(string)
		vpcName = d.Get("vnet_name").(string)
	}

	vpc := cdm.Vpc{
		Name:         vpcName,
		ID:           d.Get("tf_id").(string),
		VpcID:        d.Get("vpc_id").(string),
		CPType:       cpType,
		Region:       d.Get("region").(string),
		RoleType:     roleType,
		TopologyName: d.Get("topology_name").(string),
		ClosName:     d.Get("clos_name").(string),
		WanName:      d.Get("wan_name").(string),
		Cnps:         map[string]bool{d.Get("cnps").(string): true},
		AwsVpcInfo:   awsVpcInfo,
		AzVnetInfo:   azrVnetInfo,
		Account:      d.Get("account").(string),
	}

	vpcpb := cdu.ToResourceVpcClient(&vpc)
	if strings.EqualFold(d.Get("role").(string), "CloudLeaf") {
		vpcpb = cdu.ToResourceLeafVpc(&vpc)
	}

	addVpcRequest := clouddeploy_v1.AddVpcRequest{
		Vpc: vpcpb,
	}
	log.Printf("[CVP-INFO]AddVpcRequestPb:%s", vpcpb)

	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + vpcName + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Vpcs",
			"method":  "AddVpc",
			"body":    &addVpcRequest,
		},
	}

	_, err = client.wrpcSend(&request)
	if err != nil {
		return err
	}

	return nil
}

//DeleteVpc deletes VPC resource from Aeris
func (p *AristaProvider) DeleteVpc(d *schema.ResourceData) error {
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute DeleteVpc message")
		return err
	}
	defer client.wrpcClient.Close()

	vpcName, cpType := getCpTypeAndVpcName(d)
	vpc := cdm.Vpc{
		Name:   vpcName,
		ID:     d.Get("tf_id").(string),
		CPType: cpType,
		Region: d.Get("region").(string),
	}

	vpcpb := cdu.ToResourceVpc(&vpc)
	delVpcRequest := clouddeploy_v1.DeleteVpcRequest{
		Vpc: vpcpb,
	}

	request := wrpcRequest{
		Token:   "RPC_Token_Delete_" + vpcName + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Vpcs",
			"method":  "DeleteVpc",
			"body":    &delVpcRequest,
		},
	}

	_, err = client.wrpcSend(&request)
	if err != nil {
		return err
	}
	return nil
}

func getBgpAsn(bgpAsnRange string) (uint32, uint32, error) {
	asnRange := strings.Split(bgpAsnRange, "-")
	asnLow, err := strconv.ParseUint(asnRange[0], 10, 32)
	if err != nil {
		log.Printf("[CVP-ERROR]Can't parse bgp asn")
	}
	asnHigh, err := strconv.ParseUint(asnRange[1], 10, 32)
	if err != nil {
		log.Printf("[CVP-ERROR]Can't parse bgp asn")
	}
	log.Printf("[CVP-INFO]Bgp Asn Range %v - %v", asnLow, asnHigh)
	return uint32(asnLow), uint32(asnHigh), err
}

//ListTopology gets the Topology from Aeris which satisfy the filter
func (p *AristaProvider) ListTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVP-ERROR]Failed to create new client in ListTopology")
		return err
	}
	defer client.wrpcClient.Close()

	topoName := d.Get("topology_name").(string)
	closName := d.Get("clos_name").(string)
	wanName := d.Get("wan_name").(string)

	topoInfo := cdm.TopologyInfo{Name: topoName}
	topoInfoPb := cdu.ToResourceListTopologyInfo(&topoInfo)

	log.Printf("[CVP-INFO]ListTopologyInfoRequestPb:%s", topoInfoPb)

	listTopoInfoRequest := clouddeploy_v1.ListTopologyInfoRequest{
		Filter: []*clouddeploy_v1.TopologyInfo{topoInfoPb},
	}

	request := wrpcRequest{
		Token:   "RPC_Token_List_" + d.Get("topology_name").(string) + "_1",
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Topologyinfos",
			"method":  "ListTopologyInfo",
			"body":    &listTopoInfoRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVP : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	var metaTopoExist bool // true if base topology exists in Aeris
	var closTopoExist bool // true if clos topology exists in Aeris
	var wanTopoExist bool  // true if wan topology exists in Aeris

	resp := make(map[string]interface{})
	for {
		// read respose from clouddeploy
		err = client.wrpcClient.ReadJSON(&resp)
		if err != nil {
			log.Printf("Failed to get %s response from CVP, Error: %v",
				request.Params["method"].(string), err)
			return err
		}
		log.Printf("Received ListTopology Resp: %v", resp)

		// parse response and find the topology/clos/wan name returned
		if res, ok := resp["result"]; ok {
			if res, ok := res.(map[string]interface{}); ok {
				for key, val := range res {
					if strings.EqualFold(key, "topology_info") {
						if topo, ok := val.(map[string]interface{}); ok {
							if topo["name"] == topoName &&
								topo["topo_type"] == "TOPO_INFO_META" {
								metaTopoExist = true
							}
							if topo["name"] == topoName &&
								topo["topo_type"] == "TOPO_INFO_WAN" {
								if wan, ok := topo["wan_info"].(map[string]interface{}); ok {
									if wan["wan_name"] == wanName {
										wanTopoExist = true
									}
								}
							}
							if topo["name"] == topoName &&
								topo["topo_type"] == "TOPO_INFO_CLOS" {
								if clos, ok := topo["clos_info"].(map[string]interface{}); ok {
									if clos["clos_name"] == closName {
										closTopoExist = true
									}
								}
							}
						}
					}
				}
			}
		}
		if _, ok := resp["error"].(string); ok {
			break
		}
	}

	role := d.Get("role").(string)
	var errStr string
	if strings.EqualFold("CloudLeaf", role) {
		if metaTopoExist && closTopoExist {
			return nil
		}
		if !metaTopoExist {
			errStr = errStr + "Resource arista_topology " + topoName + " does not exist. "
		}
		if !closTopoExist {
			errStr = errStr + "Resource arista_clos " + closName + " does not exist."
		}
	} else if strings.EqualFold("CloudEdge", role) {
		if metaTopoExist && wanTopoExist && closTopoExist {
			return nil
		}
		if !metaTopoExist {
			errStr = errStr + "Resource arista_topology " + topoName + " does not exist. "
		}
		if !closTopoExist {
			errStr = errStr + "Resource arista_clos " + closName + " does not exist. "
		}
		if !wanTopoExist {
			errStr = errStr + "Resource arista_wan " + wanName + " does not exist."
		}
	}
	log.Printf("metaTopoExist: %v", metaTopoExist)
	log.Printf("wanTopoExist: %v", wanTopoExist)
	log.Printf("closTopoExist: %v", closTopoExist)
	return errors.New(errStr)
}

//AddTopology adds Topology resource to Aeris
func (p *AristaProvider) AddTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVP-ERROR]Failed to create new client in AddTopology")
		return err
	}
	defer client.wrpcClient.Close()

	// get bgp asn
	asnLow, asnHigh, err := getBgpAsn(d.Get("bgp_asn").(string))
	if err != nil {
		log.Printf("[CVP-ERROR]Failed to parse bgp asn")
		return err
	}

	// get list of managed devices
	var managedDevices []string
	if deviceSet, ok := d.GetOk("eos_managed"); ok {
		for _, device := range deviceSet.(*schema.Set).List() {
			managedDevices = append(managedDevices, device.(string))
		}
	}

	log.Printf("Current provider arista version %s", providerAristaVersion)
	topoInfo := cdm.TopologyInfo{
		Version:             providerAristaVersion,
		Name:                d.Get("topology_name").(string),
		TopoType:            cdm.TopoInfoMeta,
		BgpAsnLow:           asnLow,
		BgpAsnHigh:          asnHigh,
		VtepIPCidr:          d.Get("vtep_ip_cidr").(string),
		TerminAttrIPCidr:    d.Get("terminattr_ip_cidr").(string),
		DpsControlPlaneCidr: d.Get("dps_controlplane_cidr").(string),
		ManagedDevices:      managedDevices,
		CVaaSDomain:         p.cvaasDomain,
		CVaaSServer:         p.server,
	}

	topoInfoPb := cdu.ToResourceTopologyInfo(&topoInfo)
	log.Printf("[CVP-INFO]AddTopologyInfoRequestPb:%s", topoInfoPb)
	addTopoInfoRequest := clouddeploy_v1.AddTopologyInfoRequest{
		TopologyInfo: topoInfoPb,
	}

	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + d.Get("topology_name").(string) + "_1",
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Topologyinfos",
			"method":  "AddTopologyInfo",
			"body":    &addTopoInfoRequest,
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
				if strings.EqualFold(key, "topology_info") {
					if topoInfo, ok := val.(map[string]interface{}); ok {
						for k, v := range topoInfo {
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

//DeleteTopology deletes Topology resource from Aeris
func (p *AristaProvider) DeleteTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVP-ERROR]Failed to create new client in DeleteTopology")
		return err
	}
	defer client.wrpcClient.Close()

	topoInfo := cdm.TopologyInfo{
		Name:     d.Get("topology_name").(string),
		ID:       d.Get("tf_id").(string),
		TopoType: cdm.TopoInfoMeta,
	}

	topoInfoPb := cdu.ToResourceTopologyInfo(&topoInfo)
	delTopoInfoRequest := clouddeploy_v1.DeleteTopologyInfoRequest{
		TopologyInfo: topoInfoPb,
	}

	request := wrpcRequest{
		Token:   "RPC_Token_Delete_" + d.Get("topology_name").(string) + "_1",
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Topologyinfos",
			"method":  "DeleteTopologyInfo",
			"body":    &delTopoInfoRequest,
		},
	}

	_, err = client.wrpcSend(&request)
	if err != nil {
		return err
	}

	return nil
}

//AddClosTopology adds clos Topology resource to Aeris
func (p *AristaProvider) AddClosTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVP-ERROR]Failed to create new client in AddClosTopology")
		return err
	}
	defer client.wrpcClient.Close()

	fabricName := d.Get("fabric").(string)
	fabric := cdm.FabricTypeUnspecified
	if strings.EqualFold("full_mesh", fabricName) {
		fabric = cdm.FullMesh
	} else if strings.EqualFold("hub_spoke", fabricName) {
		fabric = cdm.HubSpoke
	}

	closInfo := cdm.ClosInfo{
		ClosName:         d.Get("name").(string),
		CPType:           cdm.Aws,
		Fabric:           fabric,
		LeafEdgePeering:  d.Get("leaf_to_edge_peering").(bool),
		LeafEdgeIgw:      d.Get("leaf_to_edge_igw").(bool),
		LeafEncryption:   d.Get("leaf_encryption").(bool),
		CvpContainerName: d.Get("cv_container_name").(string),
	}
	topoInfo := cdm.TopologyInfo{
		Name:     d.Get("topology_name").(string),
		TopoType: cdm.TopoInfoClos,
		Clos:     closInfo,
	}

	topoInfoPb := cdu.ToResourceTopologyInfo(&topoInfo)
	addTopoInfoRequest := clouddeploy_v1.AddTopologyInfoRequest{
		TopologyInfo: topoInfoPb,
	}
	log.Printf("[CVP-INFO]AddTopologyInfoRequestPb:%s", topoInfoPb)

	token := d.Get("topology_name").(string) + "_3_" + d.Get("name").(string)
	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + token,
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Topologyinfos",
			"method":  "AddTopologyInfo",
			"body":    &addTopoInfoRequest,
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
				if strings.EqualFold(key, "topology_info") {
					if topoInfo, ok := val.(map[string]interface{}); ok {
						for k, v := range topoInfo {
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

//DeleteClosTopology deletes clos Topology resource from Aeris
func (p *AristaProvider) DeleteClosTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVP-ERROR]Failed to create new client in DeleteClosTopology")
		return err
	}
	defer client.wrpcClient.Close()

	closInfo := cdm.ClosInfo{
		ClosName: d.Get("name").(string),
	}
	topoInfo := cdm.TopologyInfo{
		Name:     d.Get("topology_name").(string),
		ID:       d.Get("tf_id").(string),
		TopoType: cdm.TopoInfoClos,
		Clos:     closInfo,
	}

	topoInfoPb := cdu.ToResourceTopologyInfo(&topoInfo)
	log.Printf("[CVP-INFO]DeleteClosTopology DeleteTopologyInfoRequestPb:%s", topoInfoPb)
	delTopoInfoRequest := clouddeploy_v1.DeleteTopologyInfoRequest{
		TopologyInfo: topoInfoPb,
	}

	token := d.Get("topology_name").(string) + "_3_" + d.Get("name").(string)
	request := wrpcRequest{
		Token:   "RPC_Token_Delete_" + token,
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Topologyinfos",
			"method":  "DeleteTopologyInfo",
			"body":    &delTopoInfoRequest,
		},
	}

	_, err = client.wrpcSend(&request)
	if err != nil {
		return err
	}

	return nil
}

//AddWanTopology adds wan Topology resource to Aeris
func (p *AristaProvider) AddWanTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVP-ERROR]Failed to create new client in AddWanTopology")
		return err
	}
	defer client.wrpcClient.Close()

	wanInfo := cdm.WanInfo{
		WanName:              d.Get("name").(string),
		CPType:               cdm.Aws,
		CvpContainerName:     d.Get("cv_container_name").(string),
		EdgeEdgePeering:      d.Get("edge_to_edge_peering").(bool),
		EdgeEdgeIgw:          d.Get("edge_to_edge_igw").(bool),
		EdgeDedicatedConnect: d.Get("edge_to_edge_dedicated_connect").(bool),
	}
	topoInfo := cdm.TopologyInfo{
		Name:     d.Get("topology_name").(string),
		TopoType: cdm.TopoInfoWan,
		Wan:      wanInfo,
	}

	topoInfoPb := cdu.ToResourceTopologyInfo(&topoInfo)
	addTopoInfoRequest := clouddeploy_v1.AddTopologyInfoRequest{
		TopologyInfo: topoInfoPb,
	}
	log.Printf("[CVP-INFO]AddTopologyInfoRequestPb:%s", topoInfoPb)

	token := d.Get("topology_name").(string) + "_2_" + d.Get("name").(string)
	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + token,
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Topologyinfos",
			"method":  "AddTopologyInfo",
			"body":    &addTopoInfoRequest,
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
				if strings.EqualFold(key, "topology_info") {
					if topoInfo, ok := val.(map[string]interface{}); ok {
						for k, v := range topoInfo {
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

//DeleteWanTopology deletes wan Topology resource from Aeris
func (p *AristaProvider) DeleteWanTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVP-ERROR]Failed to create new client in DeleteClosTopology")
		return err
	}
	defer client.wrpcClient.Close()

	wanInfo := cdm.WanInfo{
		WanName: d.Get("name").(string),
	}
	topoInfo := cdm.TopologyInfo{
		Name:     d.Get("topology_name").(string),
		ID:       d.Get("tf_id").(string),
		TopoType: cdm.TopoInfoWan,
		Wan:      wanInfo,
	}

	topoInfoPb := cdu.ToResourceTopologyInfo(&topoInfo)
	delTopoInfoRequest := clouddeploy_v1.DeleteTopologyInfoRequest{
		TopologyInfo: topoInfoPb,
	}

	token := d.Get("topology_name").(string) + "_2_" + d.Get("name").(string)
	request := wrpcRequest{
		Token:   "RPC_Token_Delete_" + token,
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Topologyinfos",
			"method":  "DeleteTopologyInfo",
			"body":    &delTopoInfoRequest,
		},
	}

	_, err = client.wrpcSend(&request)
	if err != nil {
		return err
	}

	return nil
}

//AddSubnet adds subnet resource to Aeris
func (p *AristaProvider) AddSubnet(d *schema.ResourceData) error {
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute AddSubnet message")
		return err
	}
	defer client.wrpcClient.Close()

	cpName := getCloudProviderType(d)
	sub := cdm.Subnet{
		SubnetID:  d.Get("subnet_id").(string),
		CPType:    cpName,
		CidrBlock: d.Get("cidr_block").(string),
		VpcID:     d.Get("vpc_id").(string),
		Zone:      d.Get("availability_zone").(string),
	}

	subpb := cdu.ToResourceSubnet(&sub)
	addSubnetRequest := clouddeploy_v1.AddSubnetRequest{
		Subnet: subpb,
	}

	log.Printf("AddSubnetRequestPb:%s", subpb)
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
func (p *AristaProvider) DeleteSubnet(d *schema.ResourceData) error {
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute DeleteSubnet message")
		return err
	}
	defer client.wrpcClient.Close()

	cpName := getCloudProviderType(d)
	sub := cdm.Subnet{
		SubnetID: d.Get("subnet_id").(string),
		CPType:   cpName,
		ID:       d.Get("tf_id").(string),
		VpcID:    d.Get("vpc_id").(string),
	}

	subpb := cdu.ToResourceSubnet(&sub)
	delSubnetRequest := clouddeploy_v1.DeleteSubnetRequest{
		Subnet: subpb,
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

func getRouterNameFromSchema(d *schema.ResourceData) (string, error) {
	var routerName string
	if value, ok := d.GetOk("tags"); ok {
		tags := value.(map[string]interface{})
		for k, v := range tags {
			if strings.EqualFold("Name", k) {
				routerName = v.(string)
			}
		}
	} else {
		return "", fmt.Errorf("Router name not configured in tags")
	}

	return routerName, nil
}

func setBootStrapCfg(d *schema.ResourceData, cfg string) error {
	if strings.EqualFold(cfg, "") {
		log.Printf("[WARN]The CloudEOS Router is deployed but without bootstrap configuration")
	}
	//No PayG image for Azure.
	cpType := d.Get("cloud_provider")
	if cpType == "azure" {
		ipsecFile := "AristaTesting-IPSec.json"
		content, err := ioutil.ReadFile(ipsecFile)
		if err != nil {
			log.Printf("Problem reading IpSec license file: %s , %v", ipsecFile, err)
		}
		ipsecLicense := string(content)

		bwFile := "AristaTesting-vEOS.json"
		bwContent, err := ioutil.ReadFile(bwFile)
		if err != nil {
			log.Printf("Problem reading vEOS license file: %s , %v", bwFile, err)
		}
		bwLicense := string(bwContent)
		bootstrapCfg := "%EOS-STARTUP-CONFIG-START%\n" +
			cfg +
			"\n" +
			"%EOS-STARTUP-CONFIG-END%\n" +
			"%LICENSE-IPSEC-START%\n" +
			ipsecLicense +
			"%LICENSE-IPSEC-END%\n" +
			"%LICENSE-BANDWIDTH-START%\n" +
			bwLicense +
			"%LICENSE-BANDWIDTH-END%\n"
		if err := d.Set("bootstrap_cfg", bootstrapCfg); err != nil {
			return fmt.Errorf("Error bootstrap_cfg: %v", err)
		}
	} else if cpType == "aws" {
		bootstrapCfg := "%EOS-STARTUP-CONFIG-START%\n" +
			cfg +
			"\n" +
			"%EOS-STARTUP-CONFIG-END%\n"
		if err := d.Set("bootstrap_cfg", bootstrapCfg); err != nil {
			return fmt.Errorf("Error bootstrap_cfg: %v", err)
		}
	}
	return nil
}

func parseRtrResponse(rtr map[string]interface{}, d *schema.ResourceData) error {
	// Parse the bootstrap_cfg, haRtrId, peerRtTable  from response and set
	// in schema
	var bootstrapCfg string
	var haRtrID string
	var peerRtTblID []string // Internal peer route table ID
	var publicRtTblID []string
	var privateRtTblID []string
	var internalRtTblID []string

	for k, v := range rtr {
		if strings.EqualFold(k, "cv_info") {
			if cvInfo, ok := v.(map[string]interface{}); ok {
				for cvKey, cvVal := range cvInfo {
					if strings.EqualFold(cvKey, "bootstrap_cfg") {
						bootstrapCfg = cvVal.(string)
					}
					if strings.EqualFold(cvKey, "ha_rtr_id") {
						haRtrID = cvVal.(string)
					}
					if strings.EqualFold(cvKey, "peer_vpc_rt_table_id") {
						for _, id := range cvVal.([]interface{}) {
							peerRtTblID = append(peerRtTblID, id.(string))
						}
					}
					if strings.EqualFold(cvKey, "ha_rt_table_ids") {
						if rtTblIDs, ok := cvVal.(map[string]interface{}); ok {
							for rtKey, val := range rtTblIDs {
								if strings.EqualFold(rtKey, "public") {
									for _, id := range val.([]interface{}) {
										publicRtTblID = append(publicRtTblID, id.(string))
									}
								}
								if strings.EqualFold(rtKey, "private") {
									for _, id := range val.([]interface{}) {
										privateRtTblID = append(privateRtTblID, id.(string))
									}
								}
								if strings.EqualFold(rtKey, "internal") {
									for _, id := range val.([]interface{}) {
										internalRtTblID = append(internalRtTblID, id.(string))
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// set bootstrap_cfg
	if err := setBootStrapCfg(d, bootstrapCfg); err != nil {
		return err
	}

	if err := d.Set("ha_rtr_id", haRtrID); err != nil {
		return fmt.Errorf("Not able to set ha_rtr_id: %v", err)
	}
	if err := d.Set("peerroutetableid1", peerRtTblID); err != nil {
		return fmt.Errorf("Not able to set peer route table ID: %v ", err)
	}
	if err := d.Set("public_rt_table_id", publicRtTblID); err != nil {
		return fmt.Errorf("Not able to set public route table id: %v", err)
	}
	if err := d.Set("private_rt_table_id", privateRtTblID); err != nil {
		return fmt.Errorf("Not able to set private route table ID: %v", err)
	}
	if err := d.Set("internal_rt_table_id", internalRtTblID); err != nil {
		return fmt.Errorf("Not able to set internal route table ID: %v", err)
	}
	return nil
}

//ListRouter gets router details from CloudDeploy
func (p *AristaProvider) ListRouter(d *schema.ResourceData) error {
	// create new client
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute ListRouter message")
		return err
	}
	defer client.wrpcClient.Close()

	// Get params needed for Router msg
	routerName, err := getRouterNameFromSchema(d)
	if err != nil {
		return err
	}
	cloudProvider := d.Get("cloud_provider").(string)
	var cpType cdm.CloudProviderType
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = cdm.Aws
	case strings.EqualFold("azure", cloudProvider):
		cpType = cdm.Azure
	}

	rtr := cdm.Router{
		Name:   routerName,
		ID:     d.Get("tf_id").(string),
		VpcID:  d.Get("vpc_id").(string),
		CPType: cpType,
		Region: d.Get("region").(string),
		Cnps:   map[string]bool{d.Get("cnps").(string): true},
	}

	rtrpb := cdu.ToResourceListRouter(&rtr)
	log.Printf("[CVP-INFO]ListRouterRequestPb:%s", rtrpb)

	listRouterRequest := clouddeploy_v1.ListRouterRequest{
		Filter: []*clouddeploy_v1.Router{rtrpb},
	}

	request := wrpcRequest{
		Token:   "RPC_Token_List_" + routerName + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Routers",
			"method":  "ListRouter",
			"body":    &listRouterRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVP : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		return err
	}

	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key, val := range res {
				if strings.EqualFold(key, "router") {
					if rtr, ok := val.(map[string]interface{}); ok {
						err = parseRtrResponse(rtr, d)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	} else {
		// bootstrap_cfg can't be null. This will result in not
		// creation of aws_veos
		if err := setBootStrapCfg(d, ""); err != nil {
			return err
		}
	}
	log.Printf("Received Resp: %v", resp)
	return nil
}

//GetRouter gets router details from CloudDeploy
func (p *AristaProvider) GetRouter(d *schema.ResourceData) error {
	// create new client
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute GetRouter message")
		return err
	}
	defer client.wrpcClient.Close()

	rtr := cdm.Router{
		ID: d.Get("tf_id").(string),
	}

	rtrpb := cdu.ToResourceGetRouter(&rtr)
	log.Printf("[CVP-INFO]GetRouterRequestPb:%s", rtrpb)

	getRouterRequest := clouddeploy_v1.GetRouterRequest{
		Router: rtrpb,
	}

	request := wrpcRequest{
		Token:   "RPC_Token_Get_" + d.Get("tf_id").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Routers",
			"method":  "GetRouter",
			"body":    &getRouterRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVP : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		return err
	}

	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key, val := range res {
				if strings.EqualFold(key, "router") {
					if rtr, ok := val.(map[string]interface{}); ok {
						err = parseRtrResponse(rtr, d)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	} else {
		// bootstrap_cfg can't be null. This will result in not
		// creation of aws_veos
		if err := setBootStrapCfg(d, ""); err != nil {
			return err
		}
	}
	log.Printf("Received GetRouter Resp: %v", resp)
	return nil
}

//AddRouterConfig adds Router resource to Aeris
func (p *AristaProvider) AddRouterConfig(d *schema.ResourceData) error {
	enrollmentToken, err := p.getDeviceEnrollmentToken()
	if err != nil {
		log.Printf("%v", err)
		return err
	}

	// Create new client.
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute AddRouter message")
		return err
	}
	defer client.wrpcClient.Close()

	cloudProvider := d.Get("cloud_provider").(string)
	routerName, err := getRouterNameFromSchema(d)
	if err != nil {
		return err
	}

	var cpType cdm.CloudProviderType
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = cdm.Aws
	case strings.EqualFold("azure", cloudProvider):
		cpType = cdm.Azure
	}

	//Adding Intf Private IP, Type and Name to the first message.
	var intfs []cdm.NetworkInterface
	var intf cdm.NetworkInterface
	intfNameList := d.Get("intf_name").([]interface{})
	privateIPList := d.Get("intf_private_ip").([]interface{})
	intfTypeList := d.Get("intf_type").([]interface{})
	intfCount := len(intfNameList)
	for i := 0; i < intfCount; i++ {
		intf.Name = intfNameList[i].(string)
		intf.PrivateIPAddr = []string{privateIPList[i].(string)}
		switch {
		case strings.EqualFold(intfTypeList[i].(string), "public"):
			intf.IntfType = cdm.IntfPublic
		case strings.EqualFold(intfTypeList[i].(string), "private"):
			intf.IntfType = cdm.IntfPrivate
		case strings.EqualFold(intfTypeList[i].(string), "internal"):
			intf.IntfType = cdm.IntfInternal
		}
		intfs = append(intfs, intf)
	}

	rtr := cdm.Router{
		Name:                  routerName,
		VpcID:                 d.Get("vpc_id").(string),
		CPType:                cpType,
		Region:                d.Get("region").(string),
		Cnps:                  map[string]bool{d.Get("cnps").(string): true},
		DeviceEnrollmentToken: enrollmentToken,
		RouteReflector:        d.Get("is_rr").(bool),
		Intf:                  intfs,
	}

	rtrpb := cdu.ToResourceRouterClient(&rtr)
	addRouterRequest := clouddeploy_v1.AddRouterRequest{
		Router: rtrpb,
	}

	log.Printf("AddRouterRequestPb:%s", rtrpb)
	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + routerName + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Routers",
			"method":  "AddRouter",
			"body":    &addRouterRequest,
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
				if strings.EqualFold(key, "router") {
					if router, ok := val.(map[string]interface{}); ok {
						for k, v := range router {
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

//CheckEdgeRouterPresence checks if a edge router is present
func (p *AristaProvider) CheckEdgeRouterPresence(d *schema.ResourceData) error {
	// Logic to check edge router presence
	//  - Call ListVpc with region, cp_type and role=Edge and get vpc_ids
	//    of all edge vpc's.
	//  - Call ListRouter with edge vpc_id and check if there is any router.
	//  - If we found a router then that router is an edge router.

	// create new client
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute CheckEdgeRouter message")
		return err
	}
	defer client.wrpcClient.Close()

	cloudProvider := d.Get("cloud_provider").(string)
	var cpType cdm.CloudProviderType
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = cdm.Aws
	case strings.EqualFold("azure", cloudProvider):
		cpType = cdm.Azure
	}

	// Code for ListVpc request
	vpc := cdm.Vpc{
		CPType:       cpType,
		Region:       d.Get("region").(string),
		TopologyName: d.Get("topology_name").(string),
	}

	vpcpb := cdu.ToResourceListEdgeVpc(&vpc)
	listVpcRequest := clouddeploy_v1.ListVpcRequest{
		Filter: []*clouddeploy_v1.Vpc{vpcpb},
	}
	log.Printf("[CVP-INFO]ListVpcRequestPb:%s", vpcpb)

	request := wrpcRequest{
		Token:   "RPC_Token_List_" + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Vpcs",
			"method":  "ListVpc",
			"body":    &listVpcRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVP : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	var edgeVpcIDs []string // store the vpc_id of all edge VPC's
	resp1 := make(map[string]interface{})
	for {
		err = client.wrpcClient.ReadJSON(&resp1)
		if err != nil {
			log.Printf("Failed to get %s response from CVP, Error: %v",
				request.Params["method"].(string), err)
			return err
		}
		log.Printf("Received ListVpc for checkEdge resp: %v", resp1)

		// parse response and get vpc_id of edge vpc
		if res, ok := resp1["result"]; ok {
			if res, ok := res.(map[string]interface{}); ok {
				for key, val := range res {
					if strings.EqualFold(key, "vpc") {
						if vpc, ok := val.(map[string]interface{}); ok {
							if vpc["role_type"] == "ROLE_EDGE" {
								edgeVpcIDs = append(edgeVpcIDs, vpc["vpc_id"].(string))
							}
						}
					}
				}
			}
		}
		if _, ok := resp1["error"].(string); ok {
			break
		}
	}

	edgeVpcCount := len(edgeVpcIDs)
	if edgeVpcCount == 0 {
		return errors.New("No edge VPC exists")
	}

	// for each edge VPC check if a leaf router exist
	for _, edgeVpcID := range edgeVpcIDs {
		// Code for ListRouter request
		rtr := cdm.Router{
			VpcID:          edgeVpcID,
			CPType:         cpType,
			Region:         d.Get("region").(string),
			RouteReflector: false,
		}
		rtrpb := cdu.ToResourceListEdgeRouter(&rtr)
		log.Printf("[CVP-INFO]ToResourceListEdgeRouter RequestPb:%s", rtrpb)
		listRouterRequest := clouddeploy_v1.ListRouterRequest{
			Filter: []*clouddeploy_v1.Router{rtrpb},
		}

		request = wrpcRequest{
			Token:   "RPC_Token_List_Edge",
			Command: "serviceRequest",
			Params: map[string]interface{}{
				"service": "clouddeploy.Routers",
				"method":  "ListRouter",
				"body":    &listRouterRequest,
			},
		}

		err = client.wrpcClient.WriteJSON(request)
		if err != nil {
			log.Printf("Failed to send %s request to CVP : %s",
				request.Params["method"].(string), err)
			return err
		}
		log.Printf("Successfully sent %s request for %s",
			request.Params["method"].(string), request.Token)

		var rtrVpcIDs []string // stores vpc_id's of all routers in this region
		resp := make(map[string]interface{})
		for {
			// read response from clouddeploy
			err = client.wrpcClient.ReadJSON(&resp)
			if err != nil {
				log.Printf("Failed to get %s response from CVP, Error: %v",
					request.Params["method"].(string), err)
				return err
			}
			log.Printf("Received ListRouter for checkEdge Resp: %v", resp)

			// parse reponse and get vpc_id
			if res, ok := resp["result"]; ok {
				if res, ok := res.(map[string]interface{}); ok {
					for key, val := range res {
						if strings.EqualFold(key, "router") {
							if rtr, ok := val.(map[string]interface{}); ok {
								for k, v := range rtr {
									if strings.EqualFold(k, "vpc_id") {
										rtrVpcIDs = append(rtrVpcIDs, v.(string))
									}
								}
							}
						}
					}
				}
			}
			if _, ok := resp["error"].(string); ok {
				break
			}
		}

		// check if any rtrVpcIDs is present
		log.Printf("Checking for edge router")
		edgeRtrCount := len(rtrVpcIDs)
		if edgeRtrCount > 0 {
			log.Printf("Found an edge router")
			return nil
		}
	}
	return errors.New("No edge router exists")
}

func getAndCreateRouteTableIDs(d *schema.ResourceData) cdm.RouteTableIDs {
	privateRtTblList := d.Get("private_rt_table_ids").([]interface{})
	internalRtTblList := d.Get("internal_rt_table_ids").([]interface{})
	publicRtTblList := d.Get("public_rt_table_ids").([]interface{})

	priv := make([]string, len(privateRtTblList))
	for i, v := range privateRtTblList {
		priv[i] = fmt.Sprint(v)
	}
	pub := make([]string, len(publicRtTblList))
	for i, v := range publicRtTblList {
		pub[i] = fmt.Sprint(v)
	}
	internal := make([]string, len(internalRtTblList))
	for i, v := range internalRtTblList {
		internal[i] = fmt.Sprint(v)
	}
	var routeTableList cdm.RouteTableIDs
	routeTableList.Public = pub
	routeTableList.Internal = internal
	routeTableList.Private = priv

	return routeTableList
}

//AddRouter adds Router resource to Aeris
func (p *AristaProvider) AddRouter(d *schema.ResourceData) error {
	// Create new client.
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute AddRouter message")
		return err
	}
	defer client.wrpcClient.Close()

	cloudProvider := d.Get("cloud_provider").(string)
	routerName, err := getRouterNameFromSchema(d)
	if err != nil {
		return err
	}

	awsRtrDetail := cdm.AwsRouterDetail{}
	azrRtrDetail := cdm.AzureRouterDetail{}
	var cpType cdm.CloudProviderType
	switch {
	case strings.EqualFold("aws", cloudProvider):
		cpType = cdm.Aws
		awsRtrDetail.AvailabilityZone = d.Get("availability_zone").(string)
		awsRtrDetail.InstanceType = d.Get("instance_type").(string)
	case strings.EqualFold("azure", cloudProvider):
		cpType = cdm.Azure
		azrRtrDetail := new(cdm.AzureRouterDetail)
		azrRtrDetail.AvailabilityZone = d.Get("rg_location").(string)
		azrRtrDetail.ResourceGroup = d.Get("rg_name").(string)
		azrRtrDetail.InstanceType = d.Get("instance_type").(string)
	}

	var intfs []cdm.NetworkInterface
	var intf cdm.NetworkInterface
	publicIP, isPublicIP := d.GetOk("public_ip")
	intfNameList := d.Get("intf_name").([]interface{})
	intfIDList := d.Get("intf_id").([]interface{})
	privateIPList := d.Get("intf_private_ip").([]interface{})
	subnetIDList := d.Get("intf_subnet_id").([]interface{})
	intfTypeList := d.Get("intf_type").([]interface{})
	routeTableList := getAndCreateRouteTableIDs(d)

	intfCount := len(intfNameList)
	for i := 0; i < intfCount; i++ {
		intf.Name = intfNameList[i].(string)
		intf.IntfID = intfIDList[i].(string)
		intf.PrivateIPAddr = []string{privateIPList[i].(string)}
		intf.SubnetID = subnetIDList[i].(string)
		if i == 0 && isPublicIP {
			intf.PublicIPAddr = publicIP.(string)
		} else {
			intf.PublicIPAddr = ""
		}
		switch {
		case strings.EqualFold(intfTypeList[i].(string), "public"):
			intf.IntfType = cdm.IntfPublic
		case strings.EqualFold(intfTypeList[i].(string), "private"):
			intf.IntfType = cdm.IntfPrivate
		case strings.EqualFold(intfTypeList[i].(string), "internal"):
			intf.IntfType = cdm.IntfInternal
		}
		intfs = append(intfs, intf)
	}

	rtr := cdm.Router{
		Name:       routerName,
		ID:         d.Get("tf_id").(string),
		VpcID:      d.Get("vpc_id").(string),
		CPType:     cpType,
		Cnps:       map[string]bool{d.Get("cnps").(string): true},
		Region:     d.Get("region").(string),
		InstanceID: d.Get("instance_id").(string),
		//Tag: d.Get("tag_id"),
		//CVContainer: d.Get("cv_container").(string),
		AzRtrDetail:    azrRtrDetail,
		AwsRtrDetail:   awsRtrDetail,
		DepStatus:      cdm.DepStatusSuccess,
		Intf:           intfs,
		RouteTableIDs:  routeTableList,
		RouteReflector: d.Get("is_rr").(bool),
		HAName:         d.Get("ha_name").(string),
	}

	rtrpb := cdu.ToResourceRouterClient(&rtr)
	addRouterRequest := clouddeploy_v1.AddRouterRequest{
		Router: rtrpb,
	}

	log.Printf("AddRouterRequestPb:%s", rtrpb)
	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + d.Get("instance_id").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Routers",
			"method":  "AddRouter",
			"body":    &addRouterRequest,
		},
	}

	_, err = client.wrpcSend(&request)
	if err != nil {
		return err
	}

	return nil
}

//DeleteRouter deletes Router resource from Aeris
func (p *AristaProvider) DeleteRouter(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvpClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute DeleteRouter message")
		return err
	}
	defer client.wrpcClient.Close()

	cpType := getCloudProviderType(d)
	routerName, err := getRouterNameFromSchema(d)
	if err != nil {
		return err
	}

	rtr := cdm.Router{
		Name:   routerName,
		ID:     d.Get("tf_id").(string),
		VpcID:  d.Get("vpc_id").(string),
		CPType: cpType,
	}
	rtrpb := cdu.ToResourceRouterClient(&rtr)
	delRouterRequest := clouddeploy_v1.DeleteRouterRequest{
		Router: rtrpb,
	}

	request := wrpcRequest{
		Token:   "RPC_Token_Delete_" + d.Get("tf_id").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Routers",
			"method":  "DeleteRouter",
			"body":    &delRouterRequest,
		},
	}

	_, err = client.wrpcSend(&request)
	if err != nil {
		return err
	}

	return nil
}
