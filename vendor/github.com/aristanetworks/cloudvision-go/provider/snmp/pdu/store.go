// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package pdu

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aristanetworks/cloudvision-go/provider/snmp/smi"
	"github.com/soniah/gosnmp"
)

// Index represents a constraint on a PDU query. It consists of a
// name/OID indicating an index, and the value of that index.
type Index struct {
	Name  string
	Value string
}

// Store is an interface for adding SNMP PDUs and flexibly querying
// those stored PDUs.
type Store interface {
	Add(p *gosnmp.SnmpPDU) error
	Clear() error
	GetScalar(oid string) (*gosnmp.SnmpPDU, error)
	GetTabular(oid string, indexes ...Index) ([]*gosnmp.SnmpPDU, error)
}

// NewStore returns a new Store.
func NewStore(mibStore smi.Store) (Store, error) {
	return &store{
		scalars:  make(map[string]*gosnmp.SnmpPDU),
		columns:  make(map[string]*columnStore),
		mibStore: mibStore,
	}, nil
}

// allIndexMap maps the value of a set of indexes to a PDU. We store
// them by all indexes rather than just the value of the index in
// question because it allows for much faster set arithmetic.
type allIndexMap struct {
	values map[string]*gosnmp.SnmpPDU
}

// An indexStore holds all of a column's PDUs indexed by index value.
// For example, if a column "fooThing" has indexes "fooBar" and
// "fooBaz", the "fooThing" column will contain an indexStore for
// "fooBar" and "fooBaz", each of which will contain an allIndexMap
// for each index value in the column.
type indexStore struct {
	values map[string]*allIndexMap
}

// column stores columnar PDUs. It indexes those PDUs by full OID value
// (in `entries`) and by index (`indexes`).
type columnStore struct {
	indexes map[string]*indexStore
	entries map[string]*gosnmp.SnmpPDU
}

// store implements the Store interface. It holds scalar data in its
// `scalars` member, and columnar data is stored by column name in
// `columns`.
type store struct {
	scalars  map[string]*gosnmp.SnmpPDU
	columns  map[string]*columnStore
	mibStore smi.Store
	lock     sync.RWMutex
}

func (s *store) addScalar(p *gosnmp.SnmpPDU, o *smi.Object) error {
	s.scalars[o.Oid] = p
	return nil
}

func indexValues(pdu *gosnmp.SnmpPDU, o *smi.Object) []string {
	ss := strings.Split(pdu.Name, ".")
	return ss[(len(ss) - len(o.Parent.Indexes)):]
}

// IndexValues returns the index portion of the OID of the specified
// PDU.
func IndexValues(mibStore smi.Store, pdu *gosnmp.SnmpPDU) []string {
	o := mibStore.GetObject(pdu.Name)
	if o == nil {
		return nil
	}
	return indexValues(pdu, o)
}

// IndexValueByName returns the value of the index specified by
// indexName in the OID of the provided PDU.
func IndexValueByName(mibStore smi.Store, pdu *gosnmp.SnmpPDU,
	indexName string) (string, error) {
	// XXX TODO: Right now we assume that an index occupies a single
	// OID component, i.e., does not span any periods in the OID.
	// This assumption is not correct: String indexes, for example,
	// may span many OID components. We need to teach this method
	// to understand such cases.
	o := mibStore.GetObject(pdu.Name)
	if o == nil {
		return "", fmt.Errorf("No object for OID '%s'", pdu.Name)
	}
	for i, iname := range o.Parent.Indexes {
		if indexName == iname {
			return indexValues(pdu, o)[i], nil
		}
	}
	return "", fmt.Errorf("No index '%s' for OID '%s'", indexName, pdu.Name)
}

func (s *store) addTabular(p *gosnmp.SnmpPDU, o *smi.Object) error {
	if o.Parent == nil {
		return fmt.Errorf("OID %s has nil parent", p.Name)
	}
	if len(o.Parent.Indexes) == 0 {
		return fmt.Errorf("OID %s has no indexes", p.Name)
	}
	indexVals := indexValues(p, o)
	col, ok := s.columns[o.Oid]
	if !ok {
		col = &columnStore{
			indexes: make(map[string]*indexStore),
			entries: make(map[string]*gosnmp.SnmpPDU),
		}
		s.columns[o.Oid] = col
	}

	allIndexes := strings.Join(indexVals, ".")
	col.entries[allIndexes] = p

	for i, indexVal := range indexVals {
		indexName := o.Parent.Indexes[i]
		idx, ok := col.indexes[indexName]
		if !ok {
			idx = &indexStore{
				values: make(map[string]*allIndexMap),
			}
			col.indexes[indexName] = idx
		}
		if _, ok := idx.values[indexVal]; !ok {
			idx.values[indexVal] = &allIndexMap{
				values: make(map[string]*gosnmp.SnmpPDU),
			}
		}
		idx.values[indexVal].values[allIndexes] = p
	}

	return nil
}

