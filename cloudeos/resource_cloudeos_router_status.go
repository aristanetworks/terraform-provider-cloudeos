// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

//cloudeosRouterStatus: Define the cloudeosRouterStatus schema ( input and output variables )
func cloudeosRouterStatus() *schema.Resource {
	return &schema.Resource{
		Create: cloudeosRouterStatusCreate,
		Read:   cloudeosRouterStatusRead,
		Update: cloudeosRouterStatusUpdate,
		Delete: cloudeosRouterStatusDelete,

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cloud_provider": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "aws / azure / gcp",
			},
			"cv_container": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Container to which cvp should add this device",
			},
			// Set by AWS resource
			"vpc_id": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Vpc id of cloudeos",
			},
			// Set in Azure
			"rg_name": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"rg_location": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"instance_type": {
				Required: true,
				Type:     schema.TypeString,
			},
			"instance_id": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "VM instance ID",
				ForceNew:    true,
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "A mapping of tags to assign to the resource",
			},
			"availability_zone": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"primary_network_interface_id": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"availability_set_id": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Availability set if for Azure",
			},
			"public_ip": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Public ip address",
			},
			"intf_name": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Interface name",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"intf_id": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Interface id",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"intf_private_ip": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Private IP address",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"intf_subnet_id": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Subnet id attached to intf",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"intf_type": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Interface type",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"private_rt_table_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"internal_rt_table_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"public_rt_table_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ha_name": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"cnps": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"region": {
				Required: true,
				Type:     schema.TypeString,
			},
			"is_rr": {
				Optional: true,
				Type:     schema.TypeBool,
			},
			"deployment_status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Deployment Status of the CloudEOS Router",
				Computed:    true,
			},
			"tf_id": {
				Required: true,
				Type:     schema.TypeString,
			},
			"routing_resource_info": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of all route table and association resources.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"router_bgp_asn": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "BGP ASN computed on the CloudEOS Router",
			},
			"deploy_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressAttributeChange,
				Description:      "Deployment mode for the resources: provision or empty",
			},
		},
	}
}

func cloudeosRouterStatusCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.AddRouter(d)
	if err != nil {
		return err
	}

	// In the standard deploy mode, we retrieve the bgp_asn allocated for the router
	// and set it in the resource, which is then passed to aws_customer_gateway
	// resource, if the router is set to serve as a remote vpn gateway. For deploy mode
	// provision, this isn't possible since we don't really allocate asn for the routers
	// and we do not support this mode for the tgw stuff, as of now. So skip this.
	deployMode := strings.ToLower(d.Get("deploy_mode").(string))

	if deployMode != "provision" {
		err := provider.GetRouterStatusAndSetBgpAsn(d)
		if err != nil {
			return fmt.Errorf("GetRouter failed: %s", err)
		}
		bgpAsn := d.Get("router_bgp_asn").(string)
		// If we can't find an asn in the standard deploy mode in the router status entry,
		// something is wrong (since this is allocated when router_config resource is created.
		// Error out since this state means something is broken). Do we want to cleanup?
		if bgpAsn == "" {
			return fmt.Errorf("BGP ASN for the Router not returned")
		}
	}

	uuid := "cloudeos-router-status" + strings.TrimPrefix(d.Get("tf_id").(string), RtrPrefix)
	log.Print("Successfully added " + uuid)
	d.SetId(uuid)
	return nil
}

func cloudeosRouterStatusRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func cloudeosRouterStatusUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.AddRouter(d)
	if err != nil {
		return err
	}

	log.Print("Successfully updated cloudeos-router-status" +
		strings.TrimPrefix(d.Get("tf_id").(string), RtrPrefix))
	return nil
}

func cloudeosRouterStatusDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.DeleteRouter(d)
	if err != nil {
		return err
	}

	uuid := "cloudeos-router-status" + strings.TrimPrefix(d.Get("tf_id").(string), RtrPrefix)
	// wait for router deletion
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if err := provider.CheckRouterDeletionStatus(d); err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		return errors.New("Failed to destroy " + uuid + " Error: " + err.Error())
	}

	log.Print("Successfully deleted " + uuid)
	d.SetId("")
	return nil
}
