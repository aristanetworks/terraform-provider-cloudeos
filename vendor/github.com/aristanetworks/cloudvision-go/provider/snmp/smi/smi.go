// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package smi

import (
	"fmt"

	// Ensure mibs package doesn't get pruned by dependency management systems
	_ "github.com/aristanetworks/cloudvision-go/provider/snmp/smi/mibs"
)

// Object describes an SMI object.
type Object struct {
	Access      Access
	Description string
	Indexes     []string
	Kind        Kind
	Module      string
	Name        string
	Oid         string
	Status      Status
	Parent      *Object
	Children    []*Object
}

func (o *Object) String() string {
	s := fmt.Sprintf("{Name: %s", o.Name)
	s += fmt.Sprintf(", OID: %s", o.Oid)
	s += fmt.Sprintf(", Module: %s", o.Module)
	s += fmt.Sprintf(", Access: %s", o.Access)
	if len(o.Indexes) > 0 {
		s += fmt.Sprintf(", Indexes: %v", o.Indexes)
	}
	s += fmt.Sprintf(", Kind: %s", o.Kind)
	return s + "}"
}

// Import describes an imported object and the module it's imported from.
type Import struct {
	Object string
	Module string
}

func (i Import) String() string {
	return fmt.Sprintf("Object: %s, Module: %s", i.Object, i.Module)
}

// Module describes an SMI module. It contains an Object tree and a set
// of imports.
type Module struct {
	Name       string
	ObjectTree []*Object
	Imports    []Import
}

// Kind describes whether an SMI object is a table, row, column,
// scalar, or something else.
type Kind int

const (
	KindUnknown Kind = iota
	KindObject  Kind = 1 << (iota - 1)
	KindScalar
	KindTable
	KindRow
	KindColumn
	KindNotification
	KindGroup
	KindCompliance
	KindCapabilities
	KindAny Kind = 0xffff
)

func (k Kind) String() string {
	m := map[Kind]string{
		KindUnknown:      "Unknown",
		KindObject:       "Object",
		KindScalar:       "Scalar",
		KindTable:        "Table",
		KindRow:          "Row",
		KindColumn:       "Column",
		KindNotification: "Notification",
		KindGroup:        "Group",
		KindCompliance:   "Compliance",
		KindCapabilities: "Capabilities",
		KindAny:          "Any",
	}
	if p, ok := m[k]; ok {
		return p
	}
	return "Unknown"
}

// Access describes an SMI object's access value.
type Access int

const (
	AccessUnknown Access = iota
	AccessNotAccessible
	AccessNotify
	AccessReadOnly
	AccessReadWrite
	AccessReadCreate
)

func (a Access) String() string {
	m := map[Access]string{
		AccessUnknown:       "unknown",
		AccessNotAccessible: "not-accessible",
		AccessNotify:        "accessible-for-notify",
		AccessReadOnly:      "read-only",
		AccessReadWrite:     "read-write",
		AccessReadCreate:    "read-create",
	}
	if s, ok := m[a]; ok {
		return s
	}
	return m[AccessUnknown]
}

func strToAccess(s string) Access {
	m := map[string]Access{
		"accessible-for-notify": AccessNotify,
		"not-accessible":        AccessNotAccessible,
		"read-only":             AccessReadOnly,
		"read-write":            AccessReadWrite,
		"read-create":           AccessReadCreate,
	}
	if a, ok := m[s]; ok {
		return a
	}
	return AccessUnknown
}

// Status describes an SMI object's status value.
type Status int

const (
	StatusUnknown Status = iota
	StatusCurrent
	StatusDeprecated
	StatusMandatory
	StatusObsolete
	StatusOptional
)

func (s Status) String() string {
	m := map[Status]string{
		StatusUnknown:    "unknown",
		StatusCurrent:    "current",
		StatusDeprecated: "deprecated",
		StatusMandatory:  "mandatory",
		StatusObsolete:   "obsolete",
		StatusOptional:   "optional",
	}
	if st, ok := m[s]; ok {
		return st
	}
	return m[StatusUnknown]
}

func strToStatus(s string) Status {
	m := map[string]Status{
		"unknown":    StatusUnknown,
		"current":    StatusCurrent,
		"deprecated": StatusDeprecated,
		"mandatory":  StatusMandatory,
		"obsolete":   StatusObsolete,
		"optional":   StatusOptional,
	}
	if st, ok := m[s]; ok {
		return st
	}
	return StatusUnknown
}

// parseObject is effectively an augmented Object. It contains all
// the information the parser needs for a given SMI object that
// doesn't make sense to include in the Object itself.
type parseObject struct {
	object         *Object
	parent         *parseObject
	children       []*parseObject
	table          bool
	subidentifiers []string
	decl           decl
	augments       string
}

type parseModule struct {
	name       string
	objectTree []*parseObject
	orphans    []*parseObject
	imports    []Import
}

type decl int

const (
	declUnknown decl = iota
	declImplicitType
	declTypeAssignment
	declImplSequenceOf
	declValueAssignment
	declObjectType
	declObjectIdentity
	declModuleIdentity
	declNotificationType
	declTrapType
	declObjectGroup
	declNotificationGroup
	declModuleCompliance
	declAgentCapabilities
	declTextualConvention
	declMacro
	declComplGroup
	declComplObject
	declImplObject
	declModule
	declExtension
	declTypedef
	declObject
	declScalar
	declTable
	declRow
	declColumn
	declNotification
	declGroup
	declCompliance
	declIdentity
	declClass
	declAttribute
	declEvent
)
