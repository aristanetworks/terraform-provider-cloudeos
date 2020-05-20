// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package snmpoc

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	pgnmi "github.com/aristanetworks/cloudvision-go/provider/gnmi"
	"github.com/aristanetworks/cloudvision-go/provider/snmp/pdu"
	"github.com/aristanetworks/cloudvision-go/provider/snmp/smi"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/soniah/gosnmp"
)

// NewTranslator returns a Translator.
func NewTranslator(mibStore smi.Store, gs *gosnmp.GoSNMP) (*Translator, error) {
	ps, err := pdu.NewStore(mibStore)
	if err != nil {
		return nil, err
	}

	return &Translator{
		Getter:                 gs.Get,
		gosnmp:                 gs,
		gosnmpLock:             &sync.Mutex{},
		Logger:                 &nonlogger{},
		mapperData:             &sync.Map{},
		Mappings:               DefaultMappings(),
		mibStore:               mibStore,
		pathsMappingGroups:     make(map[string]map[string]*mappingGroup),
		pduStore:               ps,
		pollLock:               &sync.Mutex{},
		successfulMappings:     make(map[string]Mapper),
		successfulMappingsLock: &sync.RWMutex{},
		Walker:                 gs.BulkWalk,
	}, nil
}

// Logger defines an interface for logging.
type Logger interface {
	Info(args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})
	Debug(args ...interface{})
	Debugln(args ...interface{})
	Debugf(format string, args ...interface{})
}

// model describes a set of paths rooted at rootPath for which
// we want to produce updates.
type model struct {
	name         string
	rootPath     string
	dependencies []string
	snmpGetOIDs  []string
	snmpWalkOIDs []string
}

func (m *model) Copy() *model {
	m2 := &model{
		name:         m.name,
		rootPath:     m.rootPath,
		dependencies: make([]string, len(m.dependencies)),
		snmpGetOIDs:  make([]string, len(m.snmpGetOIDs)),
		snmpWalkOIDs: make([]string, len(m.snmpWalkOIDs)),
	}
	_ = copy(m2.dependencies, m.dependencies)
	_ = copy(m2.snmpGetOIDs, m.snmpGetOIDs)
	_ = copy(m2.snmpWalkOIDs, m.snmpWalkOIDs)

	return m2
}

// A mappingGroup contains a set of paths and their associated models.
type mappingGroup struct {
	name        string
	models      map[string]*model
	updatePaths map[string][]string
}

type nonlogger struct{}

func (n *nonlogger) Info(args ...interface{})                  {}
func (n *nonlogger) Infoln(args ...interface{})                {}
func (n *nonlogger) Infof(format string, args ...interface{})  {}
func (n *nonlogger) Debug(args ...interface{})                 {}
func (n *nonlogger) Debugln(args ...interface{})               {}
func (n *nonlogger) Debugf(format string, args ...interface{}) {}

// Translator defines an interface for producing translations from a
// set of received SNMP PDUs to a set of gNMI updates.
type Translator struct {
	// auxiliary data stores
	pduStore           pdu.Store
	mibStore           smi.Store
	mapperData         *sync.Map
	pathsMappingGroups map[string]map[string]*mappingGroup

	// mapping state
	Mappings               map[string][]Mapper
	successfulMappings     map[string]Mapper
	successfulMappingsLock *sync.RWMutex

	// gosnmp state
	gosnmp          *gosnmp.GoSNMP
	gosnmpLock      *sync.Mutex
	gosnmpConnected bool

	// to ensure non-overlapping polls
	pollLock *sync.Mutex

	// logging
	Logger Logger

	// alternative get, walk, and time.Now for testing
	Mock   bool
	Getter func([]string) (*gosnmp.SnmpPacket, error)
	Walker func(string, gosnmp.WalkFunc) error
}

// Poll generates a set of updates for the paths specified. It
// accepts exact paths and regular expressions. Any path in the
// translator's mapping list that matches the provided expression
// is added to the set of updates to produce. It performs a poll for
// any required SNMP data and translates that data into gNMI updates,
// which it then transmits via the provided gNMI client's Set method.
func (t *Translator) Poll(ctx context.Context, client gnmi.GNMIClient,
	paths []string) error {
	t.pollLock.Lock()
	defer t.pollLock.Unlock()

	mappingGroups, err := t.mappingGroupsFromPaths(paths)
	if err != nil {
		return err
	}

	if len(mappingGroups) == 0 {
		return errors.New("no models to translate")
	}

	// Clear data stores in preparation for filling them up again.
	if err := t.pduStore.Clear(); err != nil {
		return err
	}
	t.mapperData.Range(func(k interface{}, v interface{}) bool {
		// Ideally we could just do t.mapperData = &sync.Map{} but
		// the compiler says this is unsafe, so we have to use Range.
		t.mapperData.Delete(k)
		return true
	})

	// Produce updates for each mapping group.
	var wg sync.WaitGroup
	setReqCh := make(chan *gnmi.SetRequest, len(mappingGroups))
	errc := make(chan error)
	for _, mg := range mappingGroups {
		wg.Add(1)
		go t.mappingGroupUpdates(ctx, mg, &wg, setReqCh, errc)
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sr := <-setReqCh:
			if _, err := client.Set(ctx, sr); err != nil {
				return err
			}
		case <-done:
			return nil
		case err := <-errc:
			return err
		}
	}
}

