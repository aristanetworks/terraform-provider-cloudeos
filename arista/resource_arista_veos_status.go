// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

//aristaVeosStatus: Define the aristaVeosStatus schema ( input and output variables )
func aristaVeosStatus() *schema.Resource {
	return &schema.Resource{
		Create: aristaVeosStatusCreate,
		Read:   aristaVeosStatusRead,
		Update: aristaVeosStatusUpdate,
		Delete: aristaVeosStatusDelete,

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
				Description: "Vpc id of vrouter",
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tf_id": {
				Required: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func aristaVeosStatusCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddRouter(d)
	if err != nil {
		return err
	}

	log.Print("Successfully added veos_status")

	d.SetId("veos-status-" + d.Get("tf_id").(string))
	return nil
}

func aristaVeosStatusRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func aristaVeosStatusUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddRouter(d)
	if err != nil {
		return err
	}

	log.Print("Successfully updated veos_status")
	return nil
}

func aristaVeosStatusDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.DeleteRouter(d)
	if err != nil {
		return err
	}

	log.Print("Successfully deleted veos_status")
	d.SetId("")
	return nil
}
