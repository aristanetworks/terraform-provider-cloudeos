// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	api "github.com/terraform-providers/terraform-provider-cloudeos/cloudeos/internal/api"

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

//CheckForTopologyDuplicates check if there already exists an entry in CVaaS
func (p *CloudeosProvider) CheckForTopologyDuplicates(d *schema.ResourceData,
	topoType string) (bool, error) {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new CVaaS client to execute CheckForTopologyDuplicates")
		return true, err
	}
	defer client.wrpcClient.Close()

	closName := ""
	wanName := ""
	topoName := d.Get("topology_name").(string)
	if topoName == "" {
		return true, fmt.Errorf("Topology name isn't set")
	}
	if topoType == "TOPO_INFO_CLOS" {
		closName = d.Get("name").(string)
	} else if topoType == "TOPO_INFO_WAN" {
		wanName = d.Get("name").(string)
	}
	topoInfo := &api.TopologyInfo{
		Name: topoName,
	}

	fieldMask, err := getOuterFieldMask(topoInfo)
	if err != nil {
		log.Print("ListTopology: Failed to get field mask")
		return true, err
	}
	topoInfo.FieldMask = fieldMask

	log.Printf("[CVaaS-INFO]ListTopologyInfoRequestPb:%s", topoInfo)

	listTopoInfoRequest := api.ListTopologyInfoRequest{
		Filter: []*api.TopologyInfo{topoInfo},
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
		return true, fmt.Errorf("Failed to send %s request to CVaaS : %s",
			request.Params["method"].(string), err)
	}

	resp := make(map[string]interface{})
	for {
		// read respose from clouddeploy
		err = client.wrpcClient.ReadJSON(&resp)
		if err != nil {
			return true, fmt.Errorf("Failed to get %s response from CVaaS, Error: %v",
				request.Params["method"].(string), err)
		}
		log.Printf("Received ListTopology Resp: %v", resp)

		if res, ok := resp["result"]; ok {
			if res, ok := res.(map[string]interface{}); ok {
				for key, val := range res {
					if strings.EqualFold(key, "topology_info") {
						if topo, ok := val.(map[string]interface{}); ok {
							if topo["name"] == topoName && topo["topo_type"] == topoType {
								if wan, ok := topo["wan_info"].(map[string]interface{}); ok {
									if wan["wan_name"] == wanName {
										return true, fmt.Errorf("cloudeos_wan %s already exists",
											wanName)
									}
								} else if clos, ok :=
									topo["clos_info"].(map[string]interface{}); ok {
									if clos["clos_name"] == closName {
										return true, fmt.Errorf("cloudeos_clos %s already exists",
											closName)
									}
								} else {
									return true, fmt.Errorf("cloudeos_topology %s already exists",
										topoName)
								}
							}
						}
					}
				}
			} else {
				return true, fmt.Errorf("couldn't parse the ListTopology response from CVaaS")
			}
		}
		if _, ok := resp["error"].(string); ok {
			break
		}
	}
	return false, nil
}

//AddVpcConfig adds VPC resource to Aeris
func (p *CloudeosProvider) AddVpcConfig(d *schema.ResourceData) error {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to connect to CVaaS for AddVpcConfig")
		return err
	}
	defer client.wrpcClient.Close()

	vpcName, cpType := getCpTypeAndVpcName(d)
	roleType := getRoleType(d.Get("role").(string))
	vpc := &api.Vpc{
		Name:         vpcName,
		Id:           d.Get("tf_id").(string),
		CpT:          cpType,
		Region:       d.Get("region").(string),
		RoleType:     roleType,
		TopologyName: d.Get("topology_name").(string),
		ClosName:     d.Get("clos_name").(string),
		WanName:      d.Get("wan_name").(string),
		Cnps:         d.Get("cnps").(string),
	}

	fieldMask, err := getOuterFieldMask(vpc)
	if err != nil {
		log.Print("AddVpcConfig: Failed to get field mask from protobuf")
		return err
	}
	vpc.FieldMask = fieldMask

	log.Printf("[CVaaS-INFO]AddVpcRequestPb:%s", vpc)

	addVpcRequest := api.AddVpcRequest{
		Vpc: vpc,
	}

	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + vpcName + "_" + d.Get("region").(string),
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
func (p *CloudeosProvider) GetVpc(d *schema.ResourceData) error {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute GetVpc message")
		return err
	}
	defer client.wrpcClient.Close()

	vpc := &api.Vpc{
		Id: d.Get("tf_id").(string),
	}

	fieldMask, err := getOuterFieldMask(vpc)
	if err != nil {
		log.Print("GetVpc: Failed to get field mask")
		return err
	}
	vpc.FieldMask = fieldMask

	getVpcRequest := api.GetVpcRequest{
		Vpc: vpc,
	}
	log.Printf("[CVaaS-INFO]GetVpcRequestPb:%s", vpc)

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
		log.Printf("Failed to send %s request to CVaaS : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		log.Printf("Failed to get %s response from CVaaS, Error: %v",
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
										err = d.Set("peer_vpc_cidr", peer[k])
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

//CheckVpcDeletionStatus returns nil if Vpc doesn't exist
func (p *CloudeosProvider) CheckVpcDeletionStatus(d *schema.ResourceData) error {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client in CheckVpcDeletionStatus")
		return err
	}
	defer client.wrpcClient.Close()

	vpc := &api.Vpc{
		Id: d.Get("tf_id").(string),
	}

	fieldMask, err := getOuterFieldMask(vpc)
	if err != nil {
		log.Print("GetVpc: Failed to get field mask")
		return err
	}

	vpc.FieldMask = fieldMask

	getVpcRequest := api.GetVpcRequest{
		Vpc: vpc,
	}
	log.Printf("[CVaaS-INFO]GetVpcRequestPb:%s", vpc)

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
		log.Printf("Failed to send %s request to CVaaS : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		log.Printf("Failed to get %s response from CVaaS, Error: %v",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Received GetVpc Resp: %v", resp)

	vpcExists := false
	/* A response with no VPC looks like:
	   map[error:rpc error: code = NotFound desc = did not find resource "xxx"
	       status:map[code:5 message:did not find resource "xxx"] token: ... ] */
	// parse response to check if Vpc exist
	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key := range res {
				if strings.EqualFold(key, "vpc") {
					vpcExists = true
				}
			}
		}
	}

	log.Printf("vpcExist: %v", vpcExists)
	if vpcExists {
		return errors.New("Vpc resource exists")
	}
	return nil
}

