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

// cloudeosTopology: Define the cloudeos topology schema ( input and output variables )
func cloudeosTopology() *schema.Resource {
	return &schema.Resource{
		Create: cloudeosTopologyCreate,
		Read:   cloudeosTopologyRead,
		Update: cloudeosTopologyUpdate,
		Delete: cloudeosTopologyDelete,

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"topology_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the base topology",
			},
			"bgp_asn": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Range, a-b, of BGP ASN’s used for topology",
			},
			"vtep_ip_cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "CIDR block for VTEP IPs on cloudeos",
			},
			"terminattr_ip_cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Loopback IP range on cloudeos",
			},
			"dps_controlplane_cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "CIDR block for TerminAttr IPs on cloudeos",
			},
			"eos_managed": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Existing cloudeos",
				Set:         schema.HashString,
			},
			"tf_id": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func cloudeosTopologyCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	//Check for duplicates only for Add and not for Updates.
	duplicate, err := provider.CheckForTopologyDuplicates(d, "TOPO_INFO_META")
	if duplicate || err != nil {
		return err
	}
	err = provider.AddTopology(d)
	if err != nil {
		return err
	}

	uuid := "cloudeos-topology" + strings.TrimPrefix(d.Get("tf_id").(string), TopoPrefix)
	log.Print("Successfully added " + uuid)
	d.SetId(uuid)
	return nil
}

func cloudeosTopologyRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func cloudeosTopologyUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.AddTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully updated cloudeos-topology" +
		strings.TrimPrefix(d.Get("tf_id").(string), TopoPrefix))
	return nil
}

func cloudeosTopologyDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.DeleteTopology(d)
	if err != nil {
		return err
	}

	uuid := "cloudeos-topology" + strings.TrimPrefix(d.Get("tf_id").(string), TopoPrefix)
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
