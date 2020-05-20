// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package amap

import (
	"errors"
	"fmt"

	"github.com/aristanetworks/goarista/key"
)

// Map represents a map[Key]interface{} in the new Map format. This will eventually be changed to
// a key.Map.
type Map struct {
	key.Map
}

// New creates a new Map from a list of key-value pairs, so long as the list is of even length.
// keys should be of type key.Key
func New(keysAndVals ...interface{}) *Map {
	length := len(keysAndVals)
	if length%2 != 0 {
		panic("Odd number of arguments passed to New. Arguments should be of form: " +
			"key1, value1, key2, value2, ...")
	}
	m := &Map{}
	for i := 0; i < length; i += 2 {
		kk := keysAndVals[i]
		k, ok := kk.(key.Key)
		if !ok && kk != nil {
			panic(fmt.Sprintf("key %v should be of type key.Key", kk))
		}
		m.Set(k, keysAndVals[i+1])
	}
	return m
}

// KV represents a key-value pair, and is used to create new Maps in a more structured format
type KV struct {
	K key.Key
	V interface{}
}

// NewKVs creates a new Map from a list of KVs
func NewKVs(kvs ...KV) *Map {
	m := &Map{}
	for _, kv := range kvs {
		m.Set(kv.K, kv.V)
	}
	return m
}

// String will output the string representation of the map
func (m *Map) String() string {
	if m == nil {
		return ""
	}
	return m.Map.String()
}

// KeyString will output a key-formatted represntation of a Map, useful for JSON objects or path
// elements
func (m *Map) KeyString() string {
	if m == nil {
		return ""
	}
	str, _ := key.StringifyInterface(m.Map)
	return str
}

// Len returns the length of the Map
func (m *Map) Len() int {
	if m == nil {
		return 0
	}
	return m.Map.Len()
}

// Equal compares two Maps
func (m *Map) Equal(other interface{}) bool {
	mp, ok := other.(*Map)
	if !ok {
		return false
	}
	return (&m.Map).Equal(&mp.Map)
}

// Hash returns the hash value of the Map
func (m *Map) Hash() uint64 {
	if m == nil {
		return 0
	}
	var h uintptr
	_ = m.Iter(func(k key.Key, v interface{}) error {
		h += key.HashInterface(k) + key.HashInterface(v)
		return nil
	})
	return uint64(h)
}

// Set adds a key-value pair to the Map
func (m *Map) Set(k key.Key, v interface{}) {
	if m == nil {
		return
	}
	m.Map.Set(k, v)
}

// Get retrieves the value stored with key k from the Map
func (m *Map) Get(k key.Key) (interface{}, bool) {
	if m == nil {
		return nil, false
	}
	return m.Map.Get(k)
}

// GetWithoutBool retrieves the value stored with key k from the Map, but only returns the value
// or nil, without the success boolean
func (m *Map) GetWithoutBool(k key.Key) interface{} {
	v, _ := m.Get(k)
	return v
}

// Del removes an entry with key k from the Map
func (m *Map) Del(k key.Key) {
	m.Map.Del(k)
}

// Iter applies func f to every key-value pair in the Map
func (m *Map) Iter(f func(k key.Key, v interface{}) error) error {
	if m == nil {
		return nil
	}
	return m.Map.Iter(func(iterKey, iterVal interface{}) error {
		return f(iterKey.(key.Key), iterVal)
	})
}

// DeepEqual compares two maps by running a comparer function on each value of the map at each
// key, returning true if the maps are deeply equal to each other
func (m *Map) DeepEqual(other interface{}, comparer func(a, b interface{}) bool) bool {
	mp, ok := other.(*Map)
	if !ok || m.Len() != mp.Len() {
		return false
	}
	// We already know they have the same length, so if one of them is empty we can assume they
	// are equal
	if m.Len() == 0 {
		return true
	}
	err := m.Iter(func(k key.Key, v interface{}) error {
		val, ok := mp.Get(k)
		// TODO: hack to get a key that somehow doesn't hash to itself,
		// most likely when k is a ptr. rm eventually
		if !ok {
			var newkey key.Key
			if err := mp.Iter(func(kk key.Key, _ interface{}) error {
				if k.Equal(kk) {
					newkey = kk
					return errors.New("break")
				}
				return nil
			}); err != nil {
				if err.Error() != "break" {
					return err
				}
			}
			val, ok = mp.Get(newkey)
		}
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
