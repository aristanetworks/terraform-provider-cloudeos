// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package amap

import (
	"testing"

	"github.com/aristanetworks/goarista/key"
)

func TestSetEqual(t *testing.T) {
	k1 := key.New(map[string]interface{}{
		"key": New(key.New(map[string]interface{}{"a": 1}), 2)})
	k2 := key.New(map[string]interface{}{
		"key": New(key.New(map[string]interface{}{"a": 1}), 2)})
	s := NewSet(k1)
	ok := s.Get(k2)
	if !ok {
		t.Errorf("key %v does not exist in set", k1)
	}

	s2 := NewSet(k2)
	if !s.Equal(s2) {
		t.Errorf("set %s != %s", s.String(), s2.String())
	}
}
