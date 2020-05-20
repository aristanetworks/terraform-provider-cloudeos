// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package amap

import (
	"testing"

	"github.com/aristanetworks/goarista/key"
)

func TestMapEqual(t *testing.T) {
	k1 := key.New(map[string]interface{}{
		"key": New(key.New(map[string]interface{}{"a": 1}), 2)})
	k2 := key.New(map[string]interface{}{
		"key": New(key.New(map[string]interface{}{"a": 1}), 2)})

	m := New(k1, "foo")
	val, ok := m.Get(k2)
	if !ok || val != "foo" {
		t.Errorf("key %v != 'foo'", val)
	}

	m2 := NewKVs(KV{k2, "foo"})
	if !m.Equal(m2) {
		t.Errorf("map %s != %s", m.String(), m2.String())
	}
}
