// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

// aristaWan: Define the arista wan topology schema ( input and output variables )
func aristaWan() *schema.Resource {
	return &schema.Resource{
		Create: aristaWanCreate,
		Read:   aristaWanRead,
		Update: aristaWanUpdate,
		Delete: aristaWanDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Wan fabric name",
			},
			"topology_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Base topology name",
			},
			"edge_to_edge_peering": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"edge_to_edge_dedicated_connect": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"edge_to_edge_igw": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"cv_container_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Container name for edge",
				Default:     "CloudEdge",
			},
			"tf_id": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func aristaWanCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddWanTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully added wan topology")
	d.SetId(d.Get("tf_id").(string))
	return nil
}

func aristaWanRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func aristaWanUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddWanTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully updated wan topology")
	return nil
}

func aristaWanDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.DeleteWanTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully deleted wan topology")
	d.SetId("")
	return nil
}
