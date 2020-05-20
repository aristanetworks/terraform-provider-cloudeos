// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

// aristaClos: Define the arista clos schema ( input and output variables )
func aristaClos() *schema.Resource {
	return &schema.Resource{
		Create: aristaClosCreate,
		Read:   aristaClosRead,
		Update: aristaClosUpdate,
		Delete: aristaClosDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Clos toplogy name",
			},
			"topology_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Base topology name",
			},
			"fabric": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "hub_spoke",
				Description: "full_mesh or hub_spoke",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "full_mesh" && v != "hub_spoke" {
						errs = append(errs, fmt.Errorf(
							"%q must be full_mesh/hub_spoke got: %q", key, v))
					}
					return
				},
			},
			"leaf_to_edge_peering": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"leaf_to_edge_igw": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"leaf_encryption": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cv_container_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "CloudLeaf",
				Description: "Container name for leaf",
			},
			"tf_id": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func aristaClosCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddClosTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully added clos topology")
	d.SetId(d.Get("tf_id").(string))
	return nil
}

func aristaClosRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func aristaClosUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddClosTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully updated clos topology")
	return nil
}

func aristaClosDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.DeleteClosTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully deleted clos topology")
	d.SetId("")
	return nil
}
