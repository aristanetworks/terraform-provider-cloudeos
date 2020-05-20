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

// Set represents a collection of unique keys
type Set struct {
	key.Map
}

// NewSet returns a pointer to a Set
func NewSet(keys ...key.Key) *Set {
	st := &Set{}
	for _, k := range keys {
		st.Set(k)
	}
	return st
}

// String will output the string representation of the Set
func (s *Set) String() string {
	if s == nil {
		return ""
	}
	return s.Map.String()
}

// KeyString will output a key-formatted represntation of a Map, useful for JSON objects or path
// elements
func (s *Set) KeyString() string {
	if s == nil {
		return ""
	}
	str, _ := key.StringifyInterface(s.Map)
	return str
}

// Len returns the length of a Set
func (s *Set) Len() int {
	if s == nil {
		return 0
	}
	return s.Map.Len()
}

// Equal compares two Sets, returning true if they are equal, false otherwise
func (s *Set) Equal(other interface{}) bool {
	so, ok := other.(*Set)
	if !ok {
		return false
	}
	return (&s.Map).Equal(&so.Map)
}

// Hash outputs the hash value of a Set
func (s *Set) Hash() uint64 {
	if s == nil {
		return 0
	}
	var h uintptr
	_ = s.Iter(func(k key.Key) error {
		h += key.HashInterface(k)
		return nil
	})
	return uint64(h)
}

// Set sets a key in a Set
func (s *Set) Set(k key.Key) {
	if s == nil {
		return
	}

	s.Map.Set(k, struct{}{})
}

// Get returns true if the Key exists in a Set
func (s *Set) Get(k key.Key) bool {
	_, ok := s.Map.Get(k)
	return ok
}

// Del removes a key from a Set
func (s *Set) Del(k key.Key) {
	s.Map.Del(k)
}

// Iter iterates over all keys in a Set
func (s *Set) Iter(f func(k key.Key) error) error {
	if s == nil {
		return nil
	}
	return s.Map.Iter(func(iterKey, _ interface{}) error {
		return f(iterKey.(key.Key))
	})
}

// DeepEqual compares two sets, returning true if the sets are deeply equal to each other
// Implements DeepEqualer
func (s *Set) DeepEqual(other interface{}, _ func(a, b interface{}) bool) bool {
	st, ok := other.(*Set)
	if !ok || s.Len() != st.Len() {
		return false
	}
	// We already know they have the same length, so if one of them is empty we can assume they
	// are equal
	if s.Len() == 0 {
		return true
	}
	err := s.Iter(func(k key.Key) error {
		ok := st.Get(k)
		// TODO: hack to get a key that somehow doesn't hash to itself,
		// most likely when k is a ptr. rm eventually
		if !ok {
			var newkey key.Key
			if err := st.Iter(func(kk key.Key) error {
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
			ok = st.Get(newkey)
		}
		if !ok {
			return fmt.Errorf("key %v not found in other set", k)
		}
		return nil
	})
	return err == nil
}
