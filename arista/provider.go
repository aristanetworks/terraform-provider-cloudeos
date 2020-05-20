// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package main

import (
	"errors"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

//Provider function which defines the Terraform provider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"cvaas_server": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Cvp server hostname / ip address and port for terraform" +
					" client to authenticate. It must be in format of <hostname>" +
					" or <hostname>:<port>",
			},
			"service_account_web_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Service account_web_token for user 'terraform'",
			},
			"cvaas_domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "CVaaS Domain name",
				Default:     "apiserver.cv-play.corp.arista.io",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"arista_vpc_config":  aristaVpcConfig(),
			"arista_vpc":         aristaVpc(),
			"arista_veos_config": aristaVeosConfig(),
			"arista_veos_status": aristaVeosStatus(),
			"arista_subnet":      aristaSubnet(),
			"arista_topology":    aristaTopology(),
			"arista_clos":        aristaClos(),
			"arista_wan":         aristaWan(),
		},

		ConfigureFunc: configureAristaProvider,
	}
}

func configureAristaProvider(d *schema.ResourceData) (interface{}, error) {
	var cfg AristaProvider
	cfg.server = d.Get("cvaas_server").(string)
	cfg.srvcAcctToken = d.Get("service_account_web_token").(string)
	cfg.cvaasDomain = d.Get("cvaas_domain").(string)
	if cfg.server == "" || cfg.srvcAcctToken == "" {
		return nil, errors.New("Provider not configured correctly")
	}

	return cfg, nil
}
