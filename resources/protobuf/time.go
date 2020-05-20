// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package protobuf

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// UnixNano converts a timestamp proto message to a Unix nanosecond timestamp
func UnixNano(ts *timestamp.Timestamp) (time.Duration, error) {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		return 0, err
	}
	return time.Duration(t.UnixNano()) * time.Nanosecond, nil
}

// NanoTimestampProto converts a Unix nanosecond timestamp, as a duration since Unix zero time,
// to a timestamp proto message
func NanoTimestampProto(dur time.Duration) (*timestamp.Timestamp, error) {
	return ptypes.TimestampProto(time.Unix(0, int64(dur)))
}
