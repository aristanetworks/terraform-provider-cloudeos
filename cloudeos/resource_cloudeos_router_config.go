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

func cloudeosRouterConfig() *schema.Resource {
	return &schema.Resource{
		Create: cloudeosRouterConfigCreate,
		Read:   cloudeosRouterConfigRead,
		Update: cloudeosRouterConfigUpdate,
		Delete: cloudeosRouterConfigDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
				DiffSuppressFunc: suppressAttributeChange,
			},
			"cnps": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"region": {
				Required: true,
				Type:     schema.TypeString,
			},
			"topology_name": {
				Required:         true,
				Type:             schema.TypeString,
				DiffSuppressFunc: suppressAttributeChange,
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "A mapping of tags to assign to the resource",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					tags := val.(map[string]interface{})
					nameFound := false
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
			"vpc_id": {
				Required: true,
				Type:     schema.TypeString,
				ForceNew: true,
			},
			"role": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"is_rr": {
				Optional: true,
				Type:     schema.TypeBool,
			},
			"ami": {
				Optional:    true,
				Description: "CloudEOS image",
				Type:        schema.TypeString,
				ForceNew:    true,
			},
			"key_name": {
				Optional:    true,
				Description: "AWS keypair name",
				Type:        schema.TypeString,
				ForceNew:    true,
			},
			"availability_zone": {
				Optional: true,
				Type:     schema.TypeString,
				ForceNew: true,
			},
			"intf_name": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Interface name",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"intf_private_ip": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Private IP address",
				Elem:        &schema.Schema{Type: schema.TypeString},
				ForceNew:    true,
			},
			"intf_type": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Interface type",
				Elem:        &schema.Schema{Type: schema.TypeString},
				ForceNew:    true,
			},
			"peerroutetableid1": {
				Type:       schema.TypeList,
				Computed:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Deprecated: "This attribute is deprecated, use peer_routetable_id",
			},
			"peer_routetable_id": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"bootstrap_cfg": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"ha_rtr_id": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"public_rt_table_id": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"internal_rt_table_id": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"private_rt_table_id": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tf_id": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func cloudeosRouterConfigCreate(d *schema.ResourceData, m interface{}) error {
	//TBD: Call ListVpc to get deployment type( not needed for EFT )

	provider := m.(CloudeosProvider)

	//Retry ListVpc to check VPC is present in Aeris before Router.

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		err := provider.CheckVpcPresence(d)
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		return errors.New("Could not find the VPC in CVaaS.(Try terraform apply again)")
	}

	role := d.Get("role").(string)
	if strings.EqualFold("CloudLeaf", role) {
		// check if an edge router is present
		if err := provider.CheckEdgeRouterPresence(d); err != nil {
			return fmt.Errorf("Edge router should be created before leaf router: %v", err)
		}
	}

	err = provider.AddRouterConfig(d)
	if err != nil {
		return err
	}

	//Retry GetRouter for bootstrap_cfg
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		err := provider.GetRouter(d)
		if err != nil {
			return resource.RetryableError(fmt.Errorf("GetRouter failed: %s", err))
		}
		cfg := d.Get("bootstrap_cfg").(string)
		if strings.Contains(cfg, "daemon TerminAttr") {
			return nil
		}
		return resource.RetryableError(fmt.Errorf("attempting to get Bootstrap config"))
	})
	if err != nil {
		err := provider.DeleteRouter(d)
		if err != nil {
			return errors.New("bootstrap config wasn't set, failed during cleanup")
		}
		return errors.New("bootstrap config wasn't returned by CVP.(Try terraform apply again)")
	}

	uuid := "cloudeos-router-config" + strings.TrimPrefix(d.Get("tf_id").(string), RtrPrefix)
	log.Print("Successfully added " + uuid)
	d.SetId(uuid)
	return nil
}

func cloudeosRouterConfigRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func cloudeosRouterConfigUpdate(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)

	err := provider.AddRouterConfig(d)
	if err != nil {
		return err
	}

	log.Print("Successfully updated cloudeos-router-config" +
		strings.TrimPrefix(d.Get("tf_id").(string), RtrPrefix))
	return nil
}

func cloudeosRouterConfigDelete(d *schema.ResourceData, m interface{}) error {
	provider := m.(CloudeosProvider)
	err := provider.DeleteRouter(d)
	if err != nil {
		return err
	}

	uuid := "cloudeos-router-config" + strings.TrimPrefix(d.Get("tf_id").(string), RtrPrefix)
	// wait for router deletion
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if err := provider.CheckRouterDeletionStatus(d); err != nil {
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
