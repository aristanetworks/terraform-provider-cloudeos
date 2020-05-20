// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package smi

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (po *parseObject) setKind() {
	if po.object.Kind != KindUnknown {
		return
	}

	switch po.decl {
	case declModuleIdentity, declValueAssignment, declObjectIdentity:
		po.object.Kind = KindObject
	case declObjectType:
		if po.table {
			po.object.Kind = KindTable
		} else if po.object.Parent == nil {
			break
		} else if po.object.Parent.Kind == KindTable {
			po.object.Kind = KindRow
		} else if po.object.Parent.Kind == KindRow ||
			len(po.object.Parent.Indexes) > 0 {
			po.object.Kind = KindColumn
		} else if po.object.Parent.Indexes == nil {
			po.object.Kind = KindScalar
		}
	case declNotificationType, declTrapType:
		po.object.Kind = KindNotification
	case declObjectGroup, declNotificationGroup:
		po.object.Kind = KindGroup
	case declModuleCompliance:
		po.object.Kind = KindCompliance
	case declAgentCapabilities:
		po.object.Kind = KindCapabilities
	}

	if po.object.Kind != KindUnknown {
		for _, c := range po.children {
			c.setKind()
		}
	}
}

func (po *parseObject) setModule(module string) {
	if po.object.Module != "" {
		return
	}

	po.object.Module = module

	for _, c := range po.children {
		c.setModule(module)
	}
}

func (yys *yySymType) linkToParent(po *parseObject) {
	// Check whether its Parent is in the objectMap, and if so link
	// the Parent and child.
	parentId := ""
	if len(strings.Split(po.object.Oid, ".")) > 0 {
		parentId = strings.Split(po.object.Oid, ".")[0]
	}
	if o, ok := yys.objectMap[parentId]; ok {
		o.object.Children = append(o.object.Children, po.object)
		o.children = append(o.children, po)
		po.object.Parent = o.object
		po.parent = o
	} else {
		yys.orphans = append(yys.orphans, po)
	}
}

func (yys *yySymType) addObject(po *parseObject) {
	if yys.objects == nil {
		yys.objects = []*parseObject{}
	}
	if yys.objectMap == nil {
		yys.objectMap = make(map[string]*parseObject)
	}
	if yys.orphans == nil {
		yys.orphans = []*parseObject{}
	}
	if po == nil {
		return
	}

	yys.objectMap[po.object.Name] = po

	yys.linkToParent(po)

	// Set the object's Kind.
	po.setKind()

	// Add it as a top-level object if it has no parent.
	if po.parent == nil {
		yys.objects = append(yys.objects, po)
	}
}

func (yys *yySymType) addModule(module *parseModule) {
	if yys.modules == nil {
		yys.modules = []*parseModule{}
	}
	if module == nil {
		return
	}
	yys.modules = append(yys.modules, module)
}

func (yys *yySymType) setDecl(d decl) {
	if yys.object != nil {
		yys.object.decl = d
	}
}

func declStr(d decl) string {
	m := map[decl]string{
		declUnknown:           "Unknown",
		declImplicitType:      "ImplicitType",
		declTypeAssignment:    "TypeAssignment",
		declImplSequenceOf:    "ImplSequenceOf",
		declValueAssignment:   "ValueAssignment",
		declObjectType:        "ObjectType",
		declObjectIdentity:    "ObjectIdentity",
		declModuleIdentity:    "ModuleIdentity",
		declNotificationType:  "NotificationType",
		declTrapType:          "TrapType",
		declObjectGroup:       "ObjectGroup",
		declNotificationGroup: "NotificationGroup",
		declModuleCompliance:  "ModuleCompliance",
		declAgentCapabilities: "AgentCapabilities",
		declTextualConvention: "TextualConvention",
		declMacro:             "Macro",
		declComplGroup:        "ComplGroup",
		declComplObject:       "ComplObject",
		declImplObject:        "ImplObject",
		declModule:            "Module",
		declExtension:         "Extension",
		declTypedef:           "Typedef",
		declObject:            "Object",
		declScalar:            "Scalar",
		declTable:             "Table",
		declRow:               "Row",
		declColumn:            "Column",
		declNotification:      "Notification",
		declGroup:             "Group",
		declCompliance:        "Compliance",
		declIdentity:          "Identity",
		declClass:             "Class",
		declAttribute:         "Attribute",
		declEvent:             "Event",
	}
	if p, ok := m[d]; ok {
		return p
	}
	panic(fmt.Sprintf("Bad decl %d", d))
}