// updates produces updates for the provided set of paths.
func (t *Translator) updates(paths []string) ([]*gnmi.Update, error) {
	updates := []*gnmi.Update{}
	for _, path := range paths {
		// If we have a mapping that already worked, use it.
		t.successfulMappingsLock.RLock()
		mapping, ok := t.successfulMappings[path]
		t.successfulMappingsLock.RUnlock()
		if ok {
			u, err := mapping(t.mibStore, t.pduStore, t.mapperData, t.Logger)
			if err != nil {
				return nil, err
			}
			updates = append(updates, u...)
			continue
		}

		// Otherwise, try each mapping in order.
		mappings, ok := t.Mappings[path]
		if !ok {
			return nil, fmt.Errorf("No mapping supplied for path %v", path)
		}
		for _, mapping := range mappings {
			u, err := mapping(t.mibStore, t.pduStore, t.mapperData, t.Logger)
			if err != nil {
				return nil, err
			} else if len(u) == 0 {
				continue
			}
			t.successfulMappingsLock.Lock()
			t.successfulMappings[path] = mapping
			t.successfulMappingsLock.Unlock()
			updates = append(updates, u...)
			break
		}
	}
	return updates, nil
}

var supportedModels = map[string]*model{
	"interfaces": &model{
		name:         "interfaces",
		rootPath:     "/interfaces",
		snmpWalkOIDs: []string{"ifTable", "ifXTable"},
	},
	"system": &model{
		name:     "system",
		rootPath: "/system",
		snmpGetOIDs: []string{"sysName.0", "lldpLocSysName.0", "hrSystemUptime.0",
			"sysUpTimeInstance"},
	},
	"lldp": &model{
		name:         "lldp",
		rootPath:     "/lldp",
		dependencies: []string{"interfaces"},
		snmpWalkOIDs: []string{"lldpLocalSystemData", "lldpRemTable", "lldpStatistics",
			"lldpV2LocalSystemData", "lldpV2RemTable", "lldpV2Statistics"},
	},
	"platform": &model{
		name:         "platform",
		rootPath:     "/components",
		snmpWalkOIDs: []string{"entPhysicalEntry"},
	},
}

var supportedMappingGroups = map[string]*mappingGroup{
	"interfaces-lldp": &mappingGroup{
		name: "interfaces-lldp",
		models: map[string]*model{
			"interfaces": supportedModels["interfaces"],
			"lldp":       supportedModels["lldp"],
		},
	},
	"system": &mappingGroup{
		name: "system",
		models: map[string]*model{
			"system": supportedModels["system"],
		},
	},
	"platform": &mappingGroup{
		name: "platform",
		models: map[string]*model{
			"platform": supportedModels["platform"],
		},
	},
}

func (t *Translator) storePDU(pdu gosnmp.SnmpPDU) error {
	return t.pduStore.Add(&pdu)
}

func (t *Translator) getSNMPData(mg *mappingGroup) error {
	t.gosnmpLock.Lock()
	defer t.gosnmpLock.Unlock()

	// Connect to target.
	if !t.gosnmpConnected && !t.Mock {
		if err := t.gosnmp.Connect(); err != nil {
			return err
		}
		t.gosnmpConnected = true
		t.Logger.Infoln("gosnmp.Connect complete")
	}

	// Get SNMP data for each model in this mappingGroup.
	for _, model := range mg.models {
		// Walk
		for _, oid := range model.snmpWalkOIDs {
			t.Logger.Debugf("SNMP Walk (OID = %s)", oid)
			if err := t.Walker(oid, t.storePDU); err != nil {
				t.Logger.Infof("Error walking OID %s: %s", oid, err)
			} else {
				t.Logger.Debugf("SNMP Walk complete (OID = %s)", oid)
			}
		}

		// Get
		if len(model.snmpGetOIDs) == 0 {
			continue
		}
		t.Logger.Debugf("SNMP Get (OIDs = %s)",
			strings.Join(model.snmpGetOIDs, " "))
		pkt, err := t.Getter(model.snmpGetOIDs)
		if err != nil {
			t.Logger.Infof("Error getting OIDs %s: %s",
				strings.Join(model.snmpGetOIDs, " "), err)
			return nil
		} else if pkt == nil {
			t.Logger.Info("SNMP Get returned nil packet")
			return nil
		} else {
			t.Logger.Debugf("SNMP Get complete. pkt = %v, err = %v", pkt, err)
		}

		if pkt.Error != gosnmp.NoError {
			errstr, ok := SNMPErrCodes[pkt.Error]
			if !ok {
				errstr = "Unknown error"
			}
			t.Logger.Infof("SNMP Get: Error in packet (%v): %s",
				pkt, errstr)
		}

		for _, pdu := range pkt.Variables {
			if err = t.storePDU(pdu); err != nil {
				t.Logger.Infof("Error storing PDU: %s", err)
			}
		}
	}

	return nil
}

