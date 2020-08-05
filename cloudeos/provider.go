// Copyright (c) 2020 Arista Networks, Inc.
// Use of this source code is governed by the Mozilla Public License Version 2.0
// that can be found in the LICENSE file.

package cloudeos

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

//Provider function which defines the Terraform provider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"cvaas_server": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Cvp server hostname / ip address and port for terraform" +
					" client to authenticate. It must be in format of <hostname>" +
					" or <hostname>:<port>",
			},
			"service_account_web_token": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service account web token",
			},
			"cvaas_domain": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "CVaaS Domain name",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"cloudeos_vpc_config":    cloudeosVpcConfig(),
			"cloudeos_vpc_status":    cloudeosVpcStatus(),
			"cloudeos_router_config": cloudeosRouterConfig(),
			"cloudeos_router_status": cloudeosRouterStatus(),
			"cloudeos_subnet":        cloudeosSubnet(),
			"cloudeos_topology":      cloudeosTopology(),
			"cloudeos_clos":          cloudeosClos(),
			"cloudeos_wan":           cloudeosWan(),
			"cloudeos_aws_vpn":       cloudeosAwsVpn(),
		},

		ConfigureFunc: configureCloudEOSProvider,
	}
}

func configureCloudEOSProvider(d *schema.ResourceData) (interface{}, error) {
	var cfg CloudeosProvider
	cfg.server = d.Get("cvaas_server").(string)
	cfg.srvcAcctToken = d.Get("service_account_web_token").(string)
	cfg.cvaasDomain = d.Get("cvaas_domain").(string)
	if cfg.server == "" || cfg.srvcAcctToken == "" || cfg.cvaasDomain == "" {
		return nil, errors.New("Provider not configured correctly")
	}

	return cfg, nil
}
