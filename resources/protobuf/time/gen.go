// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

//go:generate protoc -I $GOPATH/src --go_out=plugins=grpc:$GOPATH/src arista/resources/protobuf/time/time.proto

package time

import (
	fmt "fmt"
	"time"

	"arista/aeris/apiserver/client"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewRequestTimeRange constructs a new RequestAtTime with a range set.
func NewRequestTimeRange(start time.Time, end time.Time) (*RequestAtTime, error) {
	s, err := ptypes.TimestampProto(start)
	if err != nil {
		return nil, fmt.Errorf("failed to convert start time: %s", err)
	}
	e, err := ptypes.TimestampProto(end)
	if err != nil {
		return nil, fmt.Errorf("failed to convert end time: %s", err)
	}
	return &RequestAtTime{
		AtTime: &RequestAtTime_Range{
			Range: &TimeRange{
				StartTime: s,
				EndTime:   e,
			},
		},
	}, nil
}

// NewRequestTime constructs a RequestAtTime with a single time set.
func NewRequestTime(at time.Time) (*RequestAtTime, error) {
	t, err := ptypes.TimestampProto(at)
	if err != nil {
		return nil, fmt.Errorf("failed to convert time: %s", err)
	}
	return &RequestAtTime{
		AtTime: &RequestAtTime_Time{
			Time: t,
		},
	}, nil
}

// IsRange returns whether the given RequestAtTime is a time range
func (t *RequestAtTime) IsRange() bool {
	if t == nil {
		return false
	}
	return t.GetRange() != nil
}

// GetStartTime returns either the starting time of a range or the single time
// if the request is not a range.
func (t *RequestAtTime) GetStartTime() (time.Time, error) {
	switch ti := t.AtTime.(type) {
	case *RequestAtTime_Range:
		return ptypes.Timestamp(ti.Range.StartTime)
	case *RequestAtTime_Time:
		return ptypes.Timestamp(ti.Time)
	default:
		return time.Time{}, fmt.Errorf("cannot extract time from type %T", ti)
	}
}

// GetEndTime returns the ending time of a range, if this is a range, or an error
func (t *RequestAtTime) GetEndTime() (time.Time, error) {
	r, ok := t.AtTime.(*RequestAtTime_Range)
	if !ok {
		return time.Time{}, fmt.Errorf("cannot fetch end-time of non-range")
	}

	return ptypes.Timestamp(r.Range.EndTime)
}

// GetStartAndEnd returns both the start and end times of a range.
// If the given RequestAtTime is not a range, or a time could not be parsed,
// then an error is returned.
func (t *RequestAtTime) GetStartAndEnd() (time.Time, time.Time, error) {
	if !t.IsRange() {
		return time.Time{}, time.Time{}, fmt.Errorf("request time is not range")
	}

	start, err := t.GetStartTime()
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to get start time: %s", err)
	}
	end, err := t.GetEndTime()
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("failed to get end time: %s", err)
	}

	return start, end, nil
}

// Contains asserts the RequestAtTime receiver is a range, and returns whether
// the passed time is within the range. If the receiver is not a range, an error
// is returned, otherwise the boolean indicates if the time is bounded by the range.
func (t *RequestAtTime) Contains(other time.Time, inclusive bool) (bool, error) {
	if !t.IsRange() {
		return false, fmt.Errorf("request-time is not a range")
	}

	start, end, err := t.GetStartAndEnd()
	if err != nil {
		return false, err
	}

	if inclusive && (start.Equal(other) || end.Equal(other)) {
		return true, nil
	}

	return other.After(start) && other.Before(end), nil
}

// ToClientOpts converts the RequestAtTime into a ClientOptions to pass to an Aeris request.
// This ensures the start (and end, if given) times are correctly converted to the options
// for a read against Aeris.
func (t *RequestAtTime) ToClientOpts() (client.Options, error) {
	var err error
	var opts client.Options

	opts.Start, err = t.GetStartTime()
	if err != nil {
		return opts, status.Error(codes.InvalidArgument, "failed to get start time from range")
	}
	if t.IsRange() {
		opts.End, err = t.GetEndTime()
		if err != nil {
			return opts, status.Error(codes.InvalidArgument, "failed to get end time from range")
		}
	}

	return opts, nil
}
