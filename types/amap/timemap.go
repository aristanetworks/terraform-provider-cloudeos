// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// Generated from github.com/aristanetworks/goarista/templates/mapkeytemplate.go

package amap

import (
	"fmt"
	"time"

	"github.com/aristanetworks/goarista/key"
)

// TimeMap represents a map with keys of type key.Key and values of type time.Time
type TimeMap struct {
	key.Map
}

// NewTimeMap returns a pointer to a TimeMap from a list of key-value pairs,
// so long as the list is of even length.
// keys should be of type key.Key, values should be of type time.Time
func NewTimeMap(keysAndVals ...interface{}) *TimeMap {
	length := len(keysAndVals)
	if length%2 != 0 {
		panic("Odd number of arguments passed to NewTimeMap. Arguments should be of form: " +
			"key1, value1, key2, value2, ...")
	}
	m := &TimeMap{}
	for i := 0; i < length; i += 2 {
		kk := keysAndVals[i]
		k, ok := kk.(key.Key)
		if !ok && kk != nil {
			panic(fmt.Sprintf("key %v should be of type key.Key", kk))
		}
		vv := keysAndVals[i+1]
		v, ok := vv.(time.Time)
		if !ok {
			panic(fmt.Sprintf("value %v should be of type time.Time", vv))
		}
		m.Set(k, v)
	}
	return m
}

// String will output the string representation of the map
func (m *TimeMap) String() string {
	if m == nil {
		return ""
	}
	return m.Map.String()
}

// KeyString will output a key-formatted represntation of a TimeMap,
// useful for JSON objects or path elements
func (m *TimeMap) KeyString() string {
	if m == nil {
		return ""
	}
	str, _ := key.StringifyInterface(m.Map)
	return str
}

// Len returns the length of the TimeMap
func (m *TimeMap) Len() int {
	if m == nil {
		return 0
	}
	return m.Map.Len()
}

// Equal compares two TimeMaps
func (m *TimeMap) Equal(other interface{}) bool {
	mp, ok := other.(*TimeMap)
	if !ok {
		return false
	}
	if (m == nil) && (mp == nil) {
		return true
	}
	if (m == nil) != (mp == nil) {
		return false
	}
	return (&m.Map).Equal(&mp.Map)
}

// Hash returns the hash value of the TimeMap
func (m *TimeMap) Hash() uint64 {
	if m == nil {
		return 0
	}
	var h uintptr
	_ = m.Iter(func(k key.Key, v time.Time) error {
		h += key.HashInterface(k) + key.HashInterface(v)
		return nil
	})
	return uint64(h)
}

// Set adds a key-value pair to the TimeMap
func (m *TimeMap) Set(k key.Key, v time.Time) {
	if m == nil {
		return
	}
	m.Map.Set(k, v)
}

// Get retrieves the value stored with key k from the TimeMap
func (m *TimeMap) Get(k key.Key) (time.Time, bool) {
	if m == nil {
		return time.Time{}, false
	}
	val, ok := m.Map.Get(k)
	if !ok {
		return time.Time{}, false
	}
	return val.(time.Time), ok
}

// GetWithoutBool retrieves the value stored with key k from the TimeMap,
// but only returns the value or nil, without the success boolean
func (m *TimeMap) GetWithoutBool(k key.Key) time.Time {
	v, _ := m.Get(k)
	return v
}

// Del removes an entry with key k from the TimeMap
func (m *TimeMap) Del(k key.Key) {
	if m == nil {
		return
	}
	m.Map.Del(k)
}

// Iter applies func f to every key-value pair in the TimeMap
func (m *TimeMap) Iter(f func(k key.Key, v time.Time) error) error {
	if m == nil {
		return nil
	}
	return m.Map.Iter(func(iterKey, iterVal interface{}) error {
		return f(iterKey.(key.Key), iterVal.(time.Time))
	})
}

// DeepEqual compares two maps by running a comparer function on each value of the map at each
// key, returning true if the maps are deeply equal to each other
func (m *TimeMap) DeepEqual(other interface{}, comparer func(a, b interface{}) bool) bool {
	mp, ok := other.(*TimeMap)
	if !ok || m.Len() != mp.Len() {
		return false
	}
	// We already know they have the same length, so if one of them is empty we can assume they
	// are equal
	if m.Len() == 0 {
		return true
	}
	err := m.Iter(func(k key.Key, v time.Time) error {
		val, ok := mp.Get(k)
		if !ok {
			return fmt.Errorf("key %v not found in other map", k)
		}
		if !comparer(v, val) {
			return fmt.Errorf("%v and %v not equal for key %v", v, val, k)
		}
		return nil
	})
	return err == nil
}
