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

// cloudeosClos: Define the cloudeos clos schema ( input and output variables )
func cloudeosClos() *schema.Resource {
	return &schema.Resource{
		Create: cloudeosClosCreate,
		Read:   cloudeosClosRead,
		Update: cloudeosClosUpdate,
		Delete: cloudeosClosDelete,

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Clos topology name",
				DiffSuppressFunc: suppressAttributeChange,
			},
			"topology_name": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Base topology name",
				DiffSuppressFunc: suppressAttributeChange,
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "CloudLeaf",
				Description:      "Container name for leaf",
				DiffSuppressFunc: suppressAttributeChange,
			},
			"tf_id": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func cloudeosClosCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	allowed, err := provider.IsValidTopoAddition(d, "TOPO_INFO_CLOS")
	if !allowed || err != nil {
		return err
	}
	err = provider.AddClosTopology(d)
	if err != nil {
		return err
	}

	uuid := "cloudeos-clos" + strings.TrimPrefix(d.Get("tf_id").(string), ClosPrefix)
	log.Print("Successfully added " + uuid)
	d.SetId(uuid)
	return nil
}

func cloudeosClosRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func cloudeosClosUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.AddClosTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully updated cloudeos-clos" +
		strings.TrimPrefix(d.Get("tf_id").(string), ClosPrefix))
	return nil
}

func cloudeosClosDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.DeleteClosTopology(d)
	if err != nil {
		return err
	}

	uuid := "cloudeos-clos" + strings.TrimPrefix(d.Get("tf_id").(string), ClosPrefix)
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
