// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

//aristaVpc: Define the aristaVpc schema ( input and output variables )
func aristaVpc() *schema.Resource {
	return &schema.Resource{
		Create: aristaVpcCreate,
		Read:   aristaVpcRead,
		Update: aristaVpcUpdate,
		Delete: aristaVpcDelete,

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
				Required: true,
				Type:     schema.TypeString,
			},
			"region": {
				Required: true,
				Type:     schema.TypeString,
			},
			"rg_name": {
				Optional: true,
				Type:     schema.TypeString,
			},
			// This is equiv to vnet_id in Azure
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Only set in Azure
			"vnet_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Security group id",
			},
			"cidr_block": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "CIDR block",
			},
			"igw": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Internet gateway id ",
			},
			"resource_group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Resource group needed by Azure",
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
			"topology_name": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Base topology name",
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
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "A mapping of tags to assign to the resource",
			},
			"tf_id": {
				Required: true,
				Type:     schema.TypeString,
			},
			"account": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier of the account",
			},
		},
	}
}

func aristaVpcCreate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddVpc(d)
	if err != nil {
		return err
	}

	log.Print("Successfully added vpc")
	d.SetId("arista-vpc-" + d.Get("tf_id").(string))
	return nil
}

func aristaVpcRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func aristaVpcUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.AddVpc(d)
	if err != nil {
		return err
	}

	log.Print("Successfully Updated vpc")
	return nil
}

func aristaVpcDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(AristaProvider)
	err := provider.DeleteVpc(d)
	if err != nil {
		return err
	}

	log.Print("Successfully deleted vpc")
	d.SetId("")
	return nil
}
