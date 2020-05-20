// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// Generated from github.com/aristanetworks/goarista/templates/mapkeytemplate.go

package amap

import (
	"fmt"

	"github.com/aristanetworks/goarista/key"
)

// TimestampedValueMap represents a map with keys of type key.Key and values of type
// TimestampedValue
type TimestampedValueMap struct {
	key.Map
}

// NewTimestampedValueMap returns a pointer to a TimestampedValueMap from a list of key-value pairs,
// so long as the list is of even length.
// keys should be of type key.Key, values should be of type TimestampedValue
func NewTimestampedValueMap(keysAndVals ...interface{}) *TimestampedValueMap {
	length := len(keysAndVals)
	if length%2 != 0 {
		panic(
			"Odd number of arguments passed to NewTimestampedValueMap. " +
				"Arguments should be of form: key1, value1, key2, value2, ...")
	}
	m := &TimestampedValueMap{}
	for i := 0; i < length; i += 2 {
		kk := keysAndVals[i]
		k, ok := kk.(key.Key)
		if !ok && kk != nil {
			panic(fmt.Sprintf("key %v should be of type key.Key", kk))
		}
		vv := keysAndVals[i+1]
		v, ok := vv.(TimestampedValue)
		if !ok {
			panic(fmt.Sprintf("value %v should be of type TimestampedValue", vv))
		}
		m.Set(k, v)
	}
	return m
}

// String will output the string representation of the map
func (m *TimestampedValueMap) String() string {
	if m == nil {
		return ""
	}
	return m.Map.String()
}

// KeyString will output a key-formatted represntation of a TimestampedValueMap,
// useful for JSON objects or path elements
func (m *TimestampedValueMap) KeyString() string {
	if m == nil {
		return ""
	}
	str, _ := key.StringifyInterface(m.Map)
	return str
}

// Len returns the length of the TimestampedValueMap
func (m *TimestampedValueMap) Len() int {
	if m == nil {
		return 0
	}
	return m.Map.Len()
}

// Equal compares two TimestampedValueMaps
func (m *TimestampedValueMap) Equal(other interface{}) bool {
	mp, ok := other.(*TimestampedValueMap)
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

// Hash returns the hash value of the TimestampedValueMap
func (m *TimestampedValueMap) Hash() uint64 {
	if m == nil {
		return 0
	}
	var h uintptr
	_ = m.Iter(func(k key.Key, v TimestampedValue) error {
		h += key.HashInterface(k) + key.HashInterface(v)
		return nil
	})
	return uint64(h)
}

// Set adds a key-value pair to the TimestampedValueMap
func (m *TimestampedValueMap) Set(k key.Key, v TimestampedValue) {
	if m == nil {
		return
	}
	m.Map.Set(k, v)
}

// Get retrieves the value stored with key k from the TimestampedValueMap
func (m *TimestampedValueMap) Get(k key.Key) (TimestampedValue, bool) {
	if m == nil {
		return TimestampedValue{}, false
	}
	val, ok := m.Map.Get(k)
	if !ok {
		return TimestampedValue{}, false
	}
	return val.(TimestampedValue), ok
}

// GetWithoutBool retrieves the value stored with key k from the TimestampedValueMap,
// but only returns the value or nil, without the success boolean
func (m *TimestampedValueMap) GetWithoutBool(k key.Key) TimestampedValue {
	v, _ := m.Get(k)
	return v
}

// Del removes an entry with key k from the TimestampedValueMap
func (m *TimestampedValueMap) Del(k key.Key) {
	if m == nil {
		return
	}
	m.Map.Del(k)
}

// Iter applies func f to every key-value pair in the TimestampedValueMap
func (m *TimestampedValueMap) Iter(f func(k key.Key, v TimestampedValue) error) error {
	if m == nil {
		return nil
	}
	return m.Map.Iter(func(iterKey, iterVal interface{}) error {
		return f(iterKey.(key.Key), iterVal.(TimestampedValue))
	})
}

// DeepEqual compares two maps by running a comparer function on each value of the map at each
// key, returning true if the maps are deeply equal to each other
func (m *TimestampedValueMap) DeepEqual(
	other interface{}, comparer func(a, b interface{}) bool) bool {
	mp, ok := other.(*TimestampedValueMap)
	if !ok || m.Len() != mp.Len() {
		return false
	}
	// We already know they have the same length, so if one of them is empty we can assume they
	// are equal
	if m.Len() == 0 {
		return true
	}
	err := m.Iter(func(k key.Key, v TimestampedValue) error {
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