//ListVpc gets all vpc which satisfy the filter
func (p *CloudeosProvider) ListVpc(d *schema.ResourceData) error {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute ListVpc message")
		return err
	}
	defer client.wrpcClient.Close()

	vpcName, cpType := getCpTypeAndVpcName(d)
	//roleType := getRoleType(d.Get("role").(string))
	vpc := &api.Vpc{
		Name:   vpcName,
		CpT:    cpType,
		Region: d.Get("region").(string),
		//RoleType: api.RoleUnknown, // BUG in clouddeploy
		//RoleType: roleType,
	}

	fieldMask, err := getOuterFieldMask(vpc)
	if err != nil {
		log.Print("ListVpc: Failed to get field mask")
		return err
	}
	vpc.FieldMask = fieldMask

	listVpcRequest := api.ListVpcRequest{
		Filter: []*api.Vpc{vpc},
	}
	log.Printf("[CVaaS-INFO]ListVpcRequestPb:%s", vpc)

	request := wrpcRequest{
		Token:   "RPC_Token_List_" + vpcName + "_" + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Vpcs",
			"method":  "ListVpc",
			"body":    &listVpcRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVaaS : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		log.Printf("Failed to get %s response from CVaaS, Error: %v",
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
func (p *CloudeosProvider) CheckVpcPresence(d *schema.ResourceData) error {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute CheckVpcPresence")
		return err
	}
	defer client.wrpcClient.Close()

	vpcID := d.Get("vpc_id").(string)
	cpType := getCloudProviderType(d)
	vpc := &api.Vpc{
		CpT:    cpType,
		Region: d.Get("region").(string),
		VpcId:  vpcID,
	}

	fieldMask, err := getOuterFieldMask(vpc)
	if err != nil {
		log.Print("CheckVpcPresence: Failed to get field mask")
		return err
	}
	vpc.FieldMask = fieldMask

	listVpcRequest := api.ListVpcRequest{
		Filter: []*api.Vpc{vpc},
	}
	log.Printf("[CVaaS-INFO]CheckVpcRequestPb:%s", vpc)

	request := wrpcRequest{
		Token:   "RPC_Token_List_" + vpcID + "_" + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Vpcs",
			"method":  "ListVpc",
			"body":    &listVpcRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVaaS : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent CheckVpcPresence %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		log.Printf("Failed to get %s response from CVaaS, Error: %v",
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
func (p *CloudeosProvider) AddVpc(d *schema.ResourceData) error {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute AddVpc message")
		return err
	}
	defer client.wrpcClient.Close()

	roleType := getRoleType(d.Get("role").(string))
	vpcName, cpType := getCpTypeAndVpcName(d)
	vpc := &api.Vpc{
		Name:         vpcName,
		Id:           d.Get("tf_id").(string),
		VpcId:        d.Get("vpc_id").(string),
		CpT:          cpType,
		Region:       d.Get("region").(string),
		RoleType:     roleType,
		TopologyName: d.Get("topology_name").(string),
		ClosName:     d.Get("clos_name").(string),
		WanName:      d.Get("wan_name").(string),
		Cnps:         d.Get("cnps").(string),
		Account:      d.Get("account").(string),
	}

	fieldMask, err := getOuterFieldMask(vpc)
	if err != nil {
		log.Print("AddVpc: Failed to get field mask")
		return err
	}

	var awsVpcInfo api.AwsVpcInfo
	var azrVnetInfo api.AzureVnetInfo
	cloudProvider := d.Get("cloud_provider").(string)
	switch {
	case strings.EqualFold("aws", cloudProvider):
		awsVpcInfo.SecurityGroup = []string{d.Get("security_group_id").(string)}
		awsVpcInfo.Cidr = d.Get("cidr_block").(string)
		vpc.AwsVpcInfo = &awsVpcInfo
		err = appendInnerFieldMask(&awsVpcInfo, fieldMask, "awsVpcInfo.")
		if err != nil {
			log.Print("AddVpc: Failed to append inner field mask for AwsVpcInfo")
			return err
		}
	case strings.EqualFold("azure", cloudProvider):
		azrVnetInfo.Nsg = []string{d.Get("security_group_id").(string)}
		azrVnetInfo.ResourceGroup = d.Get("rg_name").(string)
		azrVnetInfo.Cidr = d.Get("cidr_block").(string)
		vpc.AzVnetInfo = &azrVnetInfo
		err = appendInnerFieldMask(&azrVnetInfo, fieldMask, "azVnetInfo.")
		if err != nil {
			log.Print("AddVpc: Failed to append inner field mask for AzVnetInfo")
			return err
		}
	}
	vpc.FieldMask = fieldMask

	addVpcRequest := api.AddVpcRequest{
		Vpc: vpc,
	}
	log.Printf("[CVaaS-INFO]AddVpcRequestPb:%s", vpc)

	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + vpc.Name + "_" + d.Get("region").(string),
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
func (p *CloudeosProvider) DeleteVpc(d *schema.ResourceData) error {
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute DeleteVpc message")
		return err
	}
	defer client.wrpcClient.Close()

	vpcName, cpType := getCpTypeAndVpcName(d)
	vpc := &api.Vpc{
		Name:   vpcName,
		Id:     d.Get("tf_id").(string),
		CpT:    cpType,
		Region: d.Get("region").(string),
	}

	fieldMask, err := getOuterFieldMask(vpc)
	if err != nil {
		log.Print("DeleteVpc: Failed to get field mask")
		return err
	}
	vpc.FieldMask = fieldMask

	delVpcRequest := api.DeleteVpcRequest{
		Vpc: vpc,
	}

	log.Printf("[CVaaS-INFO]DeleteVpcRequestPb:%s", vpc)
	request := wrpcRequest{
		Token:   "RPC_Token_Delete_" + vpcName + "_" + d.Get("region").(string),
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

//ListTopology gets the Topology from Aeris which satisfy the filter
func (p *CloudeosProvider) ListTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVaaS-ERROR]Failed to create new client in ListTopology")
		return err
	}
	defer client.wrpcClient.Close()

	topoName := d.Get("topology_name").(string)
	closName := d.Get("clos_name").(string)
	wanName := d.Get("wan_name").(string)

	topoInfo := &api.TopologyInfo{
		Name: topoName,
	}

	fieldMask, err := getOuterFieldMask(topoInfo)
	if err != nil {
		log.Print("ListTopology: Failed to get field mask")
		return err
	}
	topoInfo.FieldMask = fieldMask

	log.Printf("[CVaaS-INFO]ListTopologyInfoRequestPb:%s", topoInfo)

	listTopoInfoRequest := api.ListTopologyInfoRequest{
		Filter: []*api.TopologyInfo{topoInfo},
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
		log.Printf("Failed to send %s request to CVaaS : %s",
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
			log.Printf("Failed to get %s response from CVaaS, Error: %v",
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
			errStr = errStr + "Resource cloudeos_topology " + topoName + " does not exist. "
		}
		if !closTopoExist {
			errStr = errStr + "Resource cloudeos_clos " + closName + " does not exist."
		}
	} else if strings.EqualFold("CloudEdge", role) {
		if metaTopoExist && wanTopoExist && closTopoExist {
			return nil
		}
		if !metaTopoExist {
			errStr = errStr + "Resource cloudeos_topology " + topoName + " does not exist. "
		}
		if !closTopoExist {
			errStr = errStr + "Resource cloudeos_clos " + closName + " does not exist. "
		}
		if !wanTopoExist {
			errStr = errStr + "Resource cloudeos_wan " + wanName + " does not exist."
		}
	}
	log.Printf("metaTopoExist: %v", metaTopoExist)
	log.Printf("wanTopoExist: %v", wanTopoExist)
	log.Printf("closTopoExist: %v", closTopoExist)
	return errors.New(errStr)
}

//CheckTopologyDeletionStatus returns nil if topology doesn't exist
func (p *CloudeosProvider) CheckTopologyDeletionStatus(d *schema.ResourceData) error {
	// Create new client
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVaaS-ERROR]Failed to create new client CheckTopologyDeletionStatus")
		return err
	}
	defer client.wrpcClient.Close()

	topoInfo := &api.TopologyInfo{
		Id: d.Get("tf_id").(string),
	}

	fieldMask, err := getOuterFieldMask(topoInfo)
	if err != nil {
		log.Print("CheckTopologyDeletionStatus: Failed to get field mask")
		return err
	}
	topoInfo.FieldMask = fieldMask

	log.Printf("[CVaaS-INFO]GetTopologyInfoRequestPb:%s", topoInfo)

	getTopoInfoRequest := api.GetTopologyInfoRequest{
		TopologyInfo: topoInfo,
	}

	request := wrpcRequest{
		Token:   "RPC_Token_Get_" + d.Get("tf_id").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Topologyinfos",
			"method":  "GetTopologyInfo",
			"body":    &getTopoInfoRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVaaS : %s",
			request.Params["method"].(string), err)
		return err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	// read respose from clouddeploy
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		log.Printf("Failed to get %s response from CVaaS, Error: %v",
			request.Params["method"].(string), err)
		return err
	}

	topologyExists := false
	/* A response with no topology looks like:
	   map[error:rpc error: code = NotFound desc = did not find resource "xxx"
	       status:map[code:5 message:did not find resource "xxx"] token: ... ] */

	// parse response and check if topology exist
	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key, val := range res {
				if strings.EqualFold(key, "topology_info") {
					if topo, ok := val.(map[string]interface{}); ok {
						if topo["topo_type"] == "TOPO_INFO_META" ||
							topo["topo_type"] == "TOPO_INFO_WAN" ||
							topo["topo_type"] == "TOPO_INFO_CLOS" {
							topologyExists = true
						}
					}
				}
			}
		}
	}

	log.Printf("topologyExist: %v", topologyExists)
	if topologyExists {
		return errors.New("Topology resource exists")
	}
	return nil
}

