// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package device

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Option defines a command-line option accepted by a device.
type Option struct {
	Description string
	Default     string
	Pattern     string
	Required    bool
}

// SanitizedOptions takes the map of device option keys and values
// passed in at the command line and checks it against the device
// or manager's exported list of accepted options, returning an
// error if there are inappropriate or missing options.
func SanitizedOptions(options map[string]Option,
	config map[string]string) (map[string]string, error) {
	sopt := make(map[string]string)

	// Check whether the user gave us bad options.
	for k, v := range config {
		o, ok := options[k]
		if !ok {
			return nil, fmt.Errorf("Bad option '%s'", k)
		}

		// Check whether the user's string, if non-empty, matches the
		// option's defined pattern.
		if o.Pattern != "" && v != "" {
			re, err := regexp.Compile(o.Pattern)
			if err != nil {
				return nil, err
			}
			re.Longest()
			fs := re.FindString(v)
			if fs != v {
				return nil, fmt.Errorf("Value for option '%s' ('%s') does "+
					"not match regular expression '%s'", k, v, o.Pattern)
			}
		}
		sopt[k] = v
	}

	// Check that all required options were specified, and fill in
	// any others with defaults. Also check that the defaults are
	// consistent with the provided patterns.
	for k, v := range options {
		if v.Pattern != "" && v.Default != "" {
			re, err := regexp.Compile(v.Pattern)
			if err != nil {
				return nil, err
			}
			re.Longest()
			fs := re.FindString(v.Default)
			if fs != v.Default {
				return nil, fmt.Errorf("Default value ('%s') for option "+
					"'%s' does not match regular expression '%s'",
					v.Default, k, v.Pattern)
			}
		}

		_, found := sopt[k]
		if v.Required && !found {
			return nil, fmt.Errorf("Required option '%s' not provided", k)
		}
		if !found {
			sopt[k] = v.Default
		}
	}

	return sopt, nil
}

// Create map of option key to description.
func helpDesc(options map[string]Option) map[string]string {
	hd := make(map[string]string)

	for k, v := range options {
		desc := v.Description
		// Add default if there's a non-empty one.
		if v.Default != "" {
			desc = desc + " (default " + v.Default + ")"
		}
		if v.Required {
			k = k + " (required)"
		}
		hd[k] = desc
	}
	return hd
}

// GetStringOption returns the option specified by optionName as a
// string.
func GetStringOption(optionName string,
	options map[string]string) (string, error) {
	o, ok := options[optionName]
	if !ok {
		return "", fmt.Errorf("No option '%s'", optionName)
	}
	return o, nil
}

// GetBoolOption returns the option specified by optionName as a boolean.
func GetBoolOption(optionName string,
	options map[string]string) (bool, error) {
	o, ok := options[optionName]
	if !ok {
		return false, fmt.Errorf("No option '%s'", optionName)
	}
	return strconv.ParseBool(o)
}

// GetAddressOption returns the option specified by optionName as a
// validated IP address or hostname.
func GetAddressOption(optionName string,
	options map[string]string) (string, error) {
	addr, ok := options[optionName]
	if !ok {
		return "", fmt.Errorf("No option '%s'", optionName)
	}

	// Validate IP
	ip := net.ParseIP(addr)
	if ip != nil {
		return ip.String(), nil
	}

	// Try for hostname if it's not an IP
	addrs, err := net.LookupIP(addr)
	if err != nil {
		return "", err
	}
	return addrs[0].String(), nil
}

// GetPortOption returns the option specified by optionName as a validated
// port number.
func GetPortOption(optionName string, options map[string]string) (string, error) {
	p, ok := options[optionName]
	if !ok {
		return "", fmt.Errorf("No option '%s'", optionName)
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return "", err
	}
	if port < 0 || port > 65535 {
		return "", fmt.Errorf("Invalid port number %d", port)
	}
	return p, nil
}

// GetDurationOption returns the option specified by optionName as a
// time.Duration.
func GetDurationOption(optionName string,
	options map[string]string) (time.Duration, error) {
	o, ok := options[optionName]
	if !ok {
		return 0, fmt.Errorf("No option '%s'", optionName)
	}
	return time.ParseDuration(o)
}

// GetStringListOption returns the option specified by optionName as
// a string slice.
func GetStringListOption(optionName string,
	options map[string]string) ([]string, error) {
	o, ok := options[optionName]
	if !ok {
		return nil, fmt.Errorf("No option '%s'", optionName)
	}
	return strings.Split(o, ","), nil
}
