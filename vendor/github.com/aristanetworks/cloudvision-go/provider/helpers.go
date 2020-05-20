// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package provider

import (
	"fmt"
	"math"
)

// ToInt64 converts an interface{} to an int64.
func ToInt64(valIntf interface{}) (int64, error) {
	var val int64
	switch t := valIntf.(type) {
	case int:
		val = int64(t)
	case int8:
		val = int64(t)
	case int16:
		val = int64(t)
	case int32:
		val = int64(t)
	case int64:
		val = t
	case uint:
		val = int64(t)
	case uint8:
		val = int64(t)
	case uint16:
		val = int64(t)
	case uint32:
		val = int64(t)
	case uint64:
		if t > math.MaxInt64 {
			return 0, fmt.Errorf("could not convert to int64, %d larger than max of %d",
				t, uint64(math.MaxInt64))
		}
		val = int64(t)
	default:
		return 0, fmt.Errorf("update contained value of unexpected type %T", valIntf)
	}
	return val, nil
}

// ToInt convert an interface{} to an int.
func ToInt(valIntf interface{}) (int, error) {
	v, err := ToInt64(valIntf)
	return int(v), err
}

// ToUint64 converts an interface{} to a uint64.
func ToUint64(valIntf interface{}) (uint64, error) {
	var val uint64
	switch t := valIntf.(type) {
	case int, int8, int16, int32, int64:
		v, e := ToInt64(t)
		if e != nil {
			return 0, e
		}
		if v < 0 {
			return 0, fmt.Errorf("value %d cannot be converted to uint as it is negative", v)
		}
		val = uint64(v)
	case uint:
		val = uint64(t)
	case uint8:
		val = uint64(t)
	case uint16:
		val = uint64(t)
	case uint32:
		val = uint64(t)
	case uint64:
		val = t
	default:
		return 0, fmt.Errorf("update contained value of unexpected type %T", valIntf)
	}
	return val, nil
}
