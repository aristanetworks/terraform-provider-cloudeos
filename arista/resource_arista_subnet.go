// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

//aristaSubnet: Define the aristaSubnet schema ( input and output variables )
func aristaSubnet() *schema.Resource {
	return &schema.Resource{
		Create: aristaSubnetCreate,
		Read:   aristaSubnetRead,
		Update: aristaSubnetUpdate,
		Delete: aristaSubnetDelete,

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
			// This is equivalent to rg_name in Azure
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Only set in Azure
			"vnet_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "availability zone",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "subnet id",
			},
			"computed_subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"cidr_block": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "CIDR block",
			},
			"subnet_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Subnet names",
			},
			"tf_id": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func aristaSubnetCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddSubnet(d)
	if err != nil {
		return err
	}
	log.Print("Successfully added subnet")

	err = d.Set("computed_subnet_id", d.Get("subnet_id").(string))
	if err != nil {
		return err
	}

	d.SetId("subnet-" + d.Get("tf_id").(string))
	return nil
}

func aristaSubnetRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func aristaSubnetUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddSubnet(d)
	if err != nil {
		return err
	}
	log.Print("Successfully updated subnet")
	return nil
}

func aristaSubnetDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.DeleteSubnet(d)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
