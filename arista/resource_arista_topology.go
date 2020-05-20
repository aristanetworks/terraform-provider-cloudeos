// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

// aristaTopology: Define the arista topology schema ( input and output variables )
func aristaTopology() *schema.Resource {
	return &schema.Resource{
		Create: aristaTopologyCreate,
		Read:   aristaTopologyRead,
		Update: aristaTopologyUpdate,
		Delete: aristaTopologyDelete,

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
				Description: "CIDR block for VTEP IPs on veos",
			},
			"terminattr_ip_cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Loopback IP range on veos",
			},
			"dps_controlplane_cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "CIDR block for TerminAttr IPs on veos",
			},
			"eos_managed": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Existing veos",
				Set:         schema.HashString,
			},
			"tf_id": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func aristaTopologyCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully added topology")
	d.SetId(d.Get("tf_id").(string))
	return nil
}

func aristaTopologyRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func aristaTopologyUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully updated topology")
	return nil
}

func aristaTopologyDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.DeleteTopology(d)
	if err != nil {
		return err
	}

	log.Print("Successfully deleted topology")
	d.SetId("")
	return nil
}
