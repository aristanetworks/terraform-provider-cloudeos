// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"errors"
	"fmt"
	"log"
	"os"
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
			"cloudeos_image_offer": {
				Optional:         true,
				Description:      "CloudEos Licensing Model",
				Type:             schema.TypeString,
				DiffSuppressFunc: suppressAttributeChange,
				Default:          "cloudeos-router-payg",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "cloudeos-router-byol" && v != "cloudeos-router-payg" {
						errs = append(errs, fmt.Errorf(
							"%q must be cloudeos-router-byol/cloudeos-router-payg got: %q", key, v))
					}
					return
				},
			},
			"licenses": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of licenses for cloudeos",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"hash": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
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
			"deploy_mode": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
		CustomizeDiff: func(d *schema.ResourceDiff, m interface{}) error {
			oldoffer, offer := d.GetChange("cloudeos_image_offer")
			// LicenseType : Compulsory map
			licenseNeeded := map[string]bool{
				"ipsec":     true,
				"bandwidth": true,
			}

			// Validate licenses here, as TypeSet doesn't support validateFunc.
			if v, ok := d.GetOk("licenses"); ok {
				// Licenses should be specified only with BYOL
				if offer != "cloudeos-router-byol" {
					return fmt.Errorf("Licenses not supported when using PAYG. Are you sure you are using correct cloudeos-image-offer")
				}
				licenseList := v.(*schema.Set).List()
				for _, k := range licenseList {
					license := k.(map[string]interface{})
					licenseType := license["type"].(string)
					if _, ok := licenseNeeded[licenseType]; ok {
						licenseNeeded[licenseType] = false
						_, err := os.Stat(license["path"].(string))
						if err != nil {
							return fmt.Errorf(" License %s, Unable to open file %q", licenseType, license["path"])
						}
					} else {
						supportedLicenses := " "
						for k, _ := range licenseNeeded {
							supportedLicenses += k + ", "
						}
						return fmt.Errorf("%s license isn't supported. Supported Licenses : [%s]", licenseType, supportedLicenses)
					}
				}
			}

			if offer == "cloudeos-router-byol" {

				// Plugin upgraded for already deployed topology
				oldCloudProvider, _ := d.GetChange("cloud_provider")
				if oldCloudProvider != "" && oldoffer == "" {
					return fmt.Errorf("Already exists payg topology, destroy it first before changing cloudeos_image_offer to byol")

				}

				// Some licenses are compulsory, check they are present
				missingLicenses := " "
				for k, v := range licenseNeeded {
					if v == true {
						missingLicenses += k + ", "
					}
				}
				if missingLicenses != " " {
					return fmt.Errorf("[%s] license needs to be specified when using BYOL", missingLicenses)
				}

				// Check if license attribute is updated
				oldLicenses, newLicenses := d.GetChange("licenses")
				if oldoffer == "cloudeos-router-byol" && oldLicenses != nil && newLicenses != nil {
					oldLicensesSet := oldLicenses.(*schema.Set)
					newLicensesSet := newLicenses.(*schema.Set)
					if !oldLicensesSet.Equal(newLicensesSet) {
						log.Printf("Attribute Change: licenses \n[old] : %#v \n[new] : %#v", oldLicensesSet, newLicensesSet)
						return fmt.Errorf("Updating Licenses is not supported, you need to destroy first or make changes through CVaaS")
					}
				}
			}
			return nil
		},
	}
}

func cloudeosRouterConfigCreate(d *schema.ResourceData, m interface{}) error {
	//TBD: Call ListVpc to get deployment type( not needed for EFT )
	provider := m.(CloudeosProvider)

	//Retry ListVpc to check VPC is present in Aeris before Router.
	var rtrDeployMode string
	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		deployMode, err := provider.CheckVpcPresenceAndGetDeployMode(d)
		if err != nil {
			return resource.RetryableError(err)
		}
		rtrDeployMode = deployMode
		return nil
	})
	if err != nil {
		return errors.New("Could not find the VPC in CVaaS.(Try terraform apply again)")
	}

	err = d.Set("deploy_mode", rtrDeployMode)
	if err != nil {
		return fmt.Errorf("Failed to set deploy mode in router resource :%v", err)
	}

	// Relies on deploy_mode being set
	err = validateDeployModeWithRole(d)
	if err != nil {
		return err
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
