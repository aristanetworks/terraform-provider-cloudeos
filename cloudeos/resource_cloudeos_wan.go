// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// cloudeosWan: Define the cloudeos wan topology schema ( input and output variables )
func cloudeosWan() *schema.Resource {
	return &schema.Resource{
		Create: cloudeosWanCreate,
		Read:   cloudeosWanRead,
		Update: cloudeosWanUpdate,
		Delete: cloudeosWanDelete,

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Wan fabric name",
				DiffSuppressFunc: suppressAttributeChange,
			},
			"topology_name": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Base topology name",
				DiffSuppressFunc: suppressAttributeChange,
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
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Container name for edge",
				Default:          "CloudEdge",
				DiffSuppressFunc: suppressAttributeChange,
			},
			"tf_id": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func cloudeosWanCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	allowed, err := provider.IsValidTopoAddition(d, "TOPO_INFO_WAN")
	if !allowed || err != nil {
		return err
	}

	err = provider.AddWanTopology(d)
	if err != nil {
		return err
	}

	uuid := "cloudeos-wan" + strings.TrimPrefix(d.Get("tf_id").(string), WanPrefix)
	log.Print("Successfully added " + uuid)
	d.SetId(uuid)
	return nil
}

func cloudeosWanRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func cloudeosWanUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.AddWanTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully updated cloudeos-wan" +
		strings.TrimPrefix(d.Get("tf_id").(string), WanPrefix))
	return nil
}

func cloudeosWanDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.DeleteWanTopology(d)
	if err != nil {
		return err
	}

	uuid := "cloudeos-wan" + strings.TrimPrefix(d.Get("tf_id").(string), WanPrefix)
	// wait for topology deletion
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if err := provider.CheckTopologyDeletionStatus(d); err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		return errors.New("Failed to destroy " + uuid + " error: " + err.Error())
	}

	log.Print("Successfully deleted " + uuid)
	d.SetId("")
	return nil
}