func (t *Translator) mappingGroupUpdates(ctx context.Context,
	mg *mappingGroup, wg *sync.WaitGroup, setReqCh chan *gnmi.SetRequest,
	errc chan error) {
	defer wg.Done()

	// Get SNMP data.
	if err := t.getSNMPData(mg); err != nil {
		errc <- err
	}

	// Produce updates and hand a SetRequest to the gNMI client.
	setRequest := new(gnmi.SetRequest)
	for modelName, model := range mg.models {
		setRequest.Delete = append(setRequest.Delete,
			pgnmi.PathFromString(model.rootPath))
		if up, ok := mg.updatePaths[modelName]; ok {
			updates, err := t.updates(up)
			if err != nil {
				errc <- err
				return
			}
			setRequest.Replace = append(setRequest.Replace, updates...)
			t.Logger.Debugf("Replace for mapping group %s, model %s has %d updates",
				mg.name, modelName, len(setRequest.Replace))
		} else {
			errc <- fmt.Errorf("No updatePath entries for model '%s'", modelName)
			return
		}
	}

	setReqCh <- setRequest
}

// A mappingGroup is a set of related translations that share dependencies.
func (t *Translator) mappingGroupsFromPaths(paths []string) (map[string]*mappingGroup,
	error) {
	phb := md5.Sum([]byte(strings.Join(paths, "")))
	pathHash := string(phb[:])

	// Get the paths of interest from the provided paths + patterns.
	if len(paths) == 0 {
		// If we're not given any paths, use all supported paths.
		for p := range t.Mappings {
			paths = append(paths, p)
		}
	} else {
		// Check whether we've already created mapping groups for these
		// paths.
		if mgs, ok := t.pathsMappingGroups[pathHash]; ok {
			return mgs, nil
		}

		// Expand the set of provided paths to include all paths matching
		// regexes.
		expPaths := []string{}
		for _, p := range paths {
			if _, ok := t.Mappings[p]; ok {
				expPaths = append(expPaths, p)
				continue
			} else {
				for mp := range t.Mappings {
					m, err := regexp.MatchString(p, mp)
					if err != nil {
						return nil, err
					}
					if m {
						expPaths = append(expPaths, mp)
					}
				}
			}
		}
		if len(expPaths) == 0 {
			return nil, fmt.Errorf("Provided paths expand to zero paths: %v", paths)
		}
		paths = expPaths
	}

	// Pare down mappingGroups to include only the models we need for the
	// provided paths.
	reducedMg := map[string]*mappingGroup{}
	for _, mg := range supportedMappingGroups {
		for _, mod := range mg.models {
			for _, p := range paths {
				match, err := regexp.MatchString(fmt.Sprintf("^%s/.*",
					mod.rootPath), p)
				if err != nil {
					return nil, err
				}
				if match {
					if _, ok := reducedMg[mg.name]; !ok {
						reducedMg[mg.name] = &mappingGroup{
							name:        mg.name,
							models:      map[string]*model{},
							updatePaths: map[string][]string{},
						}
					}
					cmg := reducedMg[mg.name]
					if _, ok := cmg.models[mod.name]; !ok {
						cmg.models[mod.name] = mod.Copy()

						// Swap text OIDs out for their numeric equivalents.
						for i, oid := range cmg.models[mod.name].snmpGetOIDs {
							obj := t.mibStore.GetObject(oid)
							if obj != nil {
								cmg.models[mod.name].snmpGetOIDs[i] = obj.Oid
								// Add back ".0" for scalars
								if oid[len(oid)-2:] == ".0" {
									cmg.models[mod.name].snmpGetOIDs[i] += ".0"
								}
							}
						}
						for i, oid := range cmg.models[mod.name].snmpWalkOIDs {
							obj := t.mibStore.GetObject(oid)
							if obj != nil {
								cmg.models[mod.name].snmpWalkOIDs[i] = obj.Oid
							}
						}
					}
					cmg.updatePaths[mod.name] = append(cmg.updatePaths[mod.name], p)
				}
			}
		}
	}

	// Store mapping groups for these paths.
	t.pathsMappingGroups[pathHash] = reducedMg

	return reducedMg, nil
}
