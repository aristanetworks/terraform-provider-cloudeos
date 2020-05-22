// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cloudeos

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

//cloudeosSubnet: Define the cloudeosSubnet schema ( input and output variables )
func cloudeosSubnet() *schema.Resource {
	return &schema.Resource{
		Create: cloudeosSubnetCreate,
		Read:   cloudeosSubnetRead,
		Update: cloudeosSubnetUpdate,
		Delete: cloudeosSubnetDelete,

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

func cloudeosSubnetCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.AddSubnet(d)
	if err != nil {
		return err
	}

	err = d.Set("computed_subnet_id", d.Get("subnet_id").(string))
	if err != nil {
		return err
	}

	uuid := "cloudeos-subnet" + strings.TrimPrefix(d.Get("tf_id").(string), SubnetPrefix)
	log.Print("Successfully added " + uuid)
	d.SetId(uuid)
	return nil
}

func cloudeosSubnetRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func cloudeosSubnetUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.AddSubnet(d)
	if err != nil {
		return err
	}
	log.Print("Successfully updated cloudeos-subnet" +
		strings.TrimPrefix(d.Get("tf_id").(string), SubnetPrefix))
	return nil
}

func cloudeosSubnetDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.DeleteSubnet(d)
	if err != nil {
		return err
	}

	log.Print("Successfully deleted cloudeos-subnet" +
		strings.TrimPrefix(d.Get("tf_id").(string), SubnetPrefix))
	d.SetId("")
	return nil
}