func (s *store) Add(p *gosnmp.SnmpPDU) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	o := s.mibStore.GetObject(p.Name)
	if o == nil {
		return fmt.Errorf("No corresponding object in MIB store for OID %s", p.Name)
	}
	switch o.Kind {
	case smi.KindScalar:
		return s.addScalar(p, o)
	case smi.KindObject:
		if o.Parent != nil && o.Parent.Kind == smi.KindScalar {
			return s.addScalar(p, o)
		}
	case smi.KindColumn:
		return s.addTabular(p, o)
	}
	return fmt.Errorf("Got unexpected object kind for OID %s: %d", p.Name, o.Kind)
}

func (s *store) Clear() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.scalars = make(map[string]*gosnmp.SnmpPDU)
	s.columns = make(map[string]*columnStore)
	return nil
}

func (s *store) GetScalar(oid string) (*gosnmp.SnmpPDU, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	o := s.mibStore.GetObject(oid)
	if o == nil {
		return nil,
			fmt.Errorf("No corresponding object in MIB store for OID %s", oid)
	}
	if o.Kind != smi.KindScalar &&
		(o.Parent != nil && o.Parent.Kind != smi.KindScalar) {
		return nil,
			fmt.Errorf("Object for OID %s is not a scalar (%d)", oid, o.Kind)
	}
	p, ok := s.scalars[o.Oid]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (s *store) getTabularUnconstrained(o *smi.Object) ([]*gosnmp.SnmpPDU, error) {
	col, ok := s.columns[o.Oid]
	if !ok {
		return nil, nil
	}

	pdus := []*gosnmp.SnmpPDU{}
	for _, v := range col.entries {
		pdus = append(pdus, v)
	}
	return pdus, nil
}

func (s *store) getTabularFullyConstrained(o *smi.Object,
	constraints ...Index) ([]*gosnmp.SnmpPDU, error) {
	col, ok := s.columns[o.Oid]
	if !ok {
		return nil, nil
	}

	constraintMap := make(map[string]Index)
	for _, c := range constraints {
		constraintMap[c.Name] = c
	}

	indexValues := []string{}
	for _, indexName := range o.Parent.Indexes {
		c, ok := constraintMap[indexName]
		if !ok {
			return nil, fmt.Errorf("Invalid constraint for OID %s", o.Oid)
		}
		indexValues = append(indexValues, c.Value)
	}
	p, ok := col.entries[strings.Join(indexValues, ".")]
	if !ok {
		return nil, nil
	}
	return []*gosnmp.SnmpPDU{p}, nil
}

func (s *store) getTabularPartiallyConstrained(o *smi.Object,
	constraints ...Index) ([]*gosnmp.SnmpPDU, error) {
	col, ok := s.columns[o.Oid]
	if !ok {
		return nil, nil
	}

	pdus := []*gosnmp.SnmpPDU{}
	for i, c := range constraints {
		if i == 0 {
			idx, ok := col.indexes[c.Name]
			if !ok {
				return nil, nil
			}
			aiv, ok := idx.values[c.Value]
			if !ok {
				return nil, nil
			}
			for _, v := range aiv.values {
				pdus = append(pdus, v)
			}
			if !ok {
				return nil, nil
			}
			continue
		}
		intersection := []*gosnmp.SnmpPDU{}
		for _, p := range pdus {
			allIndexes := strings.Join(indexValues(p, o), ".")
			idx, ok := col.indexes[c.Name]
			if !ok {
				return nil, nil
			}
			if aiv, ok := idx.values[c.Value]; ok {
				if _, ok := aiv.values[allIndexes]; ok {
					intersection = append(intersection, p)
				}
			}
		}
		pdus = intersection
	}
	return pdus, nil
}

func (s *store) GetTabular(oid string, constraints ...Index) ([]*gosnmp.SnmpPDU, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	o := s.mibStore.GetObject(oid)
	if o == nil {
		return nil,
			fmt.Errorf("No corresponding object in MIB store for OID %s", oid)
	}
	if o.Kind != smi.KindColumn {
		return nil,
			fmt.Errorf("Object for OID %s is not a scalar (%d)", oid, o.Kind)
	}
	if o.Parent == nil {
		return nil, fmt.Errorf("No parent for OID %s", oid)
	}
	if len(constraints) > len(o.Parent.Indexes) {
		return nil, fmt.Errorf("%d constraints is more than %d indexes",
			len(constraints), len(o.Parent.Indexes))
	}
	indexMap := make(map[string]bool)
	for _, i := range o.Parent.Indexes {
		indexMap[i] = true
	}
	for i, c := range constraints {
		co := s.mibStore.GetObject(c.Name)
		if co == nil {
			return nil, fmt.Errorf("Index '%s' not found in MIB store",
				c.Name)
		}
		constraints[i].Name = co.Name
		if _, ok := indexMap[co.Name]; !ok {
			return nil, fmt.Errorf("Invalid constraint '%s' for OID %s",
				co.Name, oid)
		}
	}

	// Unconstrained
	if len(constraints) == 0 {
		return s.getTabularUnconstrained(o)
	}

	// Fully constrained
	if len(constraints) == len(o.Parent.Indexes) {
		return s.getTabularFullyConstrained(o, constraints...)
	}

	// Partially constrained
	return s.getTabularPartiallyConstrained(o, constraints...)
}
