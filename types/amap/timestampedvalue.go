// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package amap

import "time"

// TimestampedValue associates a generic value with a timestamp representing the
// time of the value's last update.
type TimestampedValue struct {
	// Nanoseconds since Unix epoch
	Timestamp int64 `deepequal:"ignore"`
	Value     interface{}
}

// GetValue returns a TimestampedValue's value member, or nil if the TimestampedValue is nil.
func (t *TimestampedValue) GetValue() interface{} {
	if t != nil {
		return t.Value
	}
	return nil
}

// Time returns a TimestampedValue's timestamp, converted from int64 to time.Time.
func (t *TimestampedValue) Time() time.Time {
	if t == nil {
		return time.Unix(0, 0)
	}
	return time.Unix(0, t.Timestamp)
}
