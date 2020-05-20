// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package devices

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aristanetworks/cloudvision-go/device"
	"github.com/aristanetworks/cloudvision-go/provider"
	psnmp "github.com/aristanetworks/cloudvision-go/provider/snmp"
	"github.com/soniah/gosnmp"
)

var options = map[string]device.Option{
	"a": device.Option{
		Description: "SNMPv3 authentication protocol",
		Pattern:     `sha|SHA|md5|MD5`,
	},
	"A": device.Option{
		Description: "SNMPv3 authentication key",
	},
	"address": device.Option{
		Description: "Hostname or address of device",
		Required:    true,
	},
	"port": device.Option{
		Description: "Device SNMP port to use",
		Default:     "161",
	},
	"c": device.Option{
		Description: "SNMP community string",
	},
	"l": device.Option{
		Description: "SNMPv3 security level (noAuthNoPriv|authNoPriv|authPriv)",
		Default:     "authPriv",
		Pattern:     `noAuthNoPriv|authNoPriv|authPriv`,
	},
	"mibs": device.Option{
		Description: "Comma-separated list of mib files/directories",
		Required:    true,
	},
	"pollInterval": device.Option{
		Description: "Polling interval, with unit suffix (s/m/h)",
		Default:     "20s",
	},
	"u": device.Option{
		Description: "SNMPv3 security name",
	},
	"v": device.Option{
		Description: "SNMP version (2c|3)",
		Pattern:     `2c|3`,
		Default:     "2c",
	},
	"x": device.Option{
		Description: "SNMPv3 privacy protocol",
		Pattern:     `des|DES|aes|AES`,
	},
	"X": device.Option{
		Description: "SNMPv3 privacy key",
	},
}

func init() {
	device.Register("snmp", newSnmp, options)
}

type snmp struct {
	address      string
	authKey      string
	authProto    string
	community    string
	level        string
	mibs         []string
	pollInterval time.Duration
	port         uint16
	privacyKey   string
	privacyProto string
	securityName string
	systemID     string
	version      string
	v3Params     *psnmp.V3Params
	v            gosnmp.SnmpVersion
	snmpProvider provider.GNMIProvider
}

// XXX NOTE: For now, we return an error rather than just returning false. We
// may want to rethink that in the future.
func (s *snmp) Alive() (bool, error) {
	_, err := s.snmpProvider.(*psnmp.Snmp).Alive()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *snmp) DeviceID() (string, error) {
	if s.systemID != "" {
		return s.systemID, nil
	}
	systemID, err := s.snmpProvider.(*psnmp.Snmp).DeviceID()
	if err != nil {
		return "", err
	}
	s.systemID = systemID
	return s.systemID, nil
}

func (s *snmp) Providers() ([]provider.Provider, error) {
	return []provider.Provider{s.snmpProvider}, nil
}

func (s *snmp) validateOptions() error {
	if s.version == "2c" {
		if s.community == "" {
			return errors.New("community string required for version 2c")
		}
		return nil
	}

	if s.securityName == "" {
		return errors.New("v3 is configured, so a username is required")
	}

	if s.level == "authNoPriv" || s.level == "authPriv" {
		if s.authProto == "" {
			return errors.New("auth is configured, so an authentication " +
				"protocol must be specified")
		}
		if s.authKey == "" {
			return errors.New("auth is configured, so an authentication " +
				"key must be specified")
		}
	}

	if s.level == "authPriv" {
		if s.privacyProto == "" {
			return errors.New("privacy is configured, so a privacy " +
				"protocol must be specified")
		}
		if s.privacyKey == "" {
			return errors.New("privacy is configured, so a privacy " +
				"key must be specified")
		}
	}
	return nil
}

func (s *snmp) formatOptions() (gosnmp.SnmpVersion, *psnmp.V3Params) {
	if s.version == "2c" {
		return gosnmp.Version2c, nil
	}

	v3Params := &psnmp.V3Params{
		SecurityModel: gosnmp.UserSecurityModel,
		UsmParams:     &gosnmp.UsmSecurityParameters{},
	}

	v3Params.UsmParams.UserName = s.securityName

	if s.level == "noAuthNoPriv" {
		v3Params.Level = gosnmp.NoAuthNoPriv
	} else if s.level == "authNoPriv" {
		v3Params.Level = gosnmp.AuthNoPriv
	} else if s.level == "authPriv" {
		v3Params.Level = gosnmp.AuthPriv
	}

	if strings.ToLower(s.authProto) == "sha" {
		v3Params.UsmParams.AuthenticationProtocol = gosnmp.SHA
	} else if strings.ToLower(s.authProto) == "md5" {
		v3Params.UsmParams.AuthenticationProtocol = gosnmp.MD5
	}

	if strings.ToLower(s.privacyProto) == "aes" {
		v3Params.UsmParams.PrivacyProtocol = gosnmp.AES
	} else if strings.ToLower(s.privacyProto) == "des" {
		v3Params.UsmParams.PrivacyProtocol = gosnmp.DES
	}
	v3Params.UsmParams.AuthenticationPassphrase = s.authKey
	v3Params.UsmParams.PrivacyPassphrase = s.privacyKey

	return gosnmp.Version3, v3Params
}

func (s *snmp) deviceConfigErr(err error) error {
	return fmt.Errorf("Configuration error for device %s: %v",
		s.address, err)
}

// XXX NOTE: The network operations here could fail on startup, and if
// they do, the error will be passed back to Collector and it will fail.
// Are we OK with this or should we be doing retries?
func newSnmp(options map[string]string) (device.Device, error) {
	s := &snmp{}
	var err error

	s.address, err = device.GetAddressOption("address", options)
	if err != nil {
		return nil, err
	}

	s.authKey, err = device.GetStringOption("A", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}

	s.authProto, err = device.GetStringOption("a", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}

	s.community, err = device.GetStringOption("c", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}

	s.level, err = device.GetStringOption("l", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}

	s.mibs, err = device.GetStringListOption("mibs", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}

	s.pollInterval, err = device.GetDurationOption("pollInterval", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}

	port, err := device.GetPortOption("port", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}
	portint, err := strconv.Atoi(port)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}
	s.port = uint16(portint)

	s.privacyKey, err = device.GetStringOption("X", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}

	s.privacyProto, err = device.GetStringOption("x", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}

	s.securityName, err = device.GetStringOption("u", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}

	s.version, err = device.GetStringOption("v", options)
	if err != nil {
		return nil, s.deviceConfigErr(err)
	}

	if err := s.validateOptions(); err != nil {
		return nil, s.deviceConfigErr(err)
	}

	s.v, s.v3Params = s.formatOptions()

	s.snmpProvider = psnmp.NewSNMPProvider(s.address, s.port, s.community,
		s.pollInterval, s.v, s.v3Params, s.mibs, false)

	return s, nil
}
