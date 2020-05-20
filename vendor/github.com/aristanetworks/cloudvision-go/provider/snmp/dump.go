// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package snmp

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/soniah/gosnmp"
)

// This file contains functionality useful for producing a dump from
// a series of SNMP requests and responses, or for reading such a
// dump. The format of responses is what you'd see running
// `snmpwalk -O ne`, so:
//
// <OID> = <type>: <value>
//
// For example:
//
// .1.3.6.1.2.1.47.1.1.1.1.13.156025601 = STRING: Ucd90120
//
// The request format is:
//
// <request-type>: <OID>
//
// For example:
//
// WALK: .1.3.6.1.2.1.47.1.1.1.1

// SNMP PDU types of interest.
const (
	octstr              = gosnmp.OctetString
	counter             = gosnmp.Counter32
	counter64           = gosnmp.Counter64
	integer             = gosnmp.Integer
	timeticks           = gosnmp.TimeTicks // nolint: deadcode
	octstrTypeString    = "STRING"
	hexstrTypeString    = "Hex-STRING"
	integerTypeString   = "INTEGER"
	counterTypeString   = "Counter32"
	counter64TypeString = "Counter64"
	getString           = "GET"       // nolint: deadcode
	walkString          = "WALK"      // nolint: deadcode
	timeticksString     = "Timeticks" // nolint: deadcode
)

// PDU creation wrapper.
func pdu(name string, t gosnmp.Asn1BER, val interface{}) *gosnmp.SnmpPDU {
	return &gosnmp.SnmpPDU{
		Name:  name,
		Type:  t,
		Value: val,
	}
}

// Get OID, type, and value from a dumped PDU.
func parsePDU(line string) (oid, pduTypeString, value string) {
	t := strings.Split(line, " = ")
	if len(t) < 2 {
		return "", "", ""
	}
	oid = t[0]
	t = strings.Split(t[1], ": ")
	pduTypeString = t[0]
	if len(t) >= 2 {
		// Handle case where value is of format "chassis(3)".
		s := strings.Split(strings.Split(t[1], ")")[0], "(")
		value = s[0]
		if len(s) > 1 {
			value = s[1]
		}
		value = strings.Trim(value, "\"")
	} else {
		pduTypeString = strings.Split(t[0], ":")[0]
	}
	return oid, pduTypeString, value
}

// pduFromString returns a PDU from a string representation of a PDU.
// If it sees something it doesn't like, it returns nil.
func pduFromString(s string) *gosnmp.SnmpPDU {
	oid, pduTypeString, val := parsePDU(s)
	if oid == "" {
		return nil
	}

	var pduType gosnmp.Asn1BER
	var value interface{}
	switch pduTypeString {
	case integerTypeString:
		pduType = integer
		v, _ := strconv.ParseInt(val, 10, 32)
		value = int(v)
	case octstrTypeString:
		pduType = octstr
		value = []byte(val)
	case hexstrTypeString:
		pduType = octstr
		s := strings.Replace(val, " ", "", -1)
		value, _ = hex.DecodeString(s)
	case counterTypeString:
		pduType = counter
		v, _ := strconv.ParseUint(val, 10, 32)
		value = uint(v)
	case counter64TypeString:
		pduType = counter64
		v, _ := strconv.ParseUint(val, 10, 64)
		value = v
	default:
		return nil
	}
	return pdu(oid, pduType, value)
}

// PDUsFromString converts a set of formatted SNMP responses (as
// returned by snmpwalk -v3 -O ne <target> <oid>) to PDUs.
func PDUsFromString(s string) []*gosnmp.SnmpPDU {
	pdus := make([]*gosnmp.SnmpPDU, 0)
	for _, line := range strings.Split(s, "\n") {
		pdu := pduFromString(line)
		if pdu != nil {
			pdus = append(pdus, pdu)
		}
	}
	return pdus
}