type importUpgrades struct {
	// default replacement module
	defaultModule string
	// replacement module per object name
	objects map[string]string
}

// knownImportUpgrades defines a set of objects that, when imported,
// should actually redirect to newer versions in different modules.
var knownImportUpgrades = map[string]importUpgrades{
	"RFC1065-MIB": importUpgrades{defaultModule: "RFC1155-MIB"},
	"RFC1066-MIB": importUpgrades{defaultModule: "RFC1156-MIB"},
	"RFC1156-MIB": importUpgrades{defaultModule: "RFC1158-MIB"},
	"RFC1158-MIB": importUpgrades{defaultModule: "RFC1213-MIB"},
	"RFC1155-MIB": importUpgrades{defaultModule: "SNMPv2-SMI"},
	"RFC1213-MIB": importUpgrades{
		defaultModule: "RFC1213-MIB",
		objects: map[string]string{
			"mib-2":        "SNMPv2-SMI",
			"sys":          "SNMPv2-MIB",
			"if":           "IF-MIB",
			"interfaces":   "IF-MIB",
			"ip":           "IP-MIB",
			"icmp":         "IP-MIB",
			"tcp":          "TCP-MIB",
			"udp":          "UDP-MIB",
			"transmission": "SNMPv2-SMI",
			"snmp":         "SNMPv2-MIB",
		},
	},
	"RFC1231-MIB": importUpgrades{defaultModule: "TOKENRING-MIB"},
	"RFC1271-MIB": importUpgrades{defaultModule: "RMON-MIB"},
	"RFC1286-MIB": importUpgrades{
		defaultModule: "BRIDGE-MIB",
		objects: map[string]string{
			"SOURCE-ROUTING-MIB": "dot1dSr",
		},
	},
	"RFC1315-MIB": importUpgrades{defaultModule: "FRAME-RELAY-DTE-MIB"},
	"RFC1316-MIB": importUpgrades{defaultModule: "CHARACTER-MIB"},
	"RFC1406-MIB": importUpgrades{defaultModule: "DS1-MIB"},
	"RFC-1213":    importUpgrades{defaultModule: "RFC1213-MIB"},
}

func moduleUpgrade(module, object string) string {
	for {
		mr, ok := knownImportUpgrades[module]
		if !ok {
			break
		}
		for prefix, rm := range mr.objects {
			if strings.HasPrefix(object, prefix) {
				return rm
			}
		}
		if module == mr.defaultModule {
			break
		}
		module = mr.defaultModule
	}
	return module
}

func parseFile(filename string) (map[string]*parseModule, error) {
	yyErrorVerbose = true
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lx, err := newLexer(string(b))
	if err != nil {
		return nil, err
	}
	if r := yyParse(lx); r != 0 {
		return nil, fmt.Errorf("yyParse returned %d", r)
	}
	return lx.modules, nil
}

func parseFiles(files ...string) (map[string]*parseModule, error) {
	modules := make(map[string]*parseModule)
	for _, f := range files {
		err := filepath.Walk(f,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				// Don't try to parse files with extensions other than .mib/.MIB
				ext := filepath.Ext(path)
				if ext != "" && strings.ToLower(ext) != "mib" {
					return nil
				}
				m, err := parseFile(path)
				if err != nil {
					return err
				}
				for k, v := range m {
					modules[k] = v
				}
				return nil
			})
		if err != nil {
			return nil, err
		}
	}
	return modules, nil
}
