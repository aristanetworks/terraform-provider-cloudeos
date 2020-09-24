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

func cloudeosVpcConfig() *schema.Resource {
	return &schema.Resource{
		Create: cloudeosVpcConfigCreate,
		Read:   cloudeosVpcConfigRead,
		Update: cloudeosVpcConfigUpdate,
		Delete: cloudeosVpcConfigDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cloud_provider": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "aws/azure/gcp",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "aws" && v != "azure" && v != "gcp" {
						errs = append(errs, fmt.Errorf(
							"%q must be aws/azure/gcp got: %q", key, v))
					}
					return
				},
			},
			"cnps": {
				Required:         true,
				Type:             schema.TypeString,
				DiffSuppressFunc: suppressAttributeChange,
			},
			"region": {
				Required: true,
				Type:     schema.TypeString,
			},
			"topology_name": {
				Required: true,
				Type:     schema.TypeString,
			},
			"clos_name": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "ClosFabric name",
			},
			"wan_name": {
				Optional:    true, // leaf VPC won't have wan_name
				Type:        schema.TypeString,
				Description: "WanFabric name",
			},
			"rg_name": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Resource Group name",
			},
			"vnet_name": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "VNET name",
			},
			"role": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "CloudEdge/CloudLeaf",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "CloudEdge" && v != "CloudLeaf" {
						errs = append(errs, fmt.Errorf(
							"%q must be CloudEdge/CloudLeaf got: %q", key, v))
					}
					return
				},
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "A mapping of tags to assign to the resource",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					tags := val.(map[string]interface{})
					nameFound := false
					log.Printf("tags:%v", tags)
					for k := range tags {
						if "Name" == k {
							nameFound = true
						}
					}
					if !nameFound {
						errs = append(errs, fmt.Errorf("%q must contain Name", key))
					}
					return
				},
			},
			"topology_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "cloudeos_topology resource ID",
			},
			"wan_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "cloudeos_wan resource ID",
			},
			"clos_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "cloudeos_clos resource ID",
			},
			"peer_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"peervpcidr": {
				Type:       schema.TypeString,
				Computed:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Deprecated: "This attribute is deprecated, use peer_vpc_cidr",
			},
			"peer_vpc_cidr": {
				Type:     schema.TypeString,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// Consumed by Azure modules
			"peer_vnet_id": {
				Type:     schema.TypeString,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"peer_rg_name": {
				Type:     schema.TypeString,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"peer_vnet_name": {
				Type:     schema.TypeString,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tf_id": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func cloudeosVpcConfigCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if err := provider.ListTopology(d); err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = provider.AddVpcConfig(d)
	if err != nil {
		return err
	}

	role := d.Get("role").(string)
	if strings.EqualFold("CloudLeaf", role) {
		// check for Cnps in "tags"
		tags := d.Get("tags").(map[string]interface{})
		cnpsFound := false
		for k := range tags {
			if "Cnps" == k {
				cnpsFound = true
			}
		}
		if !cnpsFound {
			return errors.New("tags must contain a Cnps for Leaf Vpc")
		}

		// Call ListVpc to get peer_vpc_id, peer_vpc_cidr
		err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			if err := provider.GetVpc(d); err != nil {
				return resource.RetryableError(err)
			}
			peerVpcID := d.Get("peer_vpc_id").(string)
			if peerVpcID != "" {
				return nil
			}
			return resource.RetryableError(fmt.Errorf("attempting to get Peer's VPC ID"))
		})
		if err != nil {
			err := provider.DeleteVpc(d)
			if err != nil {
				return errors.New("Peer VPC ID not set, failed during cleanup")
			}
			return errors.New("Peer's VPC ID is not returned by CVP")
		}
	}
	uuid := "cloudeos-vpc-config" + strings.TrimPrefix(d.Get("tf_id").(string), VpcPrefix)
	log.Print("Successfully added " + uuid)
	d.SetId(uuid)
	return nil
}

func cloudeosVpcConfigRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func cloudeosVpcConfigUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.AddVpcConfig(d)
	if err != nil {
		return err
	}
	log.Print("Successfully updated cloudeos-vpc-config" +
		strings.TrimPrefix(d.Get("tf_id").(string), VpcPrefix))
	return nil
}

func cloudeosVpcConfigDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.DeleteVpc(d)
	if err != nil {
		return err
	}

	uuid := "cloudeos-vpc-config" + strings.TrimPrefix(d.Get("tf_id").(string), VpcPrefix)
	// wait for vpc deletion
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if err := provider.CheckVpcDeletionStatus(d); err != nil {
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