//AddTopology adds Topology resource to Aeris
func (p *CloudeosProvider) AddTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVaaS-ERROR]Failed to create new client in AddTopology")
		return err
	}
	defer client.wrpcClient.Close()

	// get bgp asn
	asnLow, asnHigh, err := getBgpAsn(d.Get("bgp_asn").(string))
	if err != nil {
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

	topoInfo := &api.TopologyInfo{
		Version:             providerCloudEOSVersion,
		Name:                d.Get("topology_name").(string),
		Id:                  d.Get("tf_id").(string),
		TopoType:            api.TopologyInfoType_TOPO_INFO_META,
		BgpAsnLow:           asnLow,
		BgpAsnHigh:          asnHigh,
		VtepIpCidr:          d.Get("vtep_ip_cidr").(string),
		TerminattrIpCidr:    d.Get("terminattr_ip_cidr").(string),
		DpsControlPlaneCidr: d.Get("dps_controlplane_cidr").(string),
		ManagedDevices:      managedDevices,
		CvaasDomain:         p.cvaasDomain,
		CvaasServer:         p.server,
	}

	fieldMask, err := getOuterFieldMask(topoInfo)
	if err != nil {
		log.Print("AddTopology: Failed to get field mask")
		return err
	}
	topoInfo.FieldMask = fieldMask

	log.Printf("[CVaaS-INFO]AddTopologyInfoRequestPb:%s", topoInfo)
	addTopoInfoRequest := api.AddTopologyInfoRequest{
		TopologyInfo: topoInfo,
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
func (p *CloudeosProvider) DeleteTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVaaS-ERROR]Failed to create new client in DeleteTopology")
		return err
	}
	defer client.wrpcClient.Close()

	topoInfo := &api.TopologyInfo{
		Name:     d.Get("topology_name").(string),
		Id:       d.Get("tf_id").(string),
		TopoType: api.TopologyInfoType_TOPO_INFO_META,
	}

	fieldMask, err := getOuterFieldMask(topoInfo)
	if err != nil {
		log.Print("ListTopology: Failed to get field mask")
		return err
	}
	topoInfo.FieldMask = fieldMask

	delTopoInfoRequest := api.DeleteTopologyInfoRequest{
		TopologyInfo: topoInfo,
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
func (p *CloudeosProvider) AddClosTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVaaS-ERROR]Failed to create new client in AddClosTopology")
		return err
	}
	defer client.wrpcClient.Close()

	fabricName := d.Get("fabric").(string)
	fabric := api.FabricType_FABRIC_TYPE_UNSPECIFIED
	if strings.EqualFold("full_mesh", fabricName) {
		fabric = api.FabricType_FULL_MESH
	} else if strings.EqualFold("hub_spoke", fabricName) {
		fabric = api.FabricType_HUB_SPOKE
	}

	closInfo := &api.ClosInfo{
		ClosName:         d.Get("name").(string),
		Fabric:           fabric,
		LeafEdgePeering:  d.Get("leaf_to_edge_peering").(bool),
		LeafEdgeIgw:      d.Get("leaf_to_edge_igw").(bool),
		LeafEncryption:   d.Get("leaf_encryption").(bool),
		CvpContainerName: d.Get("cv_container_name").(string),
	}

	topoInfo := &api.TopologyInfo{
		Name:     d.Get("topology_name").(string),
		Id:       d.Get("tf_id").(string),
		TopoType: api.TopologyInfoType_TOPO_INFO_CLOS,
		ClosInfo: closInfo,
	}

	fieldMask, err := getOuterFieldMask(topoInfo)
	if err != nil {
		log.Print("AddClosTopology: Failed to get topoInfo field mask")
		return err
	}

	err = appendInnerFieldMask(closInfo, fieldMask, "closInfo.")
	if err != nil {
		log.Print("AddClosTopology: Failed to get closInfo field mask")
		return err
	}
	topoInfo.FieldMask = fieldMask

	addTopoInfoRequest := api.AddTopologyInfoRequest{
		TopologyInfo: topoInfo,
	}
	log.Printf("[CVaaS-INFO]AddTopologyInfoRequestPb:%s", topoInfo)

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
func (p *CloudeosProvider) DeleteClosTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVaaS-ERROR]Failed to create new client in DeleteClosTopology")
		return err
	}
	defer client.wrpcClient.Close()

	closInfo := &api.ClosInfo{
		ClosName: d.Get("name").(string),
	}

	topoInfo := &api.TopologyInfo{
		Name:     d.Get("topology_name").(string),
		Id:       d.Get("tf_id").(string),
		TopoType: api.TopologyInfoType_TOPO_INFO_CLOS,
		ClosInfo: closInfo,
	}

	fieldMask, err := getOuterFieldMask(topoInfo)
	if err != nil {
		log.Print("DeleteClosTopology: Failed to get topoInfo field mask")
		return err
	}

	err = appendInnerFieldMask(closInfo, fieldMask, "closInfo.")
	if err != nil {
		log.Print("AddClosTopology: Failed to get closInfo field mask")
		return err
	}
	topoInfo.FieldMask = fieldMask

	log.Printf("[CVaaS-INFO]DeleteClosTopology DeleteTopologyInfoRequestPb:%s", topoInfo)
	delTopoInfoRequest := api.DeleteTopologyInfoRequest{
		TopologyInfo: topoInfo,
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
func (p *CloudeosProvider) AddWanTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVaaS-ERROR]Failed to create new client in AddWanTopology")
		return err
	}
	defer client.wrpcClient.Close()

	wanInfo := &api.WanInfo{
		WanName:              d.Get("name").(string),
		EdgeEdgePeering:      d.Get("edge_to_edge_peering").(bool),
		EdgeEdgeIgw:          d.Get("edge_to_edge_igw").(bool),
		EdgeDedicatedConnect: d.Get("edge_to_edge_dedicated_connect").(bool),
		CvpContainerName:     d.Get("cv_container_name").(string),
	}

	topoInfo := &api.TopologyInfo{
		Name:     d.Get("topology_name").(string),
		Id:       d.Get("tf_id").(string),
		TopoType: api.TopologyInfoType_TOPO_INFO_WAN,
		WanInfo:  wanInfo,
	}

	fieldMask, err := getOuterFieldMask(topoInfo)
	if err != nil {
		log.Print("AddWanTopology: Failed to get topoInfo field mask")
		return err
	}

	err = appendInnerFieldMask(wanInfo, fieldMask, "wanInfo.")
	if err != nil {
		log.Print("AddClosTopology: Failed to get wanInfo field mask")
		return err
	}
	topoInfo.FieldMask = fieldMask

	addTopoInfoRequest := api.AddTopologyInfoRequest{
		TopologyInfo: topoInfo,
	}
	log.Printf("[CVaaS-INFO]AddTopologyInfoRequestPb:%s", topoInfo)

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
func (p *CloudeosProvider) DeleteWanTopology(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("[CVaaS-ERROR]Failed to create new client in DeleteClosTopology")
		return err
	}
	defer client.wrpcClient.Close()

	wanInfo := &api.WanInfo{
		WanName: d.Get("name").(string),
	}

	topoInfo := &api.TopologyInfo{
		Name:     d.Get("topology_name").(string),
		Id:       d.Get("tf_id").(string),
		TopoType: api.TopologyInfoType_TOPO_INFO_WAN,
		WanInfo:  wanInfo,
	}

	fieldMask, err := getOuterFieldMask(topoInfo)
	if err != nil {
		log.Print("DeleteWanTopology: Failed to get field mask")
		return err
	}

	err = appendInnerFieldMask(wanInfo, fieldMask, "wanInfo.")
	if err != nil {
		log.Print("AddClosTopology: Failed to get wanInfo field mask")
		return err
	}
	topoInfo.FieldMask = fieldMask

	delTopoInfoRequest := api.DeleteTopologyInfoRequest{
		TopologyInfo: topoInfo,
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

//ListRouter gets router details from CloudDeploy
func (p *CloudeosProvider) ListRouter(d *schema.ResourceData) error {
	// create new client
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
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

	cpType := getCloudProviderType(d)
	rtr := &api.Router{
		Name:   routerName,
		Id:     d.Get("tf_id").(string),
		VpcId:  d.Get("vpc_id").(string),
		CpT:    cpType,
		Region: d.Get("region").(string),
		Cnps:   d.Get("cnps").(string),
	}

	fieldMask, err := getOuterFieldMask(rtr)
	if err != nil {
		log.Print("ListRouter: Failed to get field mask")
		return err
	}
	rtr.FieldMask = fieldMask

	log.Printf("[CVaaS-INFO]ListRouterRequestPb:%v", rtr)

	listRouterRequest := api.ListRouterRequest{
		Filter: []*api.Router{rtr},
	}

	request := wrpcRequest{
		Token:   "RPC_Token_List_" + routerName + "_" + d.Get("region").(string),
		Command: "serviceRequest",
		Params: map[string]interface{}{
			"service": "clouddeploy.Routers",
			"method":  "ListRouter",
			"body":    &listRouterRequest,
		},
	}

	err = client.wrpcClient.WriteJSON(request)
	if err != nil {
		log.Printf("Failed to send %s request to CVaaS : %s",
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
		// creation of aws_instance.cloudeosVm.
		if err := setBootStrapCfg(d, ""); err != nil {
			return err
		}
	}
	log.Printf("Received Resp: %v", resp)
	return nil
}

//GetRouterResponse - Common function to get the response from Clouddeploy.
func (p *CloudeosProvider) GetRouterResponse(d *schema.ResourceData) (map[string]interface{},
	error) {
	// create new client
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute GetRouter message")
		return nil, err
	}
	defer client.wrpcClient.Close()

	rtr := &api.Router{
		Id: d.Get("tf_id").(string),
	}

	fieldMask, err := getOuterFieldMask(rtr)
	if err != nil {
		log.Print("GetRouter: Failed to get field mask")
		return nil, err
	}
	rtr.FieldMask = fieldMask

	log.Printf("[CVaaS-INFO]GetRouterRequestPb:%s", rtr)

	getRouterRequest := api.GetRouterRequest{
		Router: rtr,
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
		log.Printf("Failed to send %s request to CVaaS : %s",
			request.Params["method"].(string), err)
		return nil, err
	}
	log.Printf("Successfully sent %s request for %s",
		request.Params["method"].(string), request.Token)

	resp := make(map[string]interface{})
	err = client.wrpcClient.ReadJSON(&resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//GetRouter gets router details from CloudDeploy
func (p *CloudeosProvider) GetRouter(d *schema.ResourceData) error {
	// create new client
	resp, err := p.GetRouterResponse(d)
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
		// creation of aws_instance.cloudeosVm
		if err := setBootStrapCfg(d, ""); err != nil {
			return err
		}
	}
	log.Printf("Received GetRouter Resp: %v", resp)
	return nil
}

func (p *CloudeosProvider) GetRouterStatus(d *schema.ResourceData) error {
	resp, err := p.GetRouterResponse(d)
	if err != nil {
		return err
	}

	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key, val := range res {
				if strings.EqualFold(key, "router") {
					if rtr, ok := val.(map[string]interface{}); ok {
						for k, v := range rtr {
							if strings.EqualFold(k, "bgp_asn") {
								routerBgpAsn := fmt.Sprint(v)
								err := d.Set("router_bgp_asn", routerBgpAsn)
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

//CheckRouterDeletionStatus returns nil if Router doesn't exist
func (p *CloudeosProvider) CheckRouterDeletionStatus(d *schema.ResourceData) error {
	// create new client
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client in CheckRouterDeletionStatus")
		return err
	}
	defer client.wrpcClient.Close()

	rtr := &api.Router{
		Id: d.Get("tf_id").(string),
	}

	fieldMask, err := getOuterFieldMask(rtr)
	if err != nil {
		log.Print("CheckRouterDeletionStatus: Failed to get field mask")
		return err
	}
	rtr.FieldMask = fieldMask

	log.Printf("[CVaaS-INFO]GetRouterRequestPb:%s", rtr)

	getRouterRequest := api.GetRouterRequest{
		Router: rtr,
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
		log.Printf("Failed to send %s request to CVaaS : %s",
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
	log.Printf("Received GetRouter Resp: %v", resp)

	routerExists := false
	/* A response with no Router looks like:
	   map[error:rpc error: code = NotFound desc = did not find resource "xxx"
	       status:map[code:5 message:did not find resource "xxx"] token: ... ] */

	// parse response to check if Router exist
	if res, ok := resp["result"]; ok {
		if res, ok := res.(map[string]interface{}); ok {
			for key := range res {
				if strings.EqualFold(key, "router") {
					routerExists = true
				}
			}
		}
	}

	log.Printf("routerExist: %v", routerExists)
	if routerExists {
		return errors.New("Router resource exists")
	}
	return nil
}

//AddRouterConfig adds Router resource to Aeris
func (p *CloudeosProvider) AddRouterConfig(d *schema.ResourceData) error {
	enrollmentToken, err := p.getDeviceEnrollmentToken()
	if err != nil {
		log.Printf("%v", err)
		return err
	}

	// Create new client.
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute AddRouter message")
		return err
	}
	defer client.wrpcClient.Close()

	routerName, err := getRouterNameFromSchema(d)
	if err != nil {
		return err
	}

	//Adding Intf Private IP, Type and Name to the first message.
	var intfs []*api.NetworkInterface
	intfNameList := d.Get("intf_name").([]interface{})
	privateIPList := d.Get("intf_private_ip").([]interface{})
	intfTypeList := d.Get("intf_type").([]interface{})
	intfCount := len(intfNameList)
	for i := 0; i < intfCount; i++ {
		var intf api.NetworkInterface
		intf.Name = intfNameList[i].(string)
		intf.PrivateIpAddr = []string{privateIPList[i].(string)}
		switch {
		case strings.EqualFold(intfTypeList[i].(string), "public"):
			intf.IntfType = api.NetworkInterfaceType_INTF_TYPE_PUBLIC
		case strings.EqualFold(intfTypeList[i].(string), "private"):
			intf.IntfType = api.NetworkInterfaceType_INTF_TYPE_PRIVATE
		case strings.EqualFold(intfTypeList[i].(string), "internal"):
			intf.IntfType = api.NetworkInterfaceType_INTF_TYPE_INTERNAL
		}
		intf.FieldMask, err = getOuterFieldMask(&intf)
		if err != nil {
			log.Print("AddRouterConfig: Failed to get field mask for intf")
			return err
		}
		intfs = append(intfs, &intf)
	}

	cpType := getCloudProviderType(d)
	rtr := &api.Router{
		Name:                  routerName,
		Id:                    d.Get("tf_id").(string),
		VpcId:                 d.Get("vpc_id").(string),
		CpT:                   cpType,
		Region:                d.Get("region").(string),
		Cnps:                  d.Get("cnps").(string),
		DeviceEnrollmentToken: enrollmentToken,
		RouteReflector:        d.Get("is_rr").(bool),
		Intf:                  intfs,
	}

	fieldMask, err := getOuterFieldMask(rtr)
	if err != nil {
		log.Print("AddRouterConfig: Failed to get field mask")
		return err
	}
	rtr.FieldMask = fieldMask

	addRouterRequest := api.AddRouterRequest{
		Router: rtr,
	}

	log.Printf("AddRouterRequestPb:%s", rtr)
	request := wrpcRequest{
		Token:   "RPC_Token_Add_" + routerName + "_" + d.Get("region").(string),
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
func (p *CloudeosProvider) CheckEdgeRouterPresence(d *schema.ResourceData) error {
	// Logic to check edge router presence
	//  - Call ListVpc with region, cp_type and role=Edge and get vpc_ids
	//    of all edge vpc's.
	//  - Call ListRouter with edge vpc_id and check if there is any router.
	//  - If we found a router then that router is an edge router.

	// create new client
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute CheckEdgeRouter message")
		return err
	}
	defer client.wrpcClient.Close()

	cpType := getCloudProviderType(d)
	// Code for ListVpc request
	vpc := &api.Vpc{
		CpT:          cpType,
		Region:       d.Get("region").(string),
		TopologyName: d.Get("topology_name").(string),
	}

	fieldMask, err := getOuterFieldMask(vpc)
	if err != nil {
		log.Print("CheckEdgeRouterPresence: Failed to get field mask for vpc")
		return err
	}
	vpc.FieldMask = fieldMask

	listVpcRequest := api.ListVpcRequest{
		Filter: []*api.Vpc{vpc},
	}
	log.Printf("[CVaaS-INFO]ListVpcRequestPb:%s", vpc)

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
		log.Printf("Failed to send %s request to CVaaS : %s",
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
			log.Printf("Failed to get %s response from CVaaS, Error: %v",
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
		rtr := &api.Router{
			VpcId:          edgeVpcID,
			CpT:            cpType,
			Region:         d.Get("region").(string),
			RouteReflector: false,
		}
		fieldMask, err := getOuterFieldMask(rtr)
		if err != nil {
			log.Print("CheckEdgeRouterPresence: Failed to get field mask for router")
			return err
		}
		rtr.FieldMask = fieldMask
		log.Printf("[CVaaS-INFO]ToResourceListEdgeRouter RequestPb:%s", rtr)
		listRouterRequest := api.ListRouterRequest{
			Filter: []*api.Router{rtr},
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
			log.Printf("Failed to send %s request to CVaaS : %s",
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
				log.Printf("Failed to get %s response from CVaaS, Error: %v",
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

//AddRouter adds Router resource to Aeris
func (p *CloudeosProvider) AddRouter(d *schema.ResourceData) error {
	// Create new client.
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
	if err != nil {
		log.Printf("Failed to create new client to execute AddRouter message")
		return err
	}
	defer client.wrpcClient.Close()

	routerName, err := getRouterNameFromSchema(d)
	if err != nil {
		return err
	}

	var intfs []*api.NetworkInterface
	publicIP, isPublicIP := d.GetOk("public_ip")
	intfNameList := d.Get("intf_name").([]interface{})
	intfIDList := d.Get("intf_id").([]interface{})
	privateIPList := d.Get("intf_private_ip").([]interface{})
	subnetIDList := d.Get("intf_subnet_id").([]interface{})
	intfTypeList := d.Get("intf_type").([]interface{})
	routeTableList := getAndCreateRouteTableIDs(d)

	intfCount := len(intfNameList)
	for i := 0; i < intfCount; i++ {
		var intf api.NetworkInterface
		intf.Name = intfNameList[i].(string)
		intf.IntfId = intfIDList[i].(string)
		intf.PrivateIpAddr = []string{privateIPList[i].(string)}
		intf.Subnet = subnetIDList[i].(string)
		if i == 0 && isPublicIP {
			intf.PublicIpAddr = publicIP.(string)
		} else {
			intf.PublicIpAddr = ""
		}
		switch {
		case strings.EqualFold(intfTypeList[i].(string), "public"):
			intf.IntfType = api.NetworkInterfaceType_INTF_TYPE_PUBLIC
		case strings.EqualFold(intfTypeList[i].(string), "private"):
			intf.IntfType = api.NetworkInterfaceType_INTF_TYPE_PRIVATE
		case strings.EqualFold(intfTypeList[i].(string), "internal"):
			intf.IntfType = api.NetworkInterfaceType_INTF_TYPE_INTERNAL
		}
		intf.FieldMask, err = getOuterFieldMask(&intf)
		if err != nil {
			log.Print("AddRouter: Failed to get field mask for intf")
			return err
		}
		intfs = append(intfs, &intf)
	}

	cpType := getCloudProviderType(d)
	rtr := &api.Router{
		Name:       routerName,
		Id:         d.Get("tf_id").(string),
		VpcId:      d.Get("vpc_id").(string),
		CpT:        cpType,
		Cnps:       d.Get("cnps").(string),
		Region:     d.Get("region").(string),
		InstanceId: d.Get("instance_id").(string),
		//Tag: d.Get("tag_id"),
		DepStatus:      api.DeploymentStatusCode_DEP_STATUS_SUCCESS,
		Intf:           intfs,
		RtTableIds:     routeTableList,
		RouteReflector: d.Get("is_rr").(bool),
		HaName:         d.Get("ha_name").(string),
	}

	fieldMask, err := getOuterFieldMask(rtr)
	if err != nil {
		log.Print("AddRouter: Failed to get field mask")
		return err
	}

	cloudProvider := d.Get("cloud_provider").(string)
	switch {
	case strings.EqualFold("aws", cloudProvider):
		awsRtrDetail := api.AwsRouterDetail{}
		awsRtrDetail.AvailZone = d.Get("availability_zone").(string)
		awsRtrDetail.InstanceType = d.Get("instance_type").(string)
		rtr.AwsRtrDetail = &awsRtrDetail
		err := appendInnerFieldMask(&awsRtrDetail, fieldMask, "awsRtrDetail.")
		if err != nil {
			log.Print("AddRouter: Failed to append field mask for awsRtrDetail")
			return err
		}
	case strings.EqualFold("azure", cloudProvider):
		azrRtrDetail := api.AzureRouterDetail{}
		azrRtrDetail.AvailZone = d.Get("rg_location").(string)
		azrRtrDetail.ResGroup = d.Get("rg_name").(string)
		azrRtrDetail.InstanceType = d.Get("instance_type").(string)
		rtr.AzRtrDetail = &azrRtrDetail
		err := appendInnerFieldMask(&azrRtrDetail, fieldMask, "azRtrDetail.")
		if err != nil {
			log.Print("AddRouter: Failed to append field mask for azRtrDetail")
			return err
		}
	}

	err = appendInnerFieldMask(routeTableList, fieldMask, "rtTableIds.")
	if err != nil {
		log.Print("AddRouter: Failed to append field mask for rtTableIds")
		return err
	}
	rtr.FieldMask = fieldMask
	log.Printf("AddRouterRequestPb:%s", rtr)

	addRouterRequest := api.AddRouterRequest{
		Router: rtr,
	}
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
func (p *CloudeosProvider) DeleteRouter(d *schema.ResourceData) error {
	// Create new client, as the client that provider created might have died.
	client, err := aristaCvaasClient(p.server, p.srvcAcctToken)
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

	rtr := &api.Router{
		Name:  routerName,
		Id:    d.Get("tf_id").(string),
		VpcId: d.Get("vpc_id").(string),
		CpT:   cpType,
	}

	fieldMask, err := getOuterFieldMask(rtr)
	if err != nil {
		log.Print("DeleteRouter: Failed to setFieldMask")
		return err
	}
	rtr.FieldMask = fieldMask

	delRouterRequest := api.DeleteRouterRequest{
		Router: rtr,
	}

	log.Printf("DeleteRouterRequestPb : %s", rtr)

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
