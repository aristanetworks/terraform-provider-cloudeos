// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package cloudeos

import (
	"fmt"
	"net"
)

func validateCIDRBlock(val interface{}, key string) (warns []string, errors []error) {
	if err := validateCIDR(val.(string)); err != nil {
		errors = append(errors, err)
	}
	return
}

func validateCIDR(cidrString string) error {
	_, _, err := net.ParseCIDR(cidrString)
	if err != nil {
		return fmt.Errorf("%s is not a valid CIDR. %w", cidrString, err)
	}
	return nil
}

func validateIPList(val interface{}, key string) (warns []string, errors []error) {
	ipList := val.([]string)
	for _, ipStr := range ipList {
		if err := validateIP(ipStr); err != nil {
			errors = append(errors, err)
			return
		}
	}
	return
}

func validateIP(ipStr string) error {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return fmt.Errorf("%s is not a valid IP address", ipStr)
	}
	return nil
}
